package cache

import (
	"fmt"
	"time"

	cache "ytproxy/cache"
	cache_default "ytproxy/cache/impl/default"
	cache_empty "ytproxy/cache/impl/empty"
	logger "ytproxy/logger"
)

func New(conf cache.ConfigT, log logger.T) (cache.T, error) {
	t, err := time.ParseDuration(*conf.ExpireTime)
	if err != nil {
		return cache_default.New(0), err
	}
	if t.Seconds() < 1 {
		log.LogDebug("", "disabled by config")
		return cache_empty.New(), nil
	}
	log.LogDebug("", fmt.Sprintf("expire time set to %s", t))
	return cache_default.New(t), nil
}
