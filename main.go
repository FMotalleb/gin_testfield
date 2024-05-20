package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"sync"
	"time"

	ratelimiter "github.com/FMotalleb/gin_testfield/rate_limiter"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis"
)

func main() {
	port := os.Getenv("PORT_NUM")
	if port == "" {
		port = "8001"
	}

	router := gin.Default()
	client := redis.NewClient(&redis.Options{
		Addr:     "127.0.0.1:6379", // Redis server address
		Password: "",               // Redis password (leave empty if no password is set)
		DB:       0,                // Redis database index (0 by default)
	})

	_, err := client.Ping().Result()
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}

	storage := RlRedisStorage{
		client,
		sync.Mutex{},
	}

	rl, e := ratelimiter.
		NewConfigBuilder().
		Limit(10).
		Timeout(time.Second * 10).
		WorkerCount(10).
		Storage(&storage).
		Build()
	fmt.Println(e)
	router.Use(rl)
	router.GET("/foo", func(c *gin.Context) {
		c.String(200, "bar")
	})

	if err := router.Run(fmt.Sprintf(":%s", port)); err != nil {
		fmt.Println("Error starting server:", err)
	}
}

type RlRedisStorage struct {
	client *redis.Client
	lock   sync.Mutex
}

// Decrease implements ratelimiter.Storage.
func (r *RlRedisStorage) Decrease(id string) {
	err := r.client.Decr(id).Err()
	if err != nil {
		log.Fatalf("Failed to Decrease value: %v", err)
	}
}

// Free implements ratelimiter.Storage.
func (r *RlRedisStorage) Free(id string) {
	err := r.client.Set(id, 0, 0).Err()
	if err != nil {
		log.Fatalf("Failed to Free value: %v", err)
	}
}

// Get implements ratelimiter.Storage.
func (r *RlRedisStorage) Get(id string) uint16 {
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
func (r *RlRedisStorage) Increase(id string) {
	err := r.client.Incr(id).Err()
	if err != nil {
		log.Fatalf("Failed to Increase value: %v", err)
	}
}
