package logger

import (
	"fmt"
	config "lib/logger/config"
	logger_default "lib/logger/impl/default"
	logger_empty "lib/logger/impl/empty"
	logger_slog "lib/logger/impl/slog"
)

func New(conf config.ConfigT) (T, error) {
	if *conf.Level == config.Nothing {
		return logger_empty.New()
	}
	if *conf.Json {
		return logger_slog.New(conf)
	} else {
		return logger_default.New(conf)
	}
}

type T interface {
	LogError(string, ...any)
	LogWarning(string, ...any)
	LogDebug(string, ...any)
	LogInfo(string, ...any)
}

type loggerLayer struct {
	impl        T
	logger_name string
}

func NewLayer(impl T, name_str string) T {
	return &loggerLayer{
		impl:        impl,
		logger_name: name_str,
	}
}

func (t *loggerLayer) f(s string) string {
	switch s {
	case "":
		return t.logger_name
	default:
		return fmt.Sprintf("%s. %s", t.logger_name, s)
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
