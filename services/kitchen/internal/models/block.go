package models

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// BlockType enumerates supported block content types.
const (
	BlockTypeRichText     = "rich_text"
	BlockTypeHeading1     = "heading_1"
	BlockTypeHeading2     = "heading_2"
	BlockTypeHeading3     = "heading_3"
	BlockTypeCode         = "code"
	BlockTypeImage        = "image"
	BlockTypeDivider      = "divider"
	BlockTypeBulletedList = "bulleted_list"
	BlockTypeNumberedList = "numbered_list"
)

// Block is the MongoDB document stored in the `blocks` collection.
// Blocks within a document form a singly linked list via InsertAfterBlockID.
// A nil InsertAfterBlockID means the block is at the top of the document.
// ContentState stores the full Automerge/Yjs binary CRDT state for this block.
// ContentMeta stores type-specific metadata (e.g. {"language":"python"} for code,
// {"url":"...","alt_text":"..."} for image).
type Block struct {
	ID                 bson.ObjectID          `bson:"_id,omitempty"               json:"id"`
	DocumentID         bson.ObjectID          `bson:"document_id"                 json:"document_id"`
	BlockType          string                 `bson:"block_type"                  json:"block_type"`
	ContentState       []byte                 `bson:"content_state"               json:"-"`
	ContentMeta        map[string]interface{} `bson:"content_meta,omitempty"      json:"content_meta,omitempty"`
	InsertAfterBlockID *bson.ObjectID         `bson:"insert_after_block_id,omitempty" json:"insert_after_block_id,omitempty"`
	DeletedAt          *time.Time             `bson:"deleted_at,omitempty"        json:"deleted_at,omitempty"`
}

// ── HTTP request / response types ────────────────────────────────────────────

// InsertBlockRequest is the request body for POST /api/v1/documents/{id}/blocks.
// ContentState is the base64-encoded initial CRDT binary state for the block.
// InsertAfterBlockID is optional; when omitted the block is inserted at the top.
type InsertBlockRequest struct {
	BlockType          string                 `json:"block_type"`
	ContentState       []byte                 `json:"content_state"`       // raw bytes; JSON clients send base64
	ContentMeta        map[string]interface{} `json:"content_meta,omitempty"`
	InsertAfterBlockID *string                `json:"insert_after_block_id,omitempty"`
}

// UpdateBlockRequest is the request body for PATCH /api/v1/blocks/{id}.
// ContentState is the new full CRDT binary state (not a delta).
type UpdateBlockRequest struct {
	ContentState []byte                 `json:"content_state"`
	ContentMeta  map[string]interface{} `json:"content_meta,omitempty"`
}

// ReorderBlockRequest is the request body for PATCH /api/v1/blocks/{id}/reorder.
// InsertAfterBlockID is the ID of the block this block should follow.
// A nil value moves the block to the top of the document.
type ReorderBlockRequest struct {
	InsertAfterBlockID *string `json:"insert_after_block_id"`
}

// BlockResponse is the public representation of a block returned by the API.
// ContentState is base64-encoded so it round-trips cleanly through JSON.
type BlockResponse struct {
	ID                 string                 `json:"id"`
	DocumentID         string                 `json:"document_id"`
	BlockType          string                 `json:"block_type"`
	ContentState       []byte                 `json:"content_state"` // JSON marshals []byte as base64
	ContentMeta        map[string]interface{} `json:"content_meta,omitempty"`
	InsertAfterBlockID *string                `json:"insert_after_block_id,omitempty"`
	DeletedAt          *time.Time             `json:"deleted_at,omitempty"`
}

// ToResponse converts a Block DB model to its API response shape.
func (b *Block) ToResponse() BlockResponse {
	resp := BlockResponse{
		ID:           b.ID.Hex(),
		DocumentID:   b.DocumentID.Hex(),
		BlockType:    b.BlockType,
		ContentState: b.ContentState,
		ContentMeta:  b.ContentMeta,
		DeletedAt:    b.DeletedAt,
	}
	if b.InsertAfterBlockID != nil {
		s := b.InsertAfterBlockID.Hex()
		resp.InsertAfterBlockID = &s
	}
	return resp
}
