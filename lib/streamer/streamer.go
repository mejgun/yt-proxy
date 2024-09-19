package streamer

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"os"
	"strings"

	extractor "lib/extractor"
	extractor_config "lib/extractor/config"
	logger "lib/logger"
)

const defaultErrorHeader = "Error-Header-"

type ConfigT struct {
	EnableErrorHeaders   *bool          `json:"error-headers"`
	IgnoreMissingHeaders *bool          `json:"ignore-missing-headers"`
	IgnoreSSLErrors      *bool          `json:"ignore-ssl-errors"`
	ErrorVideoPath       *string        `json:"error-video"`
	ErrorAudioPath       *string        `json:"error-audio"`
	SetUserAgent         *SetUserAgentT `json:"set-user-agent"`
	UserAgent            *string        `json:"user-agent"`
	Proxy                *string        `json:"proxy"`
	MinTlsVersion        *TlsVersion    `json:"min-tls-version"`
}

type TlsVersion uint16

func (u *TlsVersion) UnmarshalJSON(b []byte) error {
	var (
		s string
		i uint16
	)
	if u == nil {
		return nil
	}
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	eq := func(teststring string) bool {
		return s == teststring
	}
	if eq("") {
		return nil
	}
	for i = 0; i < math.MaxUint16; i++ {
		if eq(tls.VersionName(i)) {
			*u = TlsVersion(i)
			return nil
		}
	}
	return fmt.Errorf("cannot unmarshal %s as TLS version", b)
}

type SetUserAgentT uint8

const (
	Extractor SetUserAgentT = iota
	Request
	Config
)

func (u *SetUserAgentT) UnmarshalJSON(b []byte) error {
	var s string
	err := json.Unmarshal(b, &s)
	if err != nil {
		return err
	}
	switch s {
	case "request":
		*u = Request
	case "extractor":
		*u = Extractor
	case "config":
		*u = Config
	default:
		return fmt.Errorf("cannot unmarshal %s as user-agent", b)
	}
	return nil
}

type T interface {
	Play(http.ResponseWriter, *http.Request, extractor_config.RequestT, extractor_config.ResultT) error
	PlayError(http.ResponseWriter, extractor_config.RequestT, error) error
}

type streamer struct {
	errorVideoFile       fileT
	errorAudioFile       fileT
	httpRequest          doRequestF
	sendErrorFile        sendErrorFileF
	setHeaders           func(http.ResponseWriter, *http.Response) error
	setStreamerUserAgent func(*http.Request) string
	log                  logger.T
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

func New(conf ConfigT, log logger.T, xt extractor.T) (T, error) {
	var (
		s    streamer
		err  error
		logs []string
	)
	s.errorVideoFile, err = readFile(*conf.ErrorVideoPath)
	if err != nil {
		return &s, err
	}
	s.errorVideoFile.contentType = "video/mp4"
	s.errorAudioFile, err = readFile(*conf.ErrorAudioPath)
	if err != nil {
		return &s, err
	}
	s.errorAudioFile.contentType = "audio/mp4"
	s.httpRequest, logs, err = makeDoRequestFunc(conf)
	for _, v := range logs {
		log.LogDebug("streamer", v)
	}
	if err != nil {
		return &s, err
	}
	s.sendErrorFile = makeSendErrorVideoFunc(conf)
	s.setHeaders = makeSetHeaders(conf)
	s.setStreamerUserAgent, err = makeSetStreamerUserAgent(conf, xt, log)
	if err != nil {
		return &s, err
	}
	s.log = log
	return &s, nil
}

func (t *streamer) Play(
	w http.ResponseWriter,
	req *http.Request,
	reqst extractor_config.RequestT,
	rest extractor_config.ResultT,
) error {
	// t.log.LogDebug("Streamer request", rest)
	// fail := func(str string, err error) {
	// 	t.log.LogError(str, err)
	// 	t.PlayError(w, reqst, err)
	// }
	request, err := http.NewRequest("GET", rest.URL, nil)
	if err != nil {
		return err
	}
	if r1, ok := req.Header["Range"]; ok {
		request.Header.Set("Range", r1[0])
	}
	request.Header.Set("User-Agent", t.setStreamerUserAgent(req))
	res, err := t.httpRequest(request)
	if err != nil {
		// fail("Proxying error", err)
		return err
	}
	defer res.Body.Close()
	t.log.LogDebug("Response", res)
	err = t.setHeaders(w, res)
	if err != nil {
		// fail("Headers error", err)
		return err
	}
	_, err = io.Copy(w, res.Body)
	if err != nil {
		return err
	}
	return nil
}

func (t *streamer) PlayError(w http.ResponseWriter, req extractor_config.RequestT,
	err error) error {
	var file *fileT
	if req.FORMAT == "mp4" {
		file = &t.errorVideoFile
	} else {
		file = &t.errorAudioFile
	}
	return t.sendErrorFile(w, err, *file)
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

func makeDoRequestFunc(conf ConfigT) (doRequestF, []string, error) {
	tr := &http.Transport{}
	logs := make([]string, 0)
	func() {
		mintls := uint16(*conf.MinTlsVersion)
		tr.TLSClientConfig = &tls.Config{MinVersion: mintls}
		if mintls > 0 {
			logs = append(logs,
				fmt.Sprintf("streamer min TLS version set to: %s",
					tls.VersionName(mintls)))
		}
	}()
	if *conf.IgnoreSSLErrors {
		tr.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
		logs = append(logs, "ignoring SSL errors")
	}
	switch *conf.Proxy {
	case "":
		logs = append(logs, "no proxy set")
	case "env":
		tr.Proxy = http.ProxyFromEnvironment
		logs = append(logs, "proxy set to environment")
	default:
		logs = append(logs, fmt.Sprintf("proxy set to '%s'", *conf.Proxy))
		u, err := url.Parse(*conf.Proxy)
		if err != nil {
			return func(r *http.Request) (*http.Response, error) {
				return &http.Response{}, nil
			}, logs, err
		}
		tr.Proxy = http.ProxyURL(u)
	}
	return func(request *http.Request) (*http.Response, error) {
		client := &http.Client{Transport: tr}
		return client.Do(request)
	}, logs, nil
}

func makeSendErrorVideoFunc(conf ConfigT) sendErrorFileF {
	return func(w http.ResponseWriter, err error, file fileT) error {
		w.Header().Set("Content-Length", fmt.Sprintf("%d", file.contentLength))
		w.Header().Set("Content-Type", file.contentType)
		if *conf.EnableErrorHeaders {
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
	headersStrictCheck := !*conf.IgnoreMissingHeaders
	return func(w http.ResponseWriter, res *http.Response) error {
		h1, ok := res.Header["Content-Length"]
		if !ok && headersStrictCheck {
			return fmt.Errorf("no Content-Length header")
		}
		if ok {
			w.Header().Set("Content-Length", h1[0])
		}
		h2, ok := res.Header["Content-Type"]
		if !ok && headersStrictCheck {
			return fmt.Errorf("no Content-Type header")
		}
		if headersStrictCheck && h2[0] != "video/mp4" && h2[0] != "audio/mp4" {
			return fmt.Errorf("Content-Type is not video/mp4 or audio/mp4, but %s", h2[0])
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
		if res.StatusCode == 206 {
			w.WriteHeader(http.StatusPartialContent)
		}
		return nil
	}
}

func makeSetStreamerUserAgent(conf ConfigT, xt extractor.T, log logger.T) (func(*http.Request) string, error) {
	switch *conf.SetUserAgent {
	case Request:
		log.LogDebug("", "User-Agent", "request-set")
		return func(r *http.Request) string {
			return r.UserAgent()
		}, nil
	case Extractor:
		ua, err := xt.GetUserAgent()
		log.LogDebug("", "User-Agent", ua)
		return func(r *http.Request) string {
			return ua
		}, err
	case Config:
		ua := conf.UserAgent
		log.LogDebug("", "User-Agent", ua)
		return func(r *http.Request) string {
			return *ua
		}, nil
	default:
		return func(r *http.Request) string { return "" },
			fmt.Errorf("set-streamer-user-agent func creation error. this could not happen")
	}
}
