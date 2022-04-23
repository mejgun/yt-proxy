package ytproxy

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	config "ytproxy-config"
)

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
		log.Println("Error opening config file. ", err)
		return
	}
	var extractor extractorF
	if len(flags.customdl) > 0 {
		extractor = getCustomDL(flags.customdl)
	} else {
		extractor = getYTDL()
	}
	var requests = make(chan requestChan)
	var links linksCache
	links.cache = make(map[string]lnkT)
	debug := getDebugFunc(flags.enableDebug)
	errorVideo := readErrorVideo(flags.errorVideoPath)
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
