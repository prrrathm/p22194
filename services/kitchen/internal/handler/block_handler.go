package handler

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"
	"go.mongodb.org/mongo-driver/v2/bson"

	"p22194.prrrathm.com/kitchen/internal/models"
	"p22194.prrrathm.com/kitchen/internal/service"
)

// BlockHandler holds HTTP handlers for all block endpoints.
type BlockHandler struct {
	svc *service.BlockService
	log zerolog.Logger
}

// NewBlockHandler constructs a BlockHandler.
func NewBlockHandler(svc *service.BlockService, log zerolog.Logger) *BlockHandler {
	return &BlockHandler{svc: svc, log: log}
}

// Insert handles POST /api/v1/documents/{id}/blocks.
//
// Creates a new block inside the given document.
// The block is appended to the end of the document by default.
// Supply insert_after_block_id to place the block after a specific block;
// omit or set to null to insert at the top.
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
//	  "block_type":            string  (required) — rich_text | heading_1 | heading_2 | heading_3 | code | image | divider | bulleted_list | numbered_list
//	  "content_state":         bytes   (optional) — base64-encoded initial CRDT binary state
//	  "content_meta":          object  (optional) — type-specific metadata
//	    for code:  {"language": "python"}
//	    for image: {"url": "https://...", "alt_text": "description"}
//	  "insert_after_block_id": string  (optional) — ObjectID hex of the preceding block; null = insert at top
//	}
//
// Response 201:
//
//	{ BlockResponse }
//
// Example:
//
//	POST /api/v1/documents/665f1a2b3c4d5e6f7a8b9c0d/blocks
//	Authorization: Bearer <token>
//	{"block_type":"rich_text","content_state":"<base64>"}
//
//	→ 201 {"id":"...","document_id":"665f...","block_type":"rich_text",...}
func (h *BlockHandler) Insert(w http.ResponseWriter, r *http.Request) {
	docID, ok := parseObjectID(w, r, "id")
	if !ok {
		return
	}

	var req models.InsertBlockRequest
	if !decodeJSON(w, r, &req) {
		return
	}

	block, err := h.svc.Insert(r.Context(), docID, req)
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusCreated, block.ToResponse())
}

// Update handles PATCH /api/v1/blocks/{id}.
//
// Replaces the content_state (full CRDT binary state, not a delta) of a block.
// Optionally updates content_meta for type-specific metadata changes.
//
// Auth: Bearer JWT required.
//
// Path param:
//
//	id — ObjectID hex of the block
//
// Body (JSON):
//
//	{
//	  "content_state": bytes   (required) — base64-encoded full CRDT binary state
//	  "content_meta":  object  (optional) — updated type-specific metadata
//	}
//
// Response 200:
//
//	{ BlockResponse }
//
// Response 404:
//
//	{"error": "block not found"}
//
// Example:
//
//	PATCH /api/v1/blocks/665f1a2b3c4d5e6f7a8b9c0e
//	Authorization: Bearer <token>
//	{"content_state":"<base64-updated>"}
//
//	→ 200 {"id":"665f...","content_state":"<base64-updated>",...}
func (h *BlockHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, ok := parseObjectID(w, r, "id")
	if !ok {
		return
	}

	var req models.UpdateBlockRequest
	if !decodeJSON(w, r, &req) {
		return
	}

	block, err := h.svc.Update(r.Context(), id, req)
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, block.ToResponse())
}

// Reorder handles PATCH /api/v1/blocks/{id}/reorder.
//
// Moves a block to a new position within its document by updating its
// insert_after_block_id pointer. Pass null to move the block to the top.
// The document snapshot is rebuilt after reordering.
//
// Auth: Bearer JWT required.
//
// Path param:
//
//	id — ObjectID hex of the block to move
//
// Body (JSON):
//
//	{
//	  "insert_after_block_id": string | null  — ObjectID hex of the block this block should follow; null = move to top
//	}
//
// Response 200:
//
//	{ BlockResponse }
//
// Example (move to top):
//
//	PATCH /api/v1/blocks/665f1a2b3c4d5e6f7a8b9c0e/reorder
//	Authorization: Bearer <token>
//	{"insert_after_block_id": null}
//
//	→ 200 {"id":"665f...","insert_after_block_id":null,...}
func (h *BlockHandler) Reorder(w http.ResponseWriter, r *http.Request) {
	id, ok := parseObjectID(w, r, "id")
	if !ok {
		return
	}

	var req models.ReorderBlockRequest
	if !decodeJSON(w, r, &req) {
		return
	}

	block, err := h.svc.Reorder(r.Context(), id, req)
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, block.ToResponse())
}

// Delete handles DELETE /api/v1/blocks/{id}.
//
// Soft-deletes a block (sets deleted_at=now). The block remains in the database
// but is excluded from document reads. The document snapshot is rebuilt after deletion.
//
// Auth: Bearer JWT required.
//
// Path param:
//
//	id — ObjectID hex of the block
//
// Response 200:
//
//	{"status": "deleted"}
//
// Response 404:
//
//	{"error": "block not found"}
//
// Example:
//
//	DELETE /api/v1/blocks/665f1a2b3c4d5e6f7a8b9c0e
//	Authorization: Bearer <token>
//
//	→ 200 {"status":"deleted"}
func (h *BlockHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, ok := parseObjectID(w, r, "id")
	if !ok {
		return
	}

	if err := h.svc.Delete(r.Context(), id); err != nil {
		h.writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// ── Private helpers ──────────────────────────────────────────────────────────

func parseObjectID(w http.ResponseWriter, r *http.Request, param string) (bson.ObjectID, bool) {
	raw := chi.URLParam(r, param)
	id, err := bson.ObjectIDFromHex(raw)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorBody("invalid id"))
		return bson.ObjectID{}, false
	}
	return id, true
}

func (h *BlockHandler) writeError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, service.ErrBlockNotFound):
		writeJSON(w, http.StatusNotFound, errorBody("block not found"))
	default:
		h.log.Error().Err(err).Str("path", r.URL.Path).Msg("unhandled service error")
		writeJSON(w, http.StatusInternalServerError, errorBody("internal server error"))
	}
}
