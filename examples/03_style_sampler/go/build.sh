#!/bin/bash
set -e

# Copy templates into go/ for embed (go:embed cannot use ../paths)
cp -r ../templates .

# Build standard Go WASM
echo "Building Go WASM..."
<<<<<<< HEAD
GOOS=js GOARCH=wasm go build -o main.wasm .
=======
GOOS=js GOARCH=wasm go build -ldflags="-s" -trimpath -o main.wasm .
>>>>>>> task/WTteletype
echo "  main.wasm: $(du -h main.wasm | cut -f1)"

# Copy wasm_exec.js
cp "$(go env GOROOT)/misc/wasm/wasm_exec.js" .
echo "  wasm_exec.js copied"

<<<<<<< HEAD
# Build TinyGo WASM if available
if command -v tinygo &> /dev/null; then
    echo "Building TinyGo WASM..."
    tinygo build -o main-tinygo.wasm -target wasm .
    echo "  main-tinygo.wasm: $(du -h main-tinygo.wasm | cut -f1)"
    cp "$(tinygo env TINYGOROOT)/targets/wasm_exec.js" tinygo_wasm_exec.js
    echo "  tinygo_wasm_exec.js copied"
fi

=======
>>>>>>> task/WTteletype
# Clean up copied templates
rm -rf templates

echo "Done."
