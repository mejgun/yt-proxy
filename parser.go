package main

import (
	"fmt"
	"net/url"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

type response struct {
	url string
	err error
}

type rChan struct {
	url string
	c   chan response
}

type lnkT struct {
	url    string
	expire int64
}

type linksT map[string]lnkT

var links linksT

func getLink(vidurl string) (retURL string, retErr error) {
	t := time.Now().Unix()
	vidurl = strings.TrimSpace(vidurl)
	splitted := strings.Split(vidurl, "?/?")
	vidurl = splitted[0]
	var videoFormat string
	// defult values
	vh := "720"
	vf := "mp4"
	if len(splitted) == 2 {
		// videoFormat = "(mp4)[height<=720]"
		tOpts, tErr := url.ParseQuery(splitted[1])
		if tErr == nil {
			if tvh, ok := tOpts["vh"]; ok {
				if tvh[0] == "360" || tvh[0] == "480" || tvh[0] == "720" {
					vh = tvh[0]
				}
			}
			if tvf, ok := tOpts["vf"]; ok {
				if tvf[0] == "mp4" {
					vf = tvf[0]
				}
			}
		}

	}
	videoFormat = "(" + vf + ")[height<=" + vh + "]"
	for k, v := range links {
		if v.expire < t {
			delete(links, k)
		}
	}
	vidurlsize := vidurl + vf + vh
	debugPrint(vidurlsize)
	lnk, ok := links[vidurlsize]
	if ok {
		retURL = lnk.url
		retErr = nil
	} else {
		cmd := exec.Command("youtube-dl", "-f", videoFormat, "-g", vidurl)
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
			links[vidurlsize] = lnkT{url: stdoutStderrStr, expire: expire}
			fmt.Printf("Added url %s (%s %s) expire %d\n", vidurl, vh, vf, expire)
		}
		retURL = stdoutStderrStr
		retErr = err
	}
	return
}

func parseLinks(c <-chan rChan) {
	for {
		r := <-c
		url := r.url
		rURL, rErr := getLink(url)
		debugPrint(rURL)
		debugPrint(rErr)
		r.c <- response{url: rURL, err: rErr}
	}
}
