package logic

import (
	"fmt"
	"strings"
	"testing"
)

func TestParseQuery(t *testing.T) {
	var testPairs = map[string]string{
		"/play/youtu.be/jNQXAC9IVRw?/?vh=360&vf=mp4":   "youtu.be/jNQXAC9IVRw|360|mp4",
		"/play/youtu.be/jNQXAC9IVRw?/?vh=720&vf=avi":   "youtu.be/jNQXAC9IVRw|720|mp4",
		"/play/youtu.be/jNQXAC9IVRw":                   "youtu.be/jNQXAC9IVRw|0|mp4",
		"/play/youtu.be/jNQXAC9IVRw?/?":                "youtu.be/jNQXAC9IVRw|0|mp4",
		"/play/youtu.be/jNQXAC9IVRw?/?vf=avi":          "youtu.be/jNQXAC9IVRw|0|mp4",
		"/play/youtu.be/jNQXAC9IVRw?/?vf=mp4":          "youtu.be/jNQXAC9IVRw|0|mp4",
		"/play/youtu.be/jNQXAC9IVRw?/?vf=mp4&vh=11111": "youtu.be/jNQXAC9IVRw|11111|mp4",
	}
	for k, v := range testPairs {
		l, h, f := parseQuery(k)
		if strings.Join([]string{l, fmt.Sprintf("%d", h), f}, "|") != v {
			t.Error("For", k, "expected", v, "got", l, h, f)
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
		if r := removeHTTP(v.link); r != v.want {
			t.Error("For", v.link, "expected", v.want, "got", r)
		}
	}
}

func TestParseHost(t *testing.T) {
	for _, v := range []struct {
		link string
		want string
	}{
		{link: "youtu.be/1", want: "youtu.be"},
		{link: "youtu.be:443/?param=1&b=c", want: "youtu.be:443"},
		{link: "www.yyy.youtu.be/", want: "www.yyy.youtu.be"},
		{link: "youtu.be", want: "youtu.be"},
	} {
		if r, err := parseURLHost(v.link); r != v.want || err != nil {
			t.Error("For", v.link, "expected", v.want, "got", r, err)
		}
	}

}
