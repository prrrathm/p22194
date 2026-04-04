// Package routes registers HTTP route groups onto a chi router.
// Each file in this package owns one logical route group.
package routes

import (
	"github.com/go-chi/chi/v5"

	"p22194.prrrathm.com/api-gateway/internal/health"
)

// RegisterHealth mounts the Kubernetes-style liveness and readiness probe
// endpoints onto r. These routes intentionally bypass authentication so that
// orchestration infrastructure can reach them without credentials.
//
// Routes registered:
//
//	GET /health/live   — liveness probe; returns 200 while the process is running.
//	GET /health/ready  — readiness probe; returns 200 when all upstreams are reachable.
func RegisterHealth(r chi.Router, h *health.Handler) {
	r.Get("/health/live", h.LiveHandler)
	r.Get("/health/ready", h.ReadyHandler)
}
