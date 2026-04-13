#!/bin/bash
set -e

# Copy templates into go/ for embed (go:embed cannot use ../paths)
cp -r ../templates .

# Build standard Go WASM
echo "Building Go WASM..."
GOOS=js GOARCH=wasm go build -ldflags="-s" -trimpath -o main.wasm .
echo "  main.wasm: $(du -h main.wasm | cut -f1)"

# Copy wasm_exec.js
cp "$(go env GOROOT)/misc/wasm/wasm_exec.js" .
echo "  wasm_exec.js copied"

# Clean up copied templates
rm -rf templates

echo "Done."
