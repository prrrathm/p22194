package router

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog"

	"p22194.prrrathm.com/kitchen/internal/handler"
	"p22194.prrrathm.com/kitchen/internal/middleware"
)

// New builds the chi router for the kitchen service.
//
// Route layout:
//
//	POST   /api/v1/documents                         create document
//	GET    /api/v1/documents                         list documents (paginated)
//	GET    /api/v1/documents/{id}                    get document by ID
//	GET    /api/v1/documents/{id}/children           list sub-documents
//	DELETE /api/v1/documents/{id}                    soft-delete document
//	PATCH  /api/v1/documents/{id}/archive            archive document
//
//	POST   /api/v1/documents/{id}/blocks             insert block into document
//	PATCH  /api/v1/blocks/{id}                       update block content
//	PATCH  /api/v1/blocks/{id}/reorder               reorder block in document
//	DELETE /api/v1/blocks/{id}                       soft-delete block
//
//	POST   /api/v1/documents/{id}/shares             share document with user
//	DELETE /api/v1/documents/{id}/shares/{user_id}   remove user access
//	GET    /api/v1/documents/{id}/shares             list users with access
//
// All routes require a valid Bearer JWT (validated by the JWT middleware).
func New(
	docs *handler.DocumentHandler,
	blocks *handler.BlockHandler,
	shares *handler.SharingHandler,
	jwtSecret []byte,
	log zerolog.Logger,
) http.Handler {
	r := chi.NewRouter()

	r.Use(chimiddleware.Recoverer)
	r.Use(requestLogger(log))
	r.Use(middleware.JWT(jwtSecret))

	r.Route("/api/v1", func(r chi.Router) {
		// ── Documents ────────────────────────────────────────────────────────
		r.Route("/documents", func(r chi.Router) {
			r.Post("/", docs.Create)
			r.Get("/", docs.List)

			r.Route("/{id}", func(r chi.Router) {
				r.Get("/", docs.GetByID)
				r.Patch("/", docs.Update)
				r.Delete("/", docs.Delete)
				r.Patch("/archive", docs.Archive)
				r.Get("/children", docs.ListChildren)

				// Blocks nested under document
				r.Get("/blocks", blocks.List)
				r.Post("/blocks", blocks.Insert)

				// Sharing nested under document
				r.Post("/shares", shares.Share)
				r.Get("/shares", shares.GetShares)
				r.Delete("/shares/{user_id}", shares.RemoveAccess)
			})
		})

		// ── Blocks (standalone operations) ───────────────────────────────────
		r.Route("/blocks/{id}", func(r chi.Router) {
			r.Patch("/", blocks.Update)
			r.Patch("/reorder", blocks.Reorder)
			r.Delete("/", blocks.Delete)
		})
	})

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
