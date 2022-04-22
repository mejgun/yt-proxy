#!/bin/bash

# https://github.com/golang/go/blob/master/src/go/build/syslist.go

GoBin=go

BinDir="../bin"
FilePrefix="yt-proxy"

OSListStr="aix android darwin dragonf freebsd hurd illumos ios js linux nacl netbsd openbsd plan9 solaris windows zos"
OSList=($OSListStr)

ArchListStr="386 amd64 amd64p32 arm armbe arm64 arm64be loong64 mips mipsle mips64 mips64le mips64p32 mips64p32 ppc ppc64 ppc64le riscv riscv64 s390 s390x sparc sparc64 wasm"
ArchList=($ArchListStr)

Total=$((${#OSList[@]} * ${#ArchList[@]}))
Md5File="all.md5"

date >${Md5File}

mkdir -p ${BinDir} || exit
rm ${BinDir}/${FilePrefix}* -rf || exit
cd src || exit
for OS in ${OSListStr}; do
    for ARCH in ${ArchListStr}; do
        echo $Total ${OS}/${ARCH}
        File=${FilePrefix}-${OS}-${ARCH}
        GOOS=${OS} GOARCH=${ARCH} $GoBin build -ldflags '-s -w' -o ${BinDir}/${File} >/dev/null 2>&1 && cd ${BinDir} && md5sum ${File} >>../${Md5File} && cd -
        ((Total = Total - 1))
    done
done
