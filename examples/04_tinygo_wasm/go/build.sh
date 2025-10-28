#!/bin/bash
# Build script for compiling Go to WASM using TinyGo

set -e

echo "================================================"
echo "Building Go to WASM with TinyGo"
echo "================================================"
echo ""

# Check if TinyGo is installed
if ! command -v tinygo &> /dev/null; then
    echo "Error: TinyGo is not installed"
    echo ""
    echo "Install TinyGo:"
    echo "  macOS:  brew install tinygo"
    echo "  Linux:  https://tinygo.org/getting-started/install/linux/"
    echo "  Windows: https://tinygo.org/getting-started/install/windows/"
    echo ""
    echo "Or use Docker:"
    echo "  docker run --rm -v \$(pwd):/src tinygo/tinygo:latest tinygo build -o main.wasm -target wasm /src/main.go"
    exit 1
fi

echo "✓ TinyGo is installed: $(tinygo version)"
echo ""

# Build the WASM binary
echo "Compiling Go to WebAssembly with TinyGo..."
tinygo build -o main.wasm -target wasm main.go

if [ $? -eq 0 ]; then
    echo "✓ Build successful: main.wasm"
    echo ""

    # Get file size
    size=$(du -h main.wasm | cut -f1)
    echo "File size: $size"
    echo ""

    # Copy wasm_exec.js from TinyGo installation
    echo "Copying wasm_exec.js helper from TinyGo..."
    tinygo_root=$(tinygo env TINYGOROOT)
    if [ -f "$tinygo_root/targets/wasm_exec.js" ]; then
        cp "$tinygo_root/targets/wasm_exec.js" .
        echo "✓ Copied wasm_exec.js"
    else
        echo "Warning: Could not find wasm_exec.js at $tinygo_root/targets/wasm_exec.js"
        echo "You may need to copy it manually or download it from:"
        echo "  https://raw.githubusercontent.com/tinygo-org/tinygo/release/targets/wasm_exec.js"
    fi
    echo ""

    echo "================================================"
    echo "Build complete!"
    echo "================================================"
    echo ""
    echo "File size comparison:"
    echo "  TinyGo WASM:    $size (this build!)"
    echo "  Standard Go:    ~2MB"
    echo "  Pyodide:        ~10MB"
    echo ""
    echo "To run locally:"
    echo "  cd .. && python3 serve.py"
    echo "  or: python3 -m http.server 8000"
    echo ""
    echo "Then open http://localhost:8000"
    echo ""
else
    echo "✗ Build failed"
    exit 1
fi
