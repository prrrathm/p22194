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

// Sentinel errors for sharing operations.
var (
	ErrShareNotFound = errors.New("share not found")
	ErrShareExists   = errors.New("user already has access to this document")
)

// SharingService implements document sharing and permission management.
type SharingService struct {
	shares *repository.SharingRepo
}

// NewSharingService constructs a SharingService.
func NewSharingService(shares *repository.SharingRepo) *SharingService {
	return &SharingService{shares: shares}
}

// Share grants a user access to a document with the given role.
// Returns ErrShareExists if the user already has access.
func (s *SharingService) Share(ctx context.Context, documentID, inviterID bson.ObjectID, req models.ShareDocumentRequest) (*models.DocumentShare, error) {
	userID, err := bson.ObjectIDFromHex(req.UserID)
	if err != nil {
		return nil, fmt.Errorf("sharing_service: invalid user_id: %w", err)
	}

	share := &models.DocumentShare{
		DocumentID:      documentID,
		UserID:          userID,
		Role:            req.Role,
		InvitedByUserID: inviterID,
		CreatedAt:       time.Now().UTC(),
	}

	if err := s.shares.Create(ctx, share); err != nil {
		var we mongo.WriteException
		if errors.As(err, &we) {
			for _, e := range we.WriteErrors {
				if e.Code == 11000 {
					return nil, ErrShareExists
				}
			}
		}
		return nil, fmt.Errorf("sharing_service: share: %w", err)
	}

	return share, nil
}

// RemoveAccess removes a user's access to a document.
func (s *SharingService) RemoveAccess(ctx context.Context, documentID, userID bson.ObjectID) error {
	if err := s.shares.Delete(ctx, documentID, userID); err != nil {
		return fmt.Errorf("sharing_service: remove access: %w", err)
	}
	return nil
}

// GetShares returns all users with access to a document.
func (s *SharingService) GetShares(ctx context.Context, documentID bson.ObjectID) ([]models.DocumentShare, error) {
	shares, err := s.shares.ListByDocument(ctx, documentID)
	if err != nil {
		return nil, fmt.Errorf("sharing_service: get shares: %w", err)
	}
	return shares, nil
}
