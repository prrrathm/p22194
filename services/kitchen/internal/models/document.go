package models

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// ── MongoDB document model ────────────────────────────────────────────────────

// DocumentStatus enumerates valid values for Document.Status.
const (
	StatusActive   = "active"
	StatusArchived = "archived"
	StatusDeleted  = "deleted"
)

// Document is the MongoDB document stored in the `documents` collection.
// Blocks are ordered via a linked list (insert_after_block_id on each Block).
// Snapshot caches the resolved ordered slice of block IDs for fast full reads;
// it is rebuilt synchronously on every structural change (insert/delete/reorder).
type Document struct {
	ID               bson.ObjectID  `bson:"_id,omitempty"        json:"id"`
	CreatedByUserID  bson.ObjectID  `bson:"created_by_user_id"   json:"created_by_user_id"`
	CreatedAt        time.Time      `bson:"created_at"           json:"created_at"`
	LastUpdatedAt    time.Time      `bson:"last_updated_at"      json:"last_updated_at"`
	Title            string         `bson:"title"                json:"title"`
	TitleIcon        string         `bson:"title_icon,omitempty" json:"title_icon,omitempty"`
	DeletedAt        *time.Time     `bson:"deleted_at,omitempty" json:"deleted_at,omitempty"`
	Status           string         `bson:"status"               json:"status"`
	ParentDocumentID *bson.ObjectID `bson:"parent_document_id,omitempty" json:"parent_document_id,omitempty"`
	Snapshot         []string       `bson:"snapshot"             json:"snapshot"`
}

// ── HTTP request / response types ────────────────────────────────────────────

// CreateDocumentRequest is the request body for POST /api/v1/documents.
type CreateDocumentRequest struct {
	Title            string  `json:"title"`
	TitleIcon        string  `json:"title_icon,omitempty"`
	ParentDocumentID *string `json:"parent_document_id,omitempty"`
}

// UpdateDocumentRequest is the request body for PATCH /api/v1/documents/{id}.
// All fields are optional; only non-zero values are applied.
type UpdateDocumentRequest struct {
	Title     string `json:"title,omitempty"`
	TitleIcon string `json:"title_icon,omitempty"`
}

// DocumentResponse is the public representation of a document returned by the API.
type DocumentResponse struct {
	ID               string     `json:"id"`
	CreatedByUserID  string     `json:"created_by_user_id"`
	CreatedAt        time.Time  `json:"created_at"`
	LastUpdatedAt    time.Time  `json:"last_updated_at"`
	Title            string     `json:"title"`
	TitleIcon        string     `json:"title_icon,omitempty"`
	DeletedAt        *time.Time `json:"deleted_at,omitempty"`
	Status           string     `json:"status"`
	ParentDocumentID *string    `json:"parent_document_id,omitempty"`
	Snapshot         []string   `json:"snapshot"`
}

// ListDocumentsResponse wraps a page of documents with a cursor for the next page.
type ListDocumentsResponse struct {
	Documents  []DocumentResponse `json:"documents"`
	NextCursor string             `json:"next_cursor,omitempty"`
}

// ToResponse converts a Document DB model to its API response shape.
func (d *Document) ToResponse() DocumentResponse {
	resp := DocumentResponse{
		ID:            d.ID.Hex(),
		CreatedByUserID: d.CreatedByUserID.Hex(),
		CreatedAt:     d.CreatedAt,
		LastUpdatedAt: d.LastUpdatedAt,
		Title:         d.Title,
		TitleIcon:     d.TitleIcon,
		DeletedAt:     d.DeletedAt,
		Status:        d.Status,
		Snapshot:      d.Snapshot,
	}
	if d.ParentDocumentID != nil {
		s := d.ParentDocumentID.Hex()
		resp.ParentDocumentID = &s
	}
	return resp
}
