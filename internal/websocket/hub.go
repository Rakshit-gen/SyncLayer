// Package websocket handles real-time WebSocket connections and events.
package websocket

import (
	"context"
	"encoding/json"
	"log"
	"sync"

	"github.com/google/uuid"
)

// Hub maintains the set of active clients and broadcasts messages.
type Hub struct {
	// Registered clients, organized by board ID
	boards map[uuid.UUID]map[*Client]bool

	// Channel for registering clients
	register chan *Client

	// Channel for unregistering clients
	unregister chan *Client

	// Channel for broadcasting messages to a board
	broadcast chan *BroadcastMessage

	// Mutex for thread-safe operations
	mu sync.RWMutex

	// Redis pub/sub integration
	redisPubSub *RedisPubSub
}

// BroadcastMessage represents a message to broadcast to a board.
type BroadcastMessage struct {
	BoardID uuid.UUID
	Message []byte
	Exclude *Client // Optional: exclude this client from broadcast
}

// NewHub creates a new Hub instance.
func NewHub(redisPubSub *RedisPubSub) *Hub {
	return &Hub{
		boards:      make(map[uuid.UUID]map[*Client]bool),
		register:    make(chan *Client),
		unregister:  make(chan *Client),
		broadcast:   make(chan *BroadcastMessage, 256),
		redisPubSub: redisPubSub,
	}
}

// Run starts the hub's main loop.
func (h *Hub) Run(ctx context.Context) {
	// Start Redis subscription handler if available
	if h.redisPubSub != nil {
		go h.redisPubSub.Subscribe(ctx, h)
	}

	for {
		select {
		case <-ctx.Done():
			return

		case client := <-h.register:
			h.registerClient(client)

		case client := <-h.unregister:
			h.unregisterClient(client)

		case message := <-h.broadcast:
			h.broadcastToBoard(message)
		}
	}
}

// registerClient adds a client to a board room.
func (h *Hub) registerClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if _, ok := h.boards[client.BoardID]; !ok {
		h.boards[client.BoardID] = make(map[*Client]bool)
	}
	h.boards[client.BoardID][client] = true

	log.Printf("Client %s joined board %s (total: %d)", client.UserID, client.BoardID, len(h.boards[client.BoardID]))

	// Broadcast presence event
	h.broadcastPresenceJoined(client)
}

// unregisterClient removes a client from a board room.
func (h *Hub) unregisterClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if clients, ok := h.boards[client.BoardID]; ok {
		if _, ok := clients[client]; ok {
			delete(clients, client)
			close(client.send)

			log.Printf("Client %s left board %s (remaining: %d)", client.UserID, client.BoardID, len(clients))

			// Broadcast presence event
			h.broadcastPresenceLeft(client)

			// Clean up empty board room
			if len(clients) == 0 {
				delete(h.boards, client.BoardID)
			}
		}
	}
}

// broadcastToBoard sends a message to all clients in a board.
func (h *Hub) broadcastToBoard(message *BroadcastMessage) {
	h.mu.RLock()
	clients, ok := h.boards[message.BoardID]
	h.mu.RUnlock()

	if !ok {
		return
	}

	for client := range clients {
		// Skip excluded client if specified
		if message.Exclude != nil && client == message.Exclude {
			continue
		}

		select {
		case client.send <- message.Message:
		default:
			// Client send buffer is full, close connection
			h.mu.Lock()
			delete(h.boards[message.BoardID], client)
			close(client.send)
			h.mu.Unlock()
		}
	}
}

// broadcastPresenceJoined notifies other clients that a user joined.
func (h *Hub) broadcastPresenceJoined(client *Client) {
	event := Event{
		Type: EventPresenceJoined,
		Payload: map[string]interface{}{
			"user_id": client.UserID,
		},
	}

	data, err := json.Marshal(event)
	if err != nil {
		return
	}

	// Broadcast to other clients in the room
	if clients, ok := h.boards[client.BoardID]; ok {
		for c := range clients {
			if c != client {
				select {
				case c.send <- data:
				default:
				}
			}
		}
	}
}

// broadcastPresenceLeft notifies other clients that a user left.
func (h *Hub) broadcastPresenceLeft(client *Client) {
	event := Event{
		Type: EventPresenceLeft,
		Payload: map[string]interface{}{
			"user_id": client.UserID,
		},
	}

	data, err := json.Marshal(event)
	if err != nil {
		return
	}

	// Broadcast to remaining clients in the room
	if clients, ok := h.boards[client.BoardID]; ok {
		for c := range clients {
			select {
			case c.send <- data:
			default:
			}
		}
	}
}

// BroadcastEvent broadcasts an event to all clients in a board.
func (h *Hub) BroadcastEvent(boardID uuid.UUID, event *Event, exclude *Client) {
	data, err := json.Marshal(event)
	if err != nil {
		log.Printf("Failed to marshal event: %v", err)
		return
	}

	h.broadcast <- &BroadcastMessage{
		BoardID: boardID,
		Message: data,
		Exclude: exclude,
	}

	// Also publish to Redis for multi-instance support
	if h.redisPubSub != nil {
		h.redisPubSub.Publish(boardID, data)
	}
}

// GetBoardClients returns the list of user IDs connected to a board.
func (h *Hub) GetBoardClients(boardID uuid.UUID) []uuid.UUID {
	h.mu.RLock()
	defer h.mu.RUnlock()

	var userIDs []uuid.UUID
	if clients, ok := h.boards[boardID]; ok {
		for client := range clients {
			userIDs = append(userIDs, client.UserID)
		}
	}
	return userIDs
}

// GetClientCount returns the number of clients connected to a board.
func (h *Hub) GetClientCount(boardID uuid.UUID) int {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if clients, ok := h.boards[boardID]; ok {
		return len(clients)
	}
	return 0
}

// GetRegister returns the register channel for external registration.
func (h *Hub) GetRegister() chan *Client {
	return h.register
}
