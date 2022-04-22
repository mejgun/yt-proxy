package main

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"net/http"
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

type debugF func(string, interface{})

type corruptedT struct {
	file []byte
	size int64
}

type flagsT struct {
	version              bool
	enableDebug          bool
	enableErrorHeaders   bool
	ignoreMissingHeaders bool
	ignoreSSLErrors      bool
	portInt              uint
	errorVideoPath       string
	customdl             string
}

type extractorF func(string, string, string, debugF) (string, int64, error)

type sendErrorVideoF func(http.ResponseWriter, error)

type doRequestF func(*http.Request) (*http.Response, error)

func getYTDL() extractorF {
	return func(vURL, vHeight, vFormat string, debug debugF) (string, int64, error) {
		// videoFormat = "(mp4)[height<=720]"
		var videoFormat string
		switch vFormat {
		case "m4a":
			videoFormat = "(m4a)"
		case "mp4":
			fallthrough
		default:
			videoFormat = "(mp4)[height<=" + vHeight + "]"
		}
		cmd := exec.Command("youtube-dl", "-f", videoFormat, "-g", vURL)
		out, err := runCmd(cmd)
		if err != nil {
			return "", 0, err
		}
		var expire int64
		u, err := url.Parse(out)
		if err == nil {
			m, _ := url.ParseQuery(u.RawQuery)
			if e, ok := m["expire"]; ok {
				e1, err1 := strconv.ParseInt(e[0], 10, 64)
				if err1 == nil {
					expire = e1
				}
			}
		}
		return out, expire, err
	}
}

func getCustomDL(path string) extractorF {
	return func(vURL, vHeight, vFormat string, debug debugF) (string, int64, error) {
		cmd := exec.Command(path, vURL, vHeight, vFormat)
		out, err := runCmd(cmd)
		return out, 0, err
	}
}

func runCmd(cmd *exec.Cmd) (string, error) {
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	toS := func(s bytes.Buffer) string {
		return strings.TrimSpace(s.String())
	}
	outStr, errStr := toS(stdout), toS(stderr)
	if err != nil {
		combinedErrStr := fmt.Sprintf("%s\n%s\n%s", err.Error(), outStr, errStr)
		return "", errors.New(combinedErrStr)
	}
	return outStr, nil
}

func getLink(query string, debug debugF, links *linksCache, extractor extractorF) (string, error) {
	now := time.Now().Unix()
	vidurl, vh, vf := parseQuery(query)
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
	url, expire, err := extractor(vidurl, vh, vf, debug)
	if expire == 0 {
		expire = now + defaultExpireTime
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

func parseQuery(query string) (vURL, vh, vf string) {
	query = strings.TrimSpace(strings.TrimPrefix(query, "/play/"))
	splitted := strings.Split(query, "?/?")
	vURL = splitted[0]
	vh = defaultVideoHeight
	vf = defaultVideoFormat
	if len(splitted) != 2 {
		return
	}
	tOpts, tErr := url.ParseQuery(splitted[1])
	if tErr == nil {
		if tvh, ok := tOpts["vh"]; ok {
			if tvh[0] == "360" || tvh[0] == "480" || tvh[0] == "720" {
				vh = tvh[0]
			}
		}
		if tvf, ok := tOpts["vf"]; ok {
			if tvf[0] == "mp4" || tvf[0] == "m4a" {
				vf = tvf[0]
			}
		}
	}
	return
}
