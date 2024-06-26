package ratelimiter

import (
	"errors"
	"time"

	"github.com/FMotalleb/gin_testfield/rate_limiter/cleanup"
	rlstorage "github.com/FMotalleb/gin_testfield/rate_limiter/storage"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// Config is a struct that allows building a rate limiting middleware
// with configurable options.
type Config struct {
	limit               uint16              // The maximum number of requests allowed within the timeout duration
	workerCount         uint16              // The number of worker goroutines to handle rate limiting
	timeout             time.Duration       // The duration for which the rate limit is enforced
	tolerance           time.Duration       // The tolerance duration that will be skipped if an entry should be deleted within that window
	idSelector          IDSelector          // A function that selects the unique identifier for a request
	storage             rlstorage.RLStorage // The storage backend used for rate limiting data
	queue               chan rateEntry      // A channel to queue rate limiting entries for release
	handler             gin.HandlerFunc     // The handler function to be executed if the rate limit is exceeded
	logger              *logrus.Logger      // The logger instance for logging messages
	fullCleanupRotation time.Duration       // FullCleanup rotation time to clean whole storage to cover possible memory leak scenario
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
//	fullCleanupRotation: 24 hours (use 0 value explicitly to disable the cleanup rotation)
func NewConfigBuilder() *Config {
	logger := logrus.StandardLogger()
	return &Config{
		limit:               60,
		workerCount:         20,
		tolerance:           time.Second * 2,
		timeout:             time.Minute,
		idSelector:          defaultIdSelector,
		handler:             defaultHandler,
		queue:               make(chan rateEntry),
		storage:             rlstorage.NewHashMapStorage(logger),
		logger:              logger,
		fullCleanupRotation: time.Hour * 24,
	}
}

// Logger sets the logger for the middleware.
func (cfg *Config) Logger(logger *logrus.Logger) *Config {
	cfg.logger = logger
	return cfg
}

// Limit sets the rate limit for the middleware.
func (cfg *Config) Limit(limit uint16) *Config {
	cfg.limit = limit
	return cfg
}

// Handler sets the handler function to be executed if the rate limit is exceeded.
func (cfg *Config) Handler(handler gin.HandlerFunc) *Config {
	cfg.handler = handler
	return cfg
}

// WorkerCount sets the number of worker goroutines for the middleware.
func (cfg *Config) WorkerCount(workers uint16) *Config {
	cfg.workerCount = workers
	return cfg
}

// Timeout sets the timeout duration for the rate limit.
func (cfg *Config) Timeout(timeout time.Duration) *Config {
	cfg.timeout = timeout
	return cfg
}

// IdSelector sets the function to select the unique identifier for a request.
func (cfg *Config) IdSelector(idSelector IDSelector) *Config {
	cfg.idSelector = idSelector
	return cfg
}

// Tolerance sets the tolerance duration that will be skipped if an entry should be deleted in that window.
func (cfg *Config) Tolerance(tolerance time.Duration) *Config {
	cfg.tolerance = tolerance
	return cfg
}

// Storage sets the storage backend used for rate data.
func (cfg *Config) Storage(storage rlstorage.RLStorage) *Config {
	cfg.storage = storage
	return cfg
}

// FullCleanupRotation sets the duration for the full cleanup rotation of the rate limiting storage.
// This duration determines how often the fullCleanupWorker will remove all entries from the storage.
//
// The full cleanup rotation helps prevent potential memory leaks by periodically freeing up resources
// occupied by stale entries in the storage.
func (cfg *Config) FullCleanupRotation(rotation time.Duration) *Config {
	cfg.fullCleanupRotation = rotation
	return cfg
}

// DisableFullCleanup disables the full cleanup rotation for the rate limiting storage.
// When disabled, the fullCleanupWorker goroutine will not be started, and the storage
// will not be periodically cleared.
//
// This method should be used with caution, as it can potentially lead to memory leaks
// if stale entries are not removed from the storage over time.
func (rlb *Config) DisableFullCleanup() *Config {
	rlb.fullCleanupRotation = 0
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
//   - Ensures that the fullCleanupRotation duration is not equal or less than the timeout duration.
//   - Ensures that the workerCount is not 0.
//
// If all validations pass, it creates a new rate limiting middleware handler using the RateLimitWith function.
// If any validation fails, it returns an appropriate error message.
//
// Additionally, it starts a goroutine to run the fullCleanupWorker function, which periodically removes
// all entries from the storage to prevent potential memory leaks.
//
// Returns:
//
//	h (gin.HandlerFunc): The rate limiting middleware handler.
//	e (error): An error if any validation fails, or nil if the configuration is valid.
func (cfg *Config) Build() (h gin.HandlerFunc, e error) {
	// Check if the tolerance duration is greater than the timeout duration
	if cfg.tolerance >= cfg.timeout {
		// If true, return an error indicating that the tolerance value cannot be greater than or equal to the timeout
		e = errors.New("tolerance value cannot be greater than or equal to timeout")
		return
	}

	// Use a switch statement to validate the configuration values
	switch {
	case cfg.idSelector == nil:
		e = errors.New("`IdSelector` value cannot be nil")
	case cfg.handler == nil:
		e = errors.New("`Handler` value cannot be nil")
	case cfg.storage == nil:
		e = errors.New("`Storage` value cannot be nil")
	case cfg.limit == 0:
		e = errors.New("`Limit` value cannot be 0")
	case cfg.timeout <= time.Second:
		e = errors.New("`Timeout` cannot be less than a time.Second")
	case cfg.tolerance < 0:
		e = errors.New("`Tolerance` value cannot be less than zero")
	case cfg.timeout < cfg.tolerance:
		e = errors.New("`Tolerance` value cannot be less than `Timeout`")
	case cfg.workerCount == 0:
		e = errors.New("`WorkerCount` cannot be 0")
	case cfg.fullCleanupRotation <= cfg.timeout:
		e = errors.New("`FullCleanupRotation` cannot be less than `Timeout`")
	default:
		// If all configurations are valid, create and return a new rate limiting middleware handler
		h = RateLimitWith(cfg)
		// Start a goroutine to run the fullCleanupWorker function if rotation was set above 0
		if cfg.fullCleanupRotation > 0 {
			cleanup.
				NewWorker(cfg.storage, cfg.timeout).
				Start()
		}
	}

	return
}
