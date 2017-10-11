package main

import (
	"fmt"
	"log"
	"net/url"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

type Lnk struct {
	url    string
	expire int64
}

type Links map[string]Lnk

var links Links

func getLink(vidurl string) (string, error) {
	vidurl = strings.TrimSpace(vidurl)
	cmd := exec.Command("youtube-dl", "-g", "-f mp4", vidurl)
	stdoutStderr, err := cmd.CombinedOutput()
	stdoutStderrStr := strings.TrimSpace(string(stdoutStderr))
	if err == nil {
		var expire int64
		// err rewrited
		u, err := url.Parse(stdoutStderrStr)
		if err == nil {
			m, _ := url.ParseQuery(u.RawQuery)
			if e, ok := m["expire"]; ok {
				e1, err := strconv.ParseInt(e[0], 10, 64)
				if err == nil {
					expire = e1
				} else {
					expire = time.Now().Unix() + 10800
				}

			} else {
				expire = time.Now().Unix() + 10800
			}
		}
		links[vidurl] = Lnk{url: stdoutStderrStr, expire: expire}
		fmt.Printf("Added url %s expire %d\n", vidurl, expire)
	}
	return stdoutStderrStr, err
}

func main() {
	links = make(Links)
	url := "https://www.youtube.com/watch?v=H0hsrEhz3zo"
	//url := "https://vimeo.com/235305664"
	stdoutStderr, err := getLink(url)
	fmt.Printf("%s\n", stdoutStderr)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(links)
	fmt.Println(time.Now().Unix())
	fmt.Println(time.Now().Unix() + 10800)
}
