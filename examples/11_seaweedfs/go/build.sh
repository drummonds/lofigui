#!/bin/bash
set -e

echo "Building SeaweedFS WASM demo..."

GOOS=js GOARCH=wasm go build -ldflags="-s" -trimpath -o ../templates/main.wasm .

WASM_EXEC=$(find "$(go env GOROOT)" -name wasm_exec.js -print -quit)
if [ -z "$WASM_EXEC" ]; then
    echo "Error: wasm_exec.js not found in GOROOT"
    exit 1
fi
cp "$WASM_EXEC" ../templates/

echo "Build complete:"
ls -lh ../templates/main.wasm ../templates/wasm_exec.js
