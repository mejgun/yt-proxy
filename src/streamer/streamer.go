package streamer

import (
	"net/http"
	"os"
	extractor "ytproxy-extractor"
)

type ConfigT struct {
	EnableErrorHeaders   bool   `json:"error-headers"`
	IgnoreMissingHeaders bool   `json:"ignore-missing-headers"`
	IgnoreSSLErrors      bool   `json:"ignore-ssl-errors"`
	ErrorVideoPath       string `json:"error-video"`
	ErrorAudioPath       string `json:"error-audio"`
}

type T interface {
	Play(http.ResponseWriter, *http.Request, extractor.ResultT)
}

type streamer struct {
	errorVideoFile fileT
	errorAudioFile fileT
}

func New(conf ConfigT) (T, error) {
	var (
		s   streamer
		err error
	)
	s.errorVideoFile, err = readFile(conf.ErrorVideoPath)
	if err != nil {
		return s, err
	}
	s.errorAudioFile, err = readFile(conf.ErrorAudioPath)
	if err != nil {
		return s, err
	}

}

type fileT struct {
	content []byte
	size    int64
}

func readFile(path string) (fileT, error) {
	file, err := os.Open(path)
	if err != nil {
		return fileT{}, err
	}
	fileinfo, err := file.Stat()
	if err != nil {
		return fileT{}, err
	}
	filesize := fileinfo.Size()
	file.Close()
	content, err := os.ReadFile(path)
	if err != nil {
		return fileT{}, err
	}
	return fileT{content: content, size: filesize}, nil
}
