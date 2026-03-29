# Research: Python Implementation

lofigui has a Python implementation alongside the primary Go version. The Python API mirrors the Go API but uses FastAPI as the web server and Jinja2 for templates.

## Installation

### Using pip

```bash
pip install lofigui
```

### Using uv

```bash
uv add lofigui
```

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

```python
lg.markdown("# Heading\n\nThis is **bold** text")
```

#### `html(msg, ctx=None)`

Add raw HTML to buffer (no escaping). Only use with trusted input.

```python
lg.html("<div class='notification is-info'>Custom HTML</div>")
```

#### `table(table, header=None, ctx=None, escape=True)`

Generate an HTML table with Bulma styling.

```python
data = [
    ["Alice", 30, "Engineer"],
    ["Bob", 25, "Designer"],
]
lg.table(data, header=["Name", "Age", "Role"])
```

### Buffer Management

```python
content = lg.buffer()   # get accumulated HTML
lg.reset()              # clear buffer
```

### Context Management

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

```python
@app.get("/favicon.ico")
async def favicon():
    return lg.get_favicon_response()

# In template <head>: {{ get_favicon_html_tag()|safe }}

# Save to file
lg.save_favicon_ico("static/favicon.ico")
```

## Architecture

The Python implementation uses:

- **FastAPI** as the web server (async-ready)
- **Jinja2** for templates (vs pongo2 in Go)
- **asyncio.Queue** for the buffer (vs `strings.Builder` in Go)

### MVC Pattern

- **Model**: Functions that call `lg.print()`, etc.
- **View**: Jinja2 templates that render `{{ results | safe }}`
- **Controller**: FastAPI routes that orchestrate model and view

### Buffering

Server-side buffering: model functions write to a queue, `buffer()` drains the queue and returns HTML, templates render the complete page. Full page refresh — no partial DOM updates.

### Security

All output functions escape HTML by default:

```python
lg.print("<script>alert('xss')</script>")
# Output: <p>&lt;script&gt;alert('xss')&lt;/script&gt;</p>
```

Use `escape=False` or `html()` only with trusted input.

## Python Examples

```bash
task example-01          # Hello World (FastAPI)
task example-02          # SVG Graph (Pygal)
task example-05          # Demo App (template inheritance)
task example-06          # Notes CRUD
```

## Development

```bash
uv sync --all-extras     # Install dev dependencies
uv run pytest            # Run tests
uv run pytest --cov=lofigui --cov-report=html  # With coverage
uv run mypy lofigui      # Type checking
uv run black lofigui tests  # Code formatting
```

## Comparison with Python Alternatives

| Feature | lofigui | Streamlit | PyWebIO | Textual |
|---------|---------|-----------|---------|---------|
| JavaScript | No | Yes | Yes | No |
| Complexity | Very Low | Medium | Medium | Medium |
| Use Case | Internal tools | Data apps | Web apps | Terminal UIs |
| Learning Curve | Minimal | Moderate | Moderate | Moderate |
| Partial Updates | Via HTMX | Yes | Yes | Yes |
