package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"

	app "lib/app"
	cache "lib/cache"
	config "lib/config"
	extractor "lib/extractor"
	logger "lib/logger"
	streamer "lib/streamer"
)

const appVersion = "2.0.0"

type flagsT struct {
	version bool
	config  string
}

const (
	NoError = iota
	ConfigError
	LoggerError
	ExtractorError
	StreamerError
	WebServerError
	CacheError
)

func parseCLIFlags() flagsT {
	var f flagsT
	flag.BoolVar(&f.version, "version", false, "prints current yt-proxy version")
	flag.StringVar(&f.config, "config", "config.json", "config file path")
	flag.Parse()
	return f
}

func main() {
	stdout := func(s string) { os.Stdout.WriteString(fmt.Sprintf("%s\n", s)) }
	stderr := func(s string) { os.Stderr.WriteString(fmt.Sprintf("ERROR   %s\n", s)) }
	flags := parseCLIFlags()
	if flags.version {
		stdout(appVersion)
		os.Exit(NoError)
	}
	checkOrExit := func(err error, name string, errorcode int) {
		if err != nil {
			stderr(fmt.Sprintf("%s create error.", name))
			stderr(err.Error())
			os.Exit(errorcode)
		}
	}
	conf, err := config.Read(flags.config)
	checkOrExit(err, "Config", ConfigError)

	log, err := logger.New(conf.Log)
	checkOrExit(err, "Logger", LoggerError)
	status := func(s string) {
		log.LogDebug("App starting", "status", s)
	}
	status("logger created")

	extr, err := extractor.New(conf.Extractor, log)
	checkOrExit(err, "Extractor", ExtractorError)
	status("extractor created")

	cache, err := cache.New(conf.Cache, log)
	checkOrExit(err, "Cache", CacheError)
	status("cache created")

	restreamer, err := streamer.New(conf.Streamer, log, extr)
	checkOrExit(err, "Streamer", StreamerError)
	status("streamer created")

	app := app.New(log, cache, extr, restreamer)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.LogInfo("Bad request", r.RemoteAddr, r.RequestURI)
		log.LogDebug("Bad request", r)
		http.NotFound(w, r)
	})
	http.HandleFunc("/play/", func(w http.ResponseWriter, r *http.Request) {
		log.LogInfo("Play request", r.RemoteAddr, r.RequestURI)
		log.LogDebug("User request", r)
		app.Run(w, r)
		log.LogInfo("Player disconnected", r.RemoteAddr)
	})
	port := fmt.Sprintf("%d", conf.PortInt)
	s := &http.Server{
		Addr: ":" + port,
	}
	log.LogInfo("Starting web server", "port", port, "test")
	err = s.ListenAndServe()
	if err != nil {
		log.LogError("HTTP server start failed: ", err)
		os.Exit(WebServerError)
	}
}
