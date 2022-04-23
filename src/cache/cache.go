package cache

import (
	"sync"

	extractor "ytproxy-extractor"
)

type CacheT interface {
	Add(extractor.RequestT, extractor.ResultT)
	Get(extractor.RequestT) (extractor.ResultT, bool)
	CleanExpired(int64) []extractor.ResultT
}

type defaultCache struct {
	sync.Mutex
	cache map[string]extractor.ResultT
}

func NewMapCache() CacheT {
	return &defaultCache{}
}

func (t *defaultCache) Add(req extractor.RequestT, res extractor.ResultT) {
	t.Lock()
	t.cache[req.URL] = res
	t.Unlock()
}

func (t *defaultCache) Get(req extractor.RequestT) (extractor.ResultT, bool) {
	t.Lock()
	defer t.Unlock()
	v, ok := t.cache[req.URL]
	return v, ok
}

func (t *defaultCache) CleanExpired(now int64) []extractor.ResultT {
	deleted := make([]extractor.ResultT, 0)
	t.Lock()
	for k, v := range t.cache {
		if v.Expire < now {
			delete(t.cache, k)
			deleted = append(deleted, v)
		}
	}
	t.Unlock()
	return deleted
}
