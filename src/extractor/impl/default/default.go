// Package dedfaultextractor implements default (external command based) extractor
package dedfaultextractor

import (
	"bytes"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"sync"
	"text/template"

	extractor "ytproxy/extractor"
	logger "ytproxy/logger"
	logger_mux "ytproxy/logger/mux"
)

// New creates new default extractor implementation
func New(path string, mp4, m4a []string, getUserAgent string,
	customOptions []string) (extractor.T, error) {
	var (
		e   defaultExtractor
		err error
	)
	read := func(list []string) ([]*template.Template, error) {
		res := make([]*template.Template, 0)
		for _, v := range list {
			t, err := template.New("").Parse(v)
			if err != nil {
				return res, err
			}
			res = append(res, t)
		}
		return res, nil
	}
	e.m4a, err = read(m4a)
	if err != nil {
		return &e, err
	}
	e.mp4, err = read(mp4)
	if err != nil {
		return &e, err
	}
	e.customOptions, err = read(customOptions)
	if err != nil {
		return &e, err
	}
	e.getUserAgent = getUserAgent
	e.path = path
	return &e, nil
}

type defaultExtractor struct {
	sync.Mutex
	path          string
	mp4           []*template.Template
	m4a           []*template.Template
	customOptions []*template.Template
	getUserAgent  string
}

func (t *defaultExtractor) GetUserAgent(log logger.T) (string, error) {
	return t.runCmd([]string{t.getUserAgent}, log)
}

func (t *defaultExtractor) Extract(req extractor.RequestT, log logger.T,
) (extractor.ResultT, error) {
	var (
		buf        []string
		bufOptions []string
		err        error
	)
	execute := func(list []*template.Template) ([]string, error) {
		buf := make([]string, 0)
		for _, v := range list {
			var b bytes.Buffer
			err = v.Execute(&b, req)
			if err != nil {
				return buf, err
			}
			buf = append(buf, bytesToString(b))
		}
		return buf, nil
	}
	switch req.FORMAT {
	case "m4a":
		buf, err = execute(t.m4a)
	case "mp4":
		fallthrough
	default:
		buf, err = execute(t.mp4)
	}
	if err != nil {
		return extractor.ResultT{}, err
	}
	bufOptions, err = execute(t.customOptions)
	if err != nil {
		return extractor.ResultT{}, err
	}
	bufOptions = append(bufOptions, buf...)
	out, err := t.runCmd(bufOptions, log)
	if err != nil {
		return extractor.ResultT{}, err
	}
	return extractor.ResultT{URL: out}, err
}

func (t *defaultExtractor) runCmd(args []string, log logger.T) (string, error) {
	t.Lock()
	defer t.Unlock()
	log = logger_mux.NewLayer(log, "Extractor")
	cmd := exec.Command(t.path, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	log.LogDebug("Running", "cmd",
		fmt.Sprintf("%s '%s'", t.path, strings.Join(args, "' '")))
	err := cmd.Run()
	outStr, errStr := bytesToString(stdout), bytesToString(stderr)
	if err != nil {
		combinedErrStr := fmt.Sprintf("%s\n%s\n%s", err.Error(), outStr, errStr)
		return "", errors.New(combinedErrStr)
	}
	return outStr, nil
}

func bytesToString(s bytes.Buffer) string {
	return strings.TrimSpace(s.String())
}
