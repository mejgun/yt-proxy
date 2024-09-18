package extractor

import (
	extractor_config "lib/extractor/config"
)

func New() (*directExtractor, error) {
	return &directExtractor{}, nil
}

type directExtractor struct {
}

func (t *directExtractor) GetUserAgent() (string, error) {
	return "Mozilla", nil
}

func (t *directExtractor) Extract(req extractor_config.RequestT,
) (extractor_config.ResultT, error) {
	return extractor_config.ResultT{URL: req.URL}, nil
}
