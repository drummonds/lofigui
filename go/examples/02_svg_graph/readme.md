# Go Example 02: SVG Graph

Demonstrates rendering SVG charts using go-chart with lofigui.

## Overview

This example shows:
- Integration with go-chart library
- Rendering SVG graphics in the browser
- Using `lofigui.HTML()` for raw HTML/SVG output
- More complex model logic

## Installation

From this directory:

```bash
go mod download
```

## Running the Example

```bash
go run main.go
```

Or using Task from repo root:
```bash
task go-example-02
```

Then open your browser to: http://localhost:1340

## How It Works

### Chart Generation

The model creates a bar chart using go-chart:

```go
func model() {
    lofigui.Print("Hello to SVG graphs in Go!")

    // Create bar chart data
    fibonacci := []float64{0, 1, 1, 2, 3, 5, 8, 13, 21, 34, 55}

    barChart := chart.BarChart{
        Title: "Fibonacci Sequence",
        Width: 800,
        Height: 400,
        Bars: bars,
    }

    // Render to SVG and output as raw HTML
    svg := renderChartToSVG(&barChart)
    lofigui.HTML(svg)
}
```

Key points:
- `go-chart` creates the chart object
- `Render(chart.SVG, ...)` generates SVG
- `lofigui.HTML()` passes the SVG through without escaping

### Why `lofigui.HTML()` instead of `lofigui.Print()`?

SVG is valid HTML/XML, so it needs to be passed through without escaping:
- `lofigui.Print()` would escape `<svg>` tags (default behavior)
- `lofigui.HTML()` passes raw HTML through unchanged
- Always use `lofigui.HTML()` only with trusted content!

## Building for Production

```bash
# Build standalone binary
go build -o svg-graph main.go

# Run the binary
./svg-graph
```

## Comparison with Python Version

| Feature | Python (Pygal) | Go (go-chart) |
|---------|----------------|---------------|
| Chart Library | Pygal | go-chart |
| Performance | Slower | Much faster |
| Deployment | Needs Python | Single binary |
| Memory | Higher | Lower |
| Startup | 100-500ms | 1-10ms |

## Next Steps

- Combine charts with tables using `lofigui.Table()`
- Add user interaction with URL parameters
- Display multiple charts on one page
- Add dynamic data from databases or APIs
