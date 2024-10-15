// Package extractorconfig contains extractor constructor
package extractorconfig

import (
	"strings"

	extractor "ytproxy/extractor"
	extractor_default "ytproxy/extractor/impl/default"
	extractor_direct "ytproxy/extractor/impl/direct"
	logger "ytproxy/logger"
)

const separator = ",,"

// New creates new extractor implementation
func New(c extractor.ConfigT, log logger.T) (extractor.T, error) {
	var (
		ext layer
		err error
	)
	ext.forceHTTP = *c.ForceHTTPS
	if ext.forceHTTP {
		log.LogDebug("", "force-http", true)
	}
	ext.impl, err = realNew(c)
	return &ext, err
}

type layer struct {
	impl      extractor.T
	forceHTTP bool
}

func (t *layer) Extract(req extractor.RequestT, log logger.T) (extractor.ResultT, error) {
	if t.forceHTTP {
		req.URL = "https://" + req.URL
	}
	return t.impl.Extract(req, log)
}

func (t *layer) GetUserAgent(log logger.T) (string, error) {
	return t.impl.GetUserAgent(log)
}

func realNew(c extractor.ConfigT) (extractor.T, error) {
	switch *c.Path {
	case "direct":
		return extractor_direct.New()
	default:
		co := make([]string, 0)
		for _, v := range *c.CustomOptions {
			co = append(co, split(v)...)
		}
		return extractor_default.New(
			*c.Path,
			split(*c.MP4),
			split(*c.M4A),
			*c.GetUserAgent,
			co,
		)
	}
}

func split(s string) []string {
	return strings.Split(s, separator)
}
