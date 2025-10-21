# lofigui

**Lofi GUI** - A minimalist Python library for creating simple web-based GUIs for CLI tools and small projects.

[![Python Version](https://img.shields.io/badge/python-3.8+-blue.svg)](https://www.python.org/downloads/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

## Overview

`lofigui` provides a print-like interface for building lightweight web applications with minimal complexity. Perfect for:

- Creating quick GUIs for command-line tools
- Internal tools for small teams (1-10 users)
- Single-process or single-object front-ends
- Rapid prototyping without JavaScript overhead

## Key Features

- **Simple API**: Print-like interface (`print()`, `markdown()`, `html()`, `table()`)
- **No JavaScript**: Pure HTML/CSS using the Bulma framework
- **MVC Architecture**: Clean separation of model, view, and controller
- **Async-ready**: Built on asyncio for modern web frameworks (FastAPI, etc.)
- **Type-safe**: Full type hints and mypy support
- **Secure**: HTML escaping by default to prevent XSS attacks

## Installation

### Using pip

```bash
pip install lofigui
```

### Using Poetry

```bash
poetry add lofigui
```

### From source

```bash
git clone https://github.com/drummonds/lofigui.git
cd lofigui
poetry install
```

## Quick Start

Here's a minimal example using FastAPI:

```python
from lofigui import buffer, print, reset
import lofigui as lg
from fastapi import FastAPI, Request
from fastapi.responses import HTMLResponse
from fastapi.templating import Jinja2Templates
import uvicorn

app = FastAPI()
templates = Jinja2Templates(directory="templates")

def model():
    """Your business logic here."""
    lg.print("Hello world!")
    lg.markdown("## This is a heading")
    lg.table([["Alice", 30], ["Bob", 25]], header=["Name", "Age"])

@app.get("/", response_class=HTMLResponse)
async def root(request: Request):
    reset()  # Clear previous output
    model()  # Generate content
    return templates.TemplateResponse(
        "index.html", {"request": request, "results": buffer()}
    )

if __name__ == "__main__":
    uvicorn.run("app:app", host="127.0.0.1", port=8000, reload=True)
```

**templates/index.html:**
```html
<!DOCTYPE html>
<html>
<head>
    <link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/bulma@0.9.4/css/bulma.min.css">
</head>
<body>
    <section class="section">
        <div class="container">
            {{results | safe}}
        </div>
    </section>
</body>
</html>
```

Run the app and visit `http://127.0.0.1:8000`!

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

See the `examples/` directory for complete working examples:

- **01_hello_world**: Minimal FastAPI application
- **02_svg_graph**: Chart rendering with Pygal

To run an example:
```bash
cd examples/01_hello_world
poetry install
poetry run python hello.py
```

Visit `http://127.0.0.1:1340`

## Development

### Setup

```bash
git clone https://github.com/drummonds/lofigui.git
cd lofigui
poetry install
```

### Running Tests

```bash
poetry run pytest
```

With coverage:
```bash
poetry run pytest --cov=lofigui --cov-report=html
```

### Type Checking

```bash
poetry run mypy lofigui
```

### Code Formatting

```bash
poetry run black lofigui tests
```

## Comparison with Alternatives

| Feature | lofigui | Streamlit | PyWebIO | Textual |
|---------|---------|-----------|---------|---------|
| JavaScript | No | Yes | Yes | No |
| Complexity | Very Low | Medium | Medium | Medium |
| Use Case | Internal tools | Data apps | Web apps | Terminal UIs |
| Learning Curve | Minimal | Moderate | Moderate | Moderate |
| Partial Updates | No | Yes | Yes | Yes |

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

- **Go version**: Even simpler implementation
- **Go WASM**: Serverless deployment option
- **HTMX integration**: Optional partial page updates
- **More examples**: Forms, authentication, file uploads

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
