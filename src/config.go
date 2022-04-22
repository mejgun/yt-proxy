package main

type configT struct {
	EnableDebug          bool       `json:"debug"`
	EnableErrorHeaders   bool       `json:"error-headers"`
	IgnoreMissingHeaders bool       `json:"ignore-missing-headers"`
	IgnoreSSLErrors      bool       `json:"ignore-ssl-errors"`
	PortInt              uint16     `json:"port"`
	ErrorVideoPath       string     `json:"error-video"`
	Extractor            extractorT `json:"extractor"`
}

type extractorT struct {
	Path         string   `json:"path"`
	MP4          []string `json:"mp4"`
	M4A          []string `json:"m4a"`
	GetUserAgent []string `json:"get-user-agent"`
}
