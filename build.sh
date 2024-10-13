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

cat <<EOF >${Md5File}
see changelog

built w $(go version | awk '{print($3)}')

<details>
<summary>md5</summary>
<code>
$(md5sum -b ${FilePrefix}*)
</code>
</details>
EOF

cd ..
