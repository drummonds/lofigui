<style>
.annotation { border-left: 3px solid #3273dc; background: #f0f4ff; padding: 0.75em 1em; margin: 0.75em 0; border-radius: 0 4px 4px 0; font-size: 0.9em; }
.annotation strong { color: #3273dc; }
.screenshot { border: 1px solid #dbdbdb; border-radius: 4px; box-shadow: 0 2px 6px rgba(0,0,0,0.1); overflow: hidden; }
.screenshot img { display: block; }
</style>

# 01 — Hello World

If you can write a Go program that prints to stdout, you can write a lofigui app. The model function below is ordinary Go code — `Print()`, a loop, a sleep. The only difference is the output goes to a web page instead of a terminal. No WebSocket, no JavaScript — just the browser's built-in refresh mechanism doing the work.

<div class="buttons">
<a href="demo.html" class="button is-primary">Launch Demo</a>
<a target="_blank" href="https://codeberg.org/hum3/lofigui/src/branch/main/examples/01_hello_world" class="button is-light">Source on Codeberg</a>
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

The [live demo](demo.html) runs the same `model()` function compiled to WebAssembly — entirely in your browser, no server required. Because the model lives in its own file (`model.go`), both the server and WASM builds share it unchanged.

A separate `main_wasm.go` file (build-tagged `js && wasm`) replaces the server with a single call:

```go
//go:build js && wasm

package main

import "codeberg.org/hum3/lofigui"

func main() { lofigui.RunWASM(model) }
```

<div class="annotation">
<strong>RunWASM</strong> exports four functions to JavaScript — <code>goStart()</code>, <code>goCancel()</code>, <code>goRender()</code>, and <code>goIsRunning()</code> — then calls <code>wasmReady()</code> and blocks. A small JavaScript timer calls <code>goRender()</code> every 500ms to update the page. For apps that need custom JS exports (extra buttons, multiple render functions), write the wiring by hand instead — see examples 07-12.
</div>

<div class="annotation">
<strong>Building:</strong> <code>GOOS=js GOARCH=wasm go build -o main.wasm .</code> produces the binary. Go provides <code>wasm_exec.js</code> as a loader. The <a href="https://codeberg.org/hum3/lofigui/src/branch/main/Taskfile.yml">Taskfile.yml</a> <code>docs:build-wasm</code> task automates this for all examples.
</div>

[WASM source on Codeberg](https://codeberg.org/hum3/lofigui/src/branch/main/examples/01_hello_world/go/main_wasm.go)

### Gzipped WASM

The demo page offers two buttons: **Start** loads the uncompressed `main.wasm` (~4.8 MB), while **Start (gzipped)** loads `main.wasm.gz` (~2.1 MB) and decompresses it in the browser. The `docs:build-wasm` task creates both files automatically:

```bash
# Build step (Taskfile.yml)
find docs -name "main.wasm" -exec gzip -9 -k {} \;
```

The JavaScript uses the browser's built-in `DecompressionStream` to decompress on the fly, then feeds the result to `WebAssembly.instantiateStreaming`:

```js
const resp = await fetch('main.wasm.gz');
const ds = new DecompressionStream('gzip');
const decompressed = new Response(resp.body.pipeThrough(ds), {
    headers: {'Content-Type': 'application/wasm'}
});
const result = await WebAssembly.instantiateStreaming(decompressed, go.importObject);
```

<div class="annotation">
<strong>Why not serve gzip transparently?</strong> Most production static hosts (GitHub Pages, Cloudflare Pages, Netlify) automatically compress responses with gzip or brotli. The explicit <code>.wasm.gz</code> approach is for hosts that don't, or for demonstrating the size difference in the UI. <code>DecompressionStream</code> is supported in all modern browsers.
</div>

### Server vs WASM lifecycle

The server app and WASM app run the same model, but their lifecycles differ:

- **Server** — the model starts automatically on the first HTTP request to `/`. While the model runs, the browser polls for updates via the Refresh header. When the model completes, the server exits (unless `LOFIGUI_HOLD=1` is set).
- **WASM** — the page loads with a **Start** button. The user clicks to begin, and JavaScript polls `goRender()` every 500ms to update the output. After the model completes, clicking Start again restarts it.

<div class="annotation">
<strong>Why the difference?</strong> The server uses <code>Handle()</code> which auto-starts on an empty buffer — request-driven. WASM uses <code>RunWASM()</code> which exports <code>goStart()</code> behind a button — user-driven. The model code itself is identical in both cases.
</div>
