package cache

import (
	"lib/extractor"
	"time"
)

func New() *emptyCache {
	return &emptyCache{}
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
