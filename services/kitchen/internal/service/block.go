package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"

	"p22194.prrrathm.com/kitchen/internal/models"
	"p22194.prrrathm.com/kitchen/internal/repository"
)

// Sentinel errors for block operations.
var (
	ErrBlockNotFound = errors.New("block not found")
)

// BlockService implements block management business logic.
type BlockService struct {
	blocks *repository.BlockRepo
	docSvc *DocumentService
}

// NewBlockService constructs a BlockService.
func NewBlockService(blocks *repository.BlockRepo, docSvc *DocumentService) *BlockService {
	return &BlockService{blocks: blocks, docSvc: docSvc}
}

// Insert creates a new block in the given document.
// If req.InsertAfterBlockID is nil the block is placed at the top of the document.
// After inserting, the document snapshot is rebuilt.
func (s *BlockService) Insert(ctx context.Context, documentID bson.ObjectID, req models.InsertBlockRequest) (*models.Block, error) {
	now := time.Now().UTC()
	block := &models.Block{
		DocumentID:   documentID,
		BlockType:    req.BlockType,
		ContentState: req.ContentState,
		ContentMeta:  req.ContentMeta,
	}

	if req.InsertAfterBlockID != nil {
		id, err := bson.ObjectIDFromHex(*req.InsertAfterBlockID)
		if err != nil {
			return nil, fmt.Errorf("block_service: invalid insert_after_block_id: %w", err)
		}
		block.InsertAfterBlockID = &id
	}

	_ = now // timestamp is handled by MongoDB ObjectID ordering

	if err := s.blocks.Create(ctx, block); err != nil {
		return nil, fmt.Errorf("block_service: insert: %w", err)
	}

	if err := s.docSvc.RebuildSnapshot(ctx, documentID); err != nil {
		// Non-fatal: snapshot will be stale but block was persisted.
		_ = err
	}

	return block, nil
}

// Update replaces the content_state (and optionally content_meta) of a block.
func (s *BlockService) Update(ctx context.Context, blockID bson.ObjectID, req models.UpdateBlockRequest) (*models.Block, error) {
	block, err := s.getActive(ctx, blockID)
	if err != nil {
		return nil, err
	}

	if err := s.blocks.UpdateContent(ctx, blockID, req.ContentState, req.ContentMeta); err != nil {
		return nil, fmt.Errorf("block_service: update: %w", err)
	}

	block.ContentState = req.ContentState
	if req.ContentMeta != nil {
		block.ContentMeta = req.ContentMeta
	}
	return block, nil
}

// Reorder moves a block within its document by performing a proper linked-list
// insertion: it closes the gap at the old position and relinks the displaced
// block at the new position before rebuilding the document snapshot.
func (s *BlockService) Reorder(ctx context.Context, blockID bson.ObjectID, req models.ReorderBlockRequest) (*models.Block, error) {
	block, err := s.getActive(ctx, blockID)
	if err != nil {
		return nil, err
	}

	var insertAfter *bson.ObjectID
	if req.InsertAfterBlockID != nil {
		id, err := bson.ObjectIDFromHex(*req.InsertAfterBlockID)
		if err != nil {
			return nil, fmt.Errorf("block_service: invalid insert_after_block_id: %w", err)
		}
		insertAfter = &id
	}

	// No-op: block is already at the requested position.
	alreadyInPlace := (insertAfter == nil && block.InsertAfterBlockID == nil) ||
		(insertAfter != nil && block.InsertAfterBlockID != nil && *insertAfter == *block.InsertAfterBlockID)
	if alreadyInPlace {
		return block, nil
	}

	// Block that currently follows the moved block (will inherit its old predecessor).
	nextOfMoved, err := s.blocks.FindByInsertAfter(ctx, block.DocumentID, &blockID)
	if err != nil {
		return nil, fmt.Errorf("block_service: reorder find next of moved: %w", err)
	}

	// Block that currently follows the target position (will follow the moved block after the move).
	nextOfTarget, err := s.blocks.FindByInsertAfter(ctx, block.DocumentID, insertAfter)
	if err != nil {
		return nil, fmt.Errorf("block_service: reorder find next of target: %w", err)
	}

	// Close the gap: the block that was after M now points to M's old predecessor.
	if nextOfMoved != nil && nextOfMoved.ID != blockID {
		if err := s.blocks.UpdateInsertAfter(ctx, nextOfMoved.ID, block.InsertAfterBlockID); err != nil {
			return nil, fmt.Errorf("block_service: reorder close gap: %w", err)
		}
	}

	// Move M to its new position.
	if err := s.blocks.UpdateInsertAfter(ctx, blockID, insertAfter); err != nil {
		return nil, fmt.Errorf("block_service: reorder: %w", err)
	}

	// The block that was after the target now follows M.
	if nextOfTarget != nil && nextOfTarget.ID != blockID {
		if err := s.blocks.UpdateInsertAfter(ctx, nextOfTarget.ID, &blockID); err != nil {
			return nil, fmt.Errorf("block_service: reorder link next: %w", err)
		}
	}

	block.InsertAfterBlockID = insertAfter

	if err := s.docSvc.RebuildSnapshot(ctx, block.DocumentID); err != nil {
		_ = err
	}

	return block, nil
}

// Delete soft-deletes a block and rebuilds the document snapshot.
func (s *BlockService) Delete(ctx context.Context, blockID bson.ObjectID) error {
	block, err := s.getActive(ctx, blockID)
	if err != nil {
		return err
	}

	now := time.Now().UTC()
	if err := s.blocks.SoftDelete(ctx, blockID); err != nil {
		return fmt.Errorf("block_service: delete: %w", err)
	}

	_ = now

	if err := s.docSvc.RebuildSnapshot(ctx, block.DocumentID); err != nil {
		_ = err
	}

	return nil
}

// ListByDocument returns all active blocks for a document, ordered by the document snapshot.
func (s *BlockService) ListByDocument(ctx context.Context, documentID bson.ObjectID) ([]models.Block, error) {
	blocks, err := s.blocks.ListByDocument(ctx, documentID)
	if err != nil {
		return nil, fmt.Errorf("block_service: list by document: %w", err)
	}
	return blocks, nil
}

// getActive fetches a block and returns ErrBlockNotFound if it is missing or soft-deleted.
func (s *BlockService) getActive(ctx context.Context, id bson.ObjectID) (*models.Block, error) {
	b, err := s.blocks.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrBlockNotFound
		}
		return nil, fmt.Errorf("block_service: find: %w", err)
	}
	if b.DeletedAt != nil {
		return nil, ErrBlockNotFound
	}
	return b, nil
}
