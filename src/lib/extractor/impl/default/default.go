package extractor

import (
	"bytes"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"sync"
	"text/template"

	extractor_config "lib/extractor/config"
	logger "lib/logger"
)

func New(path string, mp4, m4a []string, get_user_agent string,
	custom_options []string, log logger.T) (*defaultExtractor, error) {
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
	e.customOptions, err = read(custom_options)
	if err != nil {
		return &e, err
	}
	e.getUserAgent = get_user_agent
	e.path = path
	e.logger = log
	return &e, nil
}

type defaultExtractor struct {
	sync.Mutex
	path          string
	mp4           []*template.Template
	m4a           []*template.Template
	customOptions []*template.Template
	getUserAgent  string
	logger        logger.T
}

func (t *defaultExtractor) GetUserAgent() (string, error) {
	return t.runCmd([]string{t.getUserAgent})
}

func (t *defaultExtractor) Extract(req extractor_config.RequestT,
) (extractor_config.ResultT, error) {
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
		return extractor_config.ResultT{}, err
	}
	bufOptions, err = execute(t.customOptions)
	if err != nil {
		return extractor_config.ResultT{}, err
	}
	bufOptions = append(bufOptions, buf...)
	out, err := t.runCmd(bufOptions)
	if err != nil {
		return extractor_config.ResultT{}, err
	}
	return extractor_config.ResultT{URL: out}, err
}

func (t *defaultExtractor) runCmd(args []string) (string, error) {
	t.Lock()
	defer t.Unlock()
	cmd := exec.Command(t.path, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	t.logger.LogDebug("Running", "path", t.path, "args", strings.Join(args, " "))
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
