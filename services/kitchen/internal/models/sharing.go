package models

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// DocumentShareRole enumerates valid values for DocumentShare.Role.
const (
	RoleOwner  = "owner"
	RoleEditor = "editor"
	RoleViewer = "viewer"
)

// DocumentShare is the MongoDB document stored in the `document_shares` collection.
// It represents a user's access grant to a document with an assigned role.
type DocumentShare struct {
	ID              bson.ObjectID `bson:"_id,omitempty"        json:"id"`
	DocumentID      bson.ObjectID `bson:"document_id"          json:"document_id"`
	UserID          bson.ObjectID `bson:"user_id"              json:"user_id"`
	Role            string        `bson:"role"                 json:"role"`
	InvitedByUserID bson.ObjectID `bson:"invited_by_user_id"   json:"invited_by_user_id"`
	CreatedAt       time.Time     `bson:"created_at"           json:"created_at"`
}

// ── HTTP request / response types ────────────────────────────────────────────

// ShareDocumentRequest is the request body for POST /api/v1/documents/{id}/shares.
type ShareDocumentRequest struct {
	UserID string `json:"user_id"`
	Role   string `json:"role"` // owner | editor | viewer
}

// DocumentShareResponse is the public representation of a share record.
type DocumentShareResponse struct {
	ID              string    `json:"id"`
	DocumentID      string    `json:"document_id"`
	UserID          string    `json:"user_id"`
	Role            string    `json:"role"`
	InvitedByUserID string    `json:"invited_by_user_id"`
	CreatedAt       time.Time `json:"created_at"`
}

// ToResponse converts a DocumentShare DB model to its API response shape.
func (s *DocumentShare) ToResponse() DocumentShareResponse {
	return DocumentShareResponse{
		ID:              s.ID.Hex(),
		DocumentID:      s.DocumentID.Hex(),
		UserID:          s.UserID.Hex(),
		Role:            s.Role,
		InvitedByUserID: s.InvitedByUserID.Hex(),
		CreatedAt:       s.CreatedAt,
	}
}
