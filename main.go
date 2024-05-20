package main

import (
	"fmt"
	"time"

	ratelimiter "github.com/FMotalleb/gin_testfield/rate_limiter"
	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.Default()
	rl, e := ratelimiter.
		NewConfigBuilder().
		Limit(10).
		Timeout(time.Second * 10).
		WorkerCount(10).
		Build()
	fmt.Println(e)
	router.Use(rl)
	router.GET("/foo", func(c *gin.Context) {
		c.String(200, "bar")
	})

	if err := router.Run(":8001"); err != nil {
		fmt.Println("Error starting server:", err)
	}
}
