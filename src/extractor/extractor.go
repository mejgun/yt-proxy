package extractor

type ConfigT struct {
	Path         string   `json:"path"`
	MP4          []string `json:"mp4"`
	M4A          []string `json:"m4a"`
	GetUserAgent []string `json:"get-user-agent"`
}
