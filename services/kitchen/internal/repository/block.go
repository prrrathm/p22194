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

// BlockRepo performs CRUD operations on the `blocks` collection.
type BlockRepo struct {
	col *mongo.Collection
}

// NewBlockRepo creates a BlockRepo backed by the given database.
func NewBlockRepo(db *mongo.Database) *BlockRepo {
	return &BlockRepo{col: db.Collection("blocks")}
}

// EnsureIndexes creates the required indexes on startup. Idempotent.
func (r *BlockRepo) EnsureIndexes(ctx context.Context) error {
	indexes := []mongo.IndexModel{
		{Keys: bson.D{{Key: "document_id", Value: 1}}},
		{
			Keys:    bson.D{{Key: "insert_after_block_id", Value: 1}},
			Options: options.Index().SetSparse(true),
		},
	}
	if _, err := r.col.Indexes().CreateMany(ctx, indexes); err != nil {
		return fmt.Errorf("block_repo: ensure indexes: %w", err)
	}
	return nil
}

// Create inserts a new block and sets b.ID on success.
func (r *BlockRepo) Create(ctx context.Context, b *models.Block) error {
	result, err := r.col.InsertOne(ctx, b)
	if err != nil {
		return fmt.Errorf("block_repo: insert: %w", err)
	}
	b.ID = result.InsertedID.(bson.ObjectID)
	return nil
}

// FindByID returns the block with the given ID.
// Returns mongo.ErrNoDocuments when not found.
func (r *BlockRepo) FindByID(ctx context.Context, id bson.ObjectID) (*models.Block, error) {
	var b models.Block
	err := r.col.FindOne(ctx, bson.M{"_id": id}).Decode(&b)
	if err != nil {
		return nil, fmt.Errorf("block_repo: find by id: %w", err)
	}
	return &b, nil
}

// ListByDocument returns all non-deleted blocks for the given document.
func (r *BlockRepo) ListByDocument(ctx context.Context, documentID bson.ObjectID) ([]models.Block, error) {
	filter := bson.M{
		"document_id": documentID,
		"deleted_at":  bson.M{"$exists": false},
	}
	cur, err := r.col.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("block_repo: list by document: %w", err)
	}
	defer cur.Close(ctx)

	var blocks []models.Block
	if err := cur.All(ctx, &blocks); err != nil {
		return nil, fmt.Errorf("block_repo: decode blocks: %w", err)
	}
	return blocks, nil
}

// UpdateContent sets content_state and optionally content_meta on a block.
func (r *BlockRepo) UpdateContent(ctx context.Context, id bson.ObjectID, contentState []byte, contentMeta map[string]interface{}) error {
	set := bson.M{"content_state": contentState}
	if contentMeta != nil {
		set["content_meta"] = contentMeta
	}
	_, err := r.col.UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$set": set})
	if err != nil {
		return fmt.Errorf("block_repo: update content: %w", err)
	}
	return nil
}

// UpdateInsertAfter sets insert_after_block_id on a block (for reordering).
// Pass nil to move the block to the top of the document.
func (r *BlockRepo) UpdateInsertAfter(ctx context.Context, id bson.ObjectID, insertAfter *bson.ObjectID) error {
	var update bson.M
	if insertAfter == nil {
		update = bson.M{"$unset": bson.M{"insert_after_block_id": ""}}
	} else {
		update = bson.M{"$set": bson.M{"insert_after_block_id": insertAfter}}
	}
	_, err := r.col.UpdateOne(ctx, bson.M{"_id": id}, update)
	if err != nil {
		return fmt.Errorf("block_repo: update insert after: %w", err)
	}
	return nil
}

// SoftDelete sets deleted_at=now on a block.
func (r *BlockRepo) SoftDelete(ctx context.Context, id bson.ObjectID) error {
	now := time.Now().UTC()
	_, err := r.col.UpdateOne(ctx,
		bson.M{"_id": id},
		bson.M{"$set": bson.M{"deleted_at": now}},
	)
	if err != nil {
		return fmt.Errorf("block_repo: soft delete: %w", err)
	}
	return nil
}

// FindByInsertAfter returns the active block whose insert_after_block_id equals predecessor
// within the given document. Pass nil predecessor to find the head block (no insert_after set).
// Returns nil, nil when no such block exists.
func (r *BlockRepo) FindByInsertAfter(ctx context.Context, documentID bson.ObjectID, predecessor *bson.ObjectID) (*models.Block, error) {
	filter := bson.M{
		"document_id": documentID,
		"deleted_at":  bson.M{"$exists": false},
	}
	if predecessor == nil {
		filter["insert_after_block_id"] = bson.M{"$exists": false}
	} else {
		filter["insert_after_block_id"] = *predecessor
	}
	var b models.Block
	err := r.col.FindOne(ctx, filter).Decode(&b)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, fmt.Errorf("block_repo: find by insert after: %w", err)
	}
	return &b, nil
}

// FindByIDs returns blocks matching the given IDs, preserving no particular order.
func (r *BlockRepo) FindByIDs(ctx context.Context, ids []bson.ObjectID) ([]models.Block, error) {
	cur, err := r.col.Find(ctx, bson.M{"_id": bson.M{"$in": ids}})
	if err != nil {
		return nil, fmt.Errorf("block_repo: find by ids: %w", err)
	}
	defer cur.Close(ctx)

	var blocks []models.Block
	if err := cur.All(ctx, &blocks); err != nil {
		return nil, fmt.Errorf("block_repo: decode blocks: %w", err)
	}
	return blocks, nil
}
