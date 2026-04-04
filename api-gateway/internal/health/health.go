package health

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"time"

	"github.com/alexliesenfeld/health"

	"p22194.prrrathm.com/api-gateway/internal/config"
)

// Handler exposes liveness and readiness HTTP endpoints.
type Handler struct {
	readinessChecker health.Checker
}

// New builds a Handler that checks each upstream for reachability on the
// readiness probe. If upstreams is empty the readiness probe always passes.
func New(upstreamCfgs map[string]config.UpstreamConfig) *Handler {
	checks := make([]health.Check, 0, len(upstreamCfgs))

	for name, cfg := range upstreamCfgs {
		name := name // capture loop var
		cfg := cfg

		checks = append(checks, health.Check{
			Name:    name,
			Timeout: 5 * time.Second,
			Check: func(ctx context.Context) error {
				u, err := url.Parse(cfg.URL)
				if err != nil {
					return err
				}
				req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
				if err != nil {
					return err
				}
				resp, err := http.DefaultClient.Do(req)
				if err != nil {
					return err
				}
				resp.Body.Close()
				return nil
			},
		})
	}

	checker := health.NewChecker(
		health.WithCacheDuration(5*time.Second),
		health.WithChecks(checks...),
	)

	return &Handler{readinessChecker: checker}
}

// LiveHandler always returns 200 — it signals that the process is running.
func (h *Handler) LiveHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "up"})
}

// ReadyHandler returns 200 when all upstream checks pass, 503 otherwise.
func (h *Handler) ReadyHandler(w http.ResponseWriter, r *http.Request) {
	health.NewHandler(h.readinessChecker).ServeHTTP(w, r)
}
