# Multi-Page Displays in lofigui

## For Developers

### What multi-page means

In lofigui, "multi-page" means separate HTTP routes that each reset the global buffer and render different content. Each page calls `lofigui.Reset()` then fills the buffer with its own output before calling `app.HandleDisplay()`.

Example 08 (Water Tank Multi-Page) has two pages:
- **`/`** — schematic with SVG tank, pump/valve controls, status tags
- **`/diagnostics`** — tables of pump cycles, valve cycles, float trips, level history

### The HTTP Refresh header

When the simulation is running, the app calls `app.WriteRefreshHeader(w)` which sets the `Refresh: N` HTTP header. The browser reloads the *current page* after N seconds — no hardcoded URL, so both `/` and `/diagnostics` auto-refresh independently.

The cycle:
1. `POST /start` starts the simulation and calls `app.StartAction()`
2. Any `GET` page renders with the Refresh header while `app.IsActionRunning()` is true
3. `POST /stop` stops the simulation and calls `app.EndAction()`
4. Subsequent `GET` responses omit the Refresh header — polling stops

### Server vs WASM

**Server** gets real URL routing, form POSTs, and auto-refresh headers. Each page is a separate URL with its own render function. The simulation runs in a goroutine on the server.

**WASM** gets zero-infrastructure deployment on GitHub Pages. Multiple "pages" become tabs in a single HTML file. The simulation runs in a goroutine inside the browser's WASM runtime. JavaScript switches which `goRenderX()` function to call based on the active tab.

### Example 08 walkthrough

Server-side (`main.go`):
```go
// GET / — schematic page
http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
    lofigui.Reset()
    sim.renderSchematic()
    app.HandleDisplay(w, r)
})

// GET /diagnostics — diagnostics page
http.HandleFunc("/diagnostics", func(w http.ResponseWriter, r *http.Request) {
    lofigui.Reset()
    sim.renderDiagnostics()
    app.HandleDisplay(w, r)
})
```

WASM equivalent (`main_wasm.go`):
```go
func goRenderSchematic(this js.Value, args []js.Value) any {
    lofigui.Reset()
    lofigui.HTML(sim.buildSVG())
    // ... status tags ...
    return js.ValueOf(lofigui.Buffer())
}

func goRenderDiagnostics(this js.Value, args []js.Value) any {
    lofigui.Reset()
    // ... diagnostic tables ...
    return js.ValueOf(lofigui.Buffer())
}
```

JavaScript tab switching (`app.js`):
```javascript
function render() {
    if (currentTab === 'schematic') {
        outputDiv.innerHTML = goRenderSchematic();
    } else {
        outputDiv.innerHTML = goRenderDiagnostics();
    }
}
```

## For AI/LLM

### Server-side multi-page template

```go
// Each page: Reset → fill buffer → HandleDisplay
http.HandleFunc("/pageA", func(w http.ResponseWriter, r *http.Request) {
    lofigui.Reset()
    renderPageA()
    app.HandleDisplay(w, r)
})
http.HandleFunc("/pageB", func(w http.ResponseWriter, r *http.Request) {
    lofigui.Reset()
    renderPageB()
    app.HandleDisplay(w, r)
})
```

### WASM multi-page template

```go
// One goRenderX per "page"
func goRenderA(this js.Value, args []js.Value) any {
    lofigui.Reset()
    // ... fill buffer ...
    return js.ValueOf(lofigui.Buffer())
}
func goRenderB(this js.Value, args []js.Value) any {
    lofigui.Reset()
    // ... fill buffer ...
    return js.ValueOf(lofigui.Buffer())
}
```

JavaScript dispatches based on tab state:
```javascript
let currentTab = 'a';
function render() {
    outputDiv.innerHTML = currentTab === 'a' ? goRenderA() : goRenderB();
}
```

### Key API calls

| Function | Purpose |
|----------|---------|
| `lofigui.Reset()` | Clear the global buffer — **must** call before each page render |
| `lofigui.Buffer()` | Return accumulated HTML |
| `app.HandleDisplay(w, r)` | Render template with buffer + polling state |
| `app.StartAction()` | Begin polling (sets Refresh header on responses) |
| `app.EndAction()` | Stop polling |

### Constraints

- **Global buffer**: `Reset()` before each render to avoid stale content from previous page
- **Build tags**: Use `//go:build !(js && wasm)` on server `main.go` and `//go:build js && wasm` on WASM `main_wasm.go` for dual-build
- **Shared code**: `simulation.go` (no build tag) is shared between server and WASM builds
