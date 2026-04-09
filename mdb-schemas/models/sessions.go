package models

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// Session is the document stored in the `sessions` MongoDB collection.
// TokenHash holds the SHA-256 hex digest of the raw UUID refresh token.
type Session struct {
	ID        bson.ObjectID `bson:"_id,omitempty"`
	UserID    bson.ObjectID `bson:"user_id"`
	TokenHash string        `bson:"token_hash"`
	ExpiresAt time.Time     `bson:"expires_at"`
	CreatedAt time.Time     `bson:"created_at"`
}
