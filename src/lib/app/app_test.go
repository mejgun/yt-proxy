package app

import (
	"testing"

	extractor_config "lib/extractor/config"
)

var testPairs = map[string]extractor_config.RequestT{
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

func TestRemoveHttp(t *testing.T) {
	for _, v := range []struct {
		link string
		want string
	}{
		{link: "youtu.be/", want: "youtu.be/"},
		{link: "https:////youtu.be/", want: "youtu.be/"},
		{link: "https:///youtu.be/", want: "youtu.be/"},
		{link: "https://youtu.be/", want: "youtu.be/"},
		{link: "https:/youtu.be/", want: "youtu.be/"},
		{link: "http:////www.youtu.be/", want: "www.youtu.be/"},
		{link: "http:///www.youtu.be/", want: "www.youtu.be/"},
		{link: "http://www.youtu.be/", want: "www.youtu.be/"},
		{link: "http:/www.youtu.be/", want: "www.youtu.be/"},
	} {
		if r := remove_http(v.link); r != v.want {
			t.Error("For", v.link, "expected", v.want, "got", r)
		}
	}
}
