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

const corFile = "corrupted.mp4"

var corrupted []byte

var filesize string

func playVideo(w http.ResponseWriter, req *http.Request, requests chan requestChan, debug debugT) {
	var success bool
	success = false
	debug(req)

	url := req.URL.Path[len("/play/"):] + "?"
	url += req.URL.RawQuery
	debug(url)

	qw := make(chan response)
	requests <- requestChan{url: url, answerChan: qw}
	r := <-qw
	debug(r.url)
	debug(r.err)

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
		debug(fmt.Sprintf("%+v\n", res))
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
		w.Header().Set("Content-Length", filesize)
		w.Header().Set("Content-Type", "video/mp4")
		w.Write(corrupted)
	}
	fmt.Printf("%s disconnected\n", req.RemoteAddr)
}

func main() {
	var version bool
	var enableDebug bool
	var portInt int

	flag.BoolVar(&version, "version", false, "prints current yt-proxy version")
	flag.BoolVar(&enableDebug, "debug", false, "turn on debug")
	flag.IntVar(&portInt, "port", 8080, "listen port")
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

	file, err := os.Open(corFile)
	if err != nil {
		log.Fatal(err)
	}
	fileinfo, err := file.Stat()
	if err != nil {
		log.Fatal(err)
	}
	filesize = fmt.Sprint(fileinfo.Size())
	file.Close()

	corrupted, err = ioutil.ReadFile(corFile)

	port := fmt.Sprintf("%d", portInt)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Println(r.RemoteAddr, r.RequestURI)
		http.NotFound(w, r)
	})
	http.HandleFunc("/play/", func(w http.ResponseWriter, r *http.Request) {
		log.Println(r.RemoteAddr, r.RequestURI)
		playVideo(w, r, requests, debug)
	})
	s := &http.Server{
		Addr: ":" + port,
	}
	s.SetKeepAlivesEnabled(true)
	fmt.Printf("starting at *:%s\n", port)
	err = s.ListenAndServe()
	if err != nil {
		log.Fatal("HTTP server start failed: ", err)
	}
}

func parseLinks(requests <-chan requestChan, debug debugT) {
	for {
		r := <-requests
		url := r.url
		rURL, rErr := getLink(url, debug)
		debug(rURL)
		debug(rErr)
		r.answerChan <- response{url: rURL, err: rErr}
	}
}

func getDebugFunc(enabled bool) func(interface{}) {
	return func(s interface{}) {
		if enabled {
			fmt.Printf("DEBUG: %+v\n", s)
		}
	}
}
