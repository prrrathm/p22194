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
