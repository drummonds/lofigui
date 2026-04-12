# Go Example 03: Hello World WASM

Go compiled to WebAssembly running in the browser - much smaller and faster than Pyodide!

## Overview

This example demonstrates:
- Go code compiled directly to WebAssembly
- Running in the browser with no server needed
- **5x smaller** than Pyodide (~2MB vs ~10MB)
- **10x faster** startup (instant vs 3-5 seconds)
- Perfect for hosting on GitHub Pages

## Structure

```
03_hello_world_wasm/
├── main.go           # Go code that compiles to WASM
├── index.html        # Main HTML page
├── app.js            # JavaScript to load and run WASM
├── build.sh          # Build script
├── serve.py          # Local development server
├── go.mod            # Go module definition
└── readme.md         # This file
```

## Build & Run

### Step 1: Build the WASM Binary

```bash
./build.sh
```

This compiles `main.go` to `main.wasm` using:
```bash
GOOS=js GOARCH=wasm go build -o main.wasm main.go
```

### Step 2: Serve Locally

```bash
python3 serve.py
```

Or using Task from repo root:
```bash
task go-wasm
```

Then open: http://localhost:8000

## How It Works

### 1. Go Code (main.go)

```go
func model() string {
    lofigui.Reset()
    lofigui.Print("Hello from Go WASM!")
    return lofigui.Buffer()
}

func main() {
    // Expose Go functions to JavaScript
    js.Global().Set("goRunModel", js.FuncOf(runModel))
    js.Global().Call("wasmReady")

    // Keep running
    <-make(chan struct{})
}
```

### 2. Compile to WASM

```bash
GOOS=js GOARCH=wasm go build -o main.wasm
```

Creates a WebAssembly binary that runs in the browser.

### 3. Load in Browser (app.js)

```javascript
const go = new Go();
const result = await WebAssembly.instantiateStreaming(
    fetch('main.wasm'),
    go.importObject
);
go.run(result.instance);
```

### 4. Call Go from JavaScript

```javascript
// Call Go function
const html = goRunModel();
document.getElementById('output').innerHTML = html;
```

## Comparison: Go WASM vs Python Pyodide

| Feature | Go WASM | Python Pyodide |
|---------|---------|----------------|
| **Size** | ~2MB | ~10MB |
| **Startup** | <100ms | 3-5 seconds |
| **Performance** | Native (compiled) | Interpreted |
| **Build Required** | Yes | No |
| **Type Safety** | Compile-time | Runtime |
| **Ecosystem** | Go packages | Python packages |
| **Best For** | Performance apps | Python ecosystem |

## When to Use Go WASM

✅ **Use Go WASM when:**
- Performance matters
- You want fast startup
- Smaller download size is important
- You're comfortable with Go
- Type safety is valuable

❌ **Use Python/Pyodide when:**
- No build step desired
- Need Python ecosystem (NumPy, Pandas, etc.)
- Rapid prototyping
- Dynamic typing preferred

## Deploying to GitHub Pages

### Option 1: Manual Deploy

```bash
# Build
./build.sh

# Copy to docs folder
mkdir -p ../../../docs/go-wasm
cp index.html app.js main.wasm wasm_exec.js ../../../docs/go-wasm/

# Commit and push
git add docs/go-wasm
git commit -m "Add Go WASM example"
git push
```

### Option 2: Use Task

```bash
task deploy-go-wasm  # If we add this task
```

Enable GitHub Pages to serve from `/docs` folder.

## File Sizes

After building:

```
main.wasm      ~2.0 MB   (Go code compiled to WASM)
wasm_exec.js   ~16 KB    (Go WASM runtime helper)
index.html     ~4 KB     (HTML page)
app.js         ~2 KB     (JavaScript glue code)
--------------------------------
Total:         ~2.02 MB  (vs ~10MB for Pyodide)
```

## Performance

- **Load time**: <1 second (vs 3-5 seconds for Pyodide)
- **Execution**: Native WASM speed
- **Memory**: Lower overhead than Pyodide
- **Startup**: Instant (no interpreter to load)

## Development Workflow

1. Edit `main.go`
2. Run `./build.sh`
3. Refresh browser
4. Repeat

For faster development, use file watchers or Task automation.

## Troubleshooting

### "WASM failed to load"
- Make sure you ran `./build.sh`
- Check that `main.wasm` exists
- Use HTTP server, not file://

### "Go is not installed"
```bash
# Install Go
# Linux/Mac:
# Download from https://golang.org/dl/
# Or use package manager
```

### Build errors
- Ensure Go 1.21+ is installed
- Check `go.mod` is present
- Run `go mod download`

## Next Steps

- Modify `model()` function in `main.go`
- Add more Go functions exposed to JavaScript
- Create interactive forms
- Add real-time data processing
- Deploy to GitHub Pages

## Resources

- [Go WASM Documentation](https://github.com/golang/go/wiki/WebAssembly)
- [syscall/js Package](https://pkg.go.dev/syscall/js)
- [WebAssembly.org](https://webassembly.org/)

## License

MIT License - same as lofigui project
