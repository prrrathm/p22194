package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"
	"go.mongodb.org/mongo-driver/v2/bson"

	"p22194.prrrathm.com/kitchen/internal/middleware"
	"p22194.prrrathm.com/kitchen/internal/models"
	"p22194.prrrathm.com/kitchen/internal/service"
)

// DocumentHandler holds HTTP handlers for all document endpoints.
type DocumentHandler struct {
	svc *service.DocumentService
	log zerolog.Logger
}

// NewDocumentHandler constructs a DocumentHandler.
func NewDocumentHandler(svc *service.DocumentService, log zerolog.Logger) *DocumentHandler {
	return &DocumentHandler{svc: svc, log: log}
}

// Create handles POST /api/v1/documents.
//
// Auth: Bearer JWT required.
//
// Body (JSON):
//
//	{
//	  "title":              string  (required) — document title
//	  "title_icon":         string  (optional) — emoji or icon identifier
//	  "parent_document_id": string  (optional) — ObjectID hex of parent document
//	}
//
// Response 201:
//
//	{ DocumentResponse }
//
// Example:
//
//	POST /api/v1/documents
//	Authorization: Bearer <token>
//	{"title": "My Notes", "title_icon": "📝"}
//
//	→ 201 {"id":"...","title":"My Notes","status":"active",...}
func (h *DocumentHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.requireUserID(w, r)
	if !ok {
		return
	}

	var req models.CreateDocumentRequest
	if !decodeJSON(w, r, &req) {
		return
	}

	doc, err := h.svc.Create(r.Context(), userID, req)
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	resp := doc.ToResponse()
	writeJSON(w, http.StatusCreated, resp)
}

// GetByID handles GET /api/v1/documents/{id}.
//
// Auth: Bearer JWT required.
//
// Path param:
//
//	id — ObjectID hex of the document
//
// Response 200:
//
//	{ DocumentResponse }
//
// Response 404:
//
//	{"error": "document not found"}
//
// Example:
//
//	GET /api/v1/documents/665f1a2b3c4d5e6f7a8b9c0d
//	Authorization: Bearer <token>
//
//	→ 200 {"id":"665f1a2b3c4d5e6f7a8b9c0d","title":"My Notes",...}
func (h *DocumentHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, ok := h.parseIDParam(w, r, "id")
	if !ok {
		return
	}

	doc, err := h.svc.GetByID(r.Context(), id)
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, doc.ToResponse())
}

// List handles GET /api/v1/documents.
//
// Auth: Bearer JWT required.
//
// Query params:
//
//	status    string  (optional) — filter by status: active | archived | deleted; default returns all non-deleted
//	parent_id string  (optional) — ObjectID hex to filter by parent document
//	limit     int     (optional) — max results per page, default 20, max 100
//	cursor    string  (optional) — ObjectID hex of last seen document (for pagination)
//
// Response 200:
//
//	{
//	  "documents":   [ DocumentResponse, ... ],
//	  "next_cursor": string  (empty when no more pages)
//	}
//
// Example:
//
//	GET /api/v1/documents?status=active&limit=10
//	Authorization: Bearer <token>
//
//	→ 200 {"documents":[...],"next_cursor":"665f..."}
func (h *DocumentHandler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.requireUserID(w, r)
	if !ok {
		return
	}

	q := r.URL.Query()
	status := q.Get("status")
	parentID := q.Get("parent_id")
	cursor := q.Get("cursor")

	limit := 20
	if l := q.Get("limit"); l != "" {
		if n, err := strconv.Atoi(l); err == nil {
			limit = n
		}
	}

	docs, err := h.svc.List(r.Context(), userID, status, parentID, cursor, limit)
	if err != nil {
		h.writeError(w, r, err)
		return
	}

	resps := make([]models.DocumentResponse, len(docs))
	for i, d := range docs {
		resps[i] = d.ToResponse()
	}

	nextCursor := ""
	if len(docs) == limit {
		nextCursor = docs[len(docs)-1].ID.Hex()
	}

	writeJSON(w, http.StatusOK, models.ListDocumentsResponse{
		Documents:  resps,
		NextCursor: nextCursor,
	})
}

// ListChildren handles GET /api/v1/documents/{id}/children.
//
// Auth: Bearer JWT required.
//
// Path param:
//
//	id — ObjectID hex of the parent document
//
// Response 200:
//
//	{ "documents": [ DocumentResponse, ... ] }
//
// Example:
//
//	GET /api/v1/documents/665f1a2b3c4d5e6f7a8b9c0d/children
//	Authorization: Bearer <token>
//
//	→ 200 {"documents":[{"id":"...","parent_document_id":"665f1a2b3c4d5e6f7a8b9c0d",...}]}
func (h *DocumentHandler) ListChildren(w http.ResponseWriter, r *http.Request) {
	id, ok := h.parseIDParam(w, r, "id")
	if !ok {
		return
	}

	docs, err := h.svc.ListChildren(r.Context(), id)
	if err != nil {
		h.writeError(w, r, err)
		return
	}

	resps := make([]models.DocumentResponse, len(docs))
	for i, d := range docs {
		resps[i] = d.ToResponse()
	}
	writeJSON(w, http.StatusOK, map[string]any{"documents": resps})
}

// Update handles PATCH /api/v1/documents/{id}.
//
// Updates title and/or title_icon. Omitted fields are left unchanged.
//
// Auth: Bearer JWT required.
//
// Response 200: { DocumentResponse }
func (h *DocumentHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, ok := h.parseIDParam(w, r, "id")
	if !ok {
		return
	}

	var req models.UpdateDocumentRequest
	if !decodeJSON(w, r, &req) {
		return
	}

	doc, err := h.svc.Update(r.Context(), id, req)
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, doc.ToResponse())
}

// Delete handles DELETE /api/v1/documents/{id}.
//
// Soft-deletes the document (sets status=deleted, deleted_at=now).
//
// Auth: Bearer JWT required.
//
// Path param:
//
//	id — ObjectID hex of the document
//
// Response 200:
//
//	{"status": "deleted"}
//
// Response 404:
//
//	{"error": "document not found"}
//
// Example:
//
//	DELETE /api/v1/documents/665f1a2b3c4d5e6f7a8b9c0d
//	Authorization: Bearer <token>
//
//	→ 200 {"status":"deleted"}
func (h *DocumentHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, ok := h.parseIDParam(w, r, "id")
	if !ok {
		return
	}

	if err := h.svc.Delete(r.Context(), id); err != nil {
		h.writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// Archive handles PATCH /api/v1/documents/{id}/archive.
//
// Sets the document status to "archived".
//
// Auth: Bearer JWT required.
//
// Path param:
//
//	id — ObjectID hex of the document
//
// Response 200:
//
//	{"status": "archived"}
//
// Example:
//
//	PATCH /api/v1/documents/665f1a2b3c4d5e6f7a8b9c0d/archive
//	Authorization: Bearer <token>
//
//	→ 200 {"status":"archived"}
func (h *DocumentHandler) Archive(w http.ResponseWriter, r *http.Request) {
	id, ok := h.parseIDParam(w, r, "id")
	if !ok {
		return
	}

	if err := h.svc.Archive(r.Context(), id); err != nil {
		h.writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "archived"})
}

// ── Private helpers ──────────────────────────────────────────────────────────

func (h *DocumentHandler) requireUserID(w http.ResponseWriter, r *http.Request) (bson.ObjectID, bool) {
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

func (h *DocumentHandler) parseIDParam(w http.ResponseWriter, r *http.Request, param string) (bson.ObjectID, bool) {
	raw := chi.URLParam(r, param)
	id, err := bson.ObjectIDFromHex(raw)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorBody("invalid id"))
		return bson.ObjectID{}, false
	}
	return id, true
}

func (h *DocumentHandler) writeError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, service.ErrDocumentNotFound):
		writeJSON(w, http.StatusNotFound, errorBody("document not found"))
	default:
		h.log.Error().Err(err).Str("path", r.URL.Path).Msg("unhandled service error")
		writeJSON(w, http.StatusInternalServerError, errorBody("internal server error"))
	}
}
