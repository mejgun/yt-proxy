module ytproxy

go 1.22

replace (
	ytproxy-config => ./config
	ytproxy-extractor => ./extractor
	ytproxy-linkscache => ./cache
	ytproxy-logger => ./logger
	ytproxy-streamer => ./streamer
)

require (
	ytproxy-config v0.0.0-00010101000000-000000000000
	ytproxy-extractor v0.0.0-00010101000000-000000000000
	ytproxy-linkscache v0.0.0-00010101000000-000000000000
)

require (
	ytproxy-logger v0.0.0-00010101000000-000000000000
	ytproxy-streamer v0.0.0-00010101000000-000000000000
)
