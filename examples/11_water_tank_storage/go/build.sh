#!/bin/bash
set -e
GOOS=js GOARCH=wasm go build -ldflags="-s" -trimpath -o main.wasm .
cp "$(go env GOROOT)/misc/wasm/wasm_exec.js" .
