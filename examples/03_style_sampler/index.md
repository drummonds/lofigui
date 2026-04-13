# 03 — Style Sampler

The first Teletype+ web example. Demonstrates layout patterns, template inheritance, and navbar styles — the building blocks for Level 2 applications. Each menu option shows a different layout style with the same teletype output.

**Interactivity level:** 2 — Teletype+ web (templates, navbars, forms)

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

All styles share a base template via pongo2 `{% extends "base.html" %}`:

```
base.html           → <html>, <head>, Bulma CSS, {% block navbar %}, {% block content %}
style_scrolling.html → extends base: teal scrolling navbar + content
style_fixed.html     → extends base: red fixed navbar + padded content
style_three_panel_nav.html → extends base: yellow navbar + columns layout
...
```

Each page overrides `{% block navbar %}` and `{% block content %}` to produce a completely different layout from the same base. The Go server creates one controller per template:

```go
ctrl, err := lofigui.NewControllerFromDir("templates", "style_scrolling.html")
```

---

## Technical: WASM + pongo2 templates

**The key insight: pongo2 is pure Go and compiles to WASM.** Template inheritance — `{% extends %}`, `{% block %}`, `{% include %}` — works identically in the browser.

### The problem

`go:embed` cannot use `..` paths. The templates live in `../templates/` relative to the Go source, but embed only looks within or below the source directory.

### The solution

The build script copies templates into `go/templates/` before compilation:

```bash
cp -r ../templates .
GOOS=js GOARCH=wasm go build -o main.wasm .
rm -rf templates
```

The embedded filesystem is then available at runtime:

```go
//go:embed templates
var templateFS embed.FS
```

### Loading templates from embed.FS

lofigui provides `NewControllerFromFS` which wraps pongo2's `FSLoader`:

```go
ctrl, err := lofigui.NewControllerFromFS(templateFS, "templates", "page.html")
```

This creates a pongo2 `TemplateSet` with an `fs.FS`-backed loader. When pongo2 encounters `{% extends "base.html" %}`, it resolves `base.html` from the embedded filesystem — no disk access, no server.

### Rendering in WASM

The server uses `ctrl.RenderTemplate(w, context)` which writes to an `http.ResponseWriter`. In WASM there is no HTTP — instead, render to a string via pongo2 directly:

```go
html, err := ctrl.GetTemplate().Execute(pongo2.Context{
    "results":      content,
    "current_path": currentPath,
})
```

JavaScript calls `goRenderPage("style_scrolling.html", "/style/scrolling")` and replaces the page content with the returned HTML. Navigation between styles is instant — no server round-trip, no page reload.

### Same code, two targets

| Aspect | Server | WASM |
|--------|--------|------|
| Template source | Filesystem (`NewControllerFromDir`) | Embedded (`NewControllerFromFS`) |
| Render target | `http.ResponseWriter` | `string` via `Execute()` |
| Navigation | HTTP request per page | JS calls `goRenderPage()` |
| Template engine | pongo2 (server-side) | pongo2 (client-side, same binary) |
| `{% extends %}` | Works | Works identically |

The model code (`model.go`) is shared between both builds with no changes.

### Binary size

<<<<<<< HEAD
The WASM binary includes pongo2 and the embedded templates. Expect ~3-4MB for standard Go WASM, ~200-500KB for TinyGo (if pongo2 is compatible).
=======
The WASM binary includes pongo2 and the embedded templates. Expect ~3-4MB for standard Go WASM.
>>>>>>> task/WTteletype

---

## Links

- [Launch Demo](demo.html)
- [Source on Codeberg](https://codeberg.org/hum3/lofigui/src/branch/main/examples/03_style_sampler)
