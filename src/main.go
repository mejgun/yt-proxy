package main

import (
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	config "ytproxy-config"
	extractor "ytproxy-extractor"
	linkscache "ytproxy-linkscache"
	logger "ytproxy-logger"
	streamer "ytproxy-streamer"
)

const appVersion = "1.0.0"

const (
	defaultVideoHeight = "720"
	defaultVideoFormat = "mp4"
	defaultExpireTime  = 3 * 60 * 60
)

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
	stderr := func(s string) { os.Stderr.WriteString(fmt.Sprintf("[ ERROR ] %s\n", s)) }
	flags := parseCLIFlags()
	if flags.version {
		stdout(appVersion)
		os.Exit(NoError)
	}
	conf, err := config.Read(flags.config)
	if err != nil {
		stderr("Config file opening error.")
		stderr(err.Error())
		os.Exit(ConfigError)
	}
	log, err := logger.New(conf.Log)
	if err != nil {
		stderr("Logger create error.")
		stderr(err.Error())
		os.Exit(LoggerError)
	}
	extr, err := extractor.New(conf.Extractor)
	if err != nil {
		stderr("Extractor make error.")
		stderr(err.Error())
		os.Exit(ExtractorError)
	}
	cache := linkscache.NewMapCache()
	restreamer, err := streamer.New(conf.Streamer, log)
	if err != nil {
		stderr("Streamer make error.")
		stderr(err.Error())
		os.Exit(StreamerError)
	}

	// errorVideo := readErrorVideo(conf.ErrorVideoPath)
	// sendErrorVideo := getSendErrorVideoFunc(flags.enableErrorHeaders, errorVideo)
	// httpRequest := getDoRequestFunc(flags.ignoreSSLErrors)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.LogInfo("Bad request", r.RemoteAddr, r.RequestURI)
		log.LogDebug("Bad request", r)
		http.NotFound(w, r)
	})
	http.HandleFunc("/play/", func(w http.ResponseWriter, r *http.Request) {
		log.LogInfo("Play request", r.RemoteAddr, r.RequestURI)
		req, res, err := getLink(r.RequestURI, log, cache, extr)
		if err != nil {
			log.LogError("URL extract error", err)
			restreamer.PlayError(w, req, err)
			log.LogInfo("Disconnecting", r.RemoteAddr)
			return
		}
		err = restreamer.Play(w, r, req, res)
		if err != nil {
			log.LogError("Restream error", err)
			restreamer.PlayError(w, req, err)
			log.LogInfo("Disconnecting", r.RemoteAddr)
			return
		}
		log.LogInfo("Player disconnected", r.RemoteAddr)
		// playVideo(w, r, requests, debug, sendErrorVideo, !flags.ignoreMissingHeaders, httpRequest)
	})
	port := fmt.Sprintf("%d", conf.PortInt)
	s := &http.Server{
		Addr: ":" + port,
	}
	s.SetKeepAlivesEnabled(true)
	log.LogInfo("Starting web server at *:%s\n", port)
	err = s.ListenAndServe()
	if err != nil {
		log.LogError("HTTP server start failed: ", err)
		os.Exit(WebServerError)
	}
}

func getLink(query string, log *logger.T, cache linkscache.T, extractor extractor.T) (extractor.RequestT, extractor.ResultT, error) {
	now := time.Now().Unix()
	req := parseQuery(query)
	for _, v := range cache.CleanExpired(now) {
		log.LogDebug("Clean expired cache", v)
	}
	log.LogInfo("Request", req)
	if lnk, ok := cache.Get(req); ok {
		return req, lnk, nil
	}
	res, err := extractor.Extract(req)
	log.LogDebug("Not cached. Extractor returned", res)
	if err != nil {
		return req, res, err
	}
	if res.Expire == 0 {
		res.Expire = now + defaultExpireTime
	}
	cache.Add(req, res)
	log.LogDebug("Cache add", res)
	return req, res, nil
}

func parseQuery(query string) extractor.RequestT {
	var req extractor.RequestT
	query = strings.TrimSpace(strings.TrimPrefix(query, "/play/"))
	splitted := strings.Split(query, "?/?")
	req.URL = splitted[0]
	req.HEIGHT = defaultVideoHeight
	req.FORMAT = defaultVideoFormat
	if len(splitted) != 2 {
		return req
	}
	tOpts, tErr := url.ParseQuery(splitted[1])
	if tErr == nil {
		if tvh, ok := tOpts["vh"]; ok {
			if tvh[0] == "360" || tvh[0] == "480" || tvh[0] == "720" {
				req.HEIGHT = tvh[0]
			}
		}
		if tvf, ok := tOpts["vf"]; ok {
			if tvf[0] == "mp4" || tvf[0] == "m4a" {
				req.FORMAT = tvf[0]
			}
		}
	}
	return req
}
