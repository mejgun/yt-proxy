package cache

import (
	"fmt"
	"time"

	cache_default "lib/cache/impl/default"
	cache_empty "lib/cache/impl/empty"
	extractor_config "lib/extractor/config"
	logger "lib/logger"
)

type T interface {
	Add(extractor_config.RequestT, extractor_config.ResultT, time.Time)
	Get(extractor_config.RequestT) (extractor_config.ResultT, bool)
	CleanExpired(time.Time) []extractor_config.RequestT
}

type ConfigT struct {
	ExpireTime *string `json:"expire-time"`
}

func New(conf ConfigT, log logger.T) (T, error) {
	t, err := time.ParseDuration(*conf.ExpireTime)
	if err != nil {
		return cache_default.New(0), err
	}
	if t.Seconds() < 1 {
		log.LogDebug("cache", "disabled by config")
		return cache_empty.New(), nil
	}
	log.LogDebug("cache", fmt.Sprintf("expire time set to %s", t))
	return cache_default.New(0), nil
}
