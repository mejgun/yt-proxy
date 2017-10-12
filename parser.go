package main

import (
	"fmt"
	//"log"
	"net/url"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

type Response struct {
	url string
	err error
}

type RChan struct {
	url string
	c   chan Response
}

type Lnk struct {
	url    string
	expire int64
}

type Links map[string]Lnk

var links Links

func getLink(vidurl string) (retUrl string, retErr error) {
	t := time.Now().Unix()
	vidurl = strings.TrimSpace(vidurl)
	lnk, ok := links[vidurl]
	if ok {
		if lnk.expire > t {
			retUrl = lnk.url
			retErr = nil
		} else {
			delete(links, vidurl)
		}
	} else {
		cmd := exec.Command("youtube-dl", "-g", "-f mp4", vidurl)
		// stdoutStderr, err := cmd.CombinedOutput()
		stdoutStderr, err := cmd.Output()
		stdoutStderrStr := strings.TrimSpace(string(stdoutStderr))
		if err == nil {
			var expire int64
			// err rewrited
			u, err := url.Parse(stdoutStderrStr)
			if err == nil {
				m, _ := url.ParseQuery(u.RawQuery)
				if e, ok := m["expire"]; ok {
					e1, err1 := strconv.ParseInt(e[0], 10, 64)
					if err1 == nil {
						expire = e1
					} else {
						expire = t + 10800
					}
				} else {
					expire = t + 10800
				}
			}
			links[vidurl] = Lnk{url: stdoutStderrStr, expire: expire}
			fmt.Printf("Added url %s expire %d\n", vidurl, expire)
		}
		retUrl = stdoutStderrStr
		retErr = err
	}
	return
}

func parseLinks(c <-chan RChan) {
	for {
		r := <-c
		url := r.url
		rUrl, rErr := getLink(url)
		//fmt.Println(rUrl)
		//fmt.Println(rErr)
		r.c <- Response{url: rUrl, err: rErr}
	}
}

//
//
// func main() {
// 	c := make(chan RChan)
// 	/*c := make(chan int)
// 	go func() { c <- 5 }()
// 	b := <-c
// 	close(c)
// 	fmt.Println(b)
// 	log.Fatal("test")*/
// 	go parseLinks(c)
// 	links = make(Links)
// 	for i := 0; i < 5; i++ {
// 		url := "https://www.youtube.com/watch?v=H0hsrEhz3zo"
// 		qw := make(chan Response)
// 		c <- RChan{url: url, c: qw}
// 		r := <-qw
// 		fmt.Println(r)
// 		/*
// 			//url := "https://vimeo.com/235305664"
// 			stdoutStderr, err := getLink(url)
// 			fmt.Printf("%s\n", stdoutStderr)
// 			if err != nil {
// 				log.Fatal(err)
// 			}
// 		*/
// 		/*fmt.Println(links)
// 		fmt.Println(time.Now().Unix())
// 		fmt.Println(time.Now().Unix() + 10800)*/
// 	}
// 	log.Fatal("ok")
// }
//
//
