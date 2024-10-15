// Package loggermux contains logger interface and logger layer constructor
package loggermux

import (
	"fmt"

	logger "ytproxy/logger"
	logger_default "ytproxy/logger/impl/default"
	logger_empty "ytproxy/logger/impl/empty"
	logger_slog "ytproxy/logger/impl/slog"
)

// New creates new logger implementation
func New(conf logger.ConfigT) (logger.T, error) {
	if *conf.Level == logger.Nothing {
		return logger_empty.New()
	}
	if *conf.JSON {
		return logger_slog.New(conf)
	}
	return logger_default.New(conf)
}

type loggerLayer struct {
	impl       logger.T
	loggerName string
}

// NewLayer creates wraps existing logger implementation
func NewLayer(impl logger.T, nameStr string) logger.T {
	return &loggerLayer{
		impl:       impl,
		loggerName: nameStr,
	}
}

func (t *loggerLayer) f(s string) string {
	switch s {
	case "":
		return t.loggerName
	default:
		return fmt.Sprintf("%s. %s", t.loggerName, s)
	}
}

func (t *loggerLayer) LogError(s string, i ...any) {
	t.impl.LogError(t.f(s), i...)
}
func (t *loggerLayer) LogWarning(s string, i ...any) {
	t.impl.LogWarning(t.f(s), i...)
}
func (t *loggerLayer) LogDebug(s string, i ...any) {
	t.impl.LogDebug(t.f(s), i...)
}
func (t *loggerLayer) LogInfo(s string, i ...any) {
	t.impl.LogInfo(t.f(s), i...)
}
func (t *loggerLayer) Close() error {
	return t.impl.Close()
}
