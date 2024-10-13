package logger

import (
	"io"
	"log/slog"
	"os"
	"sync"

	l "ytproxy/logger/config"
)

type loggerT struct {
	mu      sync.RWMutex
	lgr     *slog.Logger
	outputs []*os.File
}

func (t *loggerT) LogError(s string, i ...any) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	t.lgr.Error(s, i...)
}
func (t *loggerT) LogWarning(s string, i ...any) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	t.lgr.Warn(s, i...)
}
func (t *loggerT) LogDebug(s string, i ...any) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	t.lgr.Debug(s, i...)

}
func (t *loggerT) LogInfo(s string, i ...any) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	t.lgr.Info(s, i...)
}

func (t *loggerT) Close() {
	t.mu.Lock()
	defer t.mu.Unlock()
	for _, v := range t.outputs {
		v.Close()
	}
	t.lgr = slog.Default()
}

func New(conf l.ConfigT) (*loggerT, error) {
	open := func() (*os.File, error) {
		return os.OpenFile(
			*conf.FileName,
			os.O_APPEND|os.O_WRONLY|os.O_CREATE,
			0664)
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
	mkLogger := func(dst1, dst2 *os.File) {
		var dst io.Writer
		if dst2 == nil {
			dst = dst2
		} else {
			dst = io.MultiWriter(dst1, dst2)
		}
		lgr = slog.New(
			slog.NewJSONHandler(dst,
				&slog.HandlerOptions{Level: lvl}))
	}
	outputs := make([]*os.File, 0)
	switch *conf.Output {
	case l.Stdout:
		mkLogger(os.Stdout, nil)
	case l.File:
		f, err := open()
		if err != nil {
			return &loggerT{}, err
		}
		outputs = append(outputs, f)
		mkLogger(f, nil)
	case l.Both:
		f, err := open()
		if err != nil {
			return &loggerT{}, err
		}
		outputs = append(outputs, f)
		mkLogger(os.Stdout, f)
	}
	return &loggerT{
		lgr:     lgr,
		outputs: outputs,
	}, nil
}
