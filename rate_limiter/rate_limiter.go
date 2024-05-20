package ratelimiter

import (
	"errors"
	"time"

	"github.com/FMotalleb/gin_testfield/logger"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

var (
	rlLockList map[string]uint16 = make(map[string]uint16)
	rlQueue    chan rateLimit    = make(chan rateLimit)
	rlLog      *logrus.Entry
	rlLimit    uint16
	rlTimeout  time.Duration
)

// IDSelector is a function that selects the unique identifier for a request.
// It takes a *gin.Context as input and returns a string.
type IDSelector func(*gin.Context) string

// RateLimitBuilder is a struct that allows building a rate limiting middleware
// with configurable options.
type RateLimitBuilder struct {
	limit       uint16
	workerCount uint16
	timeout     time.Duration
	idSelector  IDSelector
}

// rateLimit is a struct that represents a rate limit entry.
// It contains the unique identifier for the request and the release time.
type rateLimit struct {
	userID      string
	releaseTime time.Time
}

// defaultIdSelector is a default implementation of the IDSelector function
// that uses the client's IP address as the unique identifier.
func defaultIdSelector(c *gin.Context) string {
	return c.ClientIP()
}

// rlWorker is a worker goroutine that processes rate limit entries from the rlQueue.
// It waits for the appropriate timeout duration and then decrements the rate limit counter for the user.
func rlWorker(workerID uint16) {
	log := rlLog.WithField("worker_id", workerID)
	log.Infoln("starting")
	for toFree := range rlQueue {
		duration := toFree.releaseTime.Sub(time.Now())
		log := log.WithField("user_id", toFree.userID)
		if duration >= time.Second {
			log.WithField("timeout", duration).Infoln("waiting for timeout")
			time.Sleep(duration)
		}
		rlDecrement(log, toFree)
	}
}

// rlDecrement decrements the rate limit counter for the given user.
// If the counter reaches 0, the user's entry is removed from the rlLockList.
func rlDecrement(log *logrus.Entry, rateLimit rateLimit) {
	log.Infoln("removing last entry")
	userLockCount := rlLockList[rateLimit.userID]
	if userLockCount <= 1 {
		log.Infoln("removed all entries")
		delete(rlLockList, rateLimit.userID)
	} else {
		rlLockList[rateLimit.userID] = userLockCount - 1
	}
}

// rlAdd adds a new rate limit entry for the given user to the rlQueue.
func rlAdd(userID string) {
	log := rlLog.WithField("user_id", userID)
	rlLockList[userID]++
	rlQueue <- rateLimit{
		userID,
		time.Now().Add(rlTimeout),
	}
	log.Infoln("worker job dispatched")
}

// RateLimiter is a Gin middleware that implements a rate limiting mechanism.
// It limits the number of requests per user (IP) per [timeout] and blocks excess requests with a 429 Too Many Requests response.
func RateLimiter(limit uint16, workerCount uint16, timeout time.Duration, idSelector IDSelector) gin.HandlerFunc {
	rlTimeout = timeout
	rlLimit = limit
	rlLog = logger.
		SetupLogger("RateLimiter").
		WithFields(
			logrus.Fields{
				"limit":   limit,
				"workers": workerCount,
				"timeout": timeout,
			},
		)
	rlLog.Infof("booting up RateLimiter with %d requests per user (IP) per minute, with %d worker goroutines", limit, workerCount)
	for i := workerCount; i > 0; i-- {
		go rlWorker(i)
	}

	return func(ctx *gin.Context) {
		// ctx.Value()
		userIP := ctx.ClientIP()
		rlLog.WithFields(
			logrus.Fields{
				"ip":            userIP,
				"current_queue": rlLockList[userIP],
			},
		).Infoln("request received")
		if rlLockList[userIP] >= limit {
			ctx.AbortWithError(429, errors.New("Too many requests"))
			return
		}
		go rlAdd(userIP)
		ctx.Next()
	}
}
