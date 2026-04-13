<style>
.annotation { border-left: 3px solid #3273dc; background: #f0f4ff; padding: 0.75em 1em; margin: 0.75em 0; border-radius: 0 4px 4px 0; font-size: 0.9em; }
.annotation strong { color: #3273dc; }
</style>

# 01a — Hello World Explicit

Same model as [example 01](../01_hello_world/index.html), but with everything unbundled. Where 01 uses `app.Run()` (one call does everything), this example shows the explicit wiring: custom template, separate route handlers, and a service worker for the WASM build.

<div class="buttons">
<a href="demo-sw.html" class="button is-primary">Launch Service Worker Demo</a>
<a target="_blank" href="https://codeberg.org/hum3/lofigui/src/branch/main/examples/01a_hello_world_explicit" class="button is-light">Source on Codeberg</a>
</div>

---

## The model — unchanged

The model is identical to example 01. It lives in `model.go` and is shared by both server and WASM builds:

```go
func model(app *lofigui.App) {
    lofigui.Print("Hello world.")
    for i := 0; i < 5; i++ {
        app.Sleep(1 * time.Second)
        lofigui.Printf("Count %d", i)
    }
    lofigui.Print("Done.")
}
```

---

## Explicit routes — setupRoutes()

Instead of `app.Run()`, all routes are registered explicitly in `setupRoutes()`, which returns a standard `*http.ServeMux`:

```go
func setupRoutes(app *lofigui.App) *http.ServeMux {
    mux := http.NewServeMux()

    // GET / — display current state
    mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
        app.HandleDisplay(w, r)
    })

    // POST /start — reset buffer, start model, redirect to /
    mux.HandleFunc("POST /start", func(w http.ResponseWriter, r *http.Request) {
        app.HandleRoot(w, r, model, true)
    })

    // POST /cancel — cancel running action, redirect to /
    mux.HandleFunc("POST /cancel", app.HandleCancel("/"))

    mux.HandleFunc("GET /favicon.ico", lofigui.ServeFavicon)
    return mux
}
```

<div class="annotation">
<strong>HandleRoot vs Handle</strong> — Example 01 uses <code>Handle()</code> which auto-starts the model on the first request and auto-shuts down the server when done. Here, <code>HandleRoot</code> starts the model explicitly on <code>POST /start</code>, and <code>HandleDisplay</code> renders the current state on <code>GET /</code>. The model is started by a form submit, not by the first page visit.
</div>

<div class="annotation">
<strong>HandleCancel</strong> — In example 01, <code>Run()</code> registers this internally. Here it's explicit: <code>POST /cancel</code> cancels the running action and redirects to <code>/</code>.
</div>

---

## Custom template

The template is defined as a Go string constant in `routes.go` (so it works in both server and WASM builds). It uses Go's `html/template` syntax with the standard lofigui variables:

```html
<nav class="navbar is-primary">
  <div class="navbar-brand">
    <span class="navbar-item has-text-weight-bold">{{.version}}</span>
  </div>
  <div class="navbar-end">
    {{if eq .polling "Running"}}
      <span class="tag is-warning">Running</span>
      <form action="/cancel" method="post">
        <button class="tag is-danger is-light" type="submit">Cancel</button>
      </form>
    {{else}}
      <span class="tag is-success">Ready</span>
    {{end}}
  </div>
</nav>
<section class="section">
  <div class="container content">
    {{.results}}
    {{if ne .polling "Running"}}
      <form action="/start" method="post">
        <button class="button is-success" type="submit">Start</button>
      </form>
    {{end}}
  </div>
</section>
```

<div class="annotation">
<strong>Form-based actions</strong> — Start and Cancel are HTML forms with <code>method="post"</code>. No JavaScript needed. The template conditionally shows the Start button when idle and the Cancel button when running.
</div>

<div class="annotation">
<strong>{{.polling}}</strong> — <code>"Running"</code> while the model is active, <code>"Stopped"</code> when idle. Combined with the HTTP <code>Refresh</code> header (set by <code>HandleDisplay</code>), the page auto-refreshes while running and stops when done.
</div>

---

## Server vs WASM: same handlers, different entry points

The key insight: `setupRoutes()` is called by both builds. Only the entry point differs.

### Server (main.go)

```go
func main() {
    app := lofigui.NewApp()
    app.Version = "Hello World Explicit v1.0"
    app.SetDisplayURL("/")

    ctrl, _ := lofigui.NewController(lofigui.ControllerConfig{
        TemplateString: helloTemplate,
        Name:           "Hello World Explicit",
    })
    app.SetController(ctrl)

    mux := setupRoutes(app)
    log.Fatal(http.ListenAndServe(":1341", mux))
}
```

### WASM (main_wasm.go)

```go
func main() {
    app := lofigui.NewApp()
    app.Version = "Hello World Explicit v1.0"
    app.SetDisplayURL("/")

    ctrl, _ := lofigui.NewController(lofigui.ControllerConfig{
        TemplateString: helloTemplate,
        Name:           "Hello World Explicit",
    })
    app.SetController(ctrl)

    mux := setupRoutes(app)
    wasmhttp.Serve(mux)  // ← only this line differs
}
```

<div class="annotation">
<strong>wasmhttp.Serve(mux)</strong> replaces <code>http.ListenAndServe()</code>. The <a href="https://github.com/nlepage/go-wasm-http-server">go-wasm-http-server</a> library registers a service worker that intercepts browser fetch events and routes them through the Go <code>*http.ServeMux</code> running inside WASM. The browser makes real HTTP requests; the service worker answers them from Go.
</div>

---

## JavaScript: direct exports vs service worker

This is the fundamental difference between example 01 and 01a's WASM builds:

### Example 01 — direct `syscall/js` exports (app.js)

```js
// Go exports functions to JavaScript via RunWASM():
//   goStart(), goRender(), goCancel(), goIsRunning()

// JavaScript actively polls Go for updates:
renderInterval = setInterval(function() {
    outputDiv.innerHTML = goRender();    // pull HTML from Go
    updateStatus();                       // check goIsRunning()
}, 500);

// Button clicks call Go directly:
startBtn.addEventListener('click', function() { goStart(); });
cancelBtn.addEventListener('click', function() { goCancel(); });
```

<div class="annotation">
<strong>01 pattern</strong>: JavaScript is the driver. It calls Go functions on a timer, manages button state, and renders HTML from Go's buffer. The page never makes HTTP requests — everything happens through direct function calls between JS and WASM.
</div>

### Example 01a — service worker (sw.js + demo-sw.html)

```js
// sw.js — the entire WASM integration:
importScripts('wasm_exec.js');
importScripts('https://cdn.jsdelivr.net/.../sw.js');

registerWasmHTTPListener('main.wasm', {
    base: scope,
    passthrough: function(request) {
        // External CDN resources pass through to network
        if (url.hostname !== self.location.hostname) return true;
        if (url.pathname.endsWith('demo-sw.html')) return true;
        return false;
    }
});
```

<div class="annotation">
<strong>01a pattern</strong>: The service worker is the driver. It loads the Go WASM binary and intercepts all fetch events. When the browser requests <code>/</code>, the service worker routes it to Go's <code>HandleDisplay</code>. When a form POSTs to <code>/start</code>, the service worker routes it to Go's <code>HandleRoot</code>. No custom JavaScript, no polling logic, no button state management — just standard HTML forms and HTTP semantics.
</div>

### When to use which

| Approach | Best for | Trade-off |
|----------|----------|-----------|
| **Direct exports** (01) | Simple apps, single-page output, no routing | Requires custom JS per app; JS manages all state |
| **Service worker** (01a) | Multi-route apps, forms, HTMX | Requires service worker support; slightly larger binary |

The service worker approach scales to complex apps (examples 09-12) because adding a new route is just adding a handler in `setupRoutes()` — no JS changes needed. The direct export approach stays simpler for single-page demos where you just need `goStart()` and `goRender()`.

---

## Running

```bash
# Server mode
task go-example:01a
# Opens http://localhost:1341 — click Start, watch output, click Cancel

# WASM demo (via docs)
task docs:build-wasm
tp pages
# Navigate to 01a_hello_world_explicit/demo-sw.html
```
