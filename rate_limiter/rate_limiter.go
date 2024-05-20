package ratelimiter

import (
	"errors"
	"time"

	"github.com/gin-gonic/gin"
)

type DataStorage map[string]uint16

// IDSelector is a function that selects the unique identifier for a request.
// It takes a *gin.Context as input and returns a string.
type IDSelector func(*gin.Context) string

// rlEntry is a struct that represents a rate limit entry.
// It contains the unique identifier for the request and the release time.
type rlEntry struct {
	userID      string
	releaseTime time.Time
}

// defaultIdSelector is a default implementation of the IDSelector function
// that uses the client's IP address as the unique identifier.
func defaultIdSelector(ctx *gin.Context) string {
	return ctx.ClientIP()
}

func defaultHandler(ctx *gin.Context) {
	ctx.AbortWithError(429, errors.New("too many requests"))
}

// rlWorker is a worker goroutine that processes rate limit entries from the rlQueue.
// It waits for the appropriate timeout duration and then decrements the rate limit counter for the user.
func rlWorker(cfg *Config, workerID uint16) {
	log := cfg.logger.
		WithField("scope", "rate-limiter").
		WithField("worker_id", workerID)
	log.Infoln("starting")
	for toFree := range cfg.queue {
		duration := toFree.releaseTime.Sub(time.Now())
		log := log.WithField("user_id", toFree.userID)
		if duration >= cfg.tolerance {
			log.WithField("timeout", duration).Debugln("waiting for timeout")
			time.Sleep(duration)
		}
		go cfg.storage.Decrease(toFree.userID)
	}
}

// RateLimiter is a Gin middleware that implements a rate limiting mechanism.
// It limits the number of requests per user (IP) per [timeout] and blocks excess requests with a 429 Too Many Requests response.
func RateLimitWith(cfg *Config) gin.HandlerFunc {
	cfg.logger.Infof("booting up RateLimiter with %d requests per user per %s, with %d worker goroutines", cfg.limit, cfg.timeout, cfg.workerCount)
	for i := cfg.workerCount; i > 0; i-- {
		go rlWorker(cfg, i)
	}

	return func(ctx *gin.Context) {
		shouldReturn := handler(cfg, ctx)
		if shouldReturn {
			return
		}
	}
}

func handler(cfg *Config, ctx *gin.Context) bool {
	id := cfg.idSelector(ctx)
	currentState := cfg.storage.Get(id)
	if currentState >= cfg.limit {
		cfg.handler(ctx)
		return true
	}
	cfg.storage.Increase(id)
	cfg.addToReleaseQueue(id)
	ctx.Next()
	return false
}
