package extractor

import (
	"strings"

	extractor_config "lib/extractor/config"
	extractor_default "lib/extractor/impl/default"
	logger "lib/logger"
)

const separator = ",,"

type T interface {
	Extract(extractor_config.RequestT) (extractor_config.ResultT, error)
	GetUserAgent() (string, error)
}

func New(c extractor_config.ConfigT, log logger.T) (T, error) {
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

func split(s string) []string {
	return strings.Split(s, separator)
}
