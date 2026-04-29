package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/rs/zerolog"

	"p22194.prrrathm.com/mail/internal/service"
)

type MailHandler struct {
	mailer service.Mailer
	log    zerolog.Logger
}

type verificationEmailRequest struct {
	ToEmail          string `json:"to_email"`
	Username         string `json:"username"`
	VerificationLink string `json:"verification_link"`
}

type documentShareEmailRequest struct {
	ToEmail       string `json:"to_email"`
	DocumentTitle string `json:"document_title"`
	DocumentLink  string `json:"document_link"`
	Role          string `json:"role"`
}

func New(mailer service.Mailer, log zerolog.Logger) *MailHandler {
	return &MailHandler{mailer: mailer, log: log}
}

func (h *MailHandler) Health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *MailHandler) SendVerificationEmail(w http.ResponseWriter, r *http.Request) {
	var req verificationEmailRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	if strings.TrimSpace(req.ToEmail) == "" || strings.TrimSpace(req.VerificationLink) == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "to_email and verification_link are required"})
		return
	}

	if err := h.mailer.SendVerificationEmail(r.Context(), req.ToEmail, req.Username, req.VerificationLink); err != nil {
		h.log.Error().Err(err).Str("to", req.ToEmail).Msg("failed to send verification email")
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": "failed to send verification email"})
		return
	}

	writeJSON(w, http.StatusAccepted, map[string]string{"status": "verification email sent"})
}

func (h *MailHandler) SendDocumentShareEmail(w http.ResponseWriter, r *http.Request) {
	var req documentShareEmailRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	if strings.TrimSpace(req.ToEmail) == "" || strings.TrimSpace(req.DocumentLink) == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "to_email and document_link are required"})
		return
	}

	if err := h.mailer.SendDocumentShareEmail(r.Context(), req.ToEmail, req.DocumentTitle, req.DocumentLink, req.Role); err != nil {
		h.log.Error().Err(err).Str("to", req.ToEmail).Msg("failed to send document share email")
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": "failed to send document share email"})
		return
	}

	writeJSON(w, http.StatusAccepted, map[string]string{"status": "document share email sent"})
}

func decodeJSON(w http.ResponseWriter, r *http.Request, v any) bool {
	if err := json.NewDecoder(r.Body).Decode(v); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return false
	}
	return true
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
