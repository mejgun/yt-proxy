package cache

import (
	"time"

	extractor_config "ytproxy/extractor/config"
)

func New() *emptyCache {
	return &emptyCache{}
}

type emptyCache struct{}

func (t *emptyCache) Add(_ extractor_config.RequestT, _ extractor_config.ResultT,
	_ time.Time) {
}

func (t *emptyCache) Get(_ extractor_config.RequestT) (extractor_config.ResultT, bool) {
	return extractor_config.ResultT{}, false
}

func (t *emptyCache) CleanExpired(_ time.Time) []extractor_config.RequestT {
	return []extractor_config.RequestT{}
}
