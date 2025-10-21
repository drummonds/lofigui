# Example 01: Hello World

A minimal example demonstrating the basic usage of lofigui with FastAPI.

## Overview

This example shows:
- Basic project structure for a lofigui application
- Integration with FastAPI and Uvicorn
- Simple print output to web page
- MVC architecture pattern

## Structure

```
01_hello_world/
├── hello.py           # Main application (Model + Controller)
├── templates/
│   └── hello.html    # View (Jinja2 template)
├── pyproject.toml    # Dependencies
└── readme.md         # This file
```

## Installation

From this directory:

```bash
uv sync
```

## Running the Example

```bash
uv run python hello.py
```

Then open your browser to: http://127.0.0.1:1340

## How It Works

### 1. Model (`model()` function)

The model contains your business logic:

```python
def model():
    lg.print("Hello world.")
```

This simply outputs "Hello world" to the buffer.

### 2. Controller (`Controller` class and route)

The controller manages state and routing:

```python
@app.get("/", response_class=HTMLResponse)
async def root(request: Request):
    lg.reset()  # Clear previous output
    model()     # Generate new output
    return templates.TemplateResponse(
        "hello.html", controller.state_dict({"request": request})
    )
```

### 3. View (`templates/hello.html`)

The view renders the buffered HTML:

```html
<h1>Test start</h1>
{{results | safe}}
```

The `results` variable contains the accumulated HTML from lofigui.

## Next Steps

- Try adding more `lg.print()` calls to the model
- Add markdown output with `lg.markdown("## Heading")`
- Add a table with `lg.table([["row1"]], header=["Column"])`
- Add hyperlinks to make it interactive

See example 02 for a more complex example with charts.
