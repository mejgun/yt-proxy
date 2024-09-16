package logger

import (
	"io"
	"log/slog"
	"os"
)

func NewSlog(conf ConfigT) (*T, error) {
	var logger = T{
		LogError:   func(s string, i ...interface{}) {},
		LogWarning: func(s string, i ...interface{}) {},
		LogDebug:   func(s string, i ...interface{}) {},
		LogInfo:    func(s string, i ...interface{}) {},
	}
	if *conf.Level == Nothing {
		return &logger, nil
	}
	open := func() (*os.File, error) {
		return os.OpenFile(
			// will never close this file :|
			// should trap exit
			*conf.FileName, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0664)
	}
	var (
		lvl slog.Level
		l   *slog.Logger
	)
	switch *conf.Level {
	case Debug:
		lvl = slog.LevelDebug
	case Info:
		lvl = slog.LevelInfo
	case Warning:
		lvl = slog.LevelWarn
	case Error:
		lvl = slog.LevelError
	}
	mkLogger := func(dst io.Writer) {
		l = slog.New(
			slog.NewTextHandler(dst,
				&slog.HandlerOptions{Level: lvl}))
	}
	switch *conf.Output {
	case Stdout:
		mkLogger(os.Stdout)
	case File:
		f, err := open()
		if err != nil {
			return &logger, err
		}
		mkLogger(f)
	case Both:
		f, err := open()
		if err != nil {
			return &logger, err
		}
		mkLogger(io.MultiWriter(os.Stdout, f))
	}
	logger.LogDebug = func(s string, i ...interface{}) {
		l.Debug(s, i...)
	}
	logger.LogInfo = func(s string, i ...interface{}) {
		l.Info(s, i...)
	}
	logger.LogWarning = func(s string, i ...interface{}) {
		l.Warn(s, i...)
	}
	logger.LogError = func(s string, i ...interface{}) {
		l.Error(s, i...)
	}
	return &logger, nil
}
