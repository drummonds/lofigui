# Go Controller — Extension Guide

`lofigui.Controller` is the template-rendering layer. Most apps don't touch it directly — `App.Run()` and `App.RunWASM()` create a default controller for you. Reach for an explicit controller when you need:

- A specific template (file, embedded string, embed.FS, or `templates/` directory with `base.html`)
- Multiple templates per app (one controller per page)
- Direct rendering with custom context — no `App` lifecycle
- Hot template reload during development

For the API reference, see [pkg.go.dev](https://pkg.go.dev/codeberg.org/hum3/lofigui) or `controller.go` directly. For the bigger architecture picture (App, buffer, Print, examples), see [`CLAUDE.md`](./CLAUDE.md). This file documents the extension patterns that aren't obvious from the godoc.

## Constructors

```go
// From a file on disk. If a base.html sits in the same directory, it's parsed
// alongside for {{block}}/{{define}} inheritance.
ctrl, err := lofigui.NewController(lofigui.ControllerConfig{
    Name:         "My Page",
    TemplatePath: "templates/page.html",
})

// From an inline string (most common for compact examples).
ctrl, err := lofigui.NewControllerFromString(`<html>{{.results}}</html>`)

// From a directory + filename. Convenience wrapper around NewController.
ctrl, err := lofigui.NewControllerFromDir("templates", "page.html")

// From an embed.FS — the only option that works under WASM, since the
// service-worker process has no host filesystem. base.html is auto-detected.
//
//go:embed templates
var templateFS embed.FS
ctrl, err := lofigui.NewControllerFromFS(templateFS, "templates", "page.html")
```

`ControllerConfig` accepts exactly one of `TemplatePath` or `TemplateString`. `Context` is optional — pass a `*lofigui.Context` from `lofigui.NewContext()` to give this controller its own buffer instead of sharing the package-global one.

## Direct rendering — no App needed

Use a `Controller` (not an `App`) when the page is fully static, when you're using HTMX (example 09) and don't want auto-refresh state, or when you need precise control over what's in the template context.

```go
ctrl, _ := lofigui.NewController(lofigui.ControllerConfig{
    TemplatePath: "templates/page.html",
})

http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
    lofigui.Reset()
    lofigui.Print("Hello")
    ctx := ctrl.StateDict(r) // {"request": r, "results": template.HTML(buffer)}
    ctx["title"] = "My Page"
    ctrl.RenderTemplate(w, ctx)
})
```

`RenderTemplate(w, ctx)` writes directly to the response. `RenderToString(ctx)` returns the rendered HTML — useful in WASM builds or when composing fragments.

## Multiple controllers per app

Each route can use a different controller. Useful when pages have very different layouts (admin vs main, dashboard vs detail page).

```go
mainCtrl, _ := lofigui.NewControllerFromFS(fs, "templates", "main.html")
adminCtrl, _ := lofigui.NewControllerFromFS(fs, "templates", "admin.html")

mux.HandleFunc("GET /{$}",   func(w http.ResponseWriter, r *http.Request) { mainCtrl.HandleDisplay(w, r, nil) })
mux.HandleFunc("GET /admin", func(w http.ResponseWriter, r *http.Request) { adminCtrl.HandleDisplay(w, r, nil) })
```

A cleaner pattern when many routes share one layout: one controller per template, indexed by path. Example 03's `loadControllers()` does this — see `examples/03_style_sampler/go/model.go`.

## Template inheritance with base.html

If a `base.html` exists in the same directory (file path or `embed.FS`) as the requested template, both are parsed together. The child uses `{{define "title"}}…{{end}}` / `{{define "content"}}…{{end}}` to override blocks declared in `base.html`. Example 03 (Style Sampler) is the canonical demo.

This is opt-in and detected automatically by every constructor — no explicit configuration.

## Hot template reload during development

```go
ctrl, _ := lofigui.NewController(lofigui.ControllerConfig{
    TemplatePath: "templates/page.html",
})

http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
    if os.Getenv("DEV") != "" {
        if err := ctrl.ReloadTemplate("templates/page.html"); err != nil {
            log.Printf("template reload failed: %v", err)
        }
    }
    ctrl.HandleDisplay(w, r, nil)
})
```

`ReloadTemplate` only works for file-backed controllers (it re-reads from disk). Embedded `embed.FS` templates can't be reloaded without rebuilding.

## Integration with third-party routers

`Controller.ServeHTTP` makes a controller a plain `http.Handler`, so it drops into any router that takes one. The standard library's `http.ServeMux` is preferred (Go 1.22+ method-and-path patterns are sufficient for everything in lofigui's examples), but the controller doesn't care:

```go
import "github.com/go-chi/chi/v5"

r := chi.NewRouter()
r.Get("/",         ctrl.ServeHTTP)            // calls HandleDisplay(w, r, nil)
r.Get("/favicon.ico", lofigui.ServeFavicon)
http.ListenAndServe(":1340", r)
```

The same shape works for Gorilla Mux, Echo, Gin (`gin.WrapH`), or any handler-compatible framework.

## Custom contexts (isolated buffers)

The package-level `lofigui.Print/Markdown/HTML/Table` write to a single global buffer. To run multiple models concurrently (e.g. one per HTMX fragment endpoint), give each its own `*lofigui.Context`:

```go
ctxA := lofigui.NewContext()
ctxB := lofigui.NewContext()

ctrlA, _ := lofigui.NewController(lofigui.ControllerConfig{
    TemplatePath: "templates/a.html",
    Context:      ctxA,
})
ctrlB, _ := lofigui.NewController(lofigui.ControllerConfig{
    TemplatePath: "templates/b.html",
    Context:      ctxB,
})

ctxA.Print("hello from A")
ctxB.Print("hello from B")
// ctrlA.StateDict(r) reads ctxA's buffer; ctrlB.StateDict(r) reads ctxB's.
```

The `*Context` methods (`Print`, `Markdown`, `HTML`, `Table`, `Buffer`, `Reset`) mirror the package-level functions exactly. Choose this when concurrent requests would otherwise interleave output in the global buffer. The HTMX examples (09, 10) sidestep the issue with a `sync.Mutex` instead — simpler when fragments are cheap.

## Template context shape

`Controller.StateDict(r)` returns the minimal context — just `request` and `results` (the buffer as `template.HTML` so it isn't escaped).

`App.StateDict(r, extra)` returns the richer context that the `App` lifecycle templates expect — adds `version`, `build_date`, `controller_name`, `polling`, `poll_count`, `refresh`. Use this when you're rendering through an `App`'s controller.

A minimal compatible template:

```html
<!DOCTYPE html>
<html>
<head>
  <link rel="stylesheet" href="/assets/bulma.min.css">
  <title>{{.version}}</title>
</head>
<body>
  {{.results}}
</body>
</html>
```

`{{.results}}` works because `StateDict` casts the buffer to `template.HTML`. Don't try to print `template.HTML` from user input — escape with `html.EscapeString` first (example 06's CRUD operations are the canonical pattern).
