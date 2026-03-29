# Example 01: Hello World

A minimal example demonstrating the basic usage of lofigui — the simplest form of UI, delivered through the web.

## Overview

This example shows:
- Basic project structure for a lofigui application
- Integration with net/http (Go) and FastAPI (Python)
- Simple real time print output to web page
- Action and display pattern of urls
- MVC architecture pattern
- Shared templates between Go (pongo2) and Python (Jinja2)

It is a little verbose but these are the bits that you will extend for your own project.

## Structure

```
01_hello_world/
├── go/
│   ├── main.go            # Go implementation
│   ├── go.mod             # Go dependencies
│   └── go.sum             # Go checksums
├── python/
│   ├── hello.py           # Python implementation
│   ├── pyproject.toml     # Python dependencies
│   └── uv.lock            # Lock file
├── templates/
│   └── hello.html         # Shared template (used by both Go and Python)
└── README.md              # This file
```

## Installation

### Go
From the `go/` directory:

```bash
cd go
go mod download
```

### Python
From the `python/` directory:

```bash
cd python
uv sync --no-install-project
```

## Running the Example

### Go
```bash
cd go
go run main.go
```

Or using Task from repo root:
```bash
task go-example:01
```

### Python
```bash
cd python
uv run --no-project python hello.py
```

Or using Task from repo root:
```bash
task example-01
```

Then open your browser to: http://127.0.0.1:1340

## How It Works

### 1. Model (`model()` function)

The model contains your business logic — a simple long-running function:

```go
func model(ctx context.Context, app *lofigui.App) {
    lofigui.Print("Hello world.")
    for i := 0; i < 5; i++ {
        select {
        case <-ctx.Done():
            lofigui.Print("Cancelled.")
            return
        case <-time.After(1 * time.Second):
        }
        lofigui.Print(fmt.Sprintf("Count %d", i))
    }
    lofigui.Markdown("<a href='/'>Restart</a>")
    lofigui.Print("Done.")
    app.EndAction()
}
```

This simply outputs "Hello world" to the buffer and count values. The loop and sleep are a facsimile of a long-running task — standing in for real work like processing files or running a simulation. On completion it gives a link to restart the application.

This is the whole point of lofigui: you write simple long-running functions that output extra information as they run, and the user gets a dynamic view of the progress.

### 2. Controller (App + routes)

The controller manages state and routing:

```go
app := lofigui.NewApp()
ctrl, _ := lofigui.NewController(lofigui.ControllerConfig{
    Name:         "Hello World Controller",
    TemplatePath: "../templates/hello.html",
})
app.SetController(ctrl)
app.SetRefreshTime(1)
app.SetDisplayURL("/display")

http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
    app.HandleRoot(w, r, model, true)
})
http.HandleFunc("/display", func(w http.ResponseWriter, r *http.Request) {
    app.HandleDisplay(w, r)
})
```

`HandleRoot` resets the buffer, starts the model in a goroutine, and redirects to `/display`. `HandleDisplay` renders the template with the current buffer content and an auto-refresh tag while the model is running.

### 3. View (`templates/hello.html`)

The view renders the buffered HTML:

```html
<h1>Test start</h1>
{{ results | safe }}
```

The `results` variable contains the accumulated HTML from lofigui.

**Note**: The Go version uses [pongo2](https://github.com/flosch/pongo2), a Django/Jinja2-compatible template engine. This allows both Go and Python implementations to share the exact same template files.

### Python equivalent

The Python implementation uses FastAPI and has the same structure. See [Python Notes](https://h3-lofigui.statichost.page/research-python.html) for the Python API reference.

See example 02 for a more complex example with charts.
