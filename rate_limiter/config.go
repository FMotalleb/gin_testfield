package ratelimiter

import (
	"errors"
	"time"

	rlstorage "github.com/FMotalleb/gin_testfield/rate_limiter/storage"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// Config is a struct that allows building a rate limiting middleware
// with configurable options.
type Config struct {
	limit       uint16              // The maximum number of requests allowed within the timeout duration
	workerCount uint16              // The number of worker goroutines to handle rate limiting
	timeout     time.Duration       // The duration for which the rate limit is enforced
	tolerance   time.Duration       // The tolerance duration that will be skipped if an entry should be deleted within that window
	idSelector  IDSelector          // A function that selects the unique identifier for a request
	storage     rlstorage.RLStorage // The storage backend used for rate limiting data
	queue       chan rateEntry      // A channel to queue rate limiting entries for release
	handler     gin.HandlerFunc     // The handler function to be executed if the rate limit is exceeded
	logger      *logrus.Logger      // The logger instance for logging messages
}

func (cfg *Config) addToReleaseQueue(id string) {
	// Adds a rate limiting entry to the release queue with the given ID
	// and a release time calculated based on the timeout duration.
	cfg.queue <- rateEntry{
		userID:      id,
		releaseTime: time.Now().Add(cfg.timeout),
	}
}

// NewConfigBuilder creates a new RateLimitBuilder with default options.
//
//	limit: 60 requests
//	workerCount: 20
//	timeout: 1 minute
//	idSelector: defaultIdSelector (selects the client IP address)
//	handler: defaultHandler (returns [429]"too many requests")
//	queue: a new unbuffered channel for rateEntry
//	storage: an in-memory HashMap storage
//	logger: the standard logger instance
func NewConfigBuilder() *Config {

	return &Config{
		limit:       60,
		workerCount: 20,
		tolerance:   time.Second * 2,
		timeout:     time.Minute,
		idSelector:  defaultIdSelector,
		handler:     defaultHandler,
		queue:       make(chan rateEntry),
		storage:     rlstorage.NewHashMapStorage(),
		logger:      logrus.StandardLogger(),
	}
}

// Logger sets the logger for the middleware.
func (rlb *Config) Logger(logger *logrus.Logger) *Config {
	rlb.logger = logger
	return rlb
}

// Limit sets the rate limit for the middleware.
func (rlb *Config) Limit(limit uint16) *Config {
	rlb.limit = limit
	return rlb
}

// Handler sets the handler function to be executed if the rate limit is exceeded.
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

// Storage sets the storage backend used for rate data.
func (rlb *Config) Storage(storage rlstorage.RLStorage) *Config {
	rlb.storage = storage
	return rlb
}

// Build validates the configuration values and creates a new rate limiting middleware handler.
// It returns the handler function (gin.HandlerFunc) and an error (if any).
//
// The method performs the following validations:
//   - Ensures that the tolerance duration is not equal or greater than the timeout duration.
//   - Ensures that the idSelector, handler, and storage are not nil.
//   - Ensures that the limit is not 0.
//   - Ensures that the timeout is greater than 1 second.
//   - Ensures that the tolerance is not less than 0.
//   - Ensures that the workerCount is not 0.
//
// If all validations pass, it creates a new rate limiting middleware handler using the RateLimitWith function.
// If any validation fails, it returns an appropriate error message.
//
// Returns:
//
//	h (gin.HandlerFunc): The rate limiting middleware handler.
//	e (error): An error if any validation fails, or nil if the configuration is valid.
func (cfg *Config) Build() (h gin.HandlerFunc, e error) {
	// Check if the tolerance duration is greater than the timeout duration
	if cfg.tolerance > cfg.timeout {
		// If true, return an error indicating that the tolerance value cannot be greater than the timeout
		e = errors.New("tolerance value cannot be less than timeout")
		return
	}

	// Use a switch statement to validate the configuration values
	switch {
	case cfg.idSelector == nil:
		// If the idSelector is nil, return an error
		e = errors.New("`IdSelector` value cannot be nil")
	case cfg.handler == nil:
		// If the handler is nil, return an error
		e = errors.New("`Handler` value cannot be nil")
	case cfg.storage == nil:
		// If the storage is nil, return an error
		e = errors.New("`Storage` value cannot be nil")
	case cfg.limit == 0:
		// If the limit is 0, return an error
		e = errors.New("`Limit` value cannot be 0")
	case cfg.timeout <= time.Second:
		// If the timeout is less than or equal to 1 second, return an error
		e = errors.New("`Timeout` cannot be less than a time.Second")
	case cfg.tolerance < 0:
		// If the tolerance is less than 0, return an error
		e = errors.New("`Tolerance` value cannot be less than zero")
	case cfg.workerCount == 0:
		// If the workerCount is 0, return an error
		e = errors.New("`WorkerCount` cannot be 0")
	default:
		// If all configurations are valid, create and return a new rate limiting middleware handler
		h = RateLimitWith(cfg)
	}

	return
}
