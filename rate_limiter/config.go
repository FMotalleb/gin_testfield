package ratelimiter

import (
	"errors"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// Config is a struct that allows building a rate limiting middleware
// with configurable options.
type Config struct {
	limit       uint16
	workerCount uint16
	timeout     time.Duration
	tolerance   time.Duration
	idSelector  IDSelector
	storage     Storage
	queue       chan rlEntry
	handler     gin.HandlerFunc
	logger      *logrus.Logger
}

func (cfg *Config) addToReleaseQueue(id string) {
	cfg.queue <- rlEntry{
		userID:      id,
		releaseTime: time.Now().Add(cfg.timeout),
	}
}

// NewConfigBuilder creates a new RateLimitBuilder with default options.
//
//	limit: 60 requests
//	workerCount: 100
//	timeout: 1 minute
//	isSelector: (ctx) => ctx.ClientIp()
func NewConfigBuilder() *Config {
	return &Config{
		limit:       60,
		workerCount: 20,
		tolerance:   time.Second * 2,
		timeout:     time.Minute,
		idSelector:  defaultIdSelector,
		handler:     defaultHandler,
		queue:       make(chan rlEntry),
		storage:     NewHashMapStorage(),
	}
}

// Limit sets the rate limit for the middleware.
func (rlb *Config) Limit(limit uint16) *Config {
	rlb.limit = limit
	return rlb
}

// Limit sets the rate limit for the middleware.
func (rlb *Config) Handler(handler gin.HandlerFunc) *Config {
	rlb.handler = handler
	return rlb
}

// WorkerCount sets the number of worker goroutines for the middleware.
func (rlb *Config) WorkerCount(workers uint16) *Config {
	rlb.workerCount = workers
	return rlb
}

// Timeout sets the timeout duration for the rate limit.
func (rlb *Config) Timeout(timeout time.Duration) *Config {
	rlb.timeout = timeout
	return rlb
}

// IdSelector sets the function to select the unique identifier for a request.
func (rlb *Config) IdSelector(idSelector IDSelector) *Config {
	rlb.idSelector = idSelector
	return rlb
}

// Tolerance sets the tolerance duration that will be skipped if an entry should be deleted in that window.
func (rlb *Config) Tolerance(tolerance time.Duration) *Config {
	rlb.tolerance = tolerance
	return rlb
}

// Storage stores the request rate data.
func (rlb *Config) Storage(storage Storage) *Config {
	rlb.storage = storage
	return rlb
}

// Build creates a new Gin middleware function that implements the rate limiting logic.
//
// will return error if:
//
//	tolerance value is too high (if its bigger than the timeout value which will cause the worker to skip everything)
//
//	tolerance value is too high (if its bigger than the timeout value which will cause the worker to skip everything)
func (rlb *Config) Build() (h gin.HandlerFunc, e error) {
	if rlb.tolerance > rlb.timeout {

		return
	}
	switch {
	case rlb.idSelector == nil:
		e = errors.New("`IdSelector` value cannot be nil")
	case rlb.handler == nil:
		e = errors.New("`Handler` value cannot be nil")
	case rlb.storage == nil:
		e = errors.New("`Storage` value cannot be nil")
	case rlb.limit == 0:
		e = errors.New("`Limit` value cannot be 0")
	case rlb.timeout <= time.Second:
		e = errors.New("`Timeout` cannot be less than a time.Second")
	case rlb.tolerance < 0:
		e = errors.New("`Tolerance` value cannot be less than zero")
	case rlb.workerCount == 0:
		e = errors.New("`WorkerCount` cannot be 0")
	case rlb.tolerance > rlb.timeout:
		e = errors.New("tolerance value cannot be less than timeout")
	default:
		h = RateLimitWith(rlb)
	}
	return
}
