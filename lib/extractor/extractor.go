package extractor

import (
	"strings"

	extractor_config "lib/extractor/config"
	extractor_default "lib/extractor/impl/default"
	extractor_direct "lib/extractor/impl/direct"
	logger "lib/logger"
)

const separator = ",,"

type T interface {
	Extract(extractor_config.RequestT) (extractor_config.ResultT, error)
	GetUserAgent() (string, error)
}

func New(c extractor_config.ConfigT, log logger.T) (T, error) {
	var (
		ext layer
		err error
	)
	ext.force_http = *c.ForceHttps
	if ext.force_http {
		log.LogDebug("", "force-http", true)
	}
	ext.impl, err = real_new(c, log)
	return &ext, err
}

type layer struct {
	impl       T
	force_http bool
}

func (t *layer) Extract(req extractor_config.RequestT) (extractor_config.ResultT, error) {
	if t.force_http {
		req.URL = "https://" + req.URL
	}
	return t.impl.Extract(req)
}

func (t *layer) GetUserAgent() (string, error) {
	return t.impl.GetUserAgent()
}

func real_new(c extractor_config.ConfigT, log logger.T) (T, error) {
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
			log,
		)
	}
}

func split(s string) []string {
	return strings.Split(s, separator)
}
