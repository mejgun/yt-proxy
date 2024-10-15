// Package logger contains logger interface type and logger constructor config
package logger

import (
	"encoding/json"
	"fmt"
)

// T is logger interface type
type T interface {
	LogError(string, ...any)
	LogWarning(string, ...any)
	LogDebug(string, ...any)
	LogInfo(string, ...any)
	Close() error
}

// ConfigT is logger constructor config
type ConfigT struct {
	Level    *LevelT  `json:"level"`
	JSON     *bool    `json:"json"`
	Output   *OutputT `json:"output"`
	FileName *string  `json:"filename"`
}

// LevelT is logger level
type LevelT uint8

// logger levels
const (
	Debug LevelT = iota
	Info
	Warning
	Error
	Nothing
)

// UnmarshalJSON for logger level json parsing, do not use directly
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

// OutputT select log destination
type OutputT uint8

// logger destination
const (
	Stdout OutputT = iota
	File
	Both
)

// UnmarshalJSON do not use directly
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
