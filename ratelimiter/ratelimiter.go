package ratelimiter

import "github.com/segmentio/rate-limiting-log-proxy/logger"

// RateLimiter is an interface for implementing a rate limiter
type RateLimiter interface {
	Limit(logger logger.Logger) bool
	Expired() bool
}
