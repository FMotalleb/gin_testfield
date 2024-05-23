package main

import (
	"fmt"
	"log"
	"os"
	"time"

	ratelimiter "github.com/FMotalleb/gin_testfield/rate_limiter"
	rlstorage "github.com/FMotalleb/gin_testfield/rate_limiter/storage"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis"
	"github.com/sirupsen/logrus"
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

	storage := rlstorage.NewRedisStorage(client, time.Second*10, logrus.StandardLogger())

	rl, e := ratelimiter.
		NewConfigBuilder().
		Limit(10).
		Timeout(time.Second * 10).
		WorkerCount(10).
		Storage(storage).
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

