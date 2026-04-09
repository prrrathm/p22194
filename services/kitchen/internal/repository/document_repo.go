package repository

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"p22194.prrrathm.com/kitchen/internal/models"
)

// DocumentRepo performs CRUD operations on the `documents` collection.
type DocumentRepo struct {
	col *mongo.Collection
}

// NewDocumentRepo creates a DocumentRepo backed by the given database.
func NewDocumentRepo(db *mongo.Database) *DocumentRepo {
	return &DocumentRepo{col: db.Collection("documents")}
}

// EnsureIndexes creates the required indexes on startup. Idempotent.
func (r *DocumentRepo) EnsureIndexes(ctx context.Context) error {
	indexes := []mongo.IndexModel{
		{Keys: bson.D{{Key: "created_by_user_id", Value: 1}, {Key: "status", Value: 1}}},
		{Keys: bson.D{{Key: "parent_document_id", Value: 1}}},
	}
	if _, err := r.col.Indexes().CreateMany(ctx, indexes); err != nil {
		return fmt.Errorf("document_repo: ensure indexes: %w", err)
	}
	return nil
}

// Create inserts a new document and sets d.ID on success.
func (r *DocumentRepo) Create(ctx context.Context, d *models.Document) error {
	result, err := r.col.InsertOne(ctx, d)
	if err != nil {
		return fmt.Errorf("document_repo: insert: %w", err)
	}
	d.ID = result.InsertedID.(bson.ObjectID)
	return nil
}

// FindByID returns the document with the given ID.
// Returns mongo.ErrNoDocuments when not found.
func (r *DocumentRepo) FindByID(ctx context.Context, id bson.ObjectID) (*models.Document, error) {
	var d models.Document
	err := r.col.FindOne(ctx, bson.M{"_id": id}).Decode(&d)
	if err != nil {
		return nil, fmt.Errorf("document_repo: find by id: %w", err)
	}
	return &d, nil
}

// ListByUser returns a page of documents owned by or shared with userID.
// status filters by document status (pass "" to include all non-deleted).
// parentID filters by parent_document_id (pass nil for root documents).
// cursor is the last seen document ID for pagination (pass zero for first page).
// limit caps the result set.
func (r *DocumentRepo) ListByUser(
	ctx context.Context,
	userID bson.ObjectID,
	status string,
	parentID *bson.ObjectID,
	cursor bson.ObjectID,
	limit int,
) ([]models.Document, error) {
	filter := bson.M{"created_by_user_id": userID}

	if status != "" {
		filter["status"] = status
	} else {
		filter["status"] = bson.M{"$ne": models.StatusDeleted}
	}

	if parentID != nil {
		filter["parent_document_id"] = *parentID
	}

	if !cursor.IsZero() {
		filter["_id"] = bson.M{"$gt": cursor}
	}

	opts := options.Find().SetSort(bson.D{{Key: "_id", Value: 1}}).SetLimit(int64(limit))
	cur, err := r.col.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("document_repo: list: %w", err)
	}
	defer cur.Close(ctx)

	var docs []models.Document
	if err := cur.All(ctx, &docs); err != nil {
		return nil, fmt.Errorf("document_repo: decode list: %w", err)
	}
	return docs, nil
}

// ListChildren returns all non-deleted direct children of parentDocumentID.
func (r *DocumentRepo) ListChildren(ctx context.Context, parentDocumentID bson.ObjectID) ([]models.Document, error) {
	filter := bson.M{
		"parent_document_id": parentDocumentID,
		"status":             bson.M{"$ne": models.StatusDeleted},
	}
	cur, err := r.col.Find(ctx, filter, options.Find().SetSort(bson.D{{Key: "created_at", Value: 1}}))
	if err != nil {
		return nil, fmt.Errorf("document_repo: list children: %w", err)
	}
	defer cur.Close(ctx)

	var docs []models.Document
	if err := cur.All(ctx, &docs); err != nil {
		return nil, fmt.Errorf("document_repo: decode children: %w", err)
	}
	return docs, nil
}

// SoftDelete sets status=deleted and deleted_at=now on the document.
func (r *DocumentRepo) SoftDelete(ctx context.Context, id bson.ObjectID) error {
	now := time.Now().UTC()
	_, err := r.col.UpdateOne(ctx,
		bson.M{"_id": id},
		bson.M{"$set": bson.M{
			"status":         models.StatusDeleted,
			"deleted_at":     now,
			"last_updated_at": now,
		}},
	)
	if err != nil {
		return fmt.Errorf("document_repo: soft delete: %w", err)
	}
	return nil
}

// Archive sets status=archived on the document.
func (r *DocumentRepo) Archive(ctx context.Context, id bson.ObjectID) error {
	now := time.Now().UTC()
	_, err := r.col.UpdateOne(ctx,
		bson.M{"_id": id},
		bson.M{"$set": bson.M{
			"status":         models.StatusArchived,
			"last_updated_at": now,
		}},
	)
	if err != nil {
		return fmt.Errorf("document_repo: archive: %w", err)
	}
	return nil
}

// UpdateSnapshot replaces the snapshot field with the given ordered block ID slice.
func (r *DocumentRepo) UpdateSnapshot(ctx context.Context, id bson.ObjectID, snapshot []string) error {
	_, err := r.col.UpdateOne(ctx,
		bson.M{"_id": id},
		bson.M{"$set": bson.M{
			"snapshot":        snapshot,
			"last_updated_at": time.Now().UTC(),
		}},
	)
	if err != nil {
		return fmt.Errorf("document_repo: update snapshot: %w", err)
	}
	return nil
}
