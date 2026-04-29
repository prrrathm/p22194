package router

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog"

	"p22194.prrrathm.com/mail/internal/handler"
)

func New(mail *handler.MailHandler, log zerolog.Logger) http.Handler {
	r := chi.NewRouter()
	r.Use(chimiddleware.Recoverer)
	r.Use(requestLogger(log))

	r.Get("/health", mail.Health)
	r.Post("/internal/mail/verification", mail.SendVerificationEmail)
	r.Post("/internal/mail/document-share", mail.SendDocumentShareEmail)

	r.NotFound(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "not found", http.StatusNotFound)
	})

	return r
}

func requestLogger(log zerolog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(w, r)
			log.Info().Str("method", r.Method).Str("path", r.URL.Path).Msg("request")
		})
	}
}
