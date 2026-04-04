package middleware

import (
	"net/http"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

// Tracer returns a middleware that starts an OpenTelemetry span for every
// request, propagates W3C trace-context headers, and records the HTTP status
// code and any errors on span completion.
func Tracer(tracer trace.Tracer, propagator propagation.TextMapPropagator) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract any incoming trace context (e.g. traceparent header).
			ctx := propagator.Extract(r.Context(), propagation.HeaderCarrier(r.Header))

			spanName := r.Method + " " + r.URL.Path
			ctx, span := tracer.Start(ctx, spanName, trace.WithSpanKind(trace.SpanKindServer))
			defer span.End()

			span.SetAttributes(
				attribute.String("http.method", r.Method),
				attribute.String("http.url", r.URL.String()),
				attribute.String("http.host", r.Host),
				attribute.String("http.user_agent", r.UserAgent()),
				attribute.String("request_id", GetRequestID(ctx)),
			)

			// Inject outgoing trace context into response headers.
			propagator.Inject(ctx, propagation.HeaderCarrier(w.Header()))

			wrapped := wrapResponseWriter(w)
			next.ServeHTTP(wrapped, r.WithContext(ctx))

			span.SetAttributes(attribute.Int("http.status_code", wrapped.status))
			if wrapped.status >= http.StatusInternalServerError {
				span.SetStatus(codes.Error, http.StatusText(wrapped.status))
			}
		})
	}
}
