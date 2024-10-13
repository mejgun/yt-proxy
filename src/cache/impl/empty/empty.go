package cache

import (
	"time"

	extractor_config "ytproxy/extractor/config"
)

func New() *emptyCache {
	return &emptyCache{}
}

type emptyCache struct{}

func (t *emptyCache) Add(req extractor_config.RequestT, res extractor_config.ResultT,
	now time.Time) {
}

func (t *emptyCache) Get(req extractor_config.RequestT) (extractor_config.ResultT, bool) {
	return extractor_config.ResultT{}, false
}

func (t *emptyCache) CleanExpired(now time.Time) []extractor_config.RequestT {
	return []extractor_config.RequestT{}
}
