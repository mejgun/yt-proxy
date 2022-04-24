package ytproxy

import (
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
	streamer "ytproxy-streamer"

	config "ytproxy-config"
	extractor "ytproxy-extractor"
	linkscache "ytproxy-linkscache"
	logger "ytproxy-logger"
)

const appVersion = "1.0.0"

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
)

func parseCLIFlags() flagsT {
	var f flagsT
	flag.BoolVar(&f.version, "version", false, "prints current yt-proxy version")
	flag.StringVar(&f.config, "config", "config.json", "config file path")
	flag.Parse()
	return f
}

func main() {
	flags := parseCLIFlags()
	if flags.version {
		fmt.Println(appVersion)
		os.Exit(NoError)
	}
	conf, err := config.Read(flags.config)
	if err != nil {
		os.Stderr.WriteString("Config file opening error. ")
		os.Stderr.WriteString(err.Error())
		os.Exit(ConfigError)
	}
	log, err := logger.New(conf.Log)
	if err != nil {
		os.Stderr.WriteString("Logger create error. ")
		os.Stderr.WriteString(err.Error())
		os.Exit(LoggerError)
	}
	extr, err := extractor.New(conf.Extractor)
	if err != nil {
		log.LogError("Extractor make", err)
		os.Exit(ExtractorError)
	}
	cache := linkscache.NewMapCache()
	s, err := streamer.New(conf.Streamer)
	if err != nil {
		log.LogError("Streamer make", err)
		os.Exit(StreamerError)
	}

	errorVideo := readErrorVideo(conf.ErrorVideoPath)
	sendErrorVideo := getSendErrorVideoFunc(flags.enableErrorHeaders, errorVideo)
	httpRequest := getDoRequestFunc(flags.ignoreSSLErrors)
	port := fmt.Sprintf("%d", conf.PortInt)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.LogInfo("Bad request", r.RemoteAddr, r.RequestURI)
		log.LogDebug("Bad request", r)
		http.NotFound(w, r)
	})
	http.HandleFunc("/play/", func(w http.ResponseWriter, r *http.Request) {
		log.LogInfo("Play request", r.RemoteAddr, r.RequestURI)
		playVideo(w, r, requests, debug, sendErrorVideo, !flags.ignoreMissingHeaders, httpRequest)
	})
	s := &http.Server{
		Addr: ":" + port,
	}
	s.SetKeepAlivesEnabled(true)
	fmt.Printf("starting at *:%s\n", port)
	err := s.ListenAndServe()
	if err != nil {
		log.Fatal("HTTP server start failed: ", err)
	}
}

func getLink(query string, log *logger.T, cache linkscache.T, extractor extractor.T) (extractor.ResultT, error) {
	now := time.Now().Unix()
	req := parseQuery(query)
	for _, v := range cache.CleanExpired(now) {
		log.LogDebug("Clean expired cache", v)
	}
	log.LogInfo("Request", req)
	if lnk, ok := cache.Get(req); ok {
		return lnk, nil
	}
	res, err := extractor.Extract(req)
	log.LogDebug("Not cached. Extractor returned", res)
	if err != nil {
		return res, err
	}
	if res.Expire == 0 {
		res.Expire = now + defaultExpireTime
	}
	cache.Add(req, res)
	log.LogDebug("Cache add", res)
	return res, nil
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
