# Example 01: Hello World

A minimal example demonstrating the basic usage of lofigui with FastAPI.

## Overview

This example shows:
- Basic project structure for a lofigui application
- Integration with FastAPI (Python) and net/http (Go)
- Simple real time print output to web page
- Action and display pattern of urls
- MVC architecture pattern
- Shared Jinja2 templates between Python and Go using pongo2

It is a litle verbose but these are the bits that you will extend for your own project.

## Structure

```
01_hello_world/
├── python/
│   ├── hello.py           # Python implementation
│   ├── pyproject.toml     # Python dependencies
│   └── uv.lock            # Lock file
├── go/
│   ├── main.go            # Go implementation
│   ├── go.mod             # Go dependencies
│   └── go.sum             # Go checksums
├── templates/
│   └── hello.html         # Shared Jinja2 template (used by both Python and Go)
└── README.md              # This file
```

## Installation

### Python
From the `python/` directory:

```bash
cd python
uv sync --no-install-project
```

### Go
From the `go/` directory:

```bash
cd go
go mod download
```

## Running the Example

### Python
```bash
cd python
uv run --no-project python hello.py
```

### Go
```bash
cd go
go run main.go
```

Or using Task from repo root:
```bash
task example-01
```

Then open your browser to: http://127.0.0.1:1340

## How It Works

### 1. Model (`model()` function)

The model contains your business logic which is a simple long running function:

```python
def model():
    lg.print("Hello world.")
    for i in range(5):
        sleep(2)
        lg.print(f"Count {i}")
    lg.markdown("<a href='/'>Restart</a>")
    lg.print("Done.")
```

This simply outputs "Hello world" to the buffer and count values.
On completion it gives a link to restart the application.

This is the whole point of lofigui so that you can write simple long running functions that output extra informaton as they run and for the user to get a dynamic
view of the progress.

### 2. Controller (`Controller` class and route)

The controller manages state and routing:

```python
# Use create_app which automatically includes favicon route
app = lg.create_app(template_dir="../templates")

@app.get("/", response_class=HTMLResponse)
async def root(request: Request):
    lg.reset()  # Clear previous output
    model()     # Generate new output
    return app.templates.TemplateResponse(
        "hello.html", controller.state_dict({"request": request})
    )
```

The `lg.create_app()` function provides:
- Automatic `/favicon.ico` endpoint
- Pre-configured Jinja2 templates accessible via `app.templates`

### 3. View (`templates/hello.html`)

The view renders the buffered HTML using Jinja2 syntax:

```html
<h1>Test start</h1>
{{results | safe}}
```

The `results` variable contains the accumulated HTML from lofigui.

**Note**: The Go version uses [pongo2](https://github.com/flosch/pongo2), a Django/Jinja2-compatible template engine for Go. This allows both Python and Go implementations to share the exact same template files.


See example 02 for a more complex example with charts.
