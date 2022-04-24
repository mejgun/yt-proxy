package config

import (
	"encoding/json"
	"os"

	extractor "ytproxy-extractor"
	logger "ytproxy-logger"
	streamer "ytproxy-streamer"
)

type configT struct {
	PortInt   uint16            `json:"port"`
	Streamer  streamer.ConfigT  `json:"streamer"`
	Extractor extractor.ConfigT `json:"extractor"`
	Log       logger.ConfigT    `json:"log"`
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
