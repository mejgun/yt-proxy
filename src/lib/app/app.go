package app

import (
	cache "lib/cache"
	extractor "lib/extractor"
	extractor_config "lib/extractor/config"
	logger "lib/logger"
	streamer "lib/streamer"

	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	defaultVideoHeight = "720"
	defaultVideoFormat = "mp4"
)

type T struct {
	log       logger.T
	cache     cache.T
	extractor extractor.T
	streamer  streamer.T
}

func New(
	l logger.T,
	c cache.T,
	x extractor.T,
	s streamer.T) *T {
	return &T{
		log:       l,
		cache:     c,
		extractor: x,
		streamer:  s,
	}
}

func (t *T) Run(w http.ResponseWriter, r *http.Request) {
	now := time.Now()
	req := parseQuery(r.RequestURI)
	t.log.LogInfo("Request", req)
	if res, ok := t.cacheCheck(req, now); ok {
		t.log.LogDebug("Link already cached", res)
		t.play(w, r, req, res)
	} else {
		res, err := t.extractor.Extract(req)
		if err != nil {
			t.log.LogError("URL extract error", err)
			t.playError(w, req, err)
		}
		t.log.LogDebug("Extractor returned", res)
		t.cacheAdd(req, res, now)
		t.play(w, r, req, res)
	}
}

func (t *T) play(
	w http.ResponseWriter,
	r *http.Request,
	req extractor_config.RequestT,
	res extractor_config.ResultT,
) {
	if err := t.streamer.Play(w, r, req, res); err != nil {
		t.log.LogError("Restream error", err)
		t.playError(w, req, err)
	}
}

func (t *T) playError(
	w http.ResponseWriter,
	req extractor_config.RequestT,
	err error,
) {
	if err := t.streamer.PlayError(w, req, err); err != nil {
		t.log.LogError("Error occured while playing error video", err)
	}
}

func (t *T) cacheCheck(req extractor_config.RequestT, now time.Time) (extractor_config.ResultT, bool) {
	for _, v := range t.cache.CleanExpired(now) {
		t.log.LogDebug("Clean expired cache", v)
	}
	return t.cache.Get(req)
}

func (t *T) cacheAdd(
	req extractor_config.RequestT,
	res extractor_config.ResultT,
	now time.Time,
) {
	t.log.LogDebug("Cache add", res)
	t.cache.Add(req, res, now)
}

func remove_http(url string) string {
	url = strings.TrimPrefix(url, "http:/")
	url = strings.TrimPrefix(url, "https:/")
	url = strings.TrimLeft(url, "/")
	return url
}

func parseQuery(query string) extractor_config.RequestT {
	var req extractor_config.RequestT
	query = strings.TrimSpace(strings.TrimPrefix(query, "/play/"))
	splitted := strings.Split(query, "?/?")
	req.URL = remove_http(splitted[0])
	req.HEIGHT = defaultVideoHeight
	req.FORMAT = defaultVideoFormat
	if len(splitted) != 2 {
		return req
	}
	tOpts, tErr := url.ParseQuery(splitted[1])
	if tErr == nil {
		if tvh, ok := tOpts["vh"]; ok {
			if tvh[0] == "360" || tvh[0] == "480" || tvh[0] == "720" {
				req.HEIGHT = tvh[0]
			}
		}
		if tvf, ok := tOpts["vf"]; ok {
			if tvf[0] == "mp4" || tvf[0] == "m4a" {
				req.FORMAT = tvf[0]
			}
		}
	}
	return req
}
