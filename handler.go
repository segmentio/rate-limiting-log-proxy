package main

import (
	"log"
	"strings"
	"sync"
	"time"

	dockerLogger "github.com/docker/docker/daemon/logger"
	"github.com/segmentio/rate-limiting-log-proxy/container"
	"github.com/segmentio/rate-limiting-log-proxy/logger"
	"github.com/segmentio/rate-limiting-log-proxy/ratelimiter"
	"gopkg.in/mcuadros/go-syslog.v2/format"
)

// RateLimitingHandler is a handler that will perform rate-limiting based
// on the container id of the incoming logs
type RateLimitingHandler struct {
	interval               time.Duration
	burst                  int
	containerLookupService container.Lookup
	loggerFactory          logger.Factory
	rateLimiters           map[string]ratelimiter.RateLimiter
	rateLimitersLock       sync.Mutex
}

// NewRateLimitingHandler creates a new RateLimitingHandler for use with a
// go-syslog server
func NewRateLimitingHandler(interval time.Duration, burst int, containerLookupService container.Lookup, loggerFactory logger.Factory) *RateLimitingHandler {
	handler := &RateLimitingHandler{
		interval: interval,
		burst:    burst,
		containerLookupService: containerLookupService,
		loggerFactory:          loggerFactory,
		rateLimiters:           map[string]ratelimiter.RateLimiter{},
	}
	go handler.purgeRateLimiters(time.Hour)
	return handler
}

// purgeRateLimiters evicts old rate limit structs which have expired
func (r *RateLimitingHandler) purgeRateLimiters(interval time.Duration) {
	for _ = range time.Tick(interval) {
		r.rateLimitersLock.Lock()
		keep := map[string]ratelimiter.RateLimiter{}
		for id, rl := range r.rateLimiters {
			if !rl.Expired() {
				keep[id] = rl
			}
		}
		r.rateLimiters = keep
		r.rateLimitersLock.Unlock()
	}
}

// Handle contains all the logic for the RateLimitHandler.  It is responsible for
// lookup up container info, creating a new logger if necessary, applying
// the rate limit, and sending the log to the downstream logger.
func (r *RateLimitingHandler) Handle(logParts format.LogParts, messageLength int64, err error) {
	tag, ok := logParts["tag"]
	if !ok {
		tag = "default"
	}
	tagStr := tag.(string)

	// older versions of docker prepend tag with "docker/"
	tagStr = strings.TrimPrefix(tagStr, "docker/")

	containerInfo, err := r.containerLookupService.Lookup(tagStr)
	if err != nil {
		containerInfo = dockerLogger.Info{}
		log.Printf("handler: %s", err)
		return
	}

	logger, err := r.loggerFactory.New(containerInfo)
	if err != nil {
		log.Printf("handler: %s", err)
		return
	}

	r.rateLimitersLock.Lock()
	rl, ok := r.rateLimiters[tagStr]
	if !ok {
		rl = ratelimiter.NewRsyslogStyle(r.interval, r.burst)
		r.rateLimiters[tagStr] = rl
	}
	r.rateLimitersLock.Unlock()

	if !rl.Limit(logger) {
		logger.Log(logParts["content"].(string))
	}
}
