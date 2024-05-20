package rlstorage

import (
	"strconv"
	"time"

	"github.com/go-redis/redis"
	"github.com/sirupsen/logrus"
)

// rlRedisStorage is a struct that implements the RLStorage interface
// and uses Redis as the underlying storage mechanism for rate limiting.
type rlRedisStorage struct {
	client *redis.Client  // Redis client instance
	ttl    time.Duration  // Time-to-live (TTL) for rate limiting keys
	logger *logrus.Logger // Logger instance for logging messages
}

// NewRedisStorage creates a new instance of rlRedisStorage with the provided
// Redis client, TTL duration, and logger instance.
func NewRedisStorage(client *redis.Client, ttl time.Duration, logger *logrus.Logger) RLStorage {
	return &rlRedisStorage{
		client: client,
		ttl:    ttl,
		logger: logger,
	}
}

// Decrease decrements the value associated with the given ID in Redis.
func (r *rlRedisStorage) Decrease(id string) {
	err := r.client.Decr(id).Err()
	if err != nil {
		r.logger.Warnf("Failed to Decrease value for ID '%s': %v", id, err)
	}
}

// Free sets the value associated with the given ID in Redis to 0.
func (r *rlRedisStorage) Free(id string) {
	err := r.client.Set(id, 0, 0).Err()
	if err != nil {
		r.logger.Warnf("Failed to Free value for ID '%s': %v", id, err)
	}
}

// Get retrieves the value associated with the given ID from Redis and
// returns it as a uint16.
func (r *rlRedisStorage) Get(id string) uint16 {
	val, err := r.client.Get(id).Result()
	if err != nil {
		r.logger.Warnf("Failed to Get value for ID '%s': %v", id, err)
		return 0
	}

	result, err := strconv.Atoi(val)
	if err != nil {
		r.logger.Warnf("Failed to convert value for ID '%s': %v", id, err)
		return 0
	}

	return uint16(result)
}

// Increase increments the value associated with the given ID in Redis
// and sets a TTL (Time-to-Live) for the key.
func (r *rlRedisStorage) Increase(id string) {
	err := r.client.Incr(id).Err()
	if err != nil {
		r.logger.Warnf("Failed to Increase value for ID '%s': %v", id, err)
		return
	}

	err = r.client.Expire(id, r.ttl).Err()
	if err != nil {
		r.logger.Warnf("Failed to SetTTL for ID '%s': %v", id, err)
	}
}

// FreeAll deletes all entries from Redis.
func (r *rlRedisStorage) FreeAll() {
	err := r.client.FlushAll().Err()
	if err != nil {
		r.logger.Warnf("Failed to flush Redis database: %v", err)
	} else {
		r.logger.Info("Flushed Redis database")
	}
}
