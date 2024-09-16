package config

import (
	"encoding/json"
	"os"
	"strings"

	extractor "ytproxy-extractor"
	cache "ytproxy-linkscache"
	logger "ytproxy-logger"
	streamer "ytproxy-streamer"
)

type configT struct {
	PortInt   uint16            `json:"port"`
	Streamer  streamer.ConfigT  `json:"streamer"`
	Extractor extractor.ConfigT `json:"extractor"`
	Log       logger.ConfigT    `json:"log"`
	Cache     cache.ConfigT     `json:"cache"`
}

func defaultConfig() configT {
	fls := false
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
	return configT{
		PortInt: 8080,
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
		Extractor: extractor.ConfigT{
			Path:          &e[0],
			MP4:           &e[1],
			M4A:           &e[2],
			GetUserAgent:  &e[2],
			CustomOptions: &co,
		},
		Log:   logger.ConfigT{},
		Cache: cache.ConfigT{},
	}
}

// add second config options to first
func appendConfig(src configT, dst configT) configT {
	// general options
	if dst.PortInt == 0 {
		dst.PortInt = src.PortInt
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
	return dst
}

func Read(path string) (configT, error) {
	var c configT
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
	return appendConfig(defaultConfig(), c), err
}
