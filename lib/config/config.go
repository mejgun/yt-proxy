package config

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	cache "lib/cache"
	extractor_config "lib/extractor/config"
	logger_config "lib/logger/config"
	streamer "lib/streamer"
)

type ConfigT struct {
	PortInt            uint16                   `json:"port"`
	Host               string                   `json:"host"`
	DefaultVideoHeight uint16                   `json:"default-video-height"`
	Streamer           streamer.ConfigT         `json:"streamer"`
	Extractor          extractor_config.ConfigT `json:"extractor"`
	Log                logger_config.ConfigT    `json:"log"`
	Cache              cache.ConfigT            `json:"cache"`
	SubConfig          []SubConfigT             `json:"sub-config"`
}

type SubConfigT struct {
	Name  string   `json:"name"`
	Sites []string `json:"sites"`
	ConfigT
}

func defaultConfig() ConfigT {
	fls := false
	tru := true
	ext := streamer.Extractor
	tv := streamer.TlsVersion(0)
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
	ll := logger_config.Info
	lo := logger_config.Stdout
	lf := "log.txt"
	exp := "3h"
	return ConfigT{
		PortInt:            8080,
		Host:               "0.0.0.0",
		DefaultVideoHeight: 720,
		Streamer: streamer.ConfigT{
			EnableErrorHeaders:   &fls,
			IgnoreMissingHeaders: &fls,
			IgnoreSSLErrors:      &fls,
			ErrorVideoPath:       &s[0],
			ErrorAudioPath:       &s[1],
			SetUserAgent:         &ext,
			UserAgent:            &s[2],
			Proxy:                &s[3],
			MinTlsVersion:        &tv,
		},
		Extractor: extractor_config.ConfigT{
			Path:          &e[0],
			MP4:           &e[1],
			M4A:           &e[2],
			GetUserAgent:  &e[3],
			CustomOptions: &co,
			ForceHttps:    &tru,
		},
		Log: logger_config.ConfigT{
			Level:    &ll,
			Json:     &fls,
			Output:   &lo,
			FileName: &lf,
		},
		Cache: cache.ConfigT{
			ExpireTime: &exp,
		},
	}
}

// add second config options to first
func appendConfig(src ConfigT, dst ConfigT) ConfigT {
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
	if dst.Streamer.MinTlsVersion == nil {
		dst.Streamer.MinTlsVersion = src.Streamer.MinTlsVersion
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
	if dst.Extractor.ForceHttps == nil {
		dst.Extractor.ForceHttps = src.Extractor.ForceHttps
	}
	// logger
	if dst.Log.Level == nil {
		dst.Log.Level = src.Log.Level
	}
	if dst.Log.Json == nil {
		dst.Log.Json = src.Log.Json
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

func Read(path string) (ConfigT, error) {
	var c ConfigT
	b, err := os.ReadFile(path)
	if err != nil {
		return c, err
	}
	func() {
		strs := make([]string, 0)
		for _, s := range strings.Split(string(b[:]), "\n") {
			s = strings.TrimSpace(s)
			if !strings.HasPrefix(s, "//") {
				strs = append(strs, s)
			}
		}
		str := strings.Join(strs, "\n")
		b = b[:0]
		b = []byte(str)

	}()
	err = json.Unmarshal(b, &c)
	if err != nil {
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
		c.SubConfig[k].ConfigT = appendConfig(c, v.ConfigT)
	}
	return c, nil
}
