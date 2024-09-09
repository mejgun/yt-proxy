module ytproxy-config

go 1.22

replace (
	ytproxy-extractor => ../extractor
	ytproxy-logger => ../logger
	ytproxy-streamer => ../streamer
)

require (
	ytproxy-extractor v0.0.0-00010101000000-000000000000
	ytproxy-logger v0.0.0-00010101000000-000000000000
	ytproxy-streamer v0.0.0-00010101000000-000000000000
)
