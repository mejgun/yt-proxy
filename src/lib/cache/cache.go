package cache

import (
	"fmt"
	"sync"
	"time"

	extractor "lib/extractor"
	logger "lib/logger"
)

type T interface {
	Add(extractor.RequestT, extractor.ResultT, time.Time)
	Get(extractor.RequestT) (extractor.ResultT, bool)
	CleanExpired(time.Time) []extractor.ResultT
}

type ConfigT struct {
	ExpireTime *string `json:"expire-time"`
}

func New(conf ConfigT, log *logger.T) (T, error) {
	defCache := func(t time.Duration) *defaultCache {
		return &defaultCache{
			cache:      make(map[extractor.RequestT]extractor.ResultT),
			expireTime: t,
		}
	}
	t, err := time.ParseDuration(*conf.ExpireTime)
	if err != nil {
		return &defaultCache{}, err
	}
	if t.Seconds() < 1 {
		log.LogDebug("cache", "disabled by config")
		return &emptyCache{}, nil
	}
	log.LogDebug("cache", fmt.Sprintf("expire time set to %s", t))
	return defCache(t), nil
}

type defaultCache struct {
	sync.Mutex
	cache      map[extractor.RequestT]extractor.ResultT
	expireTime time.Duration
}

func (t *defaultCache) Add(req extractor.RequestT, res extractor.ResultT,
	now time.Time) {
	res.Expire = now.Add(t.expireTime)
	t.Lock()
	t.cache[req] = res
	t.Unlock()
}

func (t *defaultCache) Get(req extractor.RequestT) (extractor.ResultT, bool) {
	t.Lock()
	defer t.Unlock()
	v, ok := t.cache[req]
	return v, ok
}

func (t *defaultCache) CleanExpired(now time.Time) []extractor.ResultT {
	deleted := make([]extractor.ResultT, 0)
	t.Lock()
	for k, v := range t.cache {
		if v.Expire.Before(now) {
			delete(t.cache, k)
			deleted = append(deleted, v)
		}
	}
	t.Unlock()
	return deleted
}

type emptyCache struct{}

func (t *emptyCache) Add(req extractor.RequestT, res extractor.ResultT,
	now time.Time) {
}

func (t *emptyCache) Get(req extractor.RequestT) (extractor.ResultT, bool) {
	return extractor.ResultT{}, false
}

func (t *emptyCache) CleanExpired(now time.Time) []extractor.ResultT {
	return []extractor.ResultT{}
}
