package cache

import (
	"sync"

	extractor "ytproxy-extractor"
)

type T interface {
	Add(extractor.RequestT, extractor.ResultT)
	Get(extractor.RequestT) (extractor.ResultT, bool)
	CleanExpired(int64) []extractor.ResultT
}

type ConfigT struct {
	ExpireTime *int64 `json:"expire-time"`
	Disable    bool   `json:"disable"`
}

type defaultCache struct {
	sync.Mutex
	cache map[extractor.RequestT]extractor.ResultT
}

func NewMapCache() T {
	return &defaultCache{cache: make(map[extractor.RequestT]extractor.ResultT)}
}

func (t *defaultCache) Add(req extractor.RequestT, res extractor.ResultT) {
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
