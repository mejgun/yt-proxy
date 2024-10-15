// Package extractor contains extractor interface and related types
package extractor

import (
	"time"
	logger "ytproxy/logger"
)

// T is extractor interface
type T interface {
	Extract(RequestT, logger.T) (ResultT, error)
	GetUserAgent(logger.T) (string, error)
}

// ConfigT is constructor config type
type ConfigT struct {
	Path          *string   `json:"path"`
	MP4           *string   `json:"mp4"`
	M4A           *string   `json:"m4a"`
	GetUserAgent  *string   `json:"get-user-agent"`
	CustomOptions *[]string `json:"custom-options"`
	ForceHTTPS    *bool     `json:"force-https"`
}

// ResultT is extractor's result type
type ResultT struct {
	URL    string
	Expire time.Time
}

// RequestT is request type for extractor
type RequestT struct {
	URL    string
	HEIGHT string
	FORMAT string
}
