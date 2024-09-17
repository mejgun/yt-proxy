package cache

import (
	"sync"
	"time"

	extractor_config "lib/extractor/config"
)

func New(t time.Duration) *defaultCache {
	return &defaultCache{
		cache:      make(map[extractor_config.RequestT]extractor_config.ResultT),
		expireTime: t,
	}
}

type defaultCache struct {
	sync.Mutex
	cache      map[extractor_config.RequestT]extractor_config.ResultT
	expireTime time.Duration
}

func (t *defaultCache) Add(req extractor_config.RequestT, res extractor_config.ResultT,
	now time.Time) {
	res.Expire = now.Add(t.expireTime)
	t.Lock()
	t.cache[req] = res
	t.Unlock()
}

func (t *defaultCache) Get(req extractor_config.RequestT) (extractor_config.ResultT, bool) {
	t.Lock()
	defer t.Unlock()
	v, ok := t.cache[req]
	return v, ok
}

func (t *defaultCache) CleanExpired(now time.Time) []extractor_config.ResultT {
	deleted := make([]extractor_config.ResultT, 0)
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
