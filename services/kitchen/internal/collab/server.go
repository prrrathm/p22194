package collab

import (
	"errors"
	"io"

	"github.com/rs/zerolog"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"p22194.prrrathm.com/kitchen/internal/repository"
)

// Server implements CollabServiceServer.
// It handles bidirectional streams for real-time document editing.
type Server struct {
	UnimplementedCollabServiceServer
	hub    *Hub
	blocks *repository.BlockRepo
	log    zerolog.Logger
}

// NewServer constructs a collab Server.
func NewServer(hub *Hub, blocks *repository.BlockRepo, log zerolog.Logger) *Server {
	return &Server{hub: hub, blocks: blocks, log: log}
}

// EditSession handles the bidirectional stream RPC for real-time document editing.
//
// Protocol:
//  1. Client sends an initial EditRequest with document_id and state_vector.
//     The server replies with the current content_state for every block in the
//     document as individual EditResponse messages (one per block).
//  2. Client sends subsequent EditRequest messages with block_id and crdt_operation.
//     The server persists the new content_state and broadcasts to other connected clients.
//  3. Stream closes when the client disconnects or sends EOF.
func (s *Server) EditSession(stream CollabService_EditSessionServer) error {
	ctx := stream.Context()

	// Read the first message to get document_id and state_vector.
	first, err := stream.Recv()
	if err != nil {
		return status.Errorf(codes.InvalidArgument, "expected initial sync message: %v", err)
	}

	if first.DocumentID == "" {
		return status.Error(codes.InvalidArgument, "document_id is required")
	}

	documentID := first.DocumentID

	// ── Initial sync ─────────────────────────────────────────────────────────
	// If the client supplied a state_vector, send back the current content_state
	// for all blocks in the document. The client merges them into its local replica.
	if first.StateVector != nil || first.BlockID == "" {
		docOID, err := bson.ObjectIDFromHex(documentID)
		if err != nil {
			return status.Errorf(codes.InvalidArgument, "invalid document_id: %v", err)
		}

		blocks, err := s.blocks.ListByDocument(ctx, docOID)
		if err != nil {
			s.log.Error().Err(err).Str("document_id", documentID).Msg("collab: list blocks for sync")
			return status.Error(codes.Internal, "failed to load document blocks")
		}

		for _, b := range blocks {
			if len(b.ContentState) == 0 {
				continue
			}
			if err := stream.Send(&EditResponse{
				BlockID:   b.ID.Hex(),
				CrdtDelta: b.ContentState,
			}); err != nil {
				return err
			}
		}
	}

	// ── Subscribe to hub ──────────────────────────────────────────────────────
	clientID, incoming := s.hub.Subscribe(documentID)
	defer s.hub.Unsubscribe(documentID, clientID)

	// Fan-out goroutine: forward hub messages to this client's stream.
	sendErr := make(chan error, 1)
	go func() {
		for resp := range incoming {
			if err := stream.Send(resp); err != nil {
				sendErr <- err
				return
			}
		}
	}()

	// ── Receive loop ──────────────────────────────────────────────────────────
	for {
		select {
		case err := <-sendErr:
			return err
		case <-ctx.Done():
			return nil
		default:
		}

		req, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return err
		}

		if req.BlockID == "" || len(req.CrdtOperation) == 0 {
			continue
		}

		blockOID, err := bson.ObjectIDFromHex(req.BlockID)
		if err != nil {
			s.log.Warn().Str("block_id", req.BlockID).Msg("collab: invalid block_id in edit request")
			continue
		}

		// Persist updated content_state.
		if err := s.blocks.UpdateContent(ctx, blockOID, req.CrdtOperation, nil); err != nil {
			if !errors.Is(err, mongo.ErrNoDocuments) {
				s.log.Error().Err(err).Str("block_id", req.BlockID).Msg("collab: persist content_state")
			}
			continue
		}

		// Broadcast delta to all other connected clients.
		s.hub.Broadcast(documentID, clientID, &EditResponse{
			BlockID:   req.BlockID,
			CrdtDelta: req.CrdtOperation,
		})
	}
}
