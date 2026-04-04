package repository

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"

	"p22194.prrrathm.com/mdb/models"
)

// SessionRepo performs CRUD on the `sessions` collection.
type SessionRepo struct {
	col *mongo.Collection
}

// NewSessionRepo creates a SessionRepo backed by the named database.
func NewSessionRepo(db *mongo.Database) *SessionRepo {
	return &SessionRepo{col: db.Collection("sessions")}
}

// Create inserts a new session. Populates s.ID with the generated ObjectID on success.
func (r *SessionRepo) Create(ctx context.Context, s *models.Session) error {
	result, err := r.col.InsertOne(ctx, s)
	if err != nil {
		return fmt.Errorf("session_repo: insert: %w", err)
	}
	s.ID = result.InsertedID.(bson.ObjectID)
	return nil
}

// FindByTokenHash looks up a session by its SHA-256 token hash.
// Returns mongo.ErrNoDocuments when not found.
func (r *SessionRepo) FindByTokenHash(ctx context.Context, hash string) (*models.Session, error) {
	var s models.Session
	err := r.col.FindOne(ctx, bson.M{"token_hash": hash}).Decode(&s)
	if err != nil {
		return nil, fmt.Errorf("session_repo: find: %w", err)
	}
	return &s, nil
}

// DeleteByTokenHash removes the session matching the given hash.
// Idempotent — returns nil even when no document was matched.
func (r *SessionRepo) DeleteByTokenHash(ctx context.Context, hash string) error {
	_, err := r.col.DeleteOne(ctx, bson.M{"token_hash": hash})
	if err != nil {
		return fmt.Errorf("session_repo: delete: %w", err)
	}
	return nil
}

// DeleteAllForUser removes every session belonging to userID.
func (r *SessionRepo) DeleteAllForUser(ctx context.Context, userID bson.ObjectID) error {
	_, err := r.col.DeleteMany(ctx, bson.M{"user_id": userID})
	if err != nil {
		return fmt.Errorf("session_repo: delete all: %w", err)
	}
	return nil
}
