// Package directextractor implements direct extractor,
// that just returns same url
package directextractor

import (
	extractor "ytproxy/extractor"
	logger "ytproxy/logger"
)

// New creates new direct extractor
func New() (extractor.T, error) {
	return &directExtractor{}, nil
}

type directExtractor struct {
}

func (t *directExtractor) GetUserAgent(_ logger.T) (string, error) {
	return "Mozilla", nil
}

func (t *directExtractor) Extract(req extractor.RequestT, _ logger.T,
) (extractor.ResultT, error) {
	return extractor.ResultT{URL: req.URL}, nil
}
