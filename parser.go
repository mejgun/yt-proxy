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
	for k, v := range links {
		if v.expire < t {
			delete(links, k)
		}
	}
	lnk, ok := links[vidurl]
	if ok {
		retUrl = lnk.url
		retErr = nil
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
