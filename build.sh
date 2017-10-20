#!/bin/sh

# https://github.com/golang/go/blob/master/src/go/build/syslist.go

# -ldflags '-s -w' 

GOPATH=$(pwd) GOOS=linux GOARCH=amd64 go build -o bin/amd64/yt-proxy
GOPATH=$(pwd) GOOS=linux GOARCH=386 go build -o bin/i386/yt-proxy
