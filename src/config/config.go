package config

import (
	"encoding/json"
	"os"

	extractor "ytproxy-extractor"
	logger "ytproxy-logger"
)

type configT struct {
	Log                  logger.ConfigT    `json:"log"`
	EnableErrorHeaders   bool              `json:"error-headers"`
	IgnoreMissingHeaders bool              `json:"ignore-missing-headers"`
	IgnoreSSLErrors      bool              `json:"ignore-ssl-errors"`
	PortInt              uint16            `json:"port"`
	ErrorVideoPath       string            `json:"error-video"`
	Extractor            extractor.ConfigT `json:"extractor"`
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
