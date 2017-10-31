package main

import (
	"fmt"
	"io"
	//"bytes"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

// const debug = true
const debug = false

const corFile = "corrupted.mp4"

var c chan RChan

var corrupted []byte

var filesize string

func PlayVideo(w http.ResponseWriter, req *http.Request) {
	//fmt.Println(req.UserAgent())
	//fmt.Println(req.Cookies())
	var success bool
	success = false
	if debug {
		fmt.Println(req.Write(os.Stdout))
	}
	url := req.URL.Path[len("/play/"):] + "?"
	url += req.URL.RawQuery
	if debug {
		fmt.Println(url)
	}
	qw := make(chan Response)
	c <- RChan{url: url, c: qw}
	r := <-qw
	if debug {
		fmt.Println(r.url)
		fmt.Println(r.err)
	}
	if r.err == nil {
		request, _ := http.NewRequest("GET", r.url, nil)
		r1, ok := req.Header["Range"]
		if ok {
			request.Header.Set("Range", r1[0])
		}
		request.Header.Set("User-Agent", req.UserAgent())
		tr := &http.Transport{}
		client := &http.Client{Transport: tr}
		// fmt.Println(request)
		res, err := client.Do(request)
		if err != nil {
			log.Println(err)
		}
		defer res.Body.Close()
		if debug {
			fmt.Printf("%+v\n", res)
		}
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
		log.Println("yotube-dl error:",r.err)
	}
	// w.Header().Set("Close", "true")
	//	fmt.Printf("%v\n", res.Header)
	//io.Copy(ioutil.Discard, res.Body)
	/*for _, err := io.CopyN(w, res.Body, 640000); err == nil; {
		//fmt.Print(".")
	}*/
	// res.Body.Close()

	//w.Close()
	if success == false {
		w.Header().Set("Content-Length", filesize)
		w.Header().Set("Content-Type", "video/mp4")
		w.Write(corrupted)
	}
	fmt.Printf("%s disconnected\n", req.RemoteAddr)
}

func makeHandler(fn func(http.ResponseWriter, *http.Request)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println(r.RemoteAddr, r.RequestURI)
		fn(w, r)
	}
}

func main() {
	c = make(chan RChan)
	links = make(Links)
	go parseLinks(c)

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

	var port string
	if len(os.Args) == 2 {
		port = os.Args[1]
	} else {
		port = "8080"
	}
	http.HandleFunc("/", makeHandler(http.NotFound))
	http.HandleFunc("/play/", makeHandler(PlayVideo))
	s := &http.Server{
		Addr: ":" + port,
	}
	s.SetKeepAlivesEnabled(true)
	fmt.Printf("starting at *:%s\n", port)
	err = s.ListenAndServe()
	if err != nil {
		log.Fatal("HTTP server start failed: ", err)
	}
	//	http.ListenAndServe(":8181", nil)
}
