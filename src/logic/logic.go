// Package logic implements apps client serving logic
package logic

import (
	"fmt"
	"net/http"
	"net/url"
	"slices"
	"strconv"
	"strings"
	"time"

	cache "ytproxy/cache"
	extractor "ytproxy/extractor"
	logger "ytproxy/logger"
	logger_mux "ytproxy/logger/mux"
	streamer "ytproxy/streamer"
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
	defaultVideoHeight uint64
	maxVideoHeight     uint64
}

// AppLogic is logic instance
type AppLogic struct {
	defaultApp app
	appList    []app
}

// Option is mini app, that serving selected sites
type Option struct {
	Name               string
	Sites              []string
	X                  extractor.T
	S                  streamer.T
	C                  cache.T
	DefaultVideoHeight uint64
	MaxVideoHeight     uint64
}

// New creates app logic instance
func New(def Option, opts []Option) *AppLogic {
	var t AppLogic
	t.set(def, opts)
	return &t
}

func (t *AppLogic) set(def Option, opts []Option) {
	t.defaultApp = app{
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
			defaultVideoHeight: v.DefaultVideoHeight,
			maxVideoHeight:     v.MaxVideoHeight,
		})
	}
}

func (t *AppLogic) selectApp(rawURL string) (app, error) {
	host, err := parseURLHost(rawURL)
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

func parseURLHost(rawURL string) (string, error) {
	u, err := url.Parse("https://" + rawURL)
	return u.Host, err
}

// Run serves single client
func (t *AppLogic) Run(w http.ResponseWriter, r *http.Request, log logger.T) {
	log = logger_mux.NewLayer(log, fmt.Sprintf("App %s", r.RemoteAddr))
	log.LogInfo("Play request", "url", r.RequestURI, "full", r)
	defer log.LogInfo("Player disconnected")
	now := time.Now()
	link, height, format := parseQuery(r.RequestURI)
	miniApp, err := t.selectApp(link)
	if err != nil {
		log.LogWarning("", "error", err)
		w.WriteHeader(http.StatusForbidden)
		return
	}
	miniAppLog := logger_mux.NewLayer(log, fmt.Sprintf("[%s]", miniApp.name))
	printExpired := func(links []extractor.RequestT) {
		if len(links) > 0 {
			miniAppLog.LogDebug("Expired", "links", links)
		}
	}
	req := miniApp.fixRequest(link, height, format)
	log.LogInfo("", "req", req, "app", miniApp.name)
	if res, ok, expired := miniApp.cacheCheck(req, now); ok {
		printExpired(expired)
		miniAppLog.LogDebug("Already cached", "link", res)
		miniApp.play(w, r, req, res, miniAppLog)
	} else {
		printExpired(expired)
		res, err := miniApp.extractor.Extract(req, miniAppLog)
		if err != nil {
			miniAppLog.LogError("URL extract", "error", err)
			miniApp.playError(w, req, err, miniAppLog)
			return
		}
		miniAppLog.LogDebug("Extractor returned", "link", res)
		miniApp.cacheAdd(req, res, now, miniAppLog)
		miniApp.play(w, r, req, res, miniAppLog)
	}
}

func (t *app) play(
	w http.ResponseWriter,
	r *http.Request,
	req extractor.RequestT,
	res extractor.ResultT,
	log logger.T,
) {
	if err := t.streamer.Play(w, r, res, log); err != nil {
		log.LogError("Restream", "error", err)
		t.playError(w, req, err, log)
	}
}

func (t *app) playError(
	w http.ResponseWriter,
	req extractor.RequestT,
	err error,
	log logger.T,
) {
	if err := t.streamer.PlayError(w, req, err); err != nil {
		log.LogError("Error occurred while playing error video", "error", err)
	}
}

func (t *app) cacheCheck(req extractor.RequestT, now time.Time) (extractor.ResultT, bool, []extractor.RequestT) {
	expired := t.cache.CleanExpired(now)
	res, ok := t.cache.Get(req)
	return res, ok, expired
}

func (t *app) cacheAdd(
	req extractor.RequestT,
	res extractor.ResultT,
	now time.Time,
	log logger.T,
) {
	log.LogDebug("Cache", "add", res)
	t.cache.Add(req, res, now)
}

func removeHTTP(url string) string {
	url = strings.TrimPrefix(url, "http:/")
	url = strings.TrimPrefix(url, "https:/")
	url = strings.TrimLeft(url, "/")
	return url
}

func parseQuery(query string) (string, uint64, string) {
	query = strings.TrimSpace(strings.TrimPrefix(query, "/play/"))
	split := strings.Split(query, "?/?")
	link := removeHTTP(split[0])
	format := defaultVideoFormat
	var height uint64
	if len(split) != 2 {
		return link, 0, format
	}
	tOpts, tErr := url.ParseQuery(split[1])
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

func (t *app) fixRequest(link string, height uint64, format string) extractor.RequestT {
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
	return extractor.RequestT{
		URL:    link,
		HEIGHT: h,
		FORMAT: format,
	}
}
