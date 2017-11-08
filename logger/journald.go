package logger

import (
	"log"
	"strings"
	"time"

	"github.com/coreos/go-systemd/journal"
	dockerLogger "github.com/docker/docker/daemon/logger"
	"github.com/docker/docker/daemon/logger/loggerutils"
	cache "github.com/patrickmn/go-cache"
)

const (
	TaskARNLabel = "com.amazonaws.ecs.task-arn"
)

// JournaldLoggerFactory is a factory for creating journald loggers
type JournaldLoggerFactory struct {
	loggers *cache.Cache
}

// New creates a new JournaldLogger.  It also caches loggers, returning an
// already allocated logger if one exists.
func (j *JournaldLoggerFactory) New(info dockerLogger.Info) (Logger, error) {
	logger, ok := j.loggers.Get(info.ContainerID)
	if ok {
		return logger.(*JournaldLogger), nil
	}

	newLogger := NewJournaldLogger(info)
	j.loggers.SetDefault(info.ContainerID, newLogger)

	return newLogger, nil
}

// NewJournaldLoggerFactory creates a factory for creating journald loggers
func NewJournaldLoggerFactory() *JournaldLoggerFactory {
	return &JournaldLoggerFactory{
		loggers: cache.New(30*time.Minute, 1*time.Hour),
	}
}

// JournaldLogger is a logger which sends logs to a local journald instance
type JournaldLogger struct {
	vars map[string]string
}

// NewJournaldLogger creates a journald logger for a given container
func NewJournaldLogger(info dockerLogger.Info) *JournaldLogger {
	// Load tag format from "tag" docker label, since the proxy
	// requires the actual tag config to be the container ID
	tagFormat, ok := info.ContainerLabels["tag"]
	if ok && tagFormat != "" {
		info.Config["tag"] = tagFormat
	}

	tag, err := loggerutils.ParseLogTag(info, loggerutils.DefaultTemplate)
	if err != nil {
		log.Printf("logger: failed to parse logtag: %s", err)
		tag = info.ContainerID
	}

	vars := map[string]string{
		"CONTAINER_ID":      info.ContainerID[:12],
		"CONTAINER_ID_FULL": info.ContainerID,
		"CONTAINER_NAME":    info.Name(),
		"CONTAINER_TAG":     tag,
	}

	if taskARN, ok := info.ContainerLabels[TaskARNLabel]; ok {
		vars["CONTAINER_TASK"] = taskARN
		vars["CONTAINER_TASK_UUID"] = arnToUUID(taskARN)
	}

	return &JournaldLogger{vars}
}

// Log sends a log line to journald
func (j *JournaldLogger) Log(line string) error {
	return journal.Send(line, journal.PriInfo, j.vars)
}

func arnToUUID(task string) string {
	// ECS Task ARN looks like:
	// arn:aws:ecs:<region>:<aws_account_id>:task/<UUID>
	ix := strings.IndexByte(task, '/')
	return task[ix+1:]
}
