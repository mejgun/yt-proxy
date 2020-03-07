#!/bin/sh

# https://github.com/golang/go/blob/master/src/go/build/syslist.go

GOPATH=$(pwd) GOOS=linux GOARCH=amd64 go build -ldflags '-s -w' -o bin/amd64/yt-proxy
GOPATH=$(pwd) GOOS=linux GOARCH=386 go build -ldflags '-s -w' -o bin/i386/yt-proxy
