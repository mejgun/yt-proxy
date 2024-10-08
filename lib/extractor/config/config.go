package extractor

import "time"

type ConfigT struct {
	Path          *string   `json:"path"`
	MP4           *string   `json:"mp4"`
	M4A           *string   `json:"m4a"`
	GetUserAgent  *string   `json:"get-user-agent"`
	CustomOptions *[]string `json:"custom-options"`
	ForceHttps    *bool     `json:"force-https"`
}

type ResultT struct {
	URL    string
	Expire time.Time
}
type RequestT struct {
	URL    string
	HEIGHT string
	FORMAT string
}
