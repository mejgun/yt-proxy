// Package defaultcache implements in-memory links cache
package defaultcache

import (
	"sync"
	"time"

	cache "ytproxy/cache"
	extractor "ytproxy/extractor"
)

// New creates default cache instance
func New(t time.Duration) cache.T {
	return &defaultCache{
		cache:      make(map[extractor.RequestT]extractor.ResultT),
		expireTime: t,
	}
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

func (t *defaultCache) CleanExpired(now time.Time) []extractor.RequestT {
	deleted := make([]extractor.RequestT, 0)
	t.Lock()
	for k, v := range t.cache {
		if v.Expire.Before(now) {
			delete(t.cache, k)
			deleted = append(deleted, k)
		}
	}
	t.Unlock()
	return deleted
}
