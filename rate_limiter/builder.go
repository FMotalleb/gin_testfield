package ratelimiter

import (
	"time"

	"github.com/gin-gonic/gin"
)

// NewRateLimiter creates a new RateLimitBuilder with default options.
//
//	limit: 60 requests
//	workerCount: 100
//	timeout: 1 minute
//	isSelector: (ctx) => ctx.ClientIp()
func NewRateLimiter() *RateLimitBuilder {
	return &RateLimitBuilder{
		limit:       60,
		workerCount: 20,
		timeout:     time.Minute,
		idSelector:  defaultIdSelector,
	}
}

// Limit sets the rate limit for the middleware.
func (rlb *RateLimitBuilder) Limit(limit uint16) *RateLimitBuilder {
	rlb.limit = limit
	return rlb
}

// WorkerCount sets the number of worker goroutines for the middleware.
func (rlb *RateLimitBuilder) WorkerCount(workers uint16) *RateLimitBuilder {
	rlb.workerCount = workers
	return rlb
}

// Timeout sets the timeout duration for the rate limit.
func (rlb *RateLimitBuilder) Timeout(timeout time.Duration) *RateLimitBuilder {
	rlb.timeout = timeout
	return rlb
}

// IdSelector sets the function to select the unique identifier for a request.
func (rlb *RateLimitBuilder) IdSelector(idSelector IDSelector) *RateLimitBuilder {
	rlb.idSelector = idSelector
	return rlb
}

// Build creates a new Gin middleware function that implements the rate limiting logic.
func (rlb *RateLimitBuilder) Build() gin.HandlerFunc {
	return RateLimiter(
		rlb.limit,
		rlb.workerCount,
		rlb.timeout,
		rlb.idSelector,
	)
}
