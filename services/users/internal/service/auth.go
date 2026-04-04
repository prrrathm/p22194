package service

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"golang.org/x/crypto/bcrypt"

	mdbmodels "p22194.prrrathm.com/mdb/models"
	mdbrepo "p22194.prrrathm.com/mdb/repository"
	"p22194.prrrathm.com/users/internal/models"
)

// Sentinel errors — the handler layer maps these to HTTP status codes.
var (
	ErrEmailTaken         = errors.New("email already registered")
	ErrUsernameTaken      = errors.New("username already taken")
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrSessionNotFound    = errors.New("session not found or expired")
)

// AuthService implements register, login, refresh, and logout.
type AuthService struct {
	users      *mdbrepo.UserRepo
	sessions   *mdbrepo.SessionRepo
	jwtSecret  []byte
	accessTTL  time.Duration
	refreshTTL time.Duration
}

// New constructs an AuthService.
func New(
	users *mdbrepo.UserRepo,
	sessions *mdbrepo.SessionRepo,
	jwtSecret []byte,
	accessTTL time.Duration,
	refreshTTL time.Duration,
) *AuthService {
	return &AuthService{
		users:      users,
		sessions:   sessions,
		jwtSecret:  jwtSecret,
		accessTTL:  accessTTL,
		refreshTTL: refreshTTL,
	}
}

// Register creates a new user account and returns a token pair.
// Returns ErrEmailTaken when the email is already in use.
func (s *AuthService) Register(ctx context.Context, req models.RegisterRequest) (*models.TokenPair, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), 12)
	if err != nil {
		return nil, fmt.Errorf("service: bcrypt: %w", err)
	}

	now := time.Now().UTC()
	u := &mdbmodels.User{
		Email:        req.Email,
		Username:     req.Username,
		PasswordHash: string(hash),
		Role:         "user",
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if err := s.users.Create(ctx, u); err != nil {
		switch duplicateField(err) {
		case "email":
			return nil, ErrEmailTaken
		case "username":
			return nil, ErrUsernameTaken
		case "":
			return nil, fmt.Errorf("service: create user: %w", err)
		default:
			return nil, ErrEmailTaken // unknown unique-index violation
		}
	}

	return s.issueTokenPair(ctx, u)
}

// Login verifies credentials and returns a new token pair.
// Returns ErrInvalidCredentials for unknown email or wrong password.
func (s *AuthService) Login(ctx context.Context, req models.LoginRequest) (*models.TokenPair, error) {
	u, err := s.users.FindByEmail(ctx, req.Email)
	fmt.Println("Hello this is the user", u)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrInvalidCredentials
		}
		return nil, fmt.Errorf("service: find user: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(req.Password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	return s.issueTokenPair(ctx, u)
}

// Refresh exchanges a valid refresh token for a new token pair (rotation).
// The old session is deleted before the new one is created.
// Returns ErrSessionNotFound when the token is unknown or expired.
func (s *AuthService) Refresh(ctx context.Context, req models.RefreshRequest) (*models.TokenPair, error) {
	hash := tokenHash(req.RefreshToken)

	sess, err := s.sessions.FindByTokenHash(ctx, hash)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrSessionNotFound
		}
		return nil, fmt.Errorf("service: find session: %w", err)
	}

	if time.Now().UTC().After(sess.ExpiresAt) {
		_ = s.sessions.DeleteByTokenHash(ctx, hash)
		return nil, ErrSessionNotFound
	}

	u, err := s.users.FindByID(ctx, sess.UserID)
	if err != nil {
		return nil, fmt.Errorf("service: find user: %w", err)
	}

	// Rotate: delete old session before issuing new one.
	if err := s.sessions.DeleteByTokenHash(ctx, hash); err != nil {
		return nil, fmt.Errorf("service: delete old session: %w", err)
	}

	return s.issueTokenPair(ctx, u)
}

// Logout deletes the session associated with the given refresh token.
// Idempotent — does not error if the session does not exist.
func (s *AuthService) Logout(ctx context.Context, req models.LogoutRequest) error {
	return s.sessions.DeleteByTokenHash(ctx, tokenHash(req.RefreshToken))
}

// ParseAccessToken validates and parses a signed JWT, returning (sub, email, role, err).
func (s *AuthService) ParseAccessToken(raw string) (sub, email, role string, err error) {
	token, err := jwt.ParseWithClaims(raw, &jwt.MapClaims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return s.jwtSecret, nil
	})
	if err != nil || !token.Valid {
		return "", "", "", fmt.Errorf("service: invalid token: %w", err)
	}
	claims, ok := token.Claims.(*jwt.MapClaims)
	if !ok {
		return "", "", "", fmt.Errorf("service: bad claims type")
	}
	sub, _ = (*claims)["sub"].(string)
	email, _ = (*claims)["email"].(string)
	role, _ = (*claims)["role"].(string)
	return sub, email, role, nil
}

// ── Private helpers ──────────────────────────────────────────────────────────

// issueTokenPair mints a JWT access token and creates a new refresh session.
func (s *AuthService) issueTokenPair(ctx context.Context, u *mdbmodels.User) (*models.TokenPair, error) {
	now := time.Now().UTC()
	accessExp := now.Add(s.accessTTL)

	claims := jwt.MapClaims{
		"sub":   u.ID.Hex(),
		"email": u.Email,
		"role":  u.Role,
		"iat":   now.Unix(),
		"exp":   accessExp.Unix(),
	}
	accessToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(s.jwtSecret)
	if err != nil {
		return nil, fmt.Errorf("service: sign jwt: %w", err)
	}

	rawRefresh := uuid.New().String()
	sess := &mdbmodels.Session{
		UserID:    u.ID,
		TokenHash: tokenHash(rawRefresh),
		ExpiresAt: now.Add(s.refreshTTL),
		CreatedAt: now,
	}
	if err := s.sessions.Create(ctx, sess); err != nil {
		return nil, fmt.Errorf("service: create session: %w", err)
	}

	return &models.TokenPair{
		AccessToken:  accessToken,
		RefreshToken: rawRefresh,
		ExpiresIn:    int64(s.accessTTL.Seconds()),
	}, nil
}

// tokenHash returns the lowercase hex-encoded SHA-256 digest of raw.
func tokenHash(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return fmt.Sprintf("%x", sum)
}

// duplicateField returns the field name that triggered a MongoDB duplicate-key
// error (code 11000), or "" if err is not a duplicate-key error.
// It inspects the write-error message, which MongoDB formats as:
//
//	E11000 duplicate key error … index: <name> dup key: { <field>: … }
func duplicateField(err error) string {
	var we mongo.WriteException
	if !errors.As(err, &we) {
		return ""
	}
	for _, e := range we.WriteErrors {
		if e.Code != 11000 {
			continue
		}
		msg := strings.ToLower(e.Message)
		switch {
		case strings.Contains(msg, "username"):
			return "username"
		case strings.Contains(msg, "email"):
			return "email"
		default:
			return "unknown"
		}
	}
	return ""
}
