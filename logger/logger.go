package logger

import (
	"log"

	dockerLogger "github.com/docker/docker/daemon/logger"
)

// LoggerType specifies which type of downstream logger our handlers should
// create
type LoggerType string

// These are the available LoggerType's
const (
	Journald LoggerType = "journald"
	Syslog   LoggerType = "syslog"
)

// Logger is anything that can take a log line and send it somewhere
type Logger interface {
	Log(line string) error
}

// Factory is an interface for creating new loggers
type Factory interface {
	New(info dockerLogger.Info) (Logger, error)
}

// NewLoggerFactory takes a log type and returns the appropriate factory
func NewLoggerFactory(typ LoggerType) Factory {
	switch typ {
	case Journald:
		return NewJournaldLoggerFactory()
	default:
		log.Fatalf("`%s` not yet implemented", string(typ))
	}
	return nil
}
