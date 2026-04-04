package metrics

import (
	"net/http"
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Registry holds all custom Prometheus metrics for the gateway.
type Registry struct {
	reg *prometheus.Registry

	// RequestTotal counts completed requests by method, route, and HTTP status.
	RequestTotal *prometheus.CounterVec

	// RequestDuration tracks request latency in seconds by method and route.
	RequestDuration *prometheus.HistogramVec

	// UpstreamErrors counts upstream proxy errors by service name.
	UpstreamErrors *prometheus.CounterVec

	// ActiveRequests tracks in-flight requests.
	ActiveRequests prometheus.Gauge
}

// New creates a Registry with all metrics pre-registered.
// It also registers the standard Go runtime and process collectors.
func New() *Registry {
	reg := prometheus.NewRegistry()

	reg.MustRegister(
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
	)

	r := &Registry{
		reg: reg,

		RequestTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: "gateway",
			Name:      "requests_total",
			Help:      "Total number of HTTP requests processed by the gateway.",
		}, []string{"method", "route", "status"}),

		RequestDuration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: "gateway",
			Name:      "request_duration_seconds",
			Help:      "HTTP request latency in seconds.",
			Buckets:   prometheus.DefBuckets,
		}, []string{"method", "route"}),

		UpstreamErrors: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: "gateway",
			Name:      "upstream_errors_total",
			Help:      "Total number of upstream proxy errors by service.",
		}, []string{"service"}),

		ActiveRequests: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: "gateway",
			Name:      "active_requests",
			Help:      "Number of HTTP requests currently being handled.",
		}),
	}

	reg.MustRegister(
		r.RequestTotal,
		r.RequestDuration,
		r.UpstreamErrors,
		r.ActiveRequests,
	)

	return r
}

// Handler returns an HTTP handler that serves the Prometheus metrics page.
func (r *Registry) Handler() http.Handler {
	return promhttp.HandlerFor(r.reg, promhttp.HandlerOpts{
		EnableOpenMetrics: true,
	})
}

// StatusLabel converts an integer HTTP status code to a string label.
func StatusLabel(code int) string {
	return strconv.Itoa(code)
}
