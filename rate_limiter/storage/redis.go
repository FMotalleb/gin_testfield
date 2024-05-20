package rlstorage

import (
	"log"
	"strconv"
	"time"

	"github.com/go-redis/redis"
)

// rlRedisStorage is a struct that implements the RLStorage interface
// and uses Redis as the underlying storage mechanism for rate limiting.
type rlRedisStorage struct {
	client *redis.Client // Redis client instance
	ttl    time.Duration // Time-to-live (TTL) for rate limiting keys
}

// NewRedisStorage creates a new instance of rlRedisStorage with the provided
// Redis client and TTL duration.
func NewRedisStorage(client *redis.Client, ttl time.Duration) RLStorage {
	return &rlRedisStorage{
		client,
		ttl,
	}
}

// Decrease decrements the value associated with the given ID in Redis.
// It implements the ratelimiter.Storage interface.
func (r *rlRedisStorage) Decrease(id string) {
	err := r.client.Decr(id).Err()
	if err != nil {
		log.Fatalf("Failed to Decrease value: %v", err)
	}
}

// Free sets the value associated with the given ID in Redis to 0.
// It implements the ratelimiter.Storage interface.
func (r *rlRedisStorage) Free(id string) {
	err := r.client.Set(id, 0, 0).Err()
	if err != nil {
		log.Fatalf("Failed to Free value: %v", err)
	}
}

// Get retrieves the value associated with the given ID from Redis and
// returns it as a uint16.
// It implements the ratelimiter.Storage interface.
func (r *rlRedisStorage) Get(id string) uint16 {
	val := r.client.Get(id).Val()
	if val == "" {
		return 0
	}
	result, err := strconv.Atoi(val)
	if err != nil {
		log.Fatalf("Failed to get value: %v", err)
	}
	return uint16(result)
}

// Increase increments the value associated with the given ID in Redis
// and sets a TTL (Time-to-Live) for the key.
// It implements the ratelimiter.Storage interface.
func (r *rlRedisStorage) Increase(id string) {
	if err := r.client.Incr(id).Err(); err != nil {
		log.Fatalf("Failed to Increase value: %v", err)
	}

	if err := r.client.Expire(id, r.ttl).Err(); err != nil {
		log.Printf("Failed to SetTTL for value: %v\n", err)
	}
}
