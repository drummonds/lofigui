# lofigui - LLM Reference

## What is lofigui

lofigui is a lightweight web-UI framework with dual Python and Go implementations. It provides a **print-like interface** for building server-side rendered HTML applications: you call `print()`, `markdown()`, `html()`, and `table()` to accumulate HTML in a buffer, then render it into a template. It uses **Bulma CSS** for styling, **Jinja2** (Python) or **pongo2** (Go) for templates, and **FastAPI** (Python) or **net/http** (Go) as the web server.

## Project structure

```
lofigui/
  lofigui.go              # Go: Print, Markdown, HTML, Table, Buffer, Reset, Context
  controller.go           # Go: Controller (pongo2 template wrapper)
  app.go                  # Go: App (controller lifecycle, polling, action state)
  favicon.go              # Go: ServeFavicon handler
  lofigui/                # Python package
    __init__.py            #   Public API re-exports
    print.py               #   print()
    markdown.py            #   markdown(), html(), table()
    context.py             #   PrintContext, buffer(), reset()
    controller.py          #   Controller
    app.py                 #   App (extends FastAPI), create_app()
    favicon.py             #   Favicon utilities
  examples/
    01_hello_world/        # Async with polling (Python + Go)
    02_svg_graph/          # Synchronous render (Python + Go)
    03_hello_world_wasm/   # Go WASM
    04_tinygo_wasm/        # TinyGo WASM
    05_demo_app/           # Python template inheritance
    06_notes_crud/         # CRUD app (Python + Go)
```

## Python API

### Output functions

```python
import lofigui as lg

lg.print("Hello world")                          # <p>Hello world</p>
lg.print("inline", end="")                       # &nbsp;inline&nbsp; (no paragraph wrap)
lg.print("<b>bold</b>", escape=False)             # raw HTML, no escaping
lg.markdown("## Heading\n\nParagraph")            # renders markdown to HTML
lg.html('<div class="box">raw html</div>')        # adds raw HTML to buffer
lg.table([["a","b"],["c","d"]], header=["X","Y"]) # Bulma-styled table
```

### Buffer management

```python
content = lg.buffer()   # get accumulated HTML
lg.reset()              # clear buffer
```

### PrintContext (isolated buffers)

```python
ctx = lg.PrintContext()
lg.print("scoped output", ctx=ctx)
content = lg.buffer(ctx)
lg.reset(ctx)
```

### App and Controller

```python
from fastapi import Request, BackgroundTasks
from fastapi.responses import HTMLResponse
import lofigui as lg

controller = lg.Controller()
app = lg.create_app(template_dir="templates", controller=controller)

def model():
    lg.print("Working...")
    app.end_action()

@app.get("/", response_class=HTMLResponse)
async def root(background_tasks: BackgroundTasks):
    lg.reset()
    background_tasks.add_task(model)
    app.start_action()
    return '<head><meta http-equiv="Refresh" content="0; URL=/display"/></head>'

@app.get("/display", response_class=HTMLResponse)
async def display(request: Request):
    return app.template_response(request, "hello.html")
```

Key `App` methods: `start_action(refresh_time=1)`, `end_action()`, `is_action_running()`, `template_response(request, template_name, extra={})`, `state_dict(request, extra={})`.

### Favicon

```python
# create_app(add_favicon=True) adds /favicon.ico automatically
# Manual usage:
from lofigui import get_favicon_response, get_favicon_html_tag, save_favicon_ico
```

## Go API

### Output functions

```go
import "github.com/drummonds/lofigui"

lofigui.Print("Hello world")                                       // <p>Hello world</p>
lofigui.Print("inline", lofigui.WithEnd(""))                       // &nbsp;inline&nbsp;
lofigui.Print("<b>bold</b>", lofigui.WithEscape(false))            // raw HTML
lofigui.Printf("Count: %d", 42)                                   // formatted print
lofigui.Markdown("## Heading")                                     // markdown to HTML
lofigui.HTML(`<div class="box">raw</div>`)                         // raw HTML to buffer
lofigui.Table(data, lofigui.WithHeader([]string{"Col1", "Col2"}))  // Bulma table
```

### Buffer management

```go
content := lofigui.Buffer()  // get accumulated HTML
lofigui.Reset()              // clear buffer
```

### Controller

```go
// From file
ctrl, err := lofigui.NewController(lofigui.ControllerConfig{
    Name:         "My Controller",
    TemplatePath: "templates/page.html",
})

// From embedded string
ctrl, err := lofigui.NewControllerFromString(`<html>{{ results | safe }}</html>`)

// From directory
ctrl, err := lofigui.NewControllerFromDir("templates", "page.html")
```

### App

```go
app := lofigui.NewApp()
app.Version = "My App v1.0"
app.SetController(ctrl)
app.SetRefreshTime(1)         // seconds between auto-refresh while polling
app.SetDisplayURL("/display")

// Root handler: starts model in background, resets buffer
http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
    app.HandleRoot(w, r, model, true)  // true = reset buffer
})

// Display handler: renders template with current buffer + polling state
http.HandleFunc("/display", func(w http.ResponseWriter, r *http.Request) {
    app.HandleDisplay(w, r)
})

http.HandleFunc("/favicon.ico", lofigui.ServeFavicon)
```

Model function signature: `func model(app *lofigui.App)`. Call `app.EndAction()` when done.

## Architecture patterns

### 1. Async with polling (example 01)

Model runs in background; page auto-refreshes until `EndAction()`:

```
GET /        -> reset buffer, start model in goroutine/background task, redirect to /display
GET /display -> render template (includes auto-refresh meta tag while polling)
```

### 2. Synchronous (example 02)

Model runs inline, result displayed immediately:

```
GET / -> reset buffer, run model (calls EndAction immediately), redirect to /display
```

### 3. WASM (examples 03, 04)

Go/TinyGo compiled to WASM, served via `build.sh` + `serve.py`. No server-side app.

### 4. CRUD (example 06)

Each action is a form POST that modifies state, then redirects to GET /:

```
GET /         -> reset, render list + forms
POST /create  -> create record, redirect /
POST /update  -> update record, redirect /
POST /delete  -> delete record, redirect /
```

Uses `ctrl.StateDict(r)` and `ctrl.RenderTemplate(w, context)` directly instead of `app.HandleRoot`.

## Template requirements

Templates receive these variables from `state_dict` / `StateDict`:

| Variable  | Description |
|-----------|-------------|
| `results` | Accumulated buffer HTML (use `{{ results \| safe }}`) |
| `refresh` | Auto-refresh `<meta>` tag when polling is active |
| `polling` | Boolean: whether polling is active |
| `version` | App version string |
| `name`    | Controller name |

Minimal template:

```html
<!DOCTYPE html>
<html>
<head>
  {{ refresh | safe }}
</head>
<body>
  {{ results | safe }}
</body>
</html>
```

## Running examples / Development commands

Uses [Task](https://taskfile.dev/) runner:

```bash
task test                # Run Python + Go tests
task lint                # Run linters (black, flake8)
task format              # Format with black
task check-licenses      # Check Go dependency licenses

# Python examples
task example-01          # Hello World (FastAPI)
task example-02          # SVG Graph (Pygal)
task example-05          # Demo App (template inheritance)
task example-06          # Notes CRUD

# Go examples
task go-example:01       # Hello World (net/http)
task go-example:02       # SVG Graph
task go-wasm:03          # WASM Hello World
task go-wasm:04          # TinyGo WASM
task go-example:06       # Notes CRUD
task build-wasm:03       # Build WASM binary only (no serve)
task build-wasm:04       # Build TinyGo WASM binary only

# Go module maintenance
task tidy                # Run go mod tidy for all modules
task tidy:main           # Tidy root module only
task tidy:01             # Tidy example 01 only (etc.)
```
