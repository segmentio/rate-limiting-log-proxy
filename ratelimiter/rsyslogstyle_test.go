package ratelimiter

import (
	"testing"
	"time"

	"github.com/segmentio/rate-limiting-log-proxy/logger"
	"github.com/stretchr/testify/assert"
)

func TestRsyslogStyleRateLimiter(t *testing.T) {
	limiter := NewRsyslogStyle(time.Millisecond*500, 1)
	logger := logger.NewMockLogger()

	// Should allow a single message in first 500 ms
	limitted := limiter.Limit(logger)
	assert.Equal(t, false, limitted)
	limitted = limiter.Limit(logger)
	assert.Equal(t, true, limitted)
	limitted = limiter.Limit(logger)
	assert.Equal(t, true, limitted)

	// after interval, another message should be allowed
	time.Sleep(time.Millisecond * 500)
	limitted = limiter.Limit(logger)
	assert.Equal(t, false, limitted)

	// Make sure logging of dropped messages works
	assert.Equal(t, 2, len(logger.Messages))
	assert.Equal(t, "log-proxy: beginning to drop log lines, the service is logging too much", logger.Messages[0])
	assert.Equal(t, "log-proxy: dropped 2 log lines, the service is logging too much", logger.Messages[1])
}
