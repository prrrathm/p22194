package routes

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"p22194.prrrathm.com/api-gateway/internal/config"
	"p22194.prrrathm.com/api-gateway/internal/router/controllers"
)

// RegisterAdmin mounts the administrative API under /admin, protected by the
// supplied JWT middleware. All routes in this group require a valid bearer token.
//
// Routes registered:
//
//	GET  /admin/config  — returns the running configuration as sanitised JSON.
//	POST /admin/reload  — schedules a configuration reload (stub).
func RegisterAdmin(r chi.Router, cfg *config.Config, jwtMW func(http.Handler) http.Handler) {
	r.Group(func(r chi.Router) {
		r.Use(jwtMW)
		r.Get("/admin/config", controllers.ConfigHandler(cfg))
		r.Post("/admin/reload", controllers.ReloadHandler())
	})
}
