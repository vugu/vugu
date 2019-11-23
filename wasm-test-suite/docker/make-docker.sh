#!/bin/bash

set -e

# Go build for linux
GOOS=linux GOARCH=amd64 go build -o wasm-test-suite-srv .

# Docker build and tag image
docker build -t vugu/wasm-test-suite:latest .

echo "Run with: docker run -ti --rm -p 9222:9222 -p 8846:8846 vugu/wasm-test-suite:latest"
echo "To push to DockerHub, run 'docker login' and then 'docker push vugu/wasm-test-suite:latest'"
