package handler

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"
	"go.mongodb.org/mongo-driver/v2/bson"

	"p22194.prrrathm.com/kitchen/internal/middleware"
	"p22194.prrrathm.com/kitchen/internal/models"
	"p22194.prrrathm.com/kitchen/internal/service"
)

// SharingHandler holds HTTP handlers for document sharing endpoints.
type SharingHandler struct {
	svc *service.SharingService
	log zerolog.Logger
}

// NewSharingHandler constructs a SharingHandler.
func NewSharingHandler(svc *service.SharingService, log zerolog.Logger) *SharingHandler {
	return &SharingHandler{svc: svc, log: log}
}

// Share handles POST /api/v1/documents/{id}/shares.
//
// Grants a user access to a document with the specified role.
// The requesting user becomes the inviter.
//
// Auth: Bearer JWT required.
//
// Path param:
//
//	id — ObjectID hex of the document
//
// Body (JSON):
//
//	{
//	  "user_id": string  (required) — ObjectID hex of the user to invite
//	  "role":    string  (required) — owner | editor | viewer
//	}
//
// Response 201:
//
//	{ DocumentShareResponse }
//
// Response 409:
//
//	{"error": "user already has access to this document"}
//
// Example:
//
//	POST /api/v1/documents/665f1a2b3c4d5e6f7a8b9c0d/shares
//	Authorization: Bearer <token>
//	{"user_id":"665f1a2b3c4d5e6f7a8b9c1a","role":"editor"}
//
//	→ 201 {"id":"...","document_id":"665f...","user_id":"665f...","role":"editor",...}
func (h *SharingHandler) Share(w http.ResponseWriter, r *http.Request) {
	docID, ok := parseSharingDocID(w, r)
	if !ok {
		return
	}
	inviterID, ok := h.requireUserID(w, r)
	if !ok {
		return
	}

	var req models.ShareDocumentRequest
	if !decodeJSON(w, r, &req) {
		return
	}

	share, err := h.svc.Share(r.Context(), docID, inviterID, req)
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusCreated, share.ToResponse())
}

// RemoveAccess handles DELETE /api/v1/documents/{id}/shares/{user_id}.
//
// Revokes a user's access to a document.
//
// Auth: Bearer JWT required.
//
// Path params:
//
//	id      — ObjectID hex of the document
//	user_id — ObjectID hex of the user whose access to revoke
//
// Response 200:
//
//	{"status": "removed"}
//
// Example:
//
//	DELETE /api/v1/documents/665f1a2b3c4d5e6f7a8b9c0d/shares/665f1a2b3c4d5e6f7a8b9c1a
//	Authorization: Bearer <token>
//
//	→ 200 {"status":"removed"}
func (h *SharingHandler) RemoveAccess(w http.ResponseWriter, r *http.Request) {
	docID, ok := parseSharingDocID(w, r)
	if !ok {
		return
	}

	userIDRaw := chi.URLParam(r, "user_id")
	userID, err := bson.ObjectIDFromHex(userIDRaw)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorBody("invalid user_id"))
		return
	}

	if err := h.svc.RemoveAccess(r.Context(), docID, userID); err != nil {
		h.writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "removed"})
}

// GetShares handles GET /api/v1/documents/{id}/shares.
//
// Returns all users who have been granted access to the document, including their roles.
//
// Auth: Bearer JWT required.
//
// Path param:
//
//	id — ObjectID hex of the document
//
// Response 200:
//
//	{
//	  "shares": [ DocumentShareResponse, ... ]
//	}
//
// Example:
//
//	GET /api/v1/documents/665f1a2b3c4d5e6f7a8b9c0d/shares
//	Authorization: Bearer <token>
//
//	→ 200 {"shares":[{"id":"...","user_id":"...","role":"editor",...}]}
func (h *SharingHandler) GetShares(w http.ResponseWriter, r *http.Request) {
	docID, ok := parseSharingDocID(w, r)
	if !ok {
		return
	}

	shares, err := h.svc.GetShares(r.Context(), docID)
	if err != nil {
		h.writeError(w, r, err)
		return
	}

	resps := make([]models.DocumentShareResponse, len(shares))
	for i, s := range shares {
		resps[i] = s.ToResponse()
	}
	writeJSON(w, http.StatusOK, map[string]any{"shares": resps})
}

// ── Private helpers ──────────────────────────────────────────────────────────

func parseSharingDocID(w http.ResponseWriter, r *http.Request) (bson.ObjectID, bool) {
	raw := chi.URLParam(r, "id")
	id, err := bson.ObjectIDFromHex(raw)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorBody("invalid id"))
		return bson.ObjectID{}, false
	}
	return id, true
}

func (h *SharingHandler) requireUserID(w http.ResponseWriter, r *http.Request) (bson.ObjectID, bool) {
	sub := middleware.UserIDFromContext(r.Context())
	if sub == "" {
		writeJSON(w, http.StatusUnauthorized, errorBody("unauthorized"))
		return bson.ObjectID{}, false
	}
	id, err := bson.ObjectIDFromHex(sub)
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, errorBody("invalid user id in token"))
		return bson.ObjectID{}, false
	}
	return id, true
}

func (h *SharingHandler) writeError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, service.ErrShareExists):
		writeJSON(w, http.StatusConflict, errorBody("user already has access to this document"))
	case errors.Is(err, service.ErrShareNotFound):
		writeJSON(w, http.StatusNotFound, errorBody("share not found"))
	default:
		h.log.Error().Err(err).Str("path", r.URL.Path).Msg("unhandled service error")
		writeJSON(w, http.StatusInternalServerError, errorBody("internal server error"))
	}
}
