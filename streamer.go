package main

import (
	"crypto/tls"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
)

const appVersion = "0.6.1"

const defaultVideoHeight = "720"
const defaultVideoFormat = "mp4"
const defaultExpireTime = 3 * 60 * 60
const defaultErrorHeader = "Error-Header-"

func main() {
	flags := parseCLIFlags()
	if flags.version {
		fmt.Println(appVersion)
		os.Exit(0)
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

func playVideo(
	w http.ResponseWriter,
	req *http.Request,
	requests chan requestChan,
	debug debugF,
	sendErrorVideo sendErrorVideoF,
	headersStrictCheck bool,
	httpRequest doRequestF) {
	debug("Request", req)
	debug("Query", req.URL)
	fail := func(str string, err error) {
		log.Println(str, err)
		sendErrorVideo(w, err)
	}
	qw := make(chan response)
	requests <- requestChan{url: req.URL.String(), answerChan: qw}
	r := <-qw
	if r.err != nil {
		fail("URL extractor error:", r.err)
		return
	}
	request, err := http.NewRequest("GET", r.url, nil)
	if err != nil {
		fail("Proxying error", err)
		return
	}
	if r1, ok := req.Header["Range"]; ok {
		request.Header.Set("Range", r1[0])
	}
	request.Header.Set("User-Agent", req.UserAgent())
	res, err := httpRequest(request)
	if err != nil {
		fail("Proxying error", err)
		return
	}
	defer res.Body.Close()
	debug("Response", res)
	h1, ok := res.Header["Content-Length"]
	if !ok && headersStrictCheck {
		fail("Proxying error", errors.New("No Content-Length header"))
		return
	}
	if ok {
		w.Header().Set("Content-Length", h1[0])
	}
	h2, ok := res.Header["Content-Type"]
	if !ok && headersStrictCheck {
		fail("Proxying error", errors.New("No Content-Type header"))
		return
	}
	if headersStrictCheck {
		if h2[0] != "video/mp4" {
			fail("Proxying error", errors.New("Content-Type is not video/mp4"))
			return
		}
	}
	if ok {
		w.Header().Set("Content-Type", h2[0])
	}
	if h3, ok := res.Header["Accept-Ranges"]; ok {
		w.Header().Set("Accept-Ranges", h3[0])
	}
	if h4, ok := res.Header["Content-Range"]; ok {
		w.Header().Set("Content-Range", h4[0])
	}
	if res.StatusCode == 206 {
		w.WriteHeader(http.StatusPartialContent)
	}
	_, err = io.Copy(w, res.Body)
	if err != nil {
		log.Println("Proxy error", err)
	}
	log.Printf("%s disconnected\n", req.RemoteAddr)
}

func parseLinks(requests <-chan requestChan, debug debugF, links *linksCache, extractor extractorF) {
	for {
		r := <-requests
		url := r.url
		rURL, rErr := getLink(url, debug, links, extractor)
		debug("Extractor returned URL", rURL)
		debug("Extractor returned error", rErr)
		r.answerChan <- response{url: rURL, err: rErr}
	}
}

func getDebugFunc(enabled bool) debugF {
	return func(d string, s interface{}) {
		if enabled {
			fmt.Printf("[DEBUG] %s: %+v\n", d, s)
		}
	}
}

func getSendErrorVideoFunc(errorHeaders bool, errorVideo corruptedT) sendErrorVideoF {
	return func(w http.ResponseWriter, err error) {
		w.Header().Set("Content-Length", fmt.Sprintf("%d", errorVideo.size))
		w.Header().Set("Content-Type", "video/mp4")
		if errorHeaders {
			hdrs, errs := errorToHeaders(err)
			for i := range hdrs {
				w.Header().Set(hdrs[i], errs[i])
			}
		}
		_, err = w.Write(errorVideo.file)
		if err != nil {
			log.Println("Cannot send error video", err)
		}
	}
}

func getDoRequestFunc(ignoreSSLErrors bool) doRequestF {
	var tr *http.Transport
	if ignoreSSLErrors {
		tr = &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
	} else {
		tr = &http.Transport{}
	}
	return func(request *http.Request) (*http.Response, error) {
		client := &http.Client{Transport: tr}
		return client.Do(request)
	}
}

func readErrorVideo(path string) corruptedT {
	file, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	fileinfo, err := file.Stat()
	if err != nil {
		log.Fatal(err)
	}
	filesize := fileinfo.Size()
	file.Close()
	corrupted, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatal(err)
	}
	return corruptedT{file: corrupted, size: filesize}
}

func errorToHeaders(e error) ([]string, []string) {
	splitted := strings.Split(e.Error(), "\n")
	filtered := make([]string, 0)
	for _, v := range splitted {
		v := strings.TrimSpace(v)
		if len(v) > 0 {
			filtered = append(filtered, v)
		}
	}
	count := len(fmt.Sprintf("%d", len(filtered)))
	format := fmt.Sprintf("%s%%0%dd", defaultErrorHeader, count+1)
	headers := make([]string, 0)
	for i := range filtered {
		headers = append(headers, fmt.Sprintf(format, i+1))

	}
	return headers, filtered
}

func parseCLIFlags() flagsT {
	var f flagsT
	flag.BoolVar(&f.version, "version", false, "prints current yt-proxy version")
	flag.BoolVar(&f.enableDebug, "debug", false, "turn on debug")
	flag.BoolVar(&f.enableErrorHeaders, "error-headers", false, "show errors in headers (insecure)")
	flag.BoolVar(&f.ignoreMissingHeaders, "ignore-missing-headers", false, "do not strictly check video headers")
	flag.BoolVar(&f.ignoreSSLErrors, "ignore-ssl-errors", false, "do not check video server certificate (insecure)")
	flag.UintVar(&f.portInt, "port", 8080, "listen port")
	flag.StringVar(&f.errorVideoPath, "error-video", "corrupted.mp4", "file that will be shown on errors")
	flag.StringVar(&f.customdl, "custom-extractor", "", "use custom url extractor, will be called like this: program_name url video_height video_format")
	flag.Parse()
	return f
}
