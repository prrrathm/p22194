package routes

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

// RegisterProxy mounts the reverse-proxy catch-all under /api/v1/, protected
// by the supplied JWT middleware. All HTTP methods are forwarded; the first
// path segment after /api/v1/ selects the upstream service.
//
// Routes registered:
//
//	* /api/v1/*  — forward all methods to the upstream selected by path segment.
func RegisterProxy(r chi.Router, p http.Handler, jwtMW func(http.Handler) http.Handler) {
	r.Group(func(r chi.Router) {
		r.Use(jwtMW)
		r.HandleFunc("/api/v1/*", p.ServeHTTP)
	})
}
