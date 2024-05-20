package rlstorage

import (
	"log"
	"strconv"
	"time"

	"github.com/go-redis/redis"
)

type rlRedisStorage struct {
	client *redis.Client
	ttl    time.Duration
}

func NewRedisStorage(client *redis.Client, ttl time.Duration) RLStorage {
	return &rlRedisStorage{
		client,
		ttl,
	}
}

// Decrease implements ratelimiter.Storage.
func (r *rlRedisStorage) Decrease(id string) {
	// err := r.client.Decr(id).Err()
	// if err != nil {
	// 	log.Fatalf("Failed to Decrease value: %v", err)
	// }
}

// Free implements ratelimiter.Storage.
func (r *rlRedisStorage) Free(id string) {
	err := r.client.Set(id, 0, 0).Err()
	if err != nil {
		log.Fatalf("Failed to Free value: %v", err)
	}
}

// Get implements ratelimiter.Storage.
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

// Increase implements ratelimiter.Storage.
func (r *rlRedisStorage) Increase(id string) {

	if err := r.client.Incr(id).Err(); err != nil {
		log.Fatalf("Failed to Increase value: %v", err)
	}

	if err := r.client.Expire(id, r.ttl).Err(); err != nil {
		log.Printf("Failed to SetTTL for value: %v\n", err)
	}
}
