package app

import (
	"fmt"
	cache "lib/cache"
	extractor "lib/extractor"
	extractor_config "lib/extractor/config"
	logger "lib/logger"
	streamer "lib/streamer"
	"slices"
	"strconv"
	"sync"

	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	defaultVideoFormat = "mp4"
)

type app struct {
	cache              cache.T
	extractor          extractor.T
	streamer           streamer.T
	name               string
	sites              []string
	log                logger.T
	defaultVideoHeight uint64
	maxVideoHeight     uint64
}

type AppLogic struct {
	mu         sync.RWMutex
	log        logger.T
	defaultApp app
	appList    []app
}

type Option struct {
	Name               string
	Sites              []string
	X                  extractor.T
	S                  streamer.T
	C                  cache.T
	L                  logger.T
	DefaultVideoHeight uint64
	MaxVideoHeight     uint64
}

func New(log logger.T, def Option, opts []Option) *AppLogic {
	var t AppLogic
	t.set(log, def, opts)
	return &t
}

func (t *AppLogic) set(log logger.T, def Option, opts []Option) {
	t.log = log
	t.defaultApp = app{
		log:                def.L,
		name:               "default",
		cache:              def.C,
		extractor:          def.X,
		streamer:           def.S,
		defaultVideoHeight: def.DefaultVideoHeight,
		maxVideoHeight:     def.MaxVideoHeight,
		sites:              def.Sites,
	}

	t.appList = make([]app, 0)
	for _, v := range opts {
		t.appList = append(t.appList, app{
			cache:              v.C,
			extractor:          v.X,
			streamer:           v.S,
			name:               v.Name,
			sites:              v.Sites,
			log:                v.L,
			defaultVideoHeight: v.DefaultVideoHeight,
			maxVideoHeight:     v.MaxVideoHeight,
		})
	}
}

func (t *AppLogic) selectApp(rawURL string) (app, error) {
	host, err := parseUrlHost(rawURL)
	if err == nil {
		for _, v := range t.appList {
			if slices.Contains(v.sites, host) {
				return v, nil
			}
		}
	}
	if len(t.defaultApp.sites) == 0 || slices.Contains(t.defaultApp.sites, host) {
		return t.defaultApp, nil
	}
	return app{}, fmt.Errorf("host %s did not match any sites in config or sub-configs", host)
}

func parseUrlHost(rawURL string) (string, error) {
	u, err := url.Parse("https://" + rawURL)
	return u.Host, err
}

func (t *AppLogic) Run(w http.ResponseWriter, r *http.Request) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	printExpired := func(a app, links []extractor_config.RequestT) {
		if len(links) > 0 {
			a.log.LogDebug("Expired links", links)
		}
	}
	now := time.Now()
	link, height, format := parseQuery(r.RequestURI)
	resapp, err := t.selectApp(link)
	if err != nil {
		t.log.LogWarning("", err)
		w.WriteHeader(http.StatusForbidden)
		return
	}
	req := resapp.fixRequest(link, height, format)
	t.log.LogInfo("Request", req, "app", resapp.name)
	if res, ok, expired := resapp.cacheCheck(req, now); ok {
		printExpired(resapp, expired)
		resapp.log.LogDebug("Link already cached", res)
		resapp.play(w, r, req, res, t.log)
	} else {
		printExpired(resapp, expired)
		res, err := resapp.extractor.Extract(req)
		if err != nil {
			resapp.log.LogError("URL extract error", err)
			resapp.playError(w, req, err, t.log)
			return
		}
		resapp.log.LogDebug("Extractor returned", res)
		resapp.cacheAdd(req, res, now, t.log)
		resapp.play(w, r, req, res, t.log)
	}
}

func (t *app) play(
	w http.ResponseWriter,
	r *http.Request,
	req extractor_config.RequestT,
	res extractor_config.ResultT,
	logger logger.T,
) {
	if err := t.streamer.Play(w, r, req, res); err != nil {
		logger.LogError("Restream error", err)
		t.playError(w, req, err, logger)
	}
}

func (t *app) playError(
	w http.ResponseWriter,
	req extractor_config.RequestT,
	err error,
	logger logger.T,
) {
	if err := t.streamer.PlayError(w, req, err); err != nil {
		logger.LogError("Error occured while playing error video", err)
	}
}

func (t *app) cacheCheck(req extractor_config.RequestT, now time.Time) (extractor_config.ResultT, bool, []extractor_config.RequestT) {
	expired := t.cache.CleanExpired(now)
	res, ok := t.cache.Get(req)
	return res, ok, expired
}

func (t *app) cacheAdd(
	req extractor_config.RequestT,
	res extractor_config.ResultT,
	now time.Time,
	logger logger.T,
) {
	logger.LogDebug("Cache add", res)
	t.cache.Add(req, res, now)
}

func remove_http(url string) string {
	url = strings.TrimPrefix(url, "http:/")
	url = strings.TrimPrefix(url, "https:/")
	url = strings.TrimLeft(url, "/")
	return url
}

func parseQuery(query string) (string, uint64, string) {
	query = strings.TrimSpace(strings.TrimPrefix(query, "/play/"))
	splitted := strings.Split(query, "?/?")
	link := remove_http(splitted[0])
	format := defaultVideoFormat
	var height uint64
	if len(splitted) != 2 {
		return link, 0, format
	}
	tOpts, tErr := url.ParseQuery(splitted[1])
	if tErr == nil {
		if tvh, ok := tOpts["vh"]; ok {
			height, _ = strconv.ParseUint(tvh[0], 10, 64)
		}
		if tvf, ok := tOpts["vf"]; ok {
			if tvf[0] == "mp4" || tvf[0] == "m4a" {
				format = tvf[0]
			}
		}
	}
	return link, height, format

}

func (t *app) fixRequest(link string, height uint64, format string) extractor_config.RequestT {
	var (
		h   string
		toS = func(d uint64) string {
			return fmt.Sprintf("%d", d)
		}
	)
	switch {
	case height == 0:
		h = toS(t.defaultVideoHeight)
	case height > t.maxVideoHeight:
		h = toS(t.maxVideoHeight)
	default:
		h = toS(height)
	}
	return extractor_config.RequestT{
		URL:    link,
		HEIGHT: h,
		FORMAT: format,
	}
}

func (t *AppLogic) Shutdown() {
	t.mu.Lock() // locking app forever
	t.log.LogInfo("Exiting")
	t.log.Close()
}

func (t *AppLogic) ReloadConfig(log logger.T, def Option, opts []Option) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.log.LogInfo("Reloading app")
	t.log.Close()
	t.set(log, def, opts)
	t.log.LogInfo("Reloading complete")
}
