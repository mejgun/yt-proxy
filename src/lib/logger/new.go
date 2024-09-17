package logger

import (
	config "lib/logger/config"
	def "lib/logger/impl/default"
)

func New(conf config.ConfigT) (T, error) {
	return def.New(conf)
}
