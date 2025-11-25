package websocket

import (
	"encoding/json"
	"log"
	"time"

	"github.com/gofiber/websocket/v2"
	"github.com/google/uuid"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 65536
)

// Client represents a WebSocket client connection.
type Client struct {
	hub     *Hub
	conn    *websocket.Conn
	send    chan []byte
	BoardID uuid.UUID
	UserID  uuid.UUID
	handler *EventHandler
}

// NewClient creates a new Client instance.
func NewClient(hub *Hub, conn *websocket.Conn, boardID, userID uuid.UUID, handler *EventHandler) *Client {
	return &Client{
		hub:     hub,
		conn:    conn,
		send:    make(chan []byte, 256),
		BoardID: boardID,
		UserID:  userID,
		handler: handler,
	}
}

// ReadPump pumps messages from the WebSocket connection to the hub.
func (c *Client) ReadPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		// Parse the incoming event
		var event Event
		if err := json.Unmarshal(message, &event); err != nil {
			log.Printf("Failed to parse WebSocket message: %v", err)
			continue
		}

		// Handle the event
		c.handler.HandleEvent(c, &event)
	}
}

// WritePump pumps messages from the hub to the WebSocket connection.
func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// Hub closed the channel
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// SendEvent sends an event to this specific client.
func (c *Client) SendEvent(event *Event) error {
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}

	select {
	case c.send <- data:
		return nil
	default:
		return ErrClientBufferFull
	}
}

// SendError sends an error event to the client.
func (c *Client) SendError(code, message string) {
	event := &Event{
		Type: EventError,
		Payload: map[string]interface{}{
			"code":    code,
			"message": message,
		},
	}
	c.SendEvent(event)
}

// Custom errors
var (
	ErrClientBufferFull = &ClientError{Message: "client send buffer is full"}
)

// ClientError represents a client-related error.
type ClientError struct {
	Message string
}

func (e *ClientError) Error() string {
	return e.Message
}
