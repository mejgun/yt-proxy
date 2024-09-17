package cache

import (
	"lib/extractor"
	"sync"
	"time"
)

func New(t time.Duration) *defaultCache {
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
