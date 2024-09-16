package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

func NewDefault(conf ConfigT) (*T, error) {
	var logger = T{
		LogError:   func(s string, i ...interface{}) {},
		LogWarning: func(s string, i ...interface{}) {},
		LogDebug:   func(s string, i ...interface{}) {},
		LogInfo:    func(s string, i ...interface{}) {},
	}
	if *conf.Level == Nothing {
		return &logger, nil
	}
	var (
		l *log.Logger = log.Default()
	)
	open := func() (*os.File, error) {
		return os.OpenFile(
			// will never close this file :|
			// should trap exit
			*conf.FileName, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0664)
	}
	switch *conf.Output {
	case Stdout:
		l.SetOutput(os.Stdout)
	case File:
		f, err := open()
		if err != nil {
			return &logger, err
		}
		l.SetOutput(f)
	case Both:
		f, err := open()
		if err != nil {
			return &logger, err
		}
		l.SetOutput(io.MultiWriter(os.Stdout, f))
	}
	print := func(str string, s string, i []interface{}) {
		l.Printf(
			fmt.Sprintf("[ %s ] %s:", str, s) +
				fmt.Sprintf(strings.Repeat(" %+v", len(i)), i...))
	}
	switch *conf.Level {
	case Debug:
		logger.LogDebug = func(s string, i ...interface{}) { print("DEBUG", s, i) }
		fallthrough
	case Info:
		logger.LogInfo = func(s string, i ...interface{}) { print("INFO", s, i) }
		fallthrough
	case Warning:
		logger.LogWarning = func(s string, i ...interface{}) { print("WARNING", s, i) }
		fallthrough
	case Error:
		logger.LogError = func(s string, i ...interface{}) { print("ERROR", s, i) }
	}
	return &logger, nil
}
