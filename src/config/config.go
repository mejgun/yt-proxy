// Package config contains app config related funcs
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	cache "ytproxy/cache"
	extractor "ytproxy/extractor"
	logger "ytproxy/logger"
	streamer "ytproxy/streamer"
)

// T is main app config type
type T struct {
	PortInt            uint16            `json:"port"`
	Host               string            `json:"host"`
	DefaultVideoHeight uint64            `json:"default-video-height"`
	MaxVideoHeight     uint64            `json:"max-video-height"`
	Sites              []string          `json:"sites"`
	Streamer           streamer.ConfigT  `json:"streamer"`
	Extractor          extractor.ConfigT `json:"extractor"`
	Log                logger.ConfigT    `json:"log"`
	Cache              cache.ConfigT     `json:"cache"`
	SubConfig          []SubT            `json:"sub-config"`
}

// SubT is type for extra configs
type SubT struct {
	Name string `json:"name"`
	T
}

func defaultConfig() T {
	fls := false
	tru := true
	ext := streamer.Extractor
	tv := streamer.TLSVersion(0)
	var s = [4]string{"corrupted.mp4",
		"failed.m4a",
		"Mozilla",
		"env",
	}
	var e = [4]string{"yt-dlp",
		"-f,,(mp4)[height<={{.HEIGHT}}],,-g,,{{.URL}}",
		"-f,,(m4a),,-g,,{{.URL}}",
		"--dump-user-agent",
	}
	co := make([]string, 0)
	ll := logger.Info
	lo := logger.Stdout
	lf := "log.txt"
	exp := "3h"
	return T{
		PortInt:            8080,
		Host:               "0.0.0.0",
		DefaultVideoHeight: 720,
		MaxVideoHeight:     720,
		Streamer: streamer.ConfigT{
			EnableErrorHeaders:   &fls,
			IgnoreMissingHeaders: &fls,
			IgnoreSSLErrors:      &fls,
			ErrorVideoPath:       &s[0],
			ErrorAudioPath:       &s[1],
			SetUserAgent:         &ext,
			UserAgent:            &s[2],
			Proxy:                &s[3],
			MinTLSVersion:        &tv,
		},
		Extractor: extractor.ConfigT{
			Path:          &e[0],
			MP4:           &e[1],
			M4A:           &e[2],
			GetUserAgent:  &e[3],
			CustomOptions: &co,
			ForceHTTPS:    &tru,
		},
		Log: logger.ConfigT{
			Level:    &ll,
			JSON:     &fls,
			Output:   &lo,
			FileName: &lf,
		},
		Cache: cache.ConfigT{
			ExpireTime: &exp,
		},
	}
}

// add second config options to first
func appendConfig(src T, dst T) T {
	// general options
	if dst.PortInt == 0 {
		dst.PortInt = src.PortInt
	}
	if dst.Host == "" {
		dst.Host = src.Host
	}
	if dst.DefaultVideoHeight == 0 {
		dst.DefaultVideoHeight = src.DefaultVideoHeight
	}
	if dst.MaxVideoHeight == 0 {
		dst.MaxVideoHeight = src.MaxVideoHeight
	}
	// streamer
	if dst.Streamer.EnableErrorHeaders == nil {
		dst.Streamer.EnableErrorHeaders = src.Streamer.EnableErrorHeaders
	}
	if dst.Streamer.IgnoreMissingHeaders == nil {
		dst.Streamer.IgnoreMissingHeaders = src.Streamer.IgnoreMissingHeaders
	}
	if dst.Streamer.IgnoreSSLErrors == nil {
		dst.Streamer.IgnoreSSLErrors = src.Streamer.IgnoreSSLErrors
	}
	if dst.Streamer.ErrorVideoPath == nil {
		dst.Streamer.ErrorVideoPath = src.Streamer.ErrorVideoPath
	}
	if dst.Streamer.ErrorAudioPath == nil {
		dst.Streamer.ErrorAudioPath = src.Streamer.ErrorAudioPath
	}
	if dst.Streamer.SetUserAgent == nil {
		dst.Streamer.SetUserAgent = src.Streamer.SetUserAgent
	}
	if dst.Streamer.UserAgent == nil {
		dst.Streamer.UserAgent = src.Streamer.UserAgent
	}
	if dst.Streamer.Proxy == nil {
		dst.Streamer.Proxy = src.Streamer.Proxy
	}
	if dst.Streamer.MinTLSVersion == nil {
		dst.Streamer.MinTLSVersion = src.Streamer.MinTLSVersion
	}
	// extractor
	if dst.Extractor.Path == nil {
		dst.Extractor.Path = src.Extractor.Path
	}
	if dst.Extractor.MP4 == nil {
		dst.Extractor.MP4 = src.Extractor.MP4
	}
	if dst.Extractor.M4A == nil {
		dst.Extractor.M4A = src.Extractor.M4A
	}
	if dst.Extractor.GetUserAgent == nil {
		dst.Extractor.GetUserAgent = src.Extractor.GetUserAgent
	}
	if dst.Extractor.CustomOptions == nil {
		dst.Extractor.CustomOptions = src.Extractor.CustomOptions
	}
	if dst.Extractor.ForceHTTPS == nil {
		dst.Extractor.ForceHTTPS = src.Extractor.ForceHTTPS
	}
	// logger
	if dst.Log.Level == nil {
		dst.Log.Level = src.Log.Level
	}
	if dst.Log.JSON == nil {
		dst.Log.JSON = src.Log.JSON
	}
	if dst.Log.Output == nil {
		dst.Log.Output = src.Log.Output
	}
	if dst.Log.FileName == nil {
		dst.Log.FileName = src.Log.FileName
	}
	// cache
	if dst.Cache.ExpireTime == nil {
		dst.Cache.ExpireTime = src.Cache.ExpireTime
	}
	return dst
}

// Read reads config file
func Read(path string) (T, error) {
	var c T
	b, err := os.ReadFile(path)
	if err != nil {
		return c, err
	}
	func() {
		splitStrings := make([]string, 0)
		for _, s := range strings.Split(string(b[:]), "\n") {
			s = strings.TrimSpace(s)
			if !strings.HasPrefix(s, "//") {
				splitStrings = append(splitStrings, s)
			}
		}
		str := strings.Join(splitStrings, "")
		b = b[:0]
		b = []byte(str)

	}()
	err = json.Unmarshal(b, &c)
	if err != nil {
		wrapErr := func(offset int64) error {
			l, h := offset-30, offset+20
			pre, post := "…", "…"
			if l < 0 {
				l = 0
				pre = ""
			}
			if h > int64(len(b)) {
				h = int64(len(b))
				post = ""
			}
			return fmt.Errorf("%s '%s%s%s'", err, pre, b[l:h], post)
		}
		switch e := err.(type) {
		case *json.UnmarshalTypeError:
			err = wrapErr(e.Offset)
		case *json.SyntaxError:
			err = wrapErr(e.Offset)
		}
		return c, err
	}
	c = appendConfig(defaultConfig(), c)
	for k, v := range c.SubConfig {
		if v.Name == "" {
			return c, fmt.Errorf("sub-config name empty")
		}
		if len(v.Sites) == 0 {
			return c, fmt.Errorf("sub-config sites empty")
		}
		c.SubConfig[k].T = appendConfig(c, v.T)
	}
	return c, nil
}
