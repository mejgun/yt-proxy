#!/bin/bash

set -e

BinDir="bin"
Md5File="all.md5"
FilePrefix="yt-proxy"

mkdir -p ${BinDir}
rm ${BinDir}/${FilePrefix}* -rf
cd src
go run ../build.go ${FilePrefix}
cd ../$BinDir
md5sum -b ${FilePrefix}* >${Md5File}
cd ..
