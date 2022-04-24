package ytproxy

import (
	"net/http"
)

type response struct {
	url string
	err error
}

// type requestChan struct {
// 	url        string
// 	answerChan chan response
// }

type lnkT struct {
	url    string
	expire int64
}

// type debugF func(string, interface{})

type corruptedT struct {
	file []byte
	size int64
}

type sendErrorVideoF func(http.ResponseWriter, error)

type doRequestF func(*http.Request) (*http.Response, error)
