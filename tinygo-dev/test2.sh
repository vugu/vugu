#!/bin/bash

set -e

# copy to src dir
mkdir -p src/github.com/vugu/vugu
rsync -taz --exclude=tinygo-dev --exclude=.git ../ src/github.com/vugu/vugu/

# env and dependencies
export GOPATH=`pwd`
export GO111MODULE=off
go get github.com/vugu/vjson
go get github.com/vugu/xxhash

# compile
TG=~/go/bin/tinygo
$TG build -o testpgm2.wasm -target wasm ./
