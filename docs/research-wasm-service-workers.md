# Research: WASM Service Workers

Go compiled to WebAssembly, running inside service workers as statically hosted web servers. Focus on routing, HTMX integration, and what this means for lofigui.

## The idea

A service worker intercepts every `fetch` event in its scope. If a Go program compiled to WASM handles those requests using standard `net/http`, you have a web server running entirely in the browser. The site can be hosted on GitHub Pages, statichost.eu, or any static file host --- no server process needed.

```
Browser fetch("/api/data")
  -> Service worker intercepts
  -> Routes to Go WASM (net/http handlers)
  -> Returns HTML response
  -> Browser renders it
```

The key library enabling this is [go-wasm-http-server](https://github.com/nlepage/go-wasm-http-server) (v2, Apache 2.0, ~400 stars). It bridges Go's `net/http` to service worker fetch events.

## Go WASM: two targets, one relevant

Go has two WASM compilation targets:

| Target | Since | Purpose | Browser? |
|--------|-------|---------|----------|
| `GOOS=js GOARCH=wasm` | Go 1.11 | Browser via `syscall/js` | Yes |
| `GOOS=wasip1 GOARCH=wasm` | Go 1.21 | Standalone runtimes (wasmtime, wazero) | No |

Service workers need the `js/wasm` target. The `wasip1` target is for server-side WASM runtimes and edge computing (Cloudflare Workers, etc.).

### Recent Go WASM changes

**Go 1.22 (Feb 2024):** Enhanced `ServeMux` with method matching and path parameters --- `mux.HandleFunc("GET /items/{id}", handler)`. This makes the standard library router viable for most routing needs inside WASM, eliminating the need for third-party routers.

**Go 1.24 (Feb 2025):** Added `go:wasmexport` directive and WASI reactor build mode. These are `wasip1` features (not browser-relevant) but signal Go's investment in WASM as a first-class target.

**Go 1.26 (Feb 2026):** Runtime manages heap memory in smaller increments, significantly reducing memory usage for apps with heaps under ~16 MiB. Unconditionally uses Wasm 2.0 instructions (sign extension, non-trapping float-to-int). Good news for service worker memory footprint.

**WASI P3 (proposed, mid 2026):** Composable concurrency for WASI --- goroutines that can block on I/O without blocking others. Server-side focused, not directly relevant to browser service workers, but shows the direction of Go+WASM.

## How it works: Go side

Standard `net/http` handlers, with `wasmhttp.Serve()` replacing `http.ListenAndServe()`:

```go
//go:build js && wasm

package main

import (
    "embed"
    "io/fs"
    "net/http"

    wasmhttp "github.com/nlepage/go-wasm-http-server/v2"
)

//go:embed static/*
var staticFiles embed.FS

func main() {
    mux := http.NewServeMux()

    // Standard routing --- method + path patterns (Go 1.22+)
    mux.HandleFunc("GET /",             handleIndex)
    mux.HandleFunc("GET /items/{id}",   handleGetItem)
    mux.HandleFunc("POST /items",       handleCreateItem)
    mux.HandleFunc("DELETE /items/{id}", handleDeleteItem)

    // Static files embedded in the WASM binary
    sub, _ := fs.Sub(staticFiles, "static")
    mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(sub))))

    wasmhttp.Serve(mux)
}
```

Everything that works with `net/http` works here: URL parameters (`r.PathValue("id")`), query strings (`r.URL.Query()`), form data (`r.FormValue()`), request bodies, headers, cookies. Third-party routers like Echo and Gin compile to `js/wasm`. chi and gorilla/mux should work (standard `net/http` compatible) but are untested.

## How it works: service worker side

Three files on the static host:

**`sw.js`** (service worker):
```javascript
importScripts('wasm_exec.js')
importScripts('https://cdn.jsdelivr.net/gh/nlepage/go-wasm-http-server@v2.2.1/sw.js')

registerWasmHTTPListener('server.wasm', {
    // Optional: let CDN requests pass through to the network
    passthrough: (request) => request.url.includes('cdn.jsdelivr.net')
})
```

**`index.html`** (registers the service worker):
```html
<script>
if ('serviceWorker' in navigator) {
    navigator.serviceWorker.register('sw.js')
        .then(() => navigator.serviceWorker.ready)
        .then(() => { /* WASM server is ready */ });
}
</script>
```

**`wasm_exec.js`** (from Go stdlib --- must match compiler version exactly):
```bash
cp "$(go env GOROOT)/misc/wasm/wasm_exec.js" .
```

The `passthrough` callback is important: requests for external resources (CDN CSS, CDN JS) should not be routed to the Go WASM binary.

## Binary size and compression

For a service worker, binary size directly affects cold-start time. Standard Go WASM binaries compress extremely well, and browsers transparently decompress before passing bytes to `WebAssembly.instantiateStreaming()`.

| Method | Raw size (typical lofigui) | Transfer size | Ratio |
|--------|---------------------------|---------------|-------|
| Uncompressed | 7.8 MB | 7.8 MB | 100% |
| gzip -9 | 7.8 MB | 2.1 MB | 27% |
| brotli -11 | 7.8 MB | ~1.7 MB (est.) | ~22% |

Most static hosts apply gzip automatically. Pre-compressing gives control over level and enables brotli where supported.

**Build flags** provide modest additional savings: `-ldflags="-s"` (strips WASM name section, ~186 KB) and `-trimpath` (shortens paths, ~10 KB). `-ldflags="-w"` has zero effect on WASM (no DWARF sections).

For a comparison with TinyGo's smaller binaries and why standard Go with compression was chosen instead, see [WASM TinyGo research](research-wasm-tinygo.html).

## HTMX with WASM service workers

This is the compelling integration. HTMX makes standard HTTP requests (`hx-get`, `hx-post`). Service workers intercept all fetch requests. HTMX has no idea whether responses come from a real server or a WASM binary --- it just sees HTTP responses containing HTML fragments.

```
User clicks [hx-get="/fragment/tank"]
  -> Browser fetch("/fragment/tank")
  -> Service worker intercepts
  -> Go handler renders HTML fragment
  -> Service worker returns response
  -> HTMX swaps into target div
```

**What works:**
- All HTMX attributes: `hx-get`, `hx-post`, `hx-put`, `hx-delete`, `hx-trigger`, `hx-target`, `hx-swap`, `hx-push-url`
- HTMX polling (`hx-trigger="every 1s"`) --- each poll is intercepted
- `hx-boost` for form submissions
- Server-Sent Events (SSE) --- `go-wasm-http-server` supports SSE

**What does not work:**
- `hx-ws` (WebSocket extension) --- service workers **cannot** intercept WebSocket connections (fundamental browser limitation)

**Confirmed projects:**
- [go-wasm-htmx-service-worker](https://github.com/ocomsoft/go-wasm-htmx-service-worker) --- PoC with Go WASM + HTMX + templ + service worker
- [todos-htmx-wasm](https://github.com/stackus/todos-htmx-wasm) --- Todo app with HTMX frontend and Go WASM BFF proxy
- [Local First HTMX](https://elijahm.com/posts/local_first_htmx/) --- Echo router compiled to WASM, same handlers work on server and in browser

## What this means for lofigui

### Current WASM approach

lofigui's WASM examples (01, 02, 03, 07, 08, 12) use direct `syscall/js` function exports:

```go
js.Global().Set("goRender", js.FuncOf(renderFunc))
js.Global().Set("goStart", js.FuncOf(startFunc))
```

JavaScript calls these functions directly and puts the HTML into the DOM. Routing is handled client-side in JavaScript (tab switching, SVG link interception). This works but means:

- Every new "route" requires a new exported Go function AND JavaScript glue
- The HTML/JS interaction pattern diverges from the server-side pattern
- No standard HTTP semantics (no GET/POST, no URL parameters, no form handling)

### Service worker approach

With a service worker, the **same Go HTTP handlers** serve both server and browser:

```go
// Shared handler --- works on server AND in WASM service worker
func handleTankFragment(w http.ResponseWriter, r *http.Request) {
    lofigui.Reset()
    lofigui.HTML(sim.buildSVG())
    renderTankStatus()
    w.Header().Set("Content-Type", "text/html; charset=utf-8")
    fmt.Fprint(w, lofigui.Buffer())
}
```

The HTMX template is identical in both deployments. The routing is standard `net/http`. The only difference is the entry point:

```go
// Server build
http.HandleFunc("GET /fragment/tank", handleTankFragment)
http.ListenAndServe(":8080", nil)

// WASM build
http.HandleFunc("GET /fragment/tank", handleTankFragment)
wasmhttp.Serve(nil)
```

### Trade-offs

| | Current (direct JS) | Service worker |
|---|---|---|
| Binary size (raw) | ~2 MB | ~8 MB (includes net/http) |
| Binary size (gzipped) | ~500 KB | ~2.1 MB |
| Cold start | Fast (WASM loaded directly) | Slower (service worker + WASM init) |
| Code sharing | Separate `main_wasm.go` per example | Same handlers, different entry point |
| Routing | JavaScript state + exported functions | Standard `net/http` routing |
| HTMX | Not applicable (no HTTP in browser) | Full HTMX support |
| Complexity | Lower (no service worker lifecycle) | Higher (service worker registration, caching, lifecycle) |
| Offline | Works once WASM loaded | Works once service worker cached |

### Where each approach fits on the interactivity spectrum

| Level | Current approach | Service worker approach |
|-------|-----------------|----------------------|
| 1 (Teletype) | Direct JS: `goStart()` + `setInterval(goRender, 500)` | Overkill --- adds net/http weight for no benefit |
| 2 (Teletype+ web) | Direct JS with template inheritance (example 03) | Could work --- forms POST to service worker |
| 3 (Polling) | Direct JS: `setInterval` replaces HTTP Refresh | Service worker + `<meta Refresh>` would work but is odd |
| 4 (HTMX) | Not currently possible in WASM | **Natural fit** --- HTMX requests routed through service worker |

**The service worker approach is most compelling at Level 4** (HTMX), where it enables the same server-side handlers to work in both deployments. At lower levels, direct JS exports are simpler and produce smaller binaries.

## Limitations and constraints

**Service worker lifecycle:** Browsers may terminate idle service workers. In-memory state is lost. For persistent state, use IndexedDB. The `go-wasm-http-server` keepalive demo shows approaches to extend lifetime.

**No real network access:** The Go code inside the service worker cannot make outbound HTTP requests to external APIs. It can only respond to intercepted fetch events. For apps that need a real backend, the WASM service worker acts as a BFF (Backend-For-Frontend) proxy --- but the "backend" must be reached via `passthrough`.

**No WebSockets:** Service workers fundamentally cannot intercept WebSocket connections. SSE works.

**HTTPS required:** Service workers only register over HTTPS (or localhost). Fine for statichost.eu, GitHub Pages, etc.

**Firefox Private Browsing:** Service workers are disabled. The site falls back to... nothing, unless there is a server.

**Force reload (Ctrl+Shift+R):** Bypasses service workers. The page loads without the Go server.

**`wasm_exec.js` version coupling:** Must exactly match the Go compiler version used to build the WASM binary. A version mismatch causes silent failures.

**Memory:** WASM modules are limited to 4 GB (32-bit addressing). In practice, a service worker running Go WASM uses 10--50 MB depending on the application. Go 1.26 improved heap management for sub-16 MiB heaps.

## Concurrency in browser WASM

Go WASM runs on a single thread. Goroutines are cooperatively scheduled:

- `time.Sleep()` suspends the Go runtime and returns control to the browser event loop (essential for UI responsiveness --- this is what lofigui's `Yield()` does)
- `runtime.Gosched()` yields between Go goroutines but does NOT return control to the browser
- Channel operations and select statements work normally
- There is no parallelism --- goroutines interleave, they do not run simultaneously

For service worker request handling, each request can be a goroutine. They will be processed sequentially (single thread) but the programming model is the same as server-side Go.

## Conclusion: adopting go-wasm-http-server for lofigui

### Decision: unified WASM via service workers

lofigui will adopt `go-wasm-http-server` as the standard approach for all WASM builds. The primary motivation is **eliminating drift between server and WASM code paths**.

Currently, each WASM example has a separate `main_wasm.go` that re-implements the server's behaviour using `syscall/js` function exports and custom JavaScript glue. The server version uses `net/http` handlers, the WASM version uses `js.Global().Set("goRender", ...)`. Over time these diverge --- the server gets new features, the WASM build lags behind, and every example duplicates the bridging boilerplate.

With `go-wasm-http-server`, the same `net/http` handlers serve both deployments. The only code that differs is the entry point:

```go
// +build !js !wasm
func main() { http.ListenAndServe(":8080", nil) }

// +build js,wasm
func main() { wasmhttp.Serve(nil) }
```

Routing, form handling, HTMX fragment endpoints, template rendering --- all shared. One set of handlers, two targets.

### TinyGo: out of scope

TinyGo cannot be used for the service worker approach because `net/http` panics at runtime in browser WASM. Standard Go with gzip compression closes the binary size gap sufficiently. See [WASM TinyGo research](research-wasm-tinygo.html) for the full analysis and a research proposal to make TinyGo viable in the future.

### Binary size and compression

See [Binary size and compression](#binary-size-and-compression) above for transfer-size numbers. The uncompressed binary (~8 MB) is dominated by fixed overhead:

| Component | Size contribution |
|-----------|------------------|
| Go runtime + `syscall/js` | 1.7 MB (irreducible floor) |
| `net/http` + `crypto/tls` | +1.8 MB (TLS precomputed tables) |
| pongo2 + blackfriday | +3.9 MB (template + markdown engines) |
| Application code | negligible |

**Build tag optimisation**: lofigui's `app.go`, `controller.go`, `favicon.go`, and `serve.go` currently have no build tags excluding them from `js/wasm` builds. Adding `//go:build !(js && wasm)` to server-only files would eliminate `net/http` and `crypto/tls` from non-service-worker WASM builds, dropping the binary from ~8 MB to ~4 MB (~1.2 MB gzipped). This optimisation is separate from the service worker migration and could be done first.

### Security audit of go-wasm-http-server

An audit of the library (v2.2.1, 795 lines Go + 95 lines JS) identified the following:

**Strengths:**
- Small, auditable codebase --- can be read entirely in one sitting
- Shallow dependency tree: 3 direct dependencies, zero transitive
- Architecturally sound --- leverages the browser's own security model for scope, CORS, and CSP enforcement
- Apache 2.0 license with no concerns

**Concerns:**

| Issue | Severity | Detail |
|-------|----------|--------|
| **Bus factor of 1** | Medium | Single maintainer (103 of 107 commits). No succession plan. |
| **No CI or test suite** | Medium | One example doc test exists. No functional tests for request/response serialization, edge cases, or streaming. |
| **Two unmaintained dependencies** | Low-Medium | `hack-pad/safejs` (3+ years stale), `go-js-promise` (4.5+ years stale, same author). Both are small and stable, but no one is watching for issues. |
| **Multi-value response headers dropped** | Low | `response.go` `headerValue()` uses `h.Get(k)` which returns only the first value. Multiple `Set-Cookie` headers are silently lost. |
| **No request body size limits** | Low | A malicious page in scope could POST a multi-GB body to exhaust WASM memory. This is inherent to the architecture, not unique to this library. The Go handler must enforce limits. |
| **Unreleased work on master** | Info | Multi-WASM support and Referer/Host/RequestURI fixes from Nov 2025 have no tagged release. |
| **No SECURITY.md** | Info | No responsible disclosure process documented. |

**Risk rating for lofigui: Low-Medium.** The browser sandbox is the ultimate security boundary. The main risks are reliability (no tests, single maintainer) rather than exploitable vulnerabilities. For demo and internal tool use --- lofigui's target --- the risks are acceptable.

**Mitigations to consider:**
- Vendor the library (or fork) to insulate against upstream abandonment
- Add `http.MaxBytesReader` in Go handlers to limit POST body size
- Pin `wasm_exec.js` to the exact Go compiler version in the build script
- Monitor the upstream repo; if the maintainer goes silent for 12+ months, fork

### Work plan

This is a significant migration. Recommended phases:

**Phase 1: Build tag cleanup** (small, immediate value)
- Add `//go:build !(js && wasm)` to lofigui's server-only files (`app.go`, `controller.go`, `favicon.go`, `serve.go`)
- Verify existing WASM examples still build and run
- Measure binary size reduction (~8 MB -> ~4 MB)

**Phase 2: Compression pipeline**
- Write a small Go tool or Taskfile step that gzip-compresses `.wasm` files
- Evaluate whether `andybalholm/brotli` is worth adding for ~20% better compression
- Update `build.sh` scripts and Taskfile to include compression
- Test with `WebAssembly.instantiateStreaming` to confirm transparent decompression

**Phase 3: Service worker proof of concept**
- Pick one HTMX example (09 or 10) as the pilot
- Add `go-wasm-http-server` dependency
- Create `main_sw.go` (service worker entry point) alongside existing `main.go` (server) and `main_wasm.go` (direct JS)
- Write `sw.js` with passthrough for CDN resources (Bulma, HTMX)
- Verify HTMX partial updates work through the service worker
- Measure cold start time, memory usage, and binary size

**Phase 4: Unify handlers**
- Refactor the pilot example so server and service worker share the same `net/http` handlers
- Remove the direct JS export `main_wasm.go` for that example
- Document the shared handler pattern

**Phase 5: Roll out to all examples**
- Migrate remaining WASM examples to the service worker approach
- Remove per-example `app.js` JavaScript glue where it becomes unnecessary
- Update `docs/examples.md` and the interactivity spectrum documentation

**Phase 6: Evaluate and document**
- Measure final binary sizes across all examples (uncompressed, gzipped, brotli)
- Document cold start times on target static hosts
- Write up the migration as a reference for other lofigui users
- Decide whether to vendor or fork `go-wasm-http-server`

Each phase is independently useful. Phase 1 benefits all WASM builds immediately. Phase 3 validates the approach before committing to a full migration.

## References

- [go-wasm-http-server](https://github.com/nlepage/go-wasm-http-server) --- the foundational library
- [Emulate a Go HTTP server in your browser](https://dev.to/nlepage/emulate-a-go-http-server-in-your-browser-32) --- introduction article
- [go-wasm-htmx-service-worker](https://github.com/ocomsoft/go-wasm-htmx-service-worker) --- HTMX + Go WASM PoC
- [todos-htmx-wasm](https://github.com/stackus/todos-htmx-wasm) --- HTMX + Go WASM todo app
- [Local First HTMX](https://elijahm.com/posts/local_first_htmx/) --- Echo router in WASM
- [Go Wiki: WebAssembly](https://go.dev/wiki/WebAssembly) --- official Go WASM docs
- [Go 1.24 WASM changes](https://go.dev/blog/wasmexport)
- [Go 1.26 release notes](https://go.dev/doc/go1.26)
- [WASM TinyGo research](research-wasm-tinygo.html) --- TinyGo analysis, blockers, and research proposal
- [Minimizing Go WASM binary size](https://dev.bitolog.com/minimizing-go-webassembly-binary-size/)
