#!/bin/bash

set -e

BinDir="bin"
Md5File="all.md5"
FilePrefix="yt-proxy"

date >${Md5File}
mkdir -p ${BinDir}
rm ${BinDir}/${FilePrefix}* -rf
cd cmd
go run ../build.go ${FilePrefix}
cd ../$BinDir
md5sum -b ${FilePrefix}* >${Md5File}
cd ..
