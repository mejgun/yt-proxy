package main

import (
	"testing"

	extractor "ytproxy-extractor"
)

var testPairs = map[string]extractor.RequestT{
	"/play/youtu.be/jNQXAC9IVRw?/?vh=360&vf=mp4": extractor.RequestT{"youtu.be/jNQXAC9IVRw", "360", "mp4"},
	"/play/youtu.be/jNQXAC9IVRw?/?vh=720?vf=avi": extractor.RequestT{"youtu.be/jNQXAC9IVRw", "720", defaultVideoFormat},
	"/play/youtu.be/jNQXAC9IVRw":                 extractor.RequestT{"youtu.be/jNQXAC9IVRw", defaultVideoHeight, defaultVideoFormat},
	"/play/youtu.be/jNQXAC9IVRw?/?":              extractor.RequestT{"youtu.be/jNQXAC9IVRw", defaultVideoHeight, defaultVideoFormat},
	"/play/youtu.be/jNQXAC9IVRw?/?vf=avi":        extractor.RequestT{"youtu.be/jNQXAC9IVRw", defaultVideoHeight, defaultVideoFormat},
	"/play/youtu.be/jNQXAC9IVRw?/?vf=mp4":        extractor.RequestT{"youtu.be/jNQXAC9IVRw", defaultVideoHeight, "mp4"},
}

func TestParseQuery(t *testing.T) {
	for k, v := range testPairs {
		if r := parseQuery(k); r != v {
			t.Error("For", k, "expected", v, "got", r)
		}
	}
}
