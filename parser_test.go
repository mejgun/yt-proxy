package main

import "testing"

var testPairs = map[string][3]string{
	"/play/youtu.be/jNQXAC9IVRw?/?vh=360&vf=mp4": {"youtu.be/jNQXAC9IVRw", "360", "mp4"},
	"/play/youtu.be/jNQXAC9IVRw?/?vh=720?vf=avi": {"youtu.be/jNQXAC9IVRw", "720", defaultVideoFormat},
	"/play/youtu.be/jNQXAC9IVRw":                 {"youtu.be/jNQXAC9IVRw", defaultVideoHeight, defaultVideoFormat},
}

func TestParseQuery(t *testing.T) {
	for k, v := range testPairs {
		if u, h, f := parseQuery(k); [3]string{u, h, f} != v {
			t.Error("For", k, "expected", v, "got", u, h, f)
		}
	}
}
