package main

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

const appVersion = "0.5"

type corruptedT struct {
	file []byte
	size int64
}

func playVideo(w http.ResponseWriter, req *http.Request, requests chan requestChan, debug debugT, errorVideo corruptedT) {
	var success bool
	success = false
	debug("Request", req)

	url := req.URL.Path[len("/play/"):] + "?"
	url += req.URL.RawQuery
	debug("Query", url)

	qw := make(chan response)
	requests <- requestChan{url: url, answerChan: qw}
	r := <-qw

	if r.err == nil {
		request, _ := http.NewRequest("GET", r.url, nil)
		r1, ok := req.Header["Range"]
		if ok {
			request.Header.Set("Range", r1[0])
		}
		request.Header.Set("User-Agent", req.UserAgent())
		tr := &http.Transport{}
		client := &http.Client{Transport: tr}
		res, err := client.Do(request)
		if err != nil {
			log.Println(err)
		}
		defer res.Body.Close()
		debug("Response", fmt.Sprintf("%+v\n", res))
		h1, ok1 := res.Header["Content-Length"]
		h2, ok2 := res.Header["Content-Type"]

		if ok1 && ok2 {
			if h2[0] == "video/mp4" {
				w.Header().Set("Content-Length", h1[0])
				w.Header().Set("Content-Type", h2[0])
				h3, ok := res.Header["Accept-Ranges"]
				if ok {
					w.Header().Set("Accept-Ranges", h3[0])
				}
				h4, ok := res.Header["Content-Range"]
				if ok {
					w.Header().Set("Content-Range", h4[0])
				}
				if res.StatusCode == 206 {
					w.WriteHeader(http.StatusPartialContent)
				}
				io.Copy(w, res.Body)
				success = true
			}
		}
	} else {
		log.Println("yotube-dl error:", r.err)
	}

	if success == false {
		w.Header().Set("Content-Length", fmt.Sprintf("%d", errorVideo.size))
		w.Header().Set("Content-Type", "video/mp4")
		w.Write(errorVideo.file)
	}
	fmt.Printf("%s disconnected\n", req.RemoteAddr)
}

func main() {
	var version bool
	var enableDebug bool
	var portInt uint
	var errorVideoPath string

	flag.BoolVar(&version, "version", false, "prints current yt-proxy version")
	flag.BoolVar(&enableDebug, "debug", false, "turn on debug")
	flag.UintVar(&portInt, "port", 8080, "listen port")
	flag.StringVar(&errorVideoPath, "error-video", "corrupted.mp4", "file that will be shown on errors")
	flag.Parse()
	if version {
		fmt.Println(appVersion)
		os.Exit(0)
	}
	var requests chan requestChan
	requests = make(chan requestChan)
	links = make(linksT)
	debug := getDebugFunc(enableDebug)
	go parseLinks(requests, debug)
	errorVideo := readErrorVideo(errorVideoPath)

	port := fmt.Sprintf("%d", portInt)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Println(r.RemoteAddr, r.RequestURI)
		http.NotFound(w, r)
	})
	http.HandleFunc("/play/", func(w http.ResponseWriter, r *http.Request) {
		log.Println(r.RemoteAddr, r.RequestURI)
		playVideo(w, r, requests, debug, errorVideo)
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

func parseLinks(requests <-chan requestChan, debug debugT) {
	for {
		r := <-requests
		url := r.url
		rURL, rErr := getLink(url, debug)
		debug("Extractor returned URL", rURL)
		debug("Extractor returned error", rErr)
		r.answerChan <- response{url: rURL, err: rErr}
	}
}

func getDebugFunc(enabled bool) debugT {
	return func(d string, s interface{}) {
		if enabled {
			fmt.Printf("[DEBUG] %s: %+v\n", d, s)
		}
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
