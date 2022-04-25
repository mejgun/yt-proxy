#!/bin/bash

BinDir="bin"
Md5File="all.md5"

date >${Md5File}
mkdir -p ${BinDir} || exit
rm ${BinDir}/${FilePrefix}* -rf || exit
cd src || exit
go run ../build.go || exit
cd ../$BinDir || exit
for i in $(ls -1 *); do
    md5sum $i >>../${Md5File} || exit
done
cd ..
