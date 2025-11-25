package websocket

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"taskboard-backend/internal/database"
)

const (
	// Redis channel prefix for board events
	channelPrefix = "taskboard:board:"
)

// RedisPubSub handles Redis pub/sub for WebSocket scaling.
type RedisPubSub struct {
	redis *database.RedisDB
}

// NewRedisPubSub creates a new RedisPubSub instance.
func NewRedisPubSub(redis *database.RedisDB) *RedisPubSub {
	return &RedisPubSub{redis: redis}
}

// Publish publishes an event to a board channel.
func (r *RedisPubSub) Publish(boardID uuid.UUID, message []byte) error {
	channel := fmt.Sprintf("%s%s", channelPrefix, boardID.String())
	return r.redis.Publish(context.Background(), channel, message)
}

// Subscribe subscribes to all board channels and forwards messages to the hub.
func (r *RedisPubSub) Subscribe(ctx context.Context, hub *Hub) {
	// Subscribe to pattern to catch all board channels
	pubsub := r.redis.Client.PSubscribe(ctx, channelPrefix+"*")
	defer pubsub.Close()

	ch := pubsub.Channel()

	log.Println("Redis pub/sub started, listening for board events...")

	for {
		select {
		case <-ctx.Done():
			log.Println("Redis pub/sub shutting down...")
			return

		case msg := <-ch:
			r.handleMessage(hub, msg)
		}
	}
}

// handleMessage processes incoming Redis messages.
func (r *RedisPubSub) handleMessage(hub *Hub, msg *redis.Message) {
	// Extract board ID from channel name
	boardIDStr := msg.Channel[len(channelPrefix):]
	boardID, err := uuid.Parse(boardIDStr)
	if err != nil {
		log.Printf("Invalid board ID in Redis channel: %s", boardIDStr)
		return
	}

	// Parse the event
	var event Event
	if err := json.Unmarshal([]byte(msg.Payload), &event); err != nil {
		log.Printf("Failed to parse Redis message: %v", err)
		return
	}

	// Broadcast to local clients
	hub.broadcast <- &BroadcastMessage{
		BoardID: boardID,
		Message: []byte(msg.Payload),
	}
}

// PublishNotification publishes a notification to a user's channel.
func (r *RedisPubSub) PublishNotification(userID uuid.UUID, notification interface{}) error {
	channel := fmt.Sprintf("taskboard:user:%s:notifications", userID.String())
	
	data, err := json.Marshal(notification)
	if err != nil {
		return err
	}
	
	return r.redis.Publish(context.Background(), channel, data)
}
