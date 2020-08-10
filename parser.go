package main

import (
	"log"
	"net/url"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"
)

type response struct {
	url string
	err error
}

type requestChan struct {
	url        string
	answerChan chan response
}

type lnkT struct {
	url    string
	expire int64
}

type linksCache struct {
	sync.RWMutex
	cache map[string]lnkT
}

type debugT func(string, interface{})

func youtubeDL(vURL, vHeight, vFormat string, debug debugT) (string, int64, error) {
	// videoFormat = "(mp4)[height<=720]"
	videoFormat := "(" + vFormat + ")[height<=" + vHeight + "]"
	cmd := exec.Command("youtube-dl", "-f", videoFormat, "-g", vURL)
	// stdoutStderr, err := cmd.CombinedOutput()
	stdoutStderr, err := cmd.Output()
	stdoutStderrStr := strings.TrimSpace(string(stdoutStderr))
	var expire int64
	if err == nil {
		// err rewrited
		u, err := url.Parse(stdoutStderrStr)
		if err == nil {
			m, _ := url.ParseQuery(u.RawQuery)
			if e, ok := m["expire"]; ok {
				e1, err1 := strconv.ParseInt(e[0], 10, 64)
				if err1 == nil {
					expire = e1
				} else {
					expire = 0
				}
			} else {
				expire = 0
			}
		}
	}
	return stdoutStderrStr, expire, err
}

func getLink(vidurl string, debug debugT, links *linksCache) (string, error) {
	now := time.Now().Unix()
	vidurl = strings.TrimSpace(vidurl)
	splitted := strings.Split(vidurl, "?/?")
	vidurl = splitted[0]
	vh := defaultVideoHeight
	vf := defaultVideoFormat
	if len(splitted) == 2 {
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
	links.Lock()
	for k, v := range links.cache {
		if v.expire < now {
			debug("Delete expired cached url", v.url)
			delete(links.cache, k)
		}
	}
	links.Unlock()
	links.RLock()
	vidurlsize := vidurl + vf + vh
	debug("Video URL", vidurl)
	debug("Video format", vf)
	debug("Video height", vh)
	lnk, ok := links.cache[vidurlsize]
	links.RUnlock()
	if ok {
		return lnk.url, nil
	}
	url, expire, err := youtubeDL(vidurl, vh, vf, debug)
	if expire == 0 {
		expire = now + 10800
	}
	if err == nil {
		links.Lock()
		links.cache[vidurlsize] = lnkT{url: url, expire: expire}
		links.Unlock()
		log.Printf("Caching url %s (%s %s). Expire in %dmin\n", vidurl, vh, vf, (expire-now)/60)
		return url, nil
	}
	return "", err
}
