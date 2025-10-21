# Example 03: Hello World WASM

Python running as WebAssembly in the browser using Pyodide and Web Workers.

## Overview

This example demonstrates:
- Python code compiled to WASM and running entirely in the browser
- Using Pyodide to execute Python without a server
- Web Workers for non-blocking Python execution
- Perfect for hosting on GitHub Pages
- No backend server required

## Structure

```
03_hello_world_wasm/
├── index.html         # Main HTML page
├── app.js            # Main application logic
├── worker.js         # Web Worker for Pyodide
├── hello_model.py    # Python model code (runs in browser)
├── pyproject.toml    # Project metadata
└── readme.md         # This file
```

## How It Works

### Architecture

1. **HTML Page** (`index.html`): User interface with Bulma CSS styling
2. **Main App** (`app.js`): Manages UI and communicates with the Web Worker
3. **Web Worker** (`worker.js`): Loads Pyodide and executes Python code off the main thread
4. **Python Model** (`hello_model.py`): Business logic that runs as WASM in the browser

### Execution Flow

```
User clicks button
    ↓
app.js sends message to worker
    ↓
worker.js executes Python code via Pyodide
    ↓
Python returns HTML string
    ↓
worker.js sends result back to app.js
    ↓
app.js displays HTML in the page
```

## Running Locally

Since this uses Web Workers and fetch API, you need to serve it over HTTP (not file://).

### Option 1: Python HTTP Server

```bash
cd examples/03_hello_world_wasm
python -m http.server 8000
```

Then open: http://localhost:8000

### Option 2: Using Node.js http-server

```bash
npm install -g http-server
cd examples/03_hello_world_wasm
http-server -p 8000
```

### Option 3: Using uv (if available)

```bash
cd examples/03_hello_world_wasm
uv run python -m http.server 8000
```

## Deploying to GitHub Pages

### Step 1: Enable GitHub Pages

1. Go to your repository on GitHub
2. Navigate to Settings → Pages
3. Under "Source", select the branch you want to deploy (e.g., `main`)
4. Select the folder: `/` (root) or `/docs`
5. Click Save

### Step 2: Option A - Deploy from root

Copy the example files to a `docs` folder in your repo root:

```bash
mkdir -p docs
cp examples/03_hello_world_wasm/* docs/
git add docs
git commit -m "Add WASM example to GitHub Pages"
git push
```

### Step 2: Option B - Deploy this directory directly

Set GitHub Pages to deploy from `examples/03_hello_world_wasm/` if your Pages settings allow custom paths.

### Step 3: Access Your Page

Your page will be available at:
```
https://YOUR-USERNAME.github.io/YOUR-REPO-NAME/
```

Or if in a subfolder:
```
https://YOUR-USERNAME.github.io/YOUR-REPO-NAME/examples/03_hello_world_wasm/
```

## Modifying the Python Code

Edit [`hello_model.py`](hello_model.py) to change the Python logic:

```python
def model():
    output = []
    output.append("<p>Your custom HTML here</p>")
    # Add your Python logic
    return "\n".join(output)
```

The Python code:
- Must return HTML as a string
- Can use any Python standard library features
- Can use Pyodide-compatible packages
- Executes entirely in the browser

## Adding Python Packages

Pyodide includes many popular packages. To use them:

```python
# In hello_model.py
def model_with_numpy():
    import numpy as np

    arr = np.array([1, 2, 3, 4, 5])
    output = [f"<p>Array sum: {arr.sum()}</p>"]
    output.append(f"<p>Array mean: {arr.mean()}</p>")

    return "\n".join(output)
```

Then in `worker.js`, load the package:

```javascript
// In initPyodide function
await pyodide.loadPackage('numpy');
```

See [Pyodide packages](https://pyodide.org/en/stable/usage/packages-in-pyodide.html) for available packages.

## Performance Notes

### First Load
- Pyodide (~6-10 MB) downloads on first visit
- Browser caches it for subsequent visits
- Takes 2-5 seconds to initialize

### Subsequent Runs
- Python execution is fast (compiled to WASM)
- Web Worker keeps UI responsive
- No network requests needed after initial load

## Limitations

1. **Package Size**: Pyodide is large (~6-10 MB initial download)
2. **Package Availability**: Not all Python packages work in Pyodide
3. **No Server Features**: Can't use FastAPI, uvicorn, or server-side features
4. **File I/O**: Limited file system access (virtual filesystem only)
5. **Network**: Must use fetch API, not traditional Python requests

## Benefits

1. **No Backend**: Runs entirely in browser, no server costs
2. **Free Hosting**: GitHub Pages is free
3. **Offline Capable**: Can be made into a PWA
4. **Python in Browser**: Full Python language support
5. **Secure**: Sandboxed execution environment

## Integrating with lofigui

To integrate this with the lofigui library:

1. **Server-side rendering**: Use lofigui normally with FastAPI (examples 01 & 02)
2. **Static export**: Generate static HTML from lofigui, deploy to GitHub Pages
3. **Hybrid**: Use this WASM approach for demos, server-side for production

## Troubleshooting

### "Failed to fetch worker.js"
- Make sure you're serving over HTTP, not opening file:// directly
- Check that all files are in the same directory

### "Pyodide failed to load"
- Check your internet connection (needs to download from CDN)
- Try a different browser (Chrome, Firefox, Edge recommended)
- Check browser console for detailed errors

### "Module not found"
- Make sure `hello_model.py` is in the same directory as `index.html`
- Check that the filename is exactly `hello_model.py`

### Python execution is slow
- First run is always slower (Pyodide initialization)
- Complex computations take time (WASM is fast but not native speed)
- Use Web Workers to keep UI responsive (already implemented)

## Next Steps

- Modify `hello_model.py` to create your own Python logic
- Add more Python packages via `pyodide.loadPackage()`
- Create multiple Python functions and add buttons to call them
- Style the output with Bulma CSS classes
- Deploy to GitHub Pages and share your live demo

## Resources

- [Pyodide Documentation](https://pyodide.org/)
- [Web Workers MDN](https://developer.mozilla.org/en-US/docs/Web/API/Web_Workers_API)
- [GitHub Pages Documentation](https://docs.github.com/en/pages)
- [Bulma CSS Framework](https://bulma.io/)

## Comparison with Server-Side Examples

| Feature | Example 01/02 (Server) | Example 03 (WASM) |
|---------|------------------------|-------------------|
| Requires Server | Yes (FastAPI/Uvicorn) | No |
| Hosting Cost | $$ (VPS/Cloud) | Free (GitHub Pages) |
| Python Packages | All packages | Pyodide-compatible only |
| Latency | Network dependent | Instant (after load) |
| Scalability | Limited by server | Unlimited (client-side) |
| Initial Load | Fast | Slower (downloads Pyodide) |
| Use Case | Production apps | Demos, tools, prototypes |

## License

MIT License - same as lofigui project
