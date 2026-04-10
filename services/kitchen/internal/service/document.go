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

// Sentinel errors mapped to HTTP status codes by the handler layer.
var (
	ErrDocumentNotFound = errors.New("document not found")
	ErrForbidden        = errors.New("forbidden")
)

// DocumentService implements document management business logic.
type DocumentService struct {
	docs   *repository.DocumentRepo
	blocks *repository.BlockRepo
}

// NewDocumentService constructs a DocumentService.
func NewDocumentService(docs *repository.DocumentRepo, blocks *repository.BlockRepo) *DocumentService {
	return &DocumentService{docs: docs, blocks: blocks}
}

// Create creates a new document owned by userID.
func (s *DocumentService) Create(ctx context.Context, userID bson.ObjectID, req models.CreateDocumentRequest) (*models.Document, error) {
	now := time.Now().UTC()
	doc := &models.Document{
		CreatedByUserID: userID,
		CreatedAt:       now,
		LastUpdatedAt:   now,
		Title:           req.Title,
		TitleIcon:       req.TitleIcon,
		Status:          models.StatusActive,
		Snapshot:        []string{},
	}

	if req.ParentDocumentID != nil {
		parentID, err := bson.ObjectIDFromHex(*req.ParentDocumentID)
		if err != nil {
			return nil, fmt.Errorf("document_service: invalid parent_document_id: %w", err)
		}
		doc.ParentDocumentID = &parentID
	}

	if err := s.docs.Create(ctx, doc); err != nil {
		return nil, fmt.Errorf("document_service: create: %w", err)
	}
	return doc, nil
}

// GetByID returns a document by ID. Returns ErrDocumentNotFound when missing.
func (s *DocumentService) GetByID(ctx context.Context, id bson.ObjectID) (*models.Document, error) {
	doc, err := s.docs.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrDocumentNotFound
		}
		return nil, fmt.Errorf("document_service: get by id: %w", err)
	}
	return doc, nil
}

// List returns a paginated page of documents owned by userID.
// status filters by document status; pass "" to return all non-deleted.
// parentID filters by parent; pass "" for root-level documents (no filter applied when "").
// cursor is the last seen document ID (hex); pass "" for the first page.
// limit is capped to 100.
func (s *DocumentService) List(
	ctx context.Context,
	userID bson.ObjectID,
	status string,
	parentIDHex string,
	cursorHex string,
	limit int,
) ([]models.Document, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	var parentID *bson.ObjectID
	if parentIDHex != "" {
		pid, err := bson.ObjectIDFromHex(parentIDHex)
		if err != nil {
			return nil, fmt.Errorf("document_service: invalid parent_id: %w", err)
		}
		parentID = &pid
	}

	var cursor bson.ObjectID
	if cursorHex != "" {
		c, err := bson.ObjectIDFromHex(cursorHex)
		if err != nil {
			return nil, fmt.Errorf("document_service: invalid cursor: %w", err)
		}
		cursor = c
	}

	return s.docs.ListByUser(ctx, userID, status, parentID, cursor, limit)
}

// ListChildren returns direct child documents of parentDocumentID.
func (s *DocumentService) ListChildren(ctx context.Context, parentID bson.ObjectID) ([]models.Document, error) {
	docs, err := s.docs.ListChildren(ctx, parentID)
	if err != nil {
		return nil, fmt.Errorf("document_service: list children: %w", err)
	}
	return docs, nil
}

// Delete soft-deletes a document. Returns ErrDocumentNotFound when the document does not exist.
func (s *DocumentService) Delete(ctx context.Context, id bson.ObjectID) error {
	if _, err := s.GetByID(ctx, id); err != nil {
		return err
	}
	return s.docs.SoftDelete(ctx, id)
}

// Archive sets a document's status to archived.
func (s *DocumentService) Archive(ctx context.Context, id bson.ObjectID) error {
	if _, err := s.GetByID(ctx, id); err != nil {
		return err
	}
	return s.docs.Archive(ctx, id)
}

// Update applies title/icon changes to an existing document.
func (s *DocumentService) Update(ctx context.Context, id bson.ObjectID, req models.UpdateDocumentRequest) (*models.Document, error) {
	doc, err := s.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if err := s.docs.UpdateMetadata(ctx, id, req.Title, req.TitleIcon); err != nil {
		return nil, fmt.Errorf("document_service: update: %w", err)
	}
	if req.Title != "" {
		doc.Title = req.Title
	}
	if req.TitleIcon != "" {
		doc.TitleIcon = req.TitleIcon
	}
	return doc, nil
}

// RebuildSnapshot resolves the block linked list for documentID and persists the
// ordered block ID slice to the document's snapshot field.
func (s *DocumentService) RebuildSnapshot(ctx context.Context, documentID bson.ObjectID) error {
	blocks, err := s.blocks.ListByDocument(ctx, documentID)
	if err != nil {
		return fmt.Errorf("document_service: rebuild snapshot fetch blocks: %w", err)
	}

	// Build a map from insertAfterBlockID → block for linked-list traversal.
	// Key is the hex of the predecessor block ID, or "" for the head block.
	next := make(map[string]models.Block, len(blocks))
	for _, b := range blocks {
		key := ""
		if b.InsertAfterBlockID != nil {
			key = b.InsertAfterBlockID.Hex()
		}
		next[key] = b
	}

	snapshot := make([]string, 0, len(blocks))
	cur := ""
	for {
		b, ok := next[cur]
		if !ok {
			break
		}
		snapshot = append(snapshot, b.ID.Hex())
		cur = b.ID.Hex()
	}

	return s.docs.UpdateSnapshot(ctx, documentID, snapshot)
}
