package middleware

import (
	"net/http"
	"time"

	"github.com/rs/zerolog"
)

// responseWriter wraps http.ResponseWriter to capture the status code.
type responseWriter struct {
	http.ResponseWriter
	status int
	bytes  int
}

func wrapResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{ResponseWriter: w, status: http.StatusOK}
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	n, err := rw.ResponseWriter.Write(b)
	rw.bytes += n
	return n, err
}

// Logger returns a middleware that logs each request using the provided zerolog
// logger. It records method, path, status, latency, bytes written, and
// request ID (if set by the RequestID middleware).
func Logger(log zerolog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			wrapped := wrapResponseWriter(w)

			next.ServeHTTP(wrapped, r)

			var ev *zerolog.Event
			switch {
			case wrapped.status >= 500:
				ev = log.Error()
			case wrapped.status >= 400:
				ev = log.Warn()
			default:
				ev = log.Info()
			}

			ev.
				Str("method", r.Method).
				Str("path", r.URL.Path).
				Int("status", wrapped.status).
				Dur("latency", time.Since(start)).
				Int("bytes", wrapped.bytes).
				Str("remote_addr", r.RemoteAddr).
				Str("request_id", GetRequestID(r.Context())).
				Msg("request")
		})
	}
}
