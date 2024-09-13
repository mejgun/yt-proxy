package logger

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

type logFuncT func(string, ...interface{})

type T struct {
	LogError   logFuncT
	LogWarning logFuncT
	LogDebug   logFuncT
	LogInfo    logFuncT
}

type ConfigT struct {
	Level    LevelT  `json:"level"`
	Output   OutputT `json:"output"`
	FileName string  `json:"filename"`
}

type LevelT uint8

const (
	Debug LevelT = iota
	Info
	Warning
	Error
	Nothing
)

func (l *LevelT) UnmarshalJSON(b []byte) error {
	var s string
	err := json.Unmarshal(b, &s)
	if err != nil {
		return err
	}
	switch s {
	case "debug":
		*l = Debug
	case "info":
		*l = Info
	case "warning":
		*l = Warning
	case "error":
		*l = Error
	case "nothing":
		*l = Nothing
	default:
		return fmt.Errorf("cannot unmarshal %s as log level", b)
	}
	return nil
}

type OutputT uint8

const (
	Stdout OutputT = iota
	File
	Both
)

func (o *OutputT) UnmarshalJSON(b []byte) error {
	var s string
	err := json.Unmarshal(b, &s)
	if err != nil {
		return err
	}
	switch s {
	case "stdout":
		*o = Stdout
	case "file":
		*o = File
	case "both":
		*o = Both
	default:
		return fmt.Errorf("cannot unmarshal %s as log output", b)
	}
	return nil
}

func New(conf ConfigT) (*T, error) {
	var logger = T{
		LogError:   func(s string, i ...interface{}) {},
		LogWarning: func(s string, i ...interface{}) {},
		LogDebug:   func(s string, i ...interface{}) {},
		LogInfo:    func(s string, i ...interface{}) {},
	}
	if conf.Level == Nothing {
		return &logger, nil
	}
	var (
		l *log.Logger = log.Default()
	)
	open := func() (*os.File, error) {
		return os.OpenFile(
			// will never close this file :|
			// should trap exit
			conf.FileName, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0664)
	}
	switch conf.Output {
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
	switch conf.Level {
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
