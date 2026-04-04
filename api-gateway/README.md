
| Category             | Package                               | Role                                              |
| -------------------- | ------------------------------------- | ------------------------------------------------- |
| Routing & middleware | `go-chi/chi/v5`                     | HTTP router, middleware chaining                  |
| Reverse proxy        | `net/http/httputil`(stdlib)         | Proxy requests to upstream microservices          |
| Logging              | `rs/zerolog`or `log/slog`(stdlib) | Structured, leveled logging                       |
| Rate limiting        | `golang.org/x/time/rate`            | Token-bucket rate limiter per client/route        |
| Circuit breaking     | `sony/gobreaker`                    | Stop cascading failures to unhealthy upstreams    |
| JWT auth             | `golang-jwt/jwt/v5`                 | Validate and parse JWT tokens                     |
| Tracing              | `go.opentelemetry.io/otel`          | Distributed tracing (OpenTelemetry)               |
| Metrics              | `prometheus/client_golang`          | Expose latency, error rate, request count metrics |
| Geo-IP               | `oschwald/maxminddb-golang`         | Map client IPs to geographic locations            |
| Config               | `spf13/viper`or `knadh/koanf`     | Load YAML/env config, hot-reload                  |
| CORS                 | `go-chi/cors`                       | Cross-origin request policy enforcement           |
| Request ID           | `go-chi/chi/v5/middleware`          | Inject/propagate correlation IDs                  |
| TLS                  | `crypto/tls`(stdlib)                | TLS termination at the gateway                    |
| Graceful shutdown    | `context`+`os/signal`(stdlib)     | Clean shutdown on SIGTERM                         |
| Caching              | `dgraph-io/ristretto`               | In-memory response caching                        |
| Health checks        | Custom or `alexliesenfeld/health`   | Liveness/readiness probes                         |
