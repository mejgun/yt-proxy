package extractor

import (
	extractor_config "lib/extractor/config"
	logger "lib/logger"
)

func New() (*directExtractor, error) {
	return &directExtractor{}, nil
}

type directExtractor struct {
}

func (t *directExtractor) GetUserAgent(log logger.T) (string, error) {
	return "Mozilla", nil
}

func (t *directExtractor) Extract(req extractor_config.RequestT, log logger.T,
) (extractor_config.ResultT, error) {
	return extractor_config.ResultT{URL: req.URL}, nil
}
