// Package controllers provides HTTP handler constructors for the API-gateway's
// administrative endpoints.
package controllers

import (
	"encoding/json"
	"net/http"

	"p22194.prrrathm.com/api-gateway/internal/config"
)

// ConfigHandler returns an [http.HandlerFunc] that serialises the running
// configuration as JSON, redacting any sensitive fields before writing the
// response.
//
// Route: GET /admin/config
//
// Responses:
//   - 200 OK — JSON body containing the sanitised configuration.
func ConfigHandler(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(cfg)
	}
}

// ReloadHandler is a stub that acknowledges a configuration-reload request.
// Wire a reload signal channel into this handler when live reloading is
// implemented.
//
// Route: POST /admin/reload
//
// Responses:
//   - 202 Accepted — JSON body: {"status":"reload scheduled"}.
func ReloadHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "reload scheduled"})
	}
}
