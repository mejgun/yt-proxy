package logger

import (
	"io"
	"log/slog"
	"os"

	l "lib/logger/config"
)

type loggerT struct {
	lgr *slog.Logger
}

func (t *loggerT) LogError(s string, i ...any) {
	t.lgr.Error(s, i...)
}
func (t *loggerT) LogWarning(s string, i ...any) {
	t.lgr.Warn(s, i...)
}
func (t *loggerT) LogDebug(s string, i ...any) {
	t.lgr.Debug(s, i...)

}
func (t *loggerT) LogInfo(s string, i ...any) {
	t.lgr.Info(s, i...)
}

func New(conf l.ConfigT) (*loggerT, error) {
	open := func() (*os.File, error) {
		return os.OpenFile(
			// will never close this file :|
			// should trap exit
			*conf.FileName, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0664)
	}
	var (
		lvl slog.Level
		lgr *slog.Logger
	)
	switch *conf.Level {
	case l.Debug:
		lvl = slog.LevelDebug
	case l.Info:
		lvl = slog.LevelInfo
	case l.Warning:
		lvl = slog.LevelWarn
	case l.Error:
		lvl = slog.LevelError
	}
	mkLogger := func(dst io.Writer) {
		lgr = slog.New(
			slog.NewJSONHandler(dst,
				&slog.HandlerOptions{Level: lvl}))
	}
	switch *conf.Output {
	case l.Stdout:
		mkLogger(os.Stdout)
	case l.File:
		f, err := open()
		if err != nil {
			return &loggerT{}, err
		}
		mkLogger(f)
	case l.Both:
		f, err := open()
		if err != nil {
			return &loggerT{}, err
		}
		mkLogger(io.MultiWriter(os.Stdout, f))
	}
	return &loggerT{
		lgr: lgr,
	}, nil
}
