#!/bin/bash

# https://github.com/golang/go/blob/master/src/go/build/syslist.go

BinDir="bin"
FilePrefix="yt-proxy"

OSListStr="aix android darwin dragonfly freebsd hurd illumos js linux nacl netbsd openbsd plan9 solaris windows zos"
OSList=($OSListStr)

ArchListStr="386 amd64 amd64p32 arm armbe arm64 arm64be ppc64 ppc64le mips mipsle mips64 mips64le mips64p32 mips64p32le ppc riscv riscv64 s390 s390x sparc sparc64 wasm"
ArchList=($ArchListStr)

Total=$((${#OSList[@]}*${#ArchList[@]}))

for OS in ${OSListStr}
do
    for ARCH in ${ArchListStr}
    do
        echo $Total ${OS}/${ARCH}
	File=${BinDir}/${FilePrefix}-${OS}-${ARCH}
        GOPATH=$(pwd) GOOS=${OS} GOARCH=${ARCH} go build -ldflags '-s -w' -o ${File}  > /dev/null 2>&1 && md5sum ${File} > ${File}.md5sum
        ((Total=Total-1))
    done
done

