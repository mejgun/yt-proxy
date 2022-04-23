package ytproxy

import (
	"flag"
	"fmt"
	"net/http"
	"os"

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
		os.Exit(0)
	}
	conf, err := config.Read(flags.config)
	if err != nil {
		os.Stderr.WriteString("Config file opening error. ")
		os.Stderr.WriteString(err.Error())
		os.Exit(1)
	}
	log, err := logger.New(conf.Log)
	if err != nil {
		os.Stderr.WriteString("Logger create error. ")
		os.Stderr.WriteString(err.Error())
		os.Exit(2)
	}
	extr, err := extractor.New(conf.Extractor)
	if err != nil {
		log.LogError("Extractor make", err)
		os.Exit(3)
	}
	cache := linkscache.NewMapCache()

	var requests = make(chan requestChan)
	errorVideo := readErrorVideo(conf.ErrorVideoPath)
	sendErrorVideo := getSendErrorVideoFunc(flags.enableErrorHeaders, errorVideo)
	httpRequest := getDoRequestFunc(flags.ignoreSSLErrors)
	go parseLinks(requests, debug, &links, extractor)
	port := fmt.Sprintf("%d", flags.portInt)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Println(r.RemoteAddr, r.RequestURI)
		http.NotFound(w, r)
	})
	http.HandleFunc("/play/", func(w http.ResponseWriter, r *http.Request) {
		log.Println(r.RemoteAddr, r.RequestURI)
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
