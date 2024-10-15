// Package emptycache implements dummy cache
package emptycache

import (
	"time"

	cache "ytproxy/cache"
	extractor "ytproxy/extractor"
)

// New creates dummy cache instance
func New() cache.T {
	return &emptyCache{}
}

type emptyCache struct{}

func (t *emptyCache) Add(_ extractor.RequestT, _ extractor.ResultT,
	_ time.Time) {
}

func (t *emptyCache) Get(_ extractor.RequestT) (extractor.ResultT, bool) {
	return extractor.ResultT{}, false
}

func (t *emptyCache) CleanExpired(_ time.Time) []extractor.RequestT {
	return []extractor.RequestT{}
}
