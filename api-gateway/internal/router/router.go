// Package router assembles the chi HTTP router by composing route groups and
// global middleware. Route registrations live in the ./routes sub-package and
// handler constructors in the ./controllers sub-package.
package router

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"

	"p22194.prrrathm.com/api-gateway/internal/config"
	"p22194.prrrathm.com/api-gateway/internal/health"
	"p22194.prrrathm.com/api-gateway/internal/metrics"
	"p22194.prrrathm.com/api-gateway/internal/middleware"
	"p22194.prrrathm.com/api-gateway/internal/proxy"
	"p22194.prrrathm.com/api-gateway/internal/router/routes"
)

// Deps holds all dependencies needed to construct the router.
type Deps struct {
	Config     *config.Config
	Log        zerolog.Logger
	Metrics    *metrics.Registry
	Proxy      *proxy.Handler
	Health     *health.Handler
	Tracer     trace.Tracer
	Propagator propagation.TextMapPropagator
}

// New builds and returns the fully configured chi router.
//
// Route layout:
//
//	GET  /health/live       liveness probe (no auth)
//	GET  /health/ready      readiness probe (no auth)
//	GET  /metrics           Prometheus metrics (no auth)
//	GET  /admin/config      current config JSON (JWT required)
//	POST /admin/reload      reload config (JWT required, stub)
//	POST /auth/register     create user account (no auth)
//	POST /auth/login        issue access token (no auth)
//	POST /auth/refresh      refresh access token (no auth)
//	POST /auth/logout       invalidate session (JWT required)
//	*    /api/v1/*          reverse-proxy to upstream (JWT required)
func New(d Deps) (http.Handler, error) {
	r := chi.NewRouter()

	// ── Global middleware (applied to every route) ──────────────────────────
	r.Use(chimiddleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger(d.Log))
	r.Use(middleware.Tracer(d.Tracer, d.Propagator))
	r.Use(metricsMiddleware(d.Metrics))
	r.Use(middleware.CORS(
		d.Config.CORS.AllowedOrigins,
		d.Config.CORS.AllowedMethods,
		d.Config.CORS.AllowedHeaders,
	))
	r.Use(middleware.RateLimit(
		d.Config.RateLimit.RequestsPerSecond,
		d.Config.RateLimit.Burst,
	))

	jwtMW := middleware.JWT([]byte(d.Config.Auth.JWTSecret))

	// ── Route groups ─────────────────────────────────────────────────────────
	routes.RegisterHealth(r, d.Health)
	routes.RegisterMetrics(r, d.Metrics, d.Config.Metrics.Path)
	routes.RegisterAdmin(r, d.Config, jwtMW)
	routes.RegisterAuth(r, d.Proxy.Forward("users"), jwtMW)
	routes.RegisterProxy(r, d.Proxy, jwtMW)

	// ── 404 catch-all ───────────────────────────────────────────────────────
	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "not found", http.StatusNotFound)
	})

	return r, nil
}

// metricsMiddleware records per-request Prometheus metrics using the Registry.
// It wraps the response writer to capture the HTTP status code after the
// downstream handler returns.
func metricsMiddleware(m *metrics.Registry) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			m.ActiveRequests.Inc()
			defer m.ActiveRequests.Dec()

			wrapped := &statusWriter{ResponseWriter: w, status: http.StatusOK}
			next.ServeHTTP(wrapped, r)

			route := chi.RouteContext(r.Context()).RoutePattern()
			status := metrics.StatusLabel(wrapped.status)
			m.RequestTotal.WithLabelValues(r.Method, route, status).Inc()
		})
	}
}

// statusWriter wraps [http.ResponseWriter] to capture the written HTTP status
// code so that it can be recorded in Prometheus metrics after the response
// is sent.
type statusWriter struct {
	http.ResponseWriter
	status int
}

// WriteHeader intercepts the status code before delegating to the underlying
// ResponseWriter.
func (sw *statusWriter) WriteHeader(code int) {
	sw.status = code
	sw.ResponseWriter.WriteHeader(code)
}
