package cache

import (
	"time"

	extractor_config "ytproxy/extractor/config"
)

type T interface {
	Add(extractor_config.RequestT, extractor_config.ResultT, time.Time)
	Get(extractor_config.RequestT) (extractor_config.ResultT, bool)
	CleanExpired(time.Time) []extractor_config.RequestT
}

type ConfigT struct {
	ExpireTime *string `json:"expire-time"`
}
