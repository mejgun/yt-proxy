package logger

import (
	"encoding/json"
	"fmt"
)

type logFuncT func(string, ...interface{})

type T struct {
	LogError   logFuncT
	LogWarning logFuncT
	LogDebug   logFuncT
	LogInfo    logFuncT
}

type ConfigT struct {
	Level    *LevelT  `json:"level"`
	Output   *OutputT `json:"output"`
	FileName *string  `json:"filename"`
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
