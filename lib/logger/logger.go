package logger

import (
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
