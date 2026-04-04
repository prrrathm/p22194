package routes

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

// RegisterAuth mounts the authentication flow endpoints, forwarding each
// request to the upstream users service via the provided proxy handler.
//
// Public endpoints (no JWT required):
//
//	POST /auth/register  — create a new user account.
//	POST /auth/login     — exchange credentials for an access/refresh token pair.
//	POST /auth/refresh   — exchange a refresh token for a new access token.
//
// Protected endpoints (JWT required):
//
//	POST /auth/logout       — invalidate the current session; requires a valid JWT
//	                          to prevent session-replay attacks.
//	GET  /api/v1/users/me  — return the current user's profile.
func RegisterAuth(r chi.Router, fwd http.Handler, jwtMW func(http.Handler) http.Handler) {
	// Public — no authentication needed for registration, login, and token refresh.
	r.Group(func(r chi.Router) {
		r.Post("/auth/register", fwd.ServeHTTP)
		r.Post("/auth/login", fwd.ServeHTTP)
		r.Post("/auth/refresh", fwd.ServeHTTP)
	})

	// Protected — a valid JWT is required to prevent session-replay attacks.
	r.Group(func(r chi.Router) {
		r.Use(jwtMW)
		r.Post("/auth/logout", fwd.ServeHTTP)
		r.Get("/auth/me", fwd.ServeHTTP)
	})
}
