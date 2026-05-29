#!/bin/bash

set -e

cd ..

#docker run --rm -it -v `pwd`/tinygo-dev:/go/src/testpgm tinygotest tinygo help

# build docker image with dependent packages downloaded
docker build -t tinygotest -f tinygo-dev/Dockerfile .
docker run --rm -it -v `pwd`/tinygo-dev:/go/src/testpgm -e "GOPATH=/go" tinygotest \
 tinygo build -o /go/src/testpgm/testpgm.wasm -target wasm testpgm

# copy wasm_exec.js out
if ! [ -f tinygo-dev/wasm_exec.js ]; then
echo "Copying wasm_exec.js"
docker run --rm -it -v `pwd`/tinygo-dev:/go/src/testpgm tinygotest /bin/bash -c "cp /usr/local/tinygo/targets/wasm_exec.js /go/src/testpgm/"
fi

