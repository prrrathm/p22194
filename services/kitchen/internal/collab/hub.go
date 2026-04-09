package collab

import (
	"sync"

	"github.com/google/uuid"
)

// Hub manages the set of active gRPC stream sessions per document.
// It is safe for concurrent use.
type Hub struct {
	mu      sync.RWMutex
	clients map[string]map[string]chan *EditResponse // documentID → clientID → send channel
}

// NewHub creates an empty Hub.
func NewHub() *Hub {
	return &Hub{
		clients: make(map[string]map[string]chan *EditResponse),
	}
}

// Subscribe registers a new client for the given document and returns a
// channel that receives EditResponse messages broadcast by other clients,
// along with a clientID that must be passed to Unsubscribe on disconnect.
func (h *Hub) Subscribe(documentID string) (clientID string, ch <-chan *EditResponse) {
	id := uuid.New().String()
	c := make(chan *EditResponse, 64)

	h.mu.Lock()
	if h.clients[documentID] == nil {
		h.clients[documentID] = make(map[string]chan *EditResponse)
	}
	h.clients[documentID][id] = c
	h.mu.Unlock()

	return id, c
}

// Unsubscribe removes the client from the document's subscriber set and closes its channel.
func (h *Hub) Unsubscribe(documentID, clientID string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if m, ok := h.clients[documentID]; ok {
		if ch, ok := m[clientID]; ok {
			close(ch)
			delete(m, clientID)
		}
		if len(m) == 0 {
			delete(h.clients, documentID)
		}
	}
}

// Broadcast sends resp to all subscribers of documentID except the sender identified by senderClientID.
// Sends are non-blocking: slow clients whose buffer is full are skipped.
func (h *Hub) Broadcast(documentID, senderClientID string, resp *EditResponse) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for id, ch := range h.clients[documentID] {
		if id == senderClientID {
			continue
		}
		select {
		case ch <- resp:
		default:
			// Client buffer full — skip to avoid blocking the broadcaster.
		}
	}
}
