package router

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog"

	"p22194.prrrathm.com/users/internal/handler"
)

// New builds the chi router for the users service.
//
// Route layout:
//
//	POST /auth/register    create account, return token pair
//	POST /auth/login       authenticate, return token pair
//	POST /auth/refresh     exchange refresh token for new pair
//	POST /auth/logout      invalidate refresh session (JWT validated upstream)
//	GET  /api/v1/users/me  return current user from JWT claims (JWT validated upstream)
func New(auth *handler.AuthHandler, log zerolog.Logger) http.Handler {
	r := chi.NewRouter()

	r.Use(chimiddleware.Recoverer)
	r.Use(requestLogger(log))

	r.Post("/auth/register", auth.Register)
	r.Post("/auth/login", auth.Login)
	r.Post("/auth/refresh", auth.Refresh)
	r.Post("/auth/logout", auth.Logout)

	r.Get("/auth/me", auth.Me)

	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "not found", http.StatusNotFound)
	})

	return r
}

// requestLogger is a minimal zerolog request logger.
func requestLogger(log zerolog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(w, r)
			log.Info().
				Str("method", r.Method).
				Str("path", r.URL.Path).
				Msg("request")
		})
	}
}
