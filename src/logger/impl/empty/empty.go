// Package emptylogger implements empty (null) logger
package emptylogger

import (
	l "ytproxy/logger"
)

type loggerT struct{}

func (t *loggerT) LogError(string, ...any)   {}
func (t *loggerT) LogWarning(string, ...any) {}
func (t *loggerT) LogDebug(string, ...any)   {}
func (t *loggerT) LogInfo(string, ...any)    {}
func (t *loggerT) Close() error              { return nil }

// New creates null logger
func New() (l.T, error) {
	return &loggerT{}, nil
}
