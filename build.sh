#!/bin/sh

# https://github.com/golang/go/blob/master/src/go/build/syslist.go

OSLIST="aix android darwin dragonfly freebsd hurd illumos js linux nacl netbsd openbsd plan9 solaris windows zos"
ARCHLIST="386 amd64 amd64p32 arm armbe arm64 arm64be ppc64 ppc64le mips mipsle mips64 mips64le mips64p32 mips64p32le ppc riscv riscv64 s390 s390x sparc sparc64 wasm"

for OS in ${OSLIST}
do
    for ARCH in ${ARCHLIST}
    do
    echo ${OS}/${ARCH}
    GOPATH=$(pwd) GOOS=${OS} GOARCH=${ARCH} go build -ldflags '-s -w' -o bin/yt-proxy-${OS}-${ARCH} 2>&1 > /dev/null
    done
done