<style>
.annotation { border-left: 3px solid #3273dc; background: #f0f4ff; padding: 0.75em 1em; margin: 0.75em 0; border-radius: 0 4px 4px 0; font-size: 0.9em; }
.annotation strong { color: #3273dc; }
.screenshot { border: 1px solid #dbdbdb; border-radius: 4px; box-shadow: 0 2px 6px rgba(0,0,0,0.1); overflow: hidden; }
.screenshot img { display: block; }
</style>

# 01 — Hello World

If you can write a Go program that prints to stdout, you can write a lofigui app. The model function below is ordinary Go code — `Print()`, a loop, a sleep. The only difference is the output goes to a web page instead of a terminal. No WebSocket, no JavaScript — just the browser's built-in refresh mechanism doing the work.

<div class="buttons">
<a href="wasm_demo/" class="button is-primary">Launch 01 (compact)</a>
<a href="../01a_hello_world_explicit/wasm_demo/sw/" class="button is-primary is-outlined">Launch 01a (explicit)</a>
<a href="../01b_hello_world_explicit_gzip/wasm_demo/sw/" class="button is-primary is-outlined">Launch 01b (explicit + gzip)</a>
<a target="_blank" href="https://codeberg.org/hum3/lofigui/src/branch/main/examples/01_hello_world" class="button is-light">Source on Codeberg</a>
</div>

<div class="annotation">
<strong>Three variants of the same app:</strong> 01 uses the compact <code>app.RunWASM(model)</code> call — the library auto-generates the SW bootstrap so the example itself has no <code>templates/</code> directory. <a href="../01a_hello_world_explicit/">01a</a> is the same behaviour with every wire visible (explicit <code>setupRoutes()</code>, hand-written SW bootstrap). <a href="../01b_hello_world_explicit_gzip/">01b</a> adds gzipped WASM on top of 01a, showing the <code>DecompressionStream</code> + cache plumbing.
</div>

<div class="columns is-vcentered">
<div class="column is-5">
<figure class="image screenshot">
<img src="../01_polling.svg" alt="During polling — partial output">
<figcaption class="has-text-centered has-text-grey is-size-7 mt-1">During polling</figcaption>
</figure>
</div>
<div class="column is-narrow has-text-centered">
<span style="font-size: 2rem; color: #999;">&rarr;</span>
</div>
<div class="column is-5">
<figure class="image screenshot">
<img src="../01_complete.svg" alt="After completion — full output">
<figcaption class="has-text-centered has-text-grey is-size-7 mt-1">Complete</figcaption>
</figure>
</div>
</div>

---

## The model — your application logic

A lofigui app has two parts: a **model** that does the work and a **server** that wires it to the web. This is the model:

```go
// Model function - this is your application logic.
// Just like a terminal program: print output and sleep between steps.
func model(app *lofigui.App) {
    lofigui.Print("Hello world.")
    for i := 0; i < 5; i++ {
        app.Sleep(1 * time.Second)
        lofigui.Printf("Count %d", i)
    }
    lofigui.Print("Done.")
}
```

<div class="annotation">
<strong>lofigui.Print()</strong> works like <code>fmt.Println()</code> — each call adds a line of output. The difference: instead of writing to the terminal, it appends HTML to a buffer that the browser displays.
</div>

<div class="annotation">
<strong>The loop and app.Sleep()</strong> are just a facsimile of a long-running task — standing in for real work like processing files, running a simulation, or querying an API. The model runs in a background goroutine; while it works, the browser keeps refreshing to show new output. Cancellation is transparent — if the user restarts, the framework terminates the old goroutine automatically.
</div>

<div class="annotation">
<strong>StartAction / EndAction</strong> — <code>Handle</code> calls <code>StartAction()</code> before launching the model goroutine, which enables auto-refresh polling. When the model function returns, <code>Handle</code> calls <code>EndAction()</code> automatically — the browser stops refreshing and the output stays put.
</div>

[model.go source on Codeberg](https://codeberg.org/hum3/lofigui/src/branch/main/examples/01_hello_world/go/model.go)

---

## The server — wiring it up

Two lines: create an app, run the model.

```go
func main() {
    app := lofigui.NewApp()
    app.Run(":1340", model)
}
```

<div class="annotation">
<strong>app.Run()</strong> registers the model on <code>/</code>, a cancel handler on <code>/cancel</code>, and starts the server with graceful shutdown. When the model completes, the server exits. This is the HTTP equivalent of <code>RunWASM</code> — one call does everything.
</div>

<div class="annotation">
<strong>Defaults</strong> — <code>NewApp()</code> provides a built-in template (Bulma-styled navbar with cancel button), 1-second refresh, and a <code>/favicon.ico</code> handler. Later examples unbundle <code>Run</code> into <code>Handle</code>, <code>HandleCancel</code>, and <code>ListenAndServe</code> when they need custom routes or multiple endpoints.
</div>

The full source is split across two files: [main.go](https://codeberg.org/hum3/lofigui/src/branch/main/examples/01_hello_world/go/main.go) (the server) and [model.go](https://codeberg.org/hum3/lofigui/src/branch/main/examples/01_hello_world/go/model.go) (the application logic). The model is in its own file so it can be shared with the WASM build — if you don't need a WASM version, a single `main.go` is all you need.

---

## How it works

The browser hits `/`, the server starts the model and returns a page with a Refresh header. The browser reloads every second, showing new output as the model prints. When the model returns, polling stops and the server exits cleanly.

See [technical details](technical.html) for a full sequence diagram and internals.

---

## Cancellation

Both the server and WASM builds support cancelling a running model mid-flow. The navbar shows a **Cancel** button while the model is running. `app.Run()` handles this automatically; when unbundled, register the cancel endpoint explicitly:

```go
// Unbundled form (used in later examples with custom routes):
http.HandleFunc("/cancel", app.HandleCancel("/"))

// WASM: goCancel() is exported automatically by RunWASM
```

<div class="annotation">
<strong>Transparent cancellation</strong> — when cancel is triggered, <code>EndAction()</code> cancels the context. The next call to <code>Print</code>, <code>Sleep</code>, or <code>Yield</code> in the model goroutine panics with an internal sentinel. <code>Handle</code>'s recover wrapper catches it, and the goroutine exits cleanly. The buffer retains its partial output. The model doesn't need any explicit cancellation code.
</div>

See [technical details](../research-technical.html) for the full cancel flow.

---

## WASM: running in the browser

The [live demo](wasm_demo/) runs the same `model()` function compiled to WebAssembly — entirely in your browser, no server required. A service worker intercepts HTTP requests and routes them to Go's `net/http` handlers running inside the WASM binary. The browser sees real HTTP responses — forms, redirects, Refresh headers — identical to the server version.

Because the model lives in its own file (`model.go`), both the server and WASM builds share it unchanged. A separate `main_wasm.go` file (build-tagged `js && wasm`) replaces the server with a single call:

```go
//go:build js && wasm

package main

import "codeberg.org/hum3/lofigui"

func main() {
    app := lofigui.NewApp()
    app.RunWASM(model)
}
```

<div class="annotation">
<strong>App.RunWASM</strong> is the service worker equivalent of <code>App.Run</code>. It registers the same routes (display, start, cancel, favicon) on an <code>http.ServeMux</code> and serves them via <a href="https://github.com/nlepage/go-wasm-http-server">go-wasm-http-server</a>. The browser page uses a service worker (<code>sw.js</code>) that loads the WASM binary and intercepts fetch events. No custom JavaScript polling — just standard HTTP.
</div>

<div class="annotation">
<strong>Building:</strong> <code>GOOS=js GOARCH=wasm go build -o main.wasm .</code> produces the binary. Go provides <code>wasm_exec.js</code> as a loader. The <a href="https://codeberg.org/hum3/lofigui/src/branch/main/Taskfile.yml">Taskfile.yml</a> <code>docs:build-wasm</code> task automates this for all examples.
</div>

[WASM source on Codeberg](https://codeberg.org/hum3/lofigui/src/branch/main/examples/01_hello_world/go/main_wasm.go)

### Gzipped WASM

Example 01 itself only ships the plain (~11 MB) WASM binary — the compact API accepts that trade for a one-line deployment. Projects that want a smaller download follow [example 01b](../01b_hello_world_explicit_gzip/), which layers `DecompressionStream` + a cached decompressed binary on top of 01a's explicit SW wiring. The decompression plumbing is visible enough that it's worth reading as its own tutorial rather than hidden behind a flag.

### Server vs WASM lifecycle

Both the server and WASM builds use the same HTTP handler pattern. The lifecycles differ only in how the server starts:

- **Server** — `app.Run(":1340", model)` starts a real HTTP server. The model auto-starts on the first request to `/`. When the model completes, the server exits (unless `LOFIGUI_HOLD=1` is set).
- **WASM** — `app.RunWASM(model)` registers the same routes but serves them via a service worker. The user clicks a **Start** button (HTML form POST to `/start`). The browser's Refresh header handles polling — no custom JavaScript needed.

<div class="annotation">
<strong>Explicit wiring</strong> — for custom routes, templates, or multi-page apps, see <a href="../01a_hello_world_explicit/">example 01a</a> which unbundles <code>RunWASM</code> into <code>setupRoutes()</code> + <code>wasmhttp.Serve()</code>.
</div>
