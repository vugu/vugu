#!/bin/bash

docker run -d -ti --rm -p 9222:9222 -p 8846:8846 --name wasm-test-suite vugu/wasm-test-suite:latest
