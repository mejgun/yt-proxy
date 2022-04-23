package extractor

import (
	"bytes"
	"errors"
	"fmt"
	"net/url"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"text/template"
)

const separator = ",,"

type ConfigT struct {
	Path         string `json:"path"`
	MP4          string `json:"mp4"`
	M4A          string `json:"m4a"`
	GetUserAgent string `json:"get-user-agent"`
}

type ResultT struct {
	URL    string
	Expire int64
}

type ExtractorT interface {
	Extract(RequestT) (ResultT, error)
	GetUserAgent() (string, error)
}

type defaultExtractor struct {
	sync.Mutex
	path         string
	mp4          *template.Template
	m4a          *template.Template
	getUserAgent string
}

type RequestT struct {
	URL    string
	HEIGHT string
	FORMAT string
}

func New(c ConfigT) (ExtractorT, error) {
	var (
		e   defaultExtractor
		err error
	)
	e.m4a, err = template.New("").Parse(c.M4A)
	if err != nil {
		return &e, err
	}
	e.mp4, err = template.New("").Parse(c.MP4)
	if err != nil {
		return &e, err
	}
	e.getUserAgent = c.GetUserAgent
	e.path = c.Path
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
	out, err := t.runCmd(buf.String())
	if err != nil {
		return ResultT{}, err
	}
	var expire int64
	u, err := url.Parse(out)
	if err == nil {
		m, _ := url.ParseQuery(u.RawQuery)
		if e, ok := m["expire"]; ok {
			e1, err1 := strconv.ParseInt(e[0], 10, 64)
			if err1 == nil {
				expire = e1
			}
		}
	}
	return ResultT{URL: out, Expire: expire}, err
}

func (t *defaultExtractor) runCmd(args string) (string, error) {
	t.Lock()
	defer t.Unlock()
	cmd := exec.Command(t.path, split(args)...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
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
