package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

type claimsKey struct{}

// JWT returns a middleware that validates Bearer tokens using the provided HMAC
// secret. On success the parsed jwt.MapClaims are stored in the request context
// (retrieve with GetClaims). On failure it returns 401 or 403.
func JWT(secret []byte) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			raw, ok := bearerToken(r)
			if !ok {
				http.Error(w, "missing or malformed Authorization header", http.StatusUnauthorized)
				return
			}

			token, err := jwt.ParseWithClaims(raw, &jwt.MapClaims{}, func(t *jwt.Token) (any, error) {
				if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, jwt.ErrSignatureInvalid
				}
				return secret, nil
			})
			if err != nil || !token.Valid {
				http.Error(w, "invalid or expired token", http.StatusUnauthorized)
				return
			}

			claims, ok := token.Claims.(*jwt.MapClaims)
			if !ok {
				http.Error(w, "invalid token claims", http.StatusForbidden)
				return
			}

			ctx := context.WithValue(r.Context(), claimsKey{}, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetClaims retrieves the JWT claims stored by the JWT middleware.
// Returns nil when not present.
func GetClaims(ctx context.Context) jwt.MapClaims {
	claims, _ := ctx.Value(claimsKey{}).(*jwt.MapClaims)
	if claims == nil {
		return nil
	}
	return *claims
}

// bearerToken extracts the raw token string from "Authorization: Bearer <token>".
func bearerToken(r *http.Request) (string, bool) {
	h := r.Header.Get("Authorization")
	if h == "" {
		return "", false
	}
	parts := strings.SplitN(h, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
		return "", false
	}
	return strings.TrimSpace(parts[1]), true
}
