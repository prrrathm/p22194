package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/rs/zerolog"

	"p22194.prrrathm.com/users/internal/models"
	"p22194.prrrathm.com/users/internal/service"
)

// AuthHandler holds HTTP handlers for all auth endpoints.
type AuthHandler struct {
	svc *service.AuthService
	log zerolog.Logger
}

// New constructs an AuthHandler.
func New(svc *service.AuthService, log zerolog.Logger) *AuthHandler {
	return &AuthHandler{svc: svc, log: log}
}

// Register handles POST /auth/register
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req models.RegisterRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	pair, err := h.svc.Register(r.Context(), req)
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusCreated, pair)
}

// Login handles POST /auth/login
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req models.LoginRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	pair, err := h.svc.Login(r.Context(), req)
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, pair)
}

// Refresh handles POST /auth/refresh
func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	var req models.RefreshRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	pair, err := h.svc.Refresh(r.Context(), req)
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, pair)
}

// Logout handles POST /auth/logout.
// The gateway's JWT middleware has already validated the Bearer token.
// This handler only needs to invalidate the refresh session.
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	var req models.LogoutRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	if err := h.svc.Logout(r.Context(), req); err != nil {
		h.log.Error().Err(err).Msg("logout failed")
		writeJSON(w, http.StatusInternalServerError, errorBody("internal server error"))
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "logged out"})
}

// Me handles GET /api/v1/users/me.
// Re-parses the forwarded Authorization: Bearer header using the shared JWT secret.
func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	authHeader := r.Header.Get("Authorization")
	if !strings.HasPrefix(authHeader, "Bearer ") {
		writeJSON(w, http.StatusUnauthorized, errorBody("missing authorization"))
		return
	}
	rawToken := strings.TrimPrefix(authHeader, "Bearer ")

	sub, email, role, err := h.svc.ParseAccessToken(rawToken)
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, errorBody("invalid token"))
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"id":    sub,
		"email": email,
		"role":  role,
	})
}

// ── Private helpers ──────────────────────────────────────────────────────────

func (h *AuthHandler) writeError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, service.ErrEmailTaken):
		writeJSON(w, http.StatusConflict, errorBody("email already registered"))
	case errors.Is(err, service.ErrUsernameTaken):
		writeJSON(w, http.StatusConflict, errorBody("username already taken"))
	case errors.Is(err, service.ErrInvalidCredentials):
		writeJSON(w, http.StatusUnauthorized, errorBody("invalid email or password"))
	case errors.Is(err, service.ErrSessionNotFound):
		writeJSON(w, http.StatusUnauthorized, errorBody("refresh token invalid or expired"))
	default:
		h.log.Error().Err(err).Str("path", r.URL.Path).Msg("unhandled service error")
		writeJSON(w, http.StatusInternalServerError, errorBody("internal server error"))
	}
}

func decodeJSON(w http.ResponseWriter, r *http.Request, v any) bool {
	if err := json.NewDecoder(r.Body).Decode(v); err != nil {
		writeJSON(w, http.StatusBadRequest, errorBody("invalid request body"))
		return false
	}
	return true
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func errorBody(msg string) map[string]string {
	return map[string]string{"error": msg}
}
