#!/bin/bash
set -e

# Templates live in go/templates/ so go:embed can reach them directly.
echo "Building Go WASM..."
GOOS=js GOARCH=wasm go build -ldflags="-s" -trimpath -o main.wasm .
echo "  main.wasm: $(du -h main.wasm | cut -f1)"

cp "$(go env GOROOT)/misc/wasm/wasm_exec.js" .
echo "  wasm_exec.js copied"

echo "Done."
