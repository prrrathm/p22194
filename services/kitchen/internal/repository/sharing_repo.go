package repository

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"p22194.prrrathm.com/kitchen/internal/models"
)

// SharingRepo performs CRUD operations on the `document_shares` collection.
type SharingRepo struct {
	col *mongo.Collection
}

// NewSharingRepo creates a SharingRepo backed by the given database.
func NewSharingRepo(db *mongo.Database) *SharingRepo {
	return &SharingRepo{col: db.Collection("document_shares")}
}

// EnsureIndexes creates the required indexes on startup. Idempotent.
func (r *SharingRepo) EnsureIndexes(ctx context.Context) error {
	indexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "document_id", Value: 1}, {Key: "user_id", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
		{Keys: bson.D{{Key: "document_id", Value: 1}}},
	}
	if _, err := r.col.Indexes().CreateMany(ctx, indexes); err != nil {
		return fmt.Errorf("sharing_repo: ensure indexes: %w", err)
	}
	return nil
}

// Create inserts a new share record and sets s.ID on success.
// Returns mongo.WriteException with code 11000 if the (document_id, user_id) pair already exists.
func (r *SharingRepo) Create(ctx context.Context, s *models.DocumentShare) error {
	result, err := r.col.InsertOne(ctx, s)
	if err != nil {
		return fmt.Errorf("sharing_repo: insert: %w", err)
	}
	s.ID = result.InsertedID.(bson.ObjectID)
	return nil
}

// FindByDocumentAndUser returns the share record for a (document, user) pair.
// Returns mongo.ErrNoDocuments when not found.
func (r *SharingRepo) FindByDocumentAndUser(ctx context.Context, documentID, userID bson.ObjectID) (*models.DocumentShare, error) {
	var s models.DocumentShare
	err := r.col.FindOne(ctx, bson.M{"document_id": documentID, "user_id": userID}).Decode(&s)
	if err != nil {
		return nil, fmt.Errorf("sharing_repo: find: %w", err)
	}
	return &s, nil
}

// ListByDocument returns all share records for a document.
func (r *SharingRepo) ListByDocument(ctx context.Context, documentID bson.ObjectID) ([]models.DocumentShare, error) {
	cur, err := r.col.Find(ctx, bson.M{"document_id": documentID})
	if err != nil {
		return nil, fmt.Errorf("sharing_repo: list: %w", err)
	}
	defer cur.Close(ctx)

	var shares []models.DocumentShare
	if err := cur.All(ctx, &shares); err != nil {
		return nil, fmt.Errorf("sharing_repo: decode: %w", err)
	}
	return shares, nil
}

// Delete removes the share record for a (document, user) pair.
func (r *SharingRepo) Delete(ctx context.Context, documentID, userID bson.ObjectID) error {
	_, err := r.col.DeleteOne(ctx, bson.M{"document_id": documentID, "user_id": userID})
	if err != nil {
		return fmt.Errorf("sharing_repo: delete: %w", err)
	}
	return nil
}
