#!/bin/bash
# Build script for compiling Go to WASM

set -e

echo "================================================"
echo "Building Go to WASM"
echo "================================================"
echo ""

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "Error: Go is not installed"
    echo "Please install Go from https://golang.org/dl/"
    exit 1
fi

echo "✓ Go is installed: $(go version)"
echo ""

# Build the WASM binary
echo "Compiling Go to WebAssembly..."
GOOS=js GOARCH=wasm go build -o main.wasm main.go

if [ $? -eq 0 ]; then
    echo "✓ Build successful: main.wasm"
    echo ""

    # Get file size
    size=$(du -h main.wasm | cut -f1)
    echo "File size: $size"
    echo ""

    # Copy wasm_exec.js from Go installation
    echo "Copying wasm_exec.js helper..."
    cp "$(go env GOROOT)/misc/wasm/wasm_exec.js" .
    echo "✓ Copied wasm_exec.js"
    echo ""

    echo "================================================"
    echo "Build complete!"
    echo "================================================"
    echo ""
    echo "To run locally:"
    echo "  python3 serve.py"
    echo "  or: python3 -m http.server 8000"
    echo ""
    echo "Then open http://localhost:8000"
    echo ""
else
    echo "✗ Build failed"
    exit 1
fi
