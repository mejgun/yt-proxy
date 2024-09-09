module ytproxy-streamer

go 1.22

replace (
	ytproxy-extractor => ../extractor
	ytproxy-logger => ../logger
)

require (
	ytproxy-extractor v0.0.0-00010101000000-000000000000
	ytproxy-logger v0.0.0-00010101000000-000000000000
)
