package config

import (
	"encoding/json"
	"os"
	"strings"

	extractor "ytproxy-extractor"
	cache "ytproxy-linkscache"
	logger "ytproxy-logger"
	streamer "ytproxy-streamer"
)

type configT struct {
	PortInt   uint16            `json:"port"`
	Streamer  streamer.ConfigT  `json:"streamer"`
	Extractor extractor.ConfigT `json:"extractor"`
	Log       logger.ConfigT    `json:"log"`
	Cache     cache.ConfigT     `json:"cache"`
}

func Read(path string) (configT, error) {
	var c configT
	b, err := os.ReadFile(path)
	if err != nil {
		return c, err
	}
	func() {
		strs := make([]string, 0)
		for _, s := range strings.Split(string(b[:]), "\n") {
			s = strings.TrimSpace(s)
			if !strings.HasPrefix(s, "//") {
				strs = append(strs, s)
			}
		}
		str := strings.Join(strs, "\n")
		b = b[:0]
		b = []byte(str)

	}()
	err = json.Unmarshal(b, &c)
	return c, err
}
