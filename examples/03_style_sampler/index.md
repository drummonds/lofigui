<style>
.annotation { border-left: 3px solid #3273dc; background: #f0f4ff; padding: 0.75em 1em; margin: 0.75em 0; border-radius: 0 4px 4px 0; font-size: 0.9em; }
.annotation strong { color: #3273dc; }
.screenshot { border: 1px solid #dbdbdb; border-radius: 4px; box-shadow: 0 2px 6px rgba(0,0,0,0.1); overflow: hidden; }
.screenshot img { display: block; width: 100%; height: auto; }
</style>

# 03 — Style Sampler

One Go codebase, six page layouts, two deployment targets. The same templates render a multi-page site either from a classic HTTP server (`main.go`) or entirely inside the browser via a service worker (`main_wasm.go`). Both entry points register the **same `*http.ServeMux`** — navigation, routing, and rendering code paths are identical.

The `model` function is the **same five-lines-of-teletype** used in [example 01](../01_hello_world/) — `Print("Hello world.")`, a five-count loop with one-second sleeps, `Print("Done.")`. Every layout renders the current lofigui buffer, so whichever style you click into auto-refreshes in place while the counter ticks up. The point of this example is to show that **adding layout variety on top of the same model is a matter of templates, not plumbing**.

**Interactivity level:** 6 — WASM (browser-only, service worker)

<div class="buttons">
<a href="wasm_demo/" class="button is-primary">Launch Demo</a>
<a target="_blank" href="https://codeberg.org/hum3/lofigui/src/branch/main/examples/03_style_sampler" class="button is-light">Source on Codeberg</a>
</div>

<div class="columns is-vcentered">
<div class="column is-5">
<figure class="image screenshot">
<img src="../03_loading.svg" alt="Loading — service worker bootstrapping">
<figcaption class="has-text-centered has-text-grey is-size-7 mt-1">Loading — SW bootstrap registering</figcaption>
</figure>
</div>
<div class="column is-narrow has-text-centered">
<span style="font-size: 2rem; color: #999;">&rarr;</span>
</div>
<div class="column is-5">
<figure class="image screenshot">
<img src="../03_working.svg" alt="Working — scrolling layout after model completion">
<figcaption class="has-text-centered has-text-grey is-size-7 mt-1">Working — Scrolling layout after model completes</figcaption>
</figure>
</div>
</div>

---

## How lofigui helps here

lofigui is deliberately small. For this example it contributes four things:

1. **A print-style output buffer** — `lofigui.Print` / `Printf` accumulate HTML in the global lofigui buffer. Every layout's handler reads `lofigui.Buffer()` on each request, so the teletype content is identical in every layout and grows in place as the model runs.
2. **A template-inheritance loader** — `lofigui.NewControllerFromFS` parses `base.html` alongside the named child template, so Go stdlib `{{block}}`/`{{define}}` works with embedded filesystems — the only source a WASM build can reach.
3. **One render call, same shape in both builds** — `ctrl.RenderTemplate(w, ctx)` writes to an `http.ResponseWriter` regardless of whether that writer was handed over by `net/http.ListenAndServe` or by `go-wasm-http-server` inside a service worker.
4. **Built-in `/assets/bulma.min.css`** — `lofigui.ServeBulma` is registered next to the app's routes so Bulma loads without a CDN round-trip, on server and WASM.

Everything else — the routes, the templates, the navigation — is plain Go standard library.

---

## Layout styles

| Style | Navbar | Layout | Use case |
|-------|--------|--------|----------|
| Scrolling | Default, scrolls with page | Single column | Simple tools (examples 01, 02) |
| Fixed | Pinned to top | Single column | Dashboards needing persistent nav |
| Three-Panel Nav | Top + left sidebar | Left nav, right content | Multi-page apps with page tree |
| Three-Panel Controls | Top + left sidebar | Left form, right output | CLI tools with parameter dialogs |
| Full Width | Minimal top bar | Single wide column | Maximum output area |

---

## Template inheritance

All styles share a base template using Go's `html/template` `{{block}}` / `{{define}}` inheritance:

```
base.html                        → <html>, <head>, <base>, Bulma, {{block "navbar" .}}, {{block "content" .}}
style_scrolling.html             → extends base: teal scrolling navbar + content
style_fixed.html                 → extends base: red fixed navbar + padded content
style_three_panel_nav.html       → extends base: yellow navbar + columns layout
...
```

`base.html` declares blocks with `{{block "name" .}}default{{end}}`. Each child template overrides them with `{{define "name"}}...{{end}}`:

```html
<!-- base.html -->
<head>
  <base href="{{.base}}">
  <link rel="stylesheet" href="/assets/bulma.min.css">
  …
</head>
<body>
  {{block "navbar" .}}{{end}}
  {{block "content" .}}{{end}}
</body>
```

```html
<!-- style_scrolling.html -->
{{define "navbar"}}<nav class="navbar is-primary">…</nav>{{end}}
{{define "content"}}<section class="section">{{.results}}</section>{{end}}
```

<div class="annotation">
<strong><code>&lt;base href="{{.base}}"&gt;</code></strong> lets the same template work at both the site root ("/") and under a service-worker scope ("/03_style_sampler/wasm_demo/"). The navbar/sidebar links use <em>relative</em> hrefs like <code>style/scrolling</code> and <code>""</code> — the browser resolves them against <code>&lt;base&gt;</code>, so they stay inside the SW scope in WASM and resolve to site root on the server.
</div>

---

## The model — identical to example 01

The teletype that every layout displays is the same five-line model 01 uses:

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

<div class="annotation">
<strong>Continuity with example 01.</strong> Nothing about layout changes requires a different model. The point of this example is that <em>templates</em> do the layout work; the app logic is still <code>Print</code> + <code>Sleep</code>, and you can drop it in unchanged from a simpler example.
</div>

---

## Shared mux — `model.go`

Both `main.go` and `main_wasm.go` delegate the routing table to `model.go`. It embeds `templates/`, parses every layout against `base.html`, and returns a `*http.ServeMux` with all routes wired up:

```go
//go:embed templates
var templateFS embed.FS

var pathToTemplate = map[string]string{
    "/":                           "home.html",
    "/style/scrolling":            "style_scrolling.html",
    "/style/fixed":                "style_fixed.html",
    "/style/three-panel-nav":      "style_three_panel_nav.html",
    "/style/three-panel-controls": "style_three_panel_controls.html",
    "/style/fullwidth":            "style_fullwidth.html",
}

func buildMux(app *lofigui.App, basePrefix string) *http.ServeMux {
    controllers := loadControllers()
    mux := http.NewServeMux()
    for p, name := range pathToTemplate {
        tpl := name
        pattern := "GET " + p
        if p == "/" { pattern = "GET /{$}" }
        mux.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
            app.WriteRefreshHeader(w) // HTTP Refresh while model is running
            controllers[tpl].RenderTemplate(w, lofigui.TemplateContext{
                "results":      template.HTML(lofigui.Buffer()),
                "current_path": r.URL.Path,
                "base":         basePrefix,
            })
        })
    }
    mux.HandleFunc("GET /favicon.ico",          lofigui.ServeFavicon)
    mux.HandleFunc("GET /assets/bulma.min.css", lofigui.ServeBulma)
    return mux
}
```

<div class="annotation">
<strong>Everything interesting lives here.</strong> Template parsing, route registration, and render calls all use <code>net/http</code> types — no <code>syscall/js</code>, no WASM-only imports. Both the server and WASM builds hand this mux to their respective serving runtimes.
</div>

<div class="annotation">
<strong>Auto-refresh while polling.</strong> <code>app.WriteRefreshHeader(w)</code> emits <code>Refresh: 1</code> only while the model is running; once the model returns, the header is absent and the page stops reloading. The handler reads the live <code>lofigui.Buffer()</code> on each request — no snapshotting, no per-handler state — so any layout you're looking at grows in place as the model prints.
</div>

---

## The server — `main.go`

```go
//go:build !(js && wasm)

func main() {
    app := lofigui.NewApp()
    app.SetRefreshTime(1)
    app.RunModel(model) // kick off the teletype in a background goroutine

    fmt.Println("Style Sampler running at http://localhost:1340")
    http.ListenAndServe(":1340", buildMux(app, "/"))
}
```

Four lines. `app.RunModel(model)` launches the model in a goroutine (with cancellation-aware `Sleep` recovery already wired up inside the library). `buildMux(app, "/")` is handed to `http.ListenAndServe`. Base prefix is `"/"` because the server hosts the app at the site root.

---

## The WASM entry point — `main_wasm.go`

```go
//go:build js && wasm

func main() {
    app := lofigui.NewApp()
    app.SetRefreshTime(1)
    app.RunModel(model)

    base := strings.TrimSuffix(lofigui.WASMScopePath(), "/") + "/"
    if _, err := wasmhttp.Serve(buildMux(app, base)); err != nil { panic(err) }
    select {} // keep the Go runtime alive to service SW fetches
}
```

Same three lines of setup as `main.go`, then `wasmhttp.Serve` registers every handler in the mux on the service-worker fetch pipeline. `lofigui.WASMScopePath()` returns the scope the SW is registered at (e.g. `/03_style_sampler/wasm_demo/`), which becomes the `<base href>` so the templates' relative links land back inside the scope.

<div class="annotation">
<strong>Why two files at all?</strong> <code>syscall/js</code> (indirectly imported by <code>go-wasm-http-server</code>) does not link on non-WASM targets, and <code>http.ListenAndServe</code> is a runtime no-op under WASM. <code>//go:build</code> tags pick the right entry point at compile time; no runtime switches, no stubs, no conditional imports.
</div>

---

## Same code, two targets

| Aspect | Server (`main.go`) | WASM (`main_wasm.go`) |
|--------|--------------------|------------------------|
| Build tag | `!(js && wasm)` | `js && wasm` |
| App setup | `NewApp` + `SetRefreshTime(1)` + `RunModel(model)` | identical |
| Routing table | `buildMux(app, "/")` from `model.go` | `buildMux(app, scopePath)` from `model.go` |
| Serving runtime | `http.ListenAndServe(":1340", mux)` | `wasmhttp.Serve(mux)` + `select{}` |
| Request shape | `http.Request` / `http.ResponseWriter` | `http.Request` / `http.ResponseWriter` |
| Browser navigation | HTTP request → Go handler | `fetch` intercepted by SW → Go handler |
| Template render | `ctrl.RenderTemplate(w, ctx)` | `ctrl.RenderTemplate(w, ctx)` — same call |
| `<base href>` | `"/"` | `"/03_style_sampler/wasm_demo/"` |

The `model` is unchanged between builds — literally the same five lines of Go from example 01. `buildMux` is shared. The render call is the same function. The only real differences are the entry-point setup and the `<base>` prefix.

---

## The service-worker bootstrap (how navigation actually works)

`wasmassets.Deploy` (invoked by `go run ./cmd/wasm-deploy` in the Taskfile) produces a self-contained demo directory at `docs/03_style_sampler/wasm_demo/`:

```
wasm_demo/
  index.html          — SW bootstrap (registers sw.js, loads WASM, then redirects to "./")
  sw.js               — service worker
  wasmhttp_sw.js      — go-wasm-http-server SW runtime
  main.wasm           — the Go binary
  wasm_exec.js        — Go's stdlib WASM loader
  bulma.min.css       — vendored Bulma next to the bootstrap
  demo.html           — recovery stub; unregisters the SW if the demo gets stuck
```

The sequence the browser sees:

1. GET `/03_style_sampler/wasm_demo/` → static `index.html` (bootstrap).
2. Bootstrap JS registers `sw.js` scoped at `/03_style_sampler/wasm_demo/`, loads `main.wasm`, and redirects to `./`.
3. `./` is now intercepted by the SW. `wasmhttp_sw.js` forwards the fetch as an `http.Request` into the WASM mux.
4. The mux's `GET /{$}` handler renders `home.html` and returns the HTML.
5. User clicks a nav link (`href="style/scrolling"`). Browser resolves it against `<base href="/03_style_sampler/wasm_demo/">`, fetches `/03_style_sampler/wasm_demo/style/scrolling`. SW intercepts, strips the scope prefix, Go sees `/style/scrolling`, mux calls the scrolling handler.

<div class="annotation">
<strong>No JavaScript bridge needed.</strong> The old version of this example exposed a <code>goRenderPage</code> function to JS and routed clicks through a custom <code>app.js</code>. Under the service-worker pattern the browser does normal HTTP navigation; the SW is the only integration point between JS and Go. The Go code is unaware it's running in WASM.
</div>

---

## Binary size

The WASM binary includes `html/template`, `net/http`, the embedded templates, `go-wasm-http-server`, and lofigui itself. Expect ~10–11 MB uncompressed (~2.8 MB gzipped) — a bit bigger than the old `syscall/js`-bridge version, but the entire model and routing code now compiles from the same source as the server build.

---

## Links

- [Launch Demo](wasm_demo/)
- [Source on Codeberg](https://codeberg.org/hum3/lofigui/src/branch/main/examples/03_style_sampler)
