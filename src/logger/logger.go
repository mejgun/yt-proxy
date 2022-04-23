package logger

import (
	"io"
	"log"
	"os"
)

type logFuncT func(string, interface{})

type T struct {
	LogError   logFuncT
	LogWarning logFuncT
	LogDebug   logFuncT
	LogInfo    logFuncT
}

type ConfigT struct {
	Level    LevelT
	Output   OutputT
	FileName string
}

type LevelT uint8

const (
	Debug LevelT = iota
	Info
	Warning
	Error
	Nothing
)

type OutputT uint8

const (
	Stdout OutputT = iota
	File
	Both
)

func New(conf ConfigT) (T, error) {
	var logger = T{
		LogError:   func(s string, i interface{}) {},
		LogWarning: func(s string, i interface{}) {},
		LogDebug:   func(s string, i interface{}) {},
		LogInfo:    func(s string, i interface{}) {},
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
		logger.LogDebug = func(s string, i interface{}) { l.Printf("[ DEBUG ] %s: %+v", s, i) }
		fallthrough
	case Info:
		logger.LogInfo = func(s string, i interface{}) { l.Printf("[ INFO ] %s: %+v", s, i) }
		fallthrough
	case Warning:
		logger.LogWarning = func(s string, i interface{}) { l.Printf("[ WARNING ] %s: %+v", s, i) }
		fallthrough
	case Error:
		logger.LogError = func(s string, i interface{}) { l.Printf("[ ERROR ] %s: %+v", s, i) }
	}
	return logger, nil
}
