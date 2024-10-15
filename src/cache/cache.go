// Package cache contains cache interface and config
package cache

import (
	"time"

	extractor "ytproxy/extractor"
)

// T is cache interface
type T interface {
	Add(extractor.RequestT, extractor.ResultT, time.Time)
	Get(extractor.RequestT) (extractor.ResultT, bool)
	CleanExpired(time.Time) []extractor.RequestT
}

// ConfigT is constructor config
type ConfigT struct {
	ExpireTime *string `json:"expire-time"`
}
