# lofigui

**Lofi GUI** - A minimalist library for creating really simple web-based GUIs for CLI tools and small projects. Available in both **Python** and **Go**, it provides a print-like interface for building lightweight web UIs with minimal complexity.

The aplication is where you have a single real object (eg machine or long running)
processing which then have a number of pages around it to show various aspects
of it.

[![Python Version](https://img.shields.io/badge/python-3.9+-blue.svg)](https://www.python.org/downloads/)
[![Go Version](https://img.shields.io/badge/go-1.21+-blue.svg)](https://golang.org/dl/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

## 🚀 Now Available in Go!

Check out the **[Go version](go/README.md)** for:
- **10-100x faster** performance
- **Single binary** deployment
- **WebAssembly** support (~2MB vs ~10MB for Python)
- **Type safety** at compile time

Choose the version that fits your needs - same simple API, different strengths!

## Overview

`lofigui` provides a print-like interface for building lightweight web applications with minimal complexity. Perfect for:

- Creating quick GUIs for command-line tools
- Internal tools for small teams (1-10 users)
- Single-process or single-object front-ends
- Rapid prototyping without JavaScript overhead

## Key Features

- **Simple API**: Print-like interface (`print()`, `markdown()`, `html()`, `table()`)
- **No JavaScript**: Pure HTML/CSS using the Bulma framework to make it look prettier as I am terrible at design.
- **MVC Architecture**: Model, view, and controller architecture
- **Async-ready**: Built on asyncio for modern web frameworks (FastAPI, etc.)
- **Type-safe**: Full type hints and mypy support
- **Secure**: HTML escaping by default to prevent XSS attacks

## Element
Your project is essentially a web site.  To make design simple you completely refresh pages so no code for partial refreshes.  To make things dynamic it has to be asynchonous so for python using fastapi as a server and Uvicorn to provide the https server.

Like a normal terminal program you essentially just print things to a screen but now have the ability to print enriched objects.

### model view controller architecture
All I really want to do is to write the model.  The controller and view (in the browser and templating system) are a necessary evil.  The controller includes the routing and webserver. The controller is split between the app (single instance) and a model specific controller. The view is the html templating and the browser.

### Buffer
In order to be able to decouple the display from the output and to be able to refesh you need to be able to buffer the output.  It is more efficient to buffer the output in the browser but more complicated.  Moving the buffer to the server simplifies the software but requires you to refresh the whole page.
lofigui relies on hyperlinks to perform updates.  Forms are useful for nice buttons but in general to get the right level of interactivity (click on somthing and it changes) you don't want to have forms.  HTMLx would play nicely here if you were intersted in improving interactivity and spending a bit more time on the UI.

## Installation

### Using pip

```bash
pip install lofigui
```

### Using uv

```bash
uv add lofigui
```

### From source

```bash
git clone https://github.com/drummonds/lofigui.git
cd lofigui
uv sync --all-extras
```

## Quick Start

Look at the example for a quick start.

## API Reference

### Output Functions

#### `print(msg, ctx=None, end="\n", escape=True)`

Print text to the buffer as HTML paragraphs.

**Parameters:**
- `msg` (str): Message to print
- `ctx` (PrintContext, optional): Custom context (default: global context)
- `end` (str): End character - `"\n"` for paragraphs, `""` for inline
- `escape` (bool): Escape HTML entities (default: True)

**Example:**
```python
import lofigui as lg

lg.print("Hello world")  # <p>Hello world</p>
lg.print("Inline", end="")  # &nbsp;Inline&nbsp;
lg.print("<script>alert('safe')</script>")  # Escaped by default
lg.print("<b>Bold</b>", escape=False)  # Raw HTML (use with caution!)
```

#### `markdown(msg, ctx=None)`

Convert markdown to HTML and add to buffer.

**Parameters:**
- `msg` (str): Markdown-formatted text
- `ctx` (PrintContext, optional): Custom context

**Example:**
```python
lg.markdown("# Heading\n\nThis is **bold** text")
```

#### `html(msg, ctx=None)`

Add raw HTML to buffer (no escaping).

**WARNING:** Only use with trusted input to avoid XSS vulnerabilities.

**Parameters:**
- `msg` (str): Raw HTML
- `ctx` (PrintContext, optional): Custom context

**Example:**
```python
lg.html("<div class='notification is-info'>Custom HTML</div>")
```

#### `table(table, header=None, ctx=None, escape=True)`

Generate an HTML table with Bulma styling.

**Parameters:**
- `table` (Sequence[Sequence]): Table data as nested sequences
- `header` (List[str], optional): Column headers
- `ctx` (PrintContext, optional): Custom context
- `escape` (bool): Escape cell content (default: True)

**Example:**
```python
data = [
    ["Alice", 30, "Engineer"],
    ["Bob", 25, "Designer"],
]
lg.table(data, header=["Name", "Age", "Role"])
```

### Buffer Management

#### `buffer(ctx=None)`

Get accumulated HTML output.

**Returns:** str

**Example:**
```python
content = lg.buffer()
```

#### `reset(ctx=None)`

Clear the buffer.

**Example:**
```python
lg.reset()
```

### Context Management

#### `PrintContext(max_buffer_size=None)`

Context manager for buffering HTML output.

**Parameters:**
- `max_buffer_size` (int, optional): Warn if buffer exceeds this size

**Example:**
```python
from lofigui import PrintContext, print

# Using context manager (auto-cleanup)
with PrintContext() as ctx:
    print("Hello", ctx=ctx)
    # Buffer automatically reset on exit

# Or create manually
ctx = PrintContext(max_buffer_size=10000)
```

### Favicon Support

#### `get_favicon_response()`

Get a FastAPI/Starlette Response object for serving the favicon.

**Example:**
```python
@app.get("/favicon.ico")
async def favicon():
    return lg.get_favicon_response()
```

#### `get_favicon_html_tag()`

Get an HTML link tag with embedded favicon data URI.

**Example:**
```python
# In your template <head>
{{ get_favicon_html_tag()|safe }}
```

#### `save_favicon_ico(path)`

Save the favicon to a file.

**Example:**
```python
lg.save_favicon_ico("static/favicon.ico")
```

## Architecture

### MVC Pattern

`lofigui` follows the Model-View-Controller pattern:

- **Model**: Your business logic (functions that call `lg.print()`, etc.)
- **View**: Jinja2 templates that render the buffered HTML
- **Controller**: FastAPI/Flask routes that orchestrate model and view

### Buffering Strategy

Server-side buffering simplifies the architecture:
1. Model functions write to a queue
2. `buffer()` drains the queue and returns HTML
3. Templates render the complete HTML
4. Full page refresh (no partial DOM updates)

This approach trades interactivity for simplicity - perfect for internal tools.

### Security

By default, all output functions escape HTML to prevent XSS attacks:

```python
lg.print("<script>alert('xss')</script>")
# Output: <p>&lt;script&gt;alert('xss')&lt;/script&gt;</p>
```

Use `escape=False` or `html()` only with trusted input.

## Examples

Each example introduces new features progressively. Study them in order, or jump to the one closest to your use case.

| # | Name | Features introduced | Run command |
|---|------|---------------------|-------------|
| 01 | Hello World | `App`, background model, auto-refresh polling | `task go-example:01` |
| 02 | SVG Graph | Synchronous render, SVG via `lofigui.HTML()` | `task go-example:02` |
| 03 | WASM Hello World | Go compiled to WASM, browser-only, no server | `task go-wasm:03` |
| 04 | TinyGo WASM | Smaller WASM binaries with TinyGo | `task go-wasm:04` |
| 05 | Demo App | Python template inheritance (Jinja2) | `task example-05` |
| 06 | Notes CRUD | Form POST, `Controller` directly (no App), redirect-after-POST | `task go-example:06` |
| 07 | Water Tank | SVG schematic generation, simulation goroutine, clickable SVG links, WASM-compatible | `task go-example:07` |
| 08 | Water Tank Multi | Multiple pages sharing one model, `LayoutNavbar`, HTTP Refresh polling | `task go-example:08` |
| 09 | Water Tank HTMX | HTMX partial updates, fragment endpoints, no full-page reload | `task go-example:09` |

**Starting points for new projects:**
- **Status page / dashboard** — example 01 (polling) or 02 (sync)
- **CRUD / forms app** — example 06
- **Real-time dashboard** — example 08 (HTTP Refresh) or 09 (HTMX, smoother updates)
- **Browser-only (no server)** — example 03 (WASM)

Python examples also available for 01, 02, 05, 06: `task example-01`, `task example-02`, etc.

### Using Taskfile

The project uses [Task](https://taskfile.dev/) for all development commands:

```bash
task --list              # Show all available tasks
task test                # Run Python + Go tests
task lint                # Run linters
task go-example:09       # Run any Go example by number
task example-01          # Run any Python example by number
task tidy                # go mod tidy for all modules
task clean               # Clean build artifacts
```

## Development

### Setup

```bash
git clone https://github.com/drummonds/lofigui.git
cd lofigui
uv sync --all-extras
```

### Running Tests

```bash
uv run pytest
```

With coverage:
```bash
uv run pytest --cov=lofigui --cov-report=html
```

### Type Checking

```bash
uv run mypy lofigui
```

### Code Formatting

```bash
uv run black lofigui tests
```

## Comparison with Alternatives

| Feature | lofigui | Streamlit | PyWebIO | Textual |
|---------|---------|-----------|---------|---------|
| JavaScript | No | Yes | Yes | No |
| Complexity | Very Low | Medium | Medium | Medium |
| Use Case | Internal tools | Data apps | Web apps | Terminal UIs |
| Learning Curve | Minimal | Moderate | Moderate | Moderate |
| Partial Updates | Via HTMX | Yes | Yes | Yes |

**Choose lofigui if:**
- You want maximum simplicity
- You're building internal tools
- You don't need fancy interactivity
- You want to understand every line of code

**Choose alternatives if:**
- You need rich interactivity
- You're building public-facing apps
- You want widgets and components

## Roadmap

See [ROADMAP.md](ROADMAP.md) for planned features and future direction.

## Contributing

Contributions are welcome! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

## License

MIT License - see [LICENSE](LICENSE) file for details.

## Author

Humphrey Drummond - [hum3@drummond.info](mailto:hum3@drummond.info)

## Links

- **GitHub**: https://github.com/drummonds/lofigui
- **PyPI**: https://pypi.org/project/lofigui/
- **Documentation**: https://github.com/drummonds/lofigui/blob/main/README.md
- **Issues**: https://github.com/drummonds/lofigui/issues
