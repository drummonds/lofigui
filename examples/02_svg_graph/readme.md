# Example 02: SVG Graph

Demonstrates rendering SVG charts using Pygal with lofigui.

## Overview

This example shows:
- Integration with Pygal chart library
- Rendering SVG graphics in the browser
- Using `lg.html()` for raw HTML/SVG output
- More complex model logic

## Structure

```
02_svg_graph/
├── graph.py          # Main application with chart rendering
├── templates/
│   └── hello.html   # View (Jinja2 template)
├── pyproject.toml   # Dependencies (includes pygal)
└── readme.md        # This file
```

## Installation

From this directory:

```bash
poetry install
```

This will install both lofigui and pygal.

## Running the Example

```bash
poetry run python graph.py
```

Then open your browser to: http://127.0.0.1:1340

## How It Works

### Chart Generation

The model creates a bar chart using Pygal:

```python
def model():
    lg.print("Hello to graph.")

    # Create a bar chart
    bar_chart = pygal.Bar()
    bar_chart.add('Fibonacci', [0, 1, 1, 2, 3, 5, 8, 13, 21, 34, 55])

    # Render to SVG and output as raw HTML
    lg.html(bar_chart.render().decode("utf-8"))
```

Key points:
- `pygal.Bar()` creates a bar chart object
- `bar_chart.render()` generates SVG as bytes
- `.decode("utf-8")` converts bytes to string
- `lg.html()` passes the SVG through without escaping

### Why `lg.html()` instead of `lg.print()`?

SVG is valid HTML/XML, so it needs to be passed through without escaping:
- `lg.print()` would escape `<svg>` tags (default behavior)
- `lg.html()` passes raw HTML through unchanged
- Always use `lg.html()` only with trusted content!

## Experiment

Try different chart types:

```python
# Line chart
line_chart = pygal.Line()
line_chart.add('Fibonacci', [0, 1, 1, 2, 3, 5, 8, 13, 21, 34, 55])

# Pie chart
pie_chart = pygal.Pie()
pie_chart.add('Chrome', 36.3)
pie_chart.add('Firefox', 10.8)
pie_chart.add('Safari', 8.2)

# Radar chart
radar_chart = pygal.Radar()
radar_chart.add('Alice', [10, 9, 8, 7, 6])
radar_chart.add('Bob', [8, 8, 8, 8, 8])
```

## Next Steps

- Combine charts with tables using `lg.table()`
- Add user interaction with hyperlinks (e.g., filter data)
- Display multiple charts on one page
- Add dynamic data from files or APIs

See the [Pygal documentation](http://www.pygal.org/) for more chart options.
