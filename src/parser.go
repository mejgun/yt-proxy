package ytproxy

import (
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type response struct {
	url string
	err error
}

// type requestChan struct {
// 	url        string
// 	answerChan chan response
// }

type lnkT struct {
	url    string
	expire int64
}

// type debugF func(string, interface{})

type corruptedT struct {
	file []byte
	size int64
}

type sendErrorVideoF func(http.ResponseWriter, error)

type doRequestF func(*http.Request) (*http.Response, error)

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
