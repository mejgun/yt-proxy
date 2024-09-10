package cache

import (
	"sync"
	"time"

	extractor "ytproxy-extractor"
)

type T interface {
	Add(extractor.RequestT, extractor.ResultT, time.Time)
	Get(extractor.RequestT) (extractor.ResultT, bool)
	CleanExpired(time.Time) []extractor.ResultT
}

type ConfigT struct {
	ExpireTime *string `json:"expire-time"`
}

const defaultExpireTimeInSeconds = 10800

func New(conf ConfigT) (T, error) {
	switch {
	case conf.ExpireTime == nil:
		return T{} //
	default:
		t, err := time.ParseDuration(*conf.ExpireTime)
		c := defaultCache{
			cache:      make(map[extractor.RequestT]extractor.ResultT),
			expireTime: t,
		}
		return &c, err
	}
}

type defaultCache struct {
	sync.Mutex
	cache      map[extractor.RequestT]extractor.ResultT
	expireTime time.Duration
}

var timeToInt64 = func(t time.Time) int64 { return t.Unix() }

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
