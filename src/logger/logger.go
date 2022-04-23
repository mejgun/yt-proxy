package logger

import (
	"io"
	"log"
	"os"
)

type logFuncT func(string, []interface{})

type LoggerT struct {
	LogError   logFuncT
	LogWarning logFuncT
	LogDebug   logFuncT
	LogInfo    logFuncT
}

type ConfigT struct {
	Level    LogLevelT
	Output   LogOutputT
	FileName string
}

type LogLevelT uint8

const (
	Debug LogLevelT = iota
	Info
	Warning
	Error
	Nothing
)

type LogOutputT uint8

const (
	Stdout LogOutputT = iota
	File
	Both
)

func NewLogger(conf ConfigT) (LoggerT, error) {
	var logger = LoggerT{
		LogError:   func(s string, i []interface{}) {},
		LogWarning: func(s string, i []interface{}) {},
		LogDebug:   func(s string, i []interface{}) {},
		LogInfo:    func(s string, i []interface{}) {},
	}
	if conf.Level == Nothing {
		return logger, nil
	}
	var (
		l   *log.Logger = log.Default()
		f   *os.File
		err error
	)
	switch conf.Output {
	case Stdout:
		break
	case File:
		f, err = os.OpenFile(conf.FileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			return logger, err
		}
		// will never close this file :|
		// should trap exit
		l.SetOutput(f)
		fallthrough
	case Both:
		l.SetOutput(io.MultiWriter(os.Stdout, f))
	}
	l.SetOutput(f)
	switch conf.Level {
	case Debug:
		logger.LogDebug = func(s string, i []interface{}) { l.Printf("[ DEBUG ] %s: %+v", s, i) }
		fallthrough
	case Info:
		logger.LogInfo = func(s string, i []interface{}) { l.Printf("[ INFO ] %s: %+v", s, i) }
		fallthrough
	case Warning:
		logger.LogWarning = func(s string, i []interface{}) { l.Printf("[ WARNING ] %s: %+v", s, i) }
		fallthrough
	case Error:
		logger.LogError = func(s string, i []interface{}) { l.Printf("[ ERROR ] %s: %+v", s, i) }
	}
	return logger, nil
}
