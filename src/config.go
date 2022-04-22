package main

type configT struct {
	EnableDebug          bool     `json:"debug"`
	EnableErrorHeaders   bool     `json:"error-headers"`
	IgnoreMissingHeaders bool     `json:"ignore-missing-headers"`
	IgnoreSSLErrors      bool     `json:"ignore-ssl-errors"`
	PortInt              uint16   `json:"port"`
	ErrorVideoPath       string   `json:"error-video"`
	Extractor            struct{} `json:"extractor"`
}
