package proxy

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"

	"github.com/rs/zerolog"
	"github.com/sony/gobreaker"

	"p22194.prrrathm.com/api-gateway/internal/cache"
	"p22194.prrrathm.com/api-gateway/internal/circuitbreaker"
	"p22194.prrrathm.com/api-gateway/internal/config"
	"p22194.prrrathm.com/api-gateway/internal/metrics"
)

// upstream groups the reverse proxy and circuit breaker for one backend service.
type upstream struct {
	proxy   *httputil.ReverseProxy
	breaker *gobreaker.CircuitBreaker
}

// Handler is an http.Handler that forwards requests to registered upstream
// services based on the {service} URL path parameter.
type Handler struct {
	upstreams map[string]*upstream
	cache     *cache.Cache
	metrics   *metrics.Registry
	log       zerolog.Logger
}

// New builds a Handler from the upstreams defined in config.
// Each upstream gets its own httputil.ReverseProxy and circuit breaker.
func New(
	upstreamCfgs map[string]config.UpstreamConfig,
	c *cache.Cache,
	m *metrics.Registry,
	log zerolog.Logger,
) (*Handler, error) {
	ups := make(map[string]*upstream, len(upstreamCfgs))

	for name, cfg := range upstreamCfgs {
		target, err := url.Parse(cfg.URL)
		if err != nil {
			return nil, fmt.Errorf("proxy: invalid upstream URL for %q: %w", name, err)
		}

		timeout := cfg.Timeout
		if timeout == 0 {
			timeout = 30 * time.Second
		}

		transport := &http.Transport{
			ResponseHeaderTimeout: timeout,
		}

		rp := httputil.NewSingleHostReverseProxy(target)
		rp.Transport = transport
		rp.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
			log.Error().Err(err).Str("service", name).Msg("upstream error")
			if m != nil {
				m.UpstreamErrors.WithLabelValues(name).Inc()
			}
			http.Error(w, "bad gateway", http.StatusBadGateway)
		}

		ups[name] = &upstream{
			proxy:   rp,
			breaker: circuitbreaker.New(name),
		}
	}

	return &Handler{
		upstreams: ups,
		cache:     c,
		metrics:   m,
		log:       log,
	}, nil
}

// ServeHTTP routes the request to the upstream identified by the {service}
// chi URL parameter. Returns 502 for circuit-open or unknown services.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Extract service name from URL path segment after /api/v1/.
	service := serviceFromPath(r.URL.Path)
	if service == "" {
		http.Error(w, "missing service name", http.StatusBadRequest)
		return
	}

	up, ok := h.upstreams[service]
	if !ok {
		http.Error(w, fmt.Sprintf("unknown service %q", service), http.StatusNotFound)
		return
	}

	// For GET requests, check the response cache before forwarding.
	if r.Method == http.MethodGet && h.cache != nil {
		cacheKey := r.Method + ":" + r.URL.RequestURI()
		if cached, found := h.cache.Get(cacheKey); found {
			w.Header().Set("X-Cache", "HIT")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(cached)
			return
		}
	}

	// Wrap the proxy call in the circuit breaker.
	_, err := up.breaker.Execute(func() (any, error) {
		rec := &responseRecorder{ResponseWriter: w, status: http.StatusOK}
		up.proxy.ServeHTTP(rec, r)
		if rec.status >= http.StatusInternalServerError {
			return nil, fmt.Errorf("upstream returned %d", rec.status)
		}
		return nil, nil
	})

	if err != nil {
		if isCircuitOpen(err) {
			h.log.Warn().Str("service", service).Msg("circuit open")
			http.Error(w, "service unavailable", http.StatusServiceUnavailable)
		}
		// Error already written by ErrorHandler above.
	}
}

// serviceFromPath returns the first path segment after /api/v1/.
// e.g. /api/v1/orders/123 → "orders"
func serviceFromPath(path string) string {
	const prefix = "/api/v1/"
	if len(path) <= len(prefix) {
		return ""
	}
	rest := path[len(prefix):]
	for i := 0; i < len(rest); i++ {
		if rest[i] == '/' {
			return rest[:i]
		}
	}
	return rest
}

// Forward returns an http.Handler that proxies every request to the named
// upstream regardless of URL path. Used for /auth/* routes that do not follow
// the /api/v1/{service}/ convention.
func (h *Handler) Forward(service string) http.Handler {
	up, ok := h.upstreams[service]
	if !ok {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, fmt.Sprintf("unknown service %q", service), http.StatusNotFound)
		})
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := up.breaker.Execute(func() (any, error) {
			rec := &responseRecorder{ResponseWriter: w, status: http.StatusOK}
			up.proxy.ServeHTTP(rec, r)
			if rec.status >= http.StatusInternalServerError {
				return nil, fmt.Errorf("upstream returned %d", rec.status)
			}
			return nil, nil
		})
		if err != nil && isCircuitOpen(err) {
			h.log.Warn().Str("service", service).Msg("circuit open")
			http.Error(w, "service unavailable", http.StatusServiceUnavailable)
		}
	})
}

func isCircuitOpen(err error) bool {
	return err == gobreaker.ErrOpenState || err == gobreaker.ErrTooManyRequests
}

// responseRecorder captures the status code written by the upstream proxy.
type responseRecorder struct {
	http.ResponseWriter
	status int
}

func (rr *responseRecorder) WriteHeader(code int) {
	rr.status = code
	rr.ResponseWriter.WriteHeader(code)
}
