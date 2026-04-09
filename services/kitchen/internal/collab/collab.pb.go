// Package collab provides the gRPC CollabService for real-time document editing.
//
// This file contains message types and gRPC service plumbing.
// It is hand-written to avoid a protoc dependency; the proto definition lives at
// proto/collab/collab.proto.
//
// Messages are serialised over the wire using JSON (via the registered JSON codec
// in codec.go) rather than protobuf binary, which is fine for this scale.
package collab

// EditRequest is sent by the client over the bidirectional stream.
type EditRequest struct {
	// DocumentID is the ObjectID hex of the document being edited. Required on every message.
	DocumentID string `json:"document_id"`
	// BlockID is the ObjectID hex of the block being edited. Empty on the initial sync message.
	BlockID string `json:"block_id,omitempty"`
	// CrdtOperation is the encoded CRDT delta (Automerge/Yjs binary) for BlockID.
	// Empty on the initial sync message.
	CrdtOperation []byte `json:"crdt_operation,omitempty"`
	// StateVector is the client's current CRDT state vector for the document.
	// Sent once on the initial message to request a delta sync from the server.
	StateVector []byte `json:"state_vector,omitempty"`
}

// EditResponse is broadcast by the server to all connected clients for a document.
type EditResponse struct {
	// BlockID is the ObjectID hex of the block that changed.
	BlockID string `json:"block_id"`
	// CrdtDelta is the encoded CRDT delta to apply to the client's local block replica.
	CrdtDelta []byte `json:"crdt_delta,omitempty"`
	// FromUserID is the ObjectID hex of the user who sent the originating edit.
	// Empty for server-initiated sync responses.
	FromUserID string `json:"from_user_id,omitempty"`
}
