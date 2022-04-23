module ytproxy

go 1.17

replace (
	ytproxy-config => ./config
	ytproxy-logger => ./logger
)

require ytproxy-config v0.0.0-00010101000000-000000000000

require ytproxy-logger v0.0.0-00010101000000-000000000000 // indirect
