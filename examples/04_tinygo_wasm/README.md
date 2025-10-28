# Lofigui Example 04: TinyGo WASM - Ultra-Lightweight Hello World

This example demonstrates using **TinyGo** to compile Go code to WebAssembly for running in the browser. TinyGo produces incredibly small WASM binaries - **20x smaller** than standard Go!

## 🎯 What Makes This Special

This is the **smallest** lofigui example:
- **~100KB** total size (vs ~2MB for standard Go, ~10MB for Pyodide)
- **Instant loading** - near-zero startup time
- **Perfect for mobile** - minimal bandwidth usage
- **GitHub Pages ready** - deploy anywhere static files work

## Size Comparison

```
TinyGo WASM:    ~100KB  █ (1%)
Standard Go:     ~2MB   ████████████████████ (20%)
Pyodide:        ~10MB   ████████████████████████████████████████████████████████████████████████████████████████████████████ (100%)
```

## 📁 Project Structure

```
04_tinygo_wasm/
├── go/
│   ├── main.go        # TinyGo source code
│   ├── build.sh       # Build script for TinyGo
│   ├── go.mod         # Go module file
│   ├── main.wasm      # Compiled WASM (generated)
│   └── wasm_exec.js   # TinyGo WASM helper (generated)
├── templates/
│   ├── index.html     # Main HTML page
│   └── app.js         # JavaScript to load and run WASM
├── serve.py           # Simple HTTP server
└── README.md          # This file
```

## 🚀 Quick Start

### Prerequisites

Install TinyGo:

**macOS:**
```bash
brew install tinygo
```

**Linux:**
```bash
wget https://github.com/tinygo-org/tinygo/releases/download/v0.30.0/tinygo_0.30.0_amd64.deb
sudo dpkg -i tinygo_0.30.0_amd64.deb
```

**Windows:**
See https://tinygo.org/getting-started/install/windows/

**Or use Docker (no installation needed):**
```bash
docker run --rm -v $(pwd)/go:/src tinygo/tinygo:latest \
  tinygo build -o /src/main.wasm -target wasm /src/main.go
```

### Build and Run

1. **Build the WASM:**
   ```bash
   cd go
   ./build.sh
   ```

2. **Start the server:**
   ```bash
   cd ..
   python3 serve.py
   ```

3. **Open in browser:**
   ```
   http://localhost:8000
   ```

## 💡 How It Works

### 1. Write Go Code

```go
func model() string {
    lofigui.Reset()
    lofigui.Print("Hello from TinyGo WASM!")
    lofigui.Markdown("## Features")
    // ... more code ...
    return lofigui.Buffer()
}

func main() {
    // Expose to JavaScript
    js.Global().Set("goRunModel", js.FuncOf(runModel))
    js.Global().Call("wasmReady")
    <-make(chan struct{})  // Keep running
}
```

### 2. Compile with TinyGo

```bash
tinygo build -o main.wasm -target wasm main.go
```

TinyGo optimizations:
- Aggressive dead code elimination
- Smaller runtime
- Optimized for size over speed
- No reflection overhead

### 3. Load in Browser

```javascript
const go = new Go();
const result = await WebAssembly.instantiateStreaming(
    fetch('main.wasm'),
    go.importObject
);
go.run(result.instance);
```

### 4. Call from JavaScript

```javascript
const html = goRunModel();  // Calls Go function
document.getElementById('output').innerHTML = html;
```

## 🔍 Key Differences from Standard Go

### Advantages
- **20x smaller** binaries (~100KB vs ~2MB)
- **Instant loading** - critical for mobile
- **Lower bandwidth** - perfect for edge deployment
- **Same Go syntax** - familiar to Go developers

### Limitations
- **Smaller stdlib** - not all packages supported
- **No goroutines scheduler** - single-threaded
- **Limited reflection** - some dynamic features unavailable
- **Simpler GC** - less sophisticated than standard Go

### Supported Features
- ✅ Most of `fmt`, `strings`, `strconv`
- ✅ Basic `syscall/js` for DOM interaction
- ✅ Structs, interfaces, methods
- ✅ Slices, maps, arrays
- ✅ Most standard library

### Not Supported
- ❌ Full reflection package
- ❌ Some `os` and `io` features
- ❌ Complex concurrency (goroutines work but limited)
- ❌ Some stdlib packages (check TinyGo docs)

## 📊 When to Use TinyGo vs Standard Go

### Use TinyGo When:
- Size is critical (mobile, slow networks)
- Instant loading is required
- Deploying to GitHub Pages or CDN
- Building lightweight tools/demos
- Bandwidth costs matter

### Use Standard Go When:
- Need full stdlib support
- Heavy use of reflection
- Complex concurrency required
- Size is not a concern
- Need all Go features

## 🛠️ Development Workflow

### Local Development

```bash
# Build
cd go && ./build.sh

# Serve
cd .. && python3 serve.py

# Or use task (if configured)
task tinygo-wasm
```

### Docker Build (No Installation)

```bash
docker run --rm -v $(pwd)/go:/src tinygo/tinygo:latest \
  tinygo build -o /src/main.wasm -target wasm /src/main.go

# Copy wasm_exec.js
docker run --rm -v $(pwd)/go:/src tinygo/tinygo:latest \
  cp /usr/local/tinygo/targets/wasm_exec.js /src/
```

## 🌐 Deployment

### GitHub Pages

1. Build the WASM:
   ```bash
   cd go && ./build.sh
   ```

2. Push these files to your repo:
   ```
   templates/index.html
   templates/app.js
   go/main.wasm
   go/wasm_exec.js
   ```

3. Enable GitHub Pages in repo settings

4. Access at: `https://yourusername.github.io/yourrepo/`

### Static Site Hosting

Works with any static hosting:
- Netlify
- Vercel
- AWS S3 + CloudFront
- Azure Static Web Apps
- Google Cloud Storage

Just upload the HTML, JS, and WASM files!

## 🔧 Customization

### Add New Functions

```go
// In main.go
func myCustomFunction(this js.Value, args []js.Value) interface{} {
    lofigui.Reset()
    lofigui.Print("Custom output!")
    return js.ValueOf(lofigui.Buffer())
}

func main() {
    js.Global().Set("myCustom", js.FuncOf(myCustomFunction))
    // ... rest of main
}
```

```javascript
// In app.js
function runCustom() {
    const result = myCustom();
    displayOutput(result);
}
```

## 📚 Resources

- **TinyGo Docs**: https://tinygo.org/
- **TinyGo WASM Guide**: https://tinygo.org/docs/guides/webassembly/
- **Supported Packages**: https://tinygo.org/docs/reference/lang-support/
- **Lofigui Docs**: https://github.com/drummonds/lofigui

## 🤝 Comparison with Other Examples

| Feature | TinyGo (04) | Go (03) | Pyodide (03_python) |
|---------|-------------|---------|---------------------|
| Size | ~100KB | ~2MB | ~10MB |
| Startup | Instant | <100ms | 3-5s |
| Language | Go | Go | Python |
| Stdlib | Limited | Full | Full |
| Best For | Size-critical | Full features | Python libs |

## 💻 Example Output

When you run the basic example, you'll see:
- Welcome message from TinyGo
- Feature comparison table
- Benefits of TinyGo for WASM
- Live computation examples

The advanced example demonstrates:
- Data processing in Go
- Table generation
- Number formatting
- Real-time computation

## 🐛 Troubleshooting

**TinyGo not found:**
```bash
# macOS
brew install tinygo

# Check installation
tinygo version
```

**WASM not loading:**
- Check browser console for errors
- Ensure wasm_exec.js is in the same directory
- Verify CORS headers if serving from different domain

**Package not supported:**
- Check TinyGo supported packages list
- Consider alternative approaches
- Some features may need workarounds

**Build fails:**
```bash
# Clean and rebuild
cd go
rm -f main.wasm wasm_exec.js
./build.sh
```

## 📝 License

MIT - See project root for details

## 🎉 Try It!

This example shows that you can build incredibly lightweight web applications with Go. The entire application (HTML + JS + WASM) is smaller than many JavaScript libraries!

Perfect for:
- Mobile-first applications
- GitHub Pages sites
- Documentation with live examples
- Fast-loading demos
- Bandwidth-constrained environments
