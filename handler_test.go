package main

import (
	"testing"
	"time"

	dockerLogger "github.com/docker/docker/daemon/logger"
	"github.com/segmentio/rate-limiting-log-proxy/container"
	"github.com/segmentio/rate-limiting-log-proxy/logger"
	"github.com/stretchr/testify/assert"
	"gopkg.in/mcuadros/go-syslog.v2/format"
)

func setupMockLookup() *container.MockLookup {
	mock := container.NewMockLookup()
	mock.Store = map[string]dockerLogger.Info{
		"1": dockerLogger.Info{
			ContainerID:        "1",
			ContainerImageName: "/test",
			Config: map[string]string{
				"tag": "",
			},
		},
		"2": dockerLogger.Info{
			ContainerID:        "2",
			ContainerImageName: "/test",
			Config: map[string]string{
				"tag": "",
			},
			ContainerLabels: map[string]string{
				"tag": "{{.ID}}",
			},
		},
	}
	return mock
}

func TestRateLimitingHandler(t *testing.T) {
	mockLookup := setupMockLookup()
	mockLoggerFactory := logger.NewMockLoggerFactory()
	handler := NewRateLimitingHandler(time.Millisecond*500, 1, mockLookup, mockLoggerFactory)

	containerLogs := []format.LogParts{
		{
			"content": "Look, I'm a log",
			"tag":     "1",
		},
		{
			"content": "logging from second container",
			"tag":     "2",
		},
		{
			"content": "more logs",
			"tag":     "2",
		},
		{
			"content": "container 2 is noisy",
			"tag":     "2",
		},
	}

	for _, log := range containerLogs {
		handler.Handle(log, int64(len(log["content"].(string))), nil)
	}
	time.Sleep(time.Millisecond * 500)

	// Need to send a message after limit interval to flush the "dropped x
	// messages" message
	flushMessage := format.LogParts{
		"content": "flush dropped messages message",
		"tag":     "2",
	}
	handler.Handle(flushMessage, int64(len(flushMessage["content"].(string))), nil)

	assert.Equal(t, 2, len(mockLoggerFactory.Loggers))

	firstContainerLogger := mockLoggerFactory.Loggers["1"]
	assert.Equal(t, 1, len(firstContainerLogger.Messages))
	assert.Equal(t, containerLogs[0]["content"], firstContainerLogger.Messages[0])

	secondContainerLogger := mockLoggerFactory.Loggers["2"]
	assert.Equal(t, 4, len(secondContainerLogger.Messages))
	assert.Equal(t, containerLogs[1]["content"], secondContainerLogger.Messages[0])
	assert.Equal(t, "beginning to drop messages", secondContainerLogger.Messages[1])
	assert.Equal(t, "dropped 2 messages", secondContainerLogger.Messages[2])
	assert.Equal(t, flushMessage["content"], secondContainerLogger.Messages[3])

}
