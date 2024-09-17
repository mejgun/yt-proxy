package extractor

import (
	"bytes"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"sync"
	"text/template"
	"time"

	logger "lib/logger"
)

const separator = ",,"

type ConfigT struct {
	Path          *string   `json:"path"`
	MP4           *string   `json:"mp4"`
	M4A           *string   `json:"m4a"`
	GetUserAgent  *string   `json:"get-user-agent"`
	CustomOptions *[]string `json:"custom-options"`
}

type ResultT struct {
	URL    string
	Expire time.Time
}

type T interface {
	Extract(RequestT) (ResultT, error)
	GetUserAgent() (string, error)
}

type defaultExtractor struct {
	sync.Mutex
	path          string
	mp4           *template.Template
	m4a           *template.Template
	customOptions []*template.Template
	getUserAgent  string
	logger        *logger.T
}

type RequestT struct {
	URL    string
	HEIGHT string
	FORMAT string
}

func New(c ConfigT, log *logger.T) (T, error) {
	var (
		e   defaultExtractor
		err error
	)
	e.m4a, err = template.New("").Parse(*c.M4A)
	if err != nil {
		return &e, err
	}
	e.mp4, err = template.New("").Parse(*c.MP4)
	if err != nil {
		return &e, err
	}
	e.customOptions = make([]*template.Template, 0)
	for _, v := range *c.CustomOptions {
		b, err := template.New("").Parse(v)
		if err != nil {
			return &e, err
		}
		e.customOptions = append(e.customOptions, b)
	}
	e.getUserAgent = *c.GetUserAgent
	e.path = *c.Path
	e.logger = log
	return &e, nil
}

func (t *defaultExtractor) GetUserAgent() (string, error) {
	return t.runCmd(t.getUserAgent)
}

func (t *defaultExtractor) Extract(req RequestT) (ResultT, error) {
	var (
		buf bytes.Buffer
		err error
	)
	switch req.FORMAT {
	case "m4a":
		err = t.m4a.Execute(&buf, req)
	case "mp4":
		fallthrough
	default:
		err = t.mp4.Execute(&buf, req)
	}
	if err != nil {
		return ResultT{}, err
	}
	bufOptions := make([]string, 0)
	for _, v := range t.customOptions {
		var b bytes.Buffer
		err = v.Execute(&b, req)
		if err != nil {
			return ResultT{}, err
		}
		bufOptions = append(bufOptions, bytesToString(b))
	}
	bufOptions = append(bufOptions, bytesToString(buf))
	out, err := t.runCmd(strings.Join(bufOptions, separator))
	if err != nil {
		return ResultT{}, err
	}
	return ResultT{URL: out}, err
}

func (t *defaultExtractor) runCmd(args string) (string, error) {
	realargs := split(args)
	t.Lock()
	defer t.Unlock()
	cmd := exec.Command(t.path, realargs...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	t.logger.LogDebug("Running", t.path, strings.Join(realargs, " "))
	err := cmd.Run()
	outStr, errStr := bytesToString(stdout), bytesToString(stderr)
	if err != nil {
		combinedErrStr := fmt.Sprintf("%s\n%s\n%s", err.Error(), outStr, errStr)
		return "", errors.New(combinedErrStr)
	}
	return outStr, nil
}

func split(s string) []string {
	return strings.Split(s, separator)
}

func bytesToString(s bytes.Buffer) string {
	return strings.TrimSpace(s.String())
}
