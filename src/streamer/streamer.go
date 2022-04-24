package streamer

import (
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	extractor "ytproxy-extractor"
	logger "ytproxy-logger"
)

const defaultErrorHeader = "Error-Header-"

type ConfigT struct {
	EnableErrorHeaders   bool   `json:"error-headers"`
	IgnoreMissingHeaders bool   `json:"ignore-missing-headers"`
	IgnoreSSLErrors      bool   `json:"ignore-ssl-errors"`
	ErrorVideoPath       string `json:"error-video"`
	ErrorAudioPath       string `json:"error-audio"`
}

type T interface {
	Play(http.ResponseWriter, *http.Request, extractor.RequestT, extractor.ResultT)
	PlayError(http.ResponseWriter, extractor.RequestT)
}

type streamer struct {
	errorVideoFile fileT
	errorAudioFile fileT
	httpRequest    doRequestF
	sendErrorFile  sendErrorFileF
	setHeaders     func(http.ResponseWriter, *http.Response) error
	log            *logger.T
}

type (
	doRequestF     func(*http.Request) (*http.Response, error)
	sendErrorFileF func(http.ResponseWriter, error, fileT) error
)

type fileT struct {
	content       []byte
	contentType   string
	contentLength int64
}

func New(conf ConfigT, log *logger.T) (T, error) {
	var (
		s   streamer
		err error
	)
	s.errorVideoFile, err = readFile(conf.ErrorVideoPath)
	if err != nil {
		return &s, err
	}
	s.errorVideoFile.contentType = "video/mp4"
	s.errorAudioFile, err = readFile(conf.ErrorAudioPath)
	if err != nil {
		return &s, err
	}
	s.errorAudioFile.contentType = "audio/mp4"
	s.httpRequest = makeDoRequestFunc(conf)
	s.sendErrorFile = makeSendErrorVideoFunc(conf)
	s.setHeaders = makeSetHeaders(conf)
	return &s, nil
}

func (t *streamer) Play(
	w http.ResponseWriter,
	req *http.Request,
	reqst extractor.RequestT,
	rest extractor.ResultT,
) {
	t.log.LogDebug("Streamer request", rest)
	fail := func(str string, err error) {
		t.log.LogError(str, err)
		var file *fileT
		if reqst.FORMAT == "mp4" {
			file = &t.errorVideoFile
		} else {
			file = &t.errorAudioFile
		}
		t.sendErrorFile(w, err, *file)
	}
	request, err := http.NewRequest("GET", rest.URL, nil)
	if err != nil {
		fail("Proxying error", err)
		return
	}
	if r1, ok := req.Header["Range"]; ok {
		request.Header.Set("Range", r1[0])
	}
	request.Header.Set("User-Agent", req.UserAgent())
	res, err := t.httpRequest(request)
	if err != nil {
		fail("Proxying error", err)
		return
	}
	defer res.Body.Close()
	t.log.LogDebug("Response", res)
	err = t.setHeaders(w, res)
	if err != nil {
		fail("Headers error", err)
	}
	if res.StatusCode == 206 {
		w.WriteHeader(http.StatusPartialContent)
	}
	_, err = io.Copy(w, res.Body)
	if err != nil {
		log.Println("Proxy error", err)
	}
}

func (t *streamer) PlayError(w http.ResponseWriter, req extractor.RequestT) {}

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
	return fileT{content: content, contentLength: filesize}, nil
}

func errorToHeaders(e error) ([]string, []string) {
	splitted := strings.Split(e.Error(), "\n")
	filtered := make([]string, 0)
	for _, v := range splitted {
		v := strings.TrimSpace(v)
		if len(v) > 0 {
			filtered = append(filtered, v)
		}
	}
	count := len(fmt.Sprintf("%d", len(filtered)))
	format := fmt.Sprintf("%s%%0%dd", defaultErrorHeader, count+1)
	headers := make([]string, 0)
	for i := range filtered {
		headers = append(headers, fmt.Sprintf(format, i+1))

	}
	return headers, filtered
}

func makeDoRequestFunc(conf ConfigT) doRequestF {
	var tr *http.Transport
	if conf.IgnoreSSLErrors {
		tr = &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
	} else {
		tr = &http.Transport{}
	}
	return func(request *http.Request) (*http.Response, error) {
		client := &http.Client{Transport: tr}
		return client.Do(request)
	}
}

func makeSendErrorVideoFunc(conf ConfigT) sendErrorFileF {
	return func(w http.ResponseWriter, err error, file fileT) error {
		w.Header().Set("Content-Length", fmt.Sprintf("%d", file.contentLength))
		w.Header().Set("Content-Type", file.contentType)
		if conf.EnableErrorHeaders {
			hdrs, errs := errorToHeaders(err)
			for i := range hdrs {
				w.Header().Set(hdrs[i], errs[i])
			}
		}
		_, err = w.Write(file.content)
		return err
	}
}

func makeSetHeaders(conf ConfigT) func(http.ResponseWriter, *http.Response) error {
	headersStrictCheck := !conf.IgnoreMissingHeaders
	return func(w http.ResponseWriter, res *http.Response) error {
		h1, ok := res.Header["Content-Length"]
		if !ok && headersStrictCheck {
			return errors.New("no Content-Length header")
		}
		if ok {
			w.Header().Set("Content-Length", h1[0])
		}
		h2, ok := res.Header["Content-Type"]
		if !ok && headersStrictCheck {
			return errors.New("no Content-Type header")
		}
		if headersStrictCheck && h2[0] != "video/mp4" && h2[0] != "audio/mp4" {
			return errors.New("Content-Type is not video/mp4 or audio/mp4")
		}
		if ok {
			w.Header().Set("Content-Type", h2[0])
		}
		if h3, ok := res.Header["Accept-Ranges"]; ok {
			w.Header().Set("Accept-Ranges", h3[0])
		}
		if h4, ok := res.Header["Content-Range"]; ok {
			w.Header().Set("Content-Range", h4[0])
		}
		return nil
	}
}
