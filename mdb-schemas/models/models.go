package models

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// User is the document stored in the `users` MongoDB collection.
type User struct {
	ID           bson.ObjectID `bson:"_id,omitempty"`
	Email        string        `bson:"email"`
	Username     string        `bson:"username"`
	PasswordHash string        `bson:"password_hash"`
	Role         string        `bson:"role"`
	CreatedAt    time.Time     `bson:"created_at"`
	UpdatedAt    time.Time     `bson:"updated_at"`
}

// Session is the document stored in the `sessions` MongoDB collection.
// TokenHash holds the SHA-256 hex digest of the raw UUID refresh token.
type Session struct {
	ID        bson.ObjectID `bson:"_id,omitempty"`
	UserID    bson.ObjectID `bson:"user_id"`
	TokenHash string        `bson:"token_hash"`
	ExpiresAt time.Time     `bson:"expires_at"`
	CreatedAt time.Time     `bson:"created_at"`
}
