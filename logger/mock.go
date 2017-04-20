package logger

import (
	dockerLogger "github.com/docker/docker/daemon/logger"
)

type MockLogger struct {
	Messages []string
}

func NewMockLogger() *MockLogger {
	return &MockLogger{Messages: []string{}}
}

func (m *MockLogger) Log(line string) error {
	m.Messages = append(m.Messages, line)
	return nil
}

type MockLoggerFactory struct {
	Loggers map[string]*MockLogger
}

func NewMockLoggerFactory() *MockLoggerFactory {
	return &MockLoggerFactory{Loggers: map[string]*MockLogger{}}
}

func (m *MockLoggerFactory) New(info dockerLogger.Info) (Logger, error) {
	logger, ok := m.Loggers[info.ContainerID]
	if ok {
		return logger, nil
	}

	newLogger := NewMockLogger()
	m.Loggers[info.ContainerID] = newLogger
	return newLogger, nil
}
