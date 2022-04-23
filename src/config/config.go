package config

import (
	"encoding/json"
	"os"

	logger "ytproxy-logger"
)

type configT struct {
	Log                  logger.LogConfigT `json:"log"`
	EnableErrorHeaders   bool              `json:"error-headers"`
	IgnoreMissingHeaders bool              `json:"ignore-missing-headers"`
	IgnoreSSLErrors      bool              `json:"ignore-ssl-errors"`
	PortInt              uint16            `json:"port"`
	ErrorVideoPath       string            `json:"error-video"`
	Extractor            extractorT        `json:"extractor"`
}

type extractorT struct {
	Path         string   `json:"path"`
	MP4          []string `json:"mp4"`
	M4A          []string `json:"m4a"`
	GetUserAgent []string `json:"get-user-agent"`
}

func Read(path string) (configT, error) {
	var c configT
	b, err := os.ReadFile(path)
	if err != nil {
		return c, err
	}
	err = json.Unmarshal(b, &c)
	return c, err
}
