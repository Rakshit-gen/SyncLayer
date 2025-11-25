package database

import (
	"context"
	"crypto/tls"
	"fmt"
	"time"

	"taskboard-backend/internal/config"

	"github.com/redis/go-redis/v9"
)

// RedisDB wraps the Redis client.
type RedisDB struct {
	Client *redis.Client
}

// NewRedisDB creates a new Redis client connection.
func NewRedisDB(cfg config.RedisConfig) (*RedisDB, error) {
	options := &redis.Options{
		Addr:         fmt.Sprintf("%s:%s", cfg.Host, cfg.Port),
		Password:     cfg.Password,
		DB:           cfg.DB,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		PoolSize:     10,
		MinIdleConns: 5,
	}

	// Enable TLS if configured
	if cfg.TLS {
		options.TLSConfig = &tls.Config{
			ServerName: cfg.Host,
		}
	}

	client := redis.NewClient(options)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Verify connection
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &RedisDB{Client: client}, nil
}

// Close closes the Redis client connection.
func (r *RedisDB) Close() error {
	if r.Client != nil {
		return r.Client.Close()
	}
	return nil
}

// HealthCheck verifies the Redis connection is healthy.
func (r *RedisDB) HealthCheck(ctx context.Context) error {
	return r.Client.Ping(ctx).Err()
}

// Publish sends a message to a Redis channel.
func (r *RedisDB) Publish(ctx context.Context, channel string, message interface{}) error {
	return r.Client.Publish(ctx, channel, message).Err()
}

// Subscribe creates a subscription to one or more Redis channels.
func (r *RedisDB) Subscribe(ctx context.Context, channels ...string) *redis.PubSub {
	return r.Client.Subscribe(ctx, channels...)
}
