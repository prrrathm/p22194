package routes

import (
	"github.com/go-chi/chi/v5"

	"p22194.prrrathm.com/api-gateway/internal/metrics"
)

// RegisterMetrics mounts the Prometheus scrape endpoint at the path configured
// in the application settings. The endpoint requires no authentication so that
// Prometheus agents can scrape without a bearer token.
//
// Routes registered:
//
//	GET {path}  — Prometheus metrics exposition (text/plain; version=0.0.4).
func RegisterMetrics(r chi.Router, m *metrics.Registry, path string) {
	r.Get(path, m.Handler().ServeHTTP)
}
