package main

import (
	"testing"

	extractor "ytproxy-extractor"
)

var testPairs = map[string]extractor.RequestT{
	"/play/youtu.be/jNQXAC9IVRw?/?vh=360&vf=mp4": {
		URL:    "youtu.be/jNQXAC9IVRw",
		HEIGHT: "360",
		FORMAT: "mp4",
	},
	"/play/youtu.be/jNQXAC9IVRw?/?vh=720?vf=avi": {
		URL:    "youtu.be/jNQXAC9IVRw",
		HEIGHT: "720",
		FORMAT: defaultVideoFormat,
	},
	"/play/youtu.be/jNQXAC9IVRw": {
		URL:    "youtu.be/jNQXAC9IVRw",
		HEIGHT: defaultVideoHeight,
		FORMAT: defaultVideoFormat,
	},
	"/play/youtu.be/jNQXAC9IVRw?/?": {
		URL:    "youtu.be/jNQXAC9IVRw",
		HEIGHT: defaultVideoHeight,
		FORMAT: defaultVideoFormat,
	},
	"/play/youtu.be/jNQXAC9IVRw?/?vf=avi": {
		URL:    "youtu.be/jNQXAC9IVRw",
		HEIGHT: defaultVideoHeight,
		FORMAT: defaultVideoFormat,
	},
	"/play/youtu.be/jNQXAC9IVRw?/?vf=mp4": {
		URL:    "youtu.be/jNQXAC9IVRw",
		HEIGHT: defaultVideoHeight,
		FORMAT: "mp4",
	},
}

func TestParseQuery(t *testing.T) {
	for k, v := range testPairs {
		if r := parseQuery(k); r != v {
			t.Error("For", k, "expected", v, "got", r)
		}
	}
}
