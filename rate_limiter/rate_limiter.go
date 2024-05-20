package ratelimiter

import (
	"errors"
	"time"

	"github.com/gin-gonic/gin"
)

// IDSelector is a function type that selects a unique identifier for the client of a request.
// It takes a *gin.Context and returns a string identifier.
type IDSelector func(*gin.Context) string

// rateEntry represents an entry in the rate limiting queue.
type rateEntry struct {
	userID      string    // The user ID or identifier
	releaseTime time.Time // The time when the rate limiting entry should be released
}

// defaultIdSelector is the default implementation of the IDSelector function.
// It selects the client IP address as the identifier.
func defaultIdSelector(ctx *gin.Context) string {
	return ctx.ClientIP()
}

// defaultHandler is the default handler function that is called when the rate limit is exceeded.
// It aborts the request with a [429]"Too Many Requests" status code and an error message.
func defaultHandler(ctx *gin.Context) {
	ctx.AbortWithError(429, errors.New("too many requests"))
}

// rlWorker is a worker goroutine that processes rate limiting entries in the queue.
// It frees (decreases) the rate limiting entries when their release time is reached.
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
		cfg.storage.Decrease(toFree.userID)
	}
}

// RateLimitWith creates a new rate limiting middleware handler based on the provided configuration.
func RateLimitWith(cfg *Config) gin.HandlerFunc {
	cfg.logger.Infof("booting up RateLimiter with %d requests per user per %s, with %d worker goroutines", cfg.limit, cfg.timeout, cfg.workerCount)

	// Start the worker goroutines
	for i := cfg.workerCount; i > 0; i-- {
		go rlWorker(cfg, i)
	}

	return func(ctx *gin.Context) {
		id := cfg.idSelector(ctx)
		blocked := isBlocked(cfg, id)
		if blocked {
			cfg.handler(ctx)
			return
		}
		ctx.Next()
	}
}

// isBlocked checks if the current request should be blocked based on the rate limiting configuration.
// It returns true if the request should be blocked, and false otherwise.
func isBlocked(cfg *Config, id string) bool {
	currentState := cfg.storage.Get(id)
	if currentState >= cfg.limit {
		return true
	}
	cfg.storage.Increase(id)
	cfg.addToReleaseQueue(id)
	return false
}
