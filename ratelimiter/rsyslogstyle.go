package ratelimiter

import (
	"fmt"
	"sync"
	"time"

	"github.com/segmentio/rate-limiting-log-proxy/logger"
)

// RsyslogStyle is a rate limiter that works similar to rsyslog (and journald's)
// rate limiters, allowing you to set an interval and a burst.  If there are more
// messages than burst in a given interval, all further messages are dropped
// until the end of that interval.
type RsyslogStyle struct {
	interval time.Duration
	burst    int

	start  time.Time
	count  int
	missed int
	mu     sync.Mutex
}

// NewRsyslogStyle returns a RateLimiter based on rsyslog's rate limiter
func NewRsyslogStyle(interval time.Duration, burst int) *RsyslogStyle {
	return &RsyslogStyle{
		interval: interval,
		burst:    burst,
	}
}

// Limit should be called each time you want to test if you've reached your rate
// limit.  It will return true if you are over the limit and false otherwise.
func (r *RsyslogStyle) Limit(logger logger.Logger) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()

	if r.start.IsZero() {
		r.start = now
	}

	if now.After(r.start.Add(r.interval)) {
		if r.missed > 0 {
			// TODO: send datadog stats about dropped messages
			logger.Log(fmt.Sprintf("log-proxy: dropped %d log lines, the service is logging too much", r.missed))
		}
		r.count = 0
		r.missed = 0
		r.start = now
	}
	r.count++

	if r.count <= r.burst {
		return false
	}

	r.missed++
	if r.missed == 1 {
		logger.Log("log-proxy: beginning to drop log lines, the service is logging too much")
	}
	return true
}

// Expired returns whether or not this rate limiter is able to be removed.  We
// mark an RsyslogStyle rate limiter expired if there have been no checks
// issued for an hour.
func (r *RsyslogStyle) Expired() bool {
	now := time.Now()

	// Expire rate limiter if it hasn't been used in an hour
	if r.start.Add(time.Hour).Before(now) {
		return true
	}
	return false
}
