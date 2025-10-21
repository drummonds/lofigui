# Go Example 01: Hello World

A minimal example demonstrating the basic usage of lofigui with Go's net/http.

## Overview

This example shows:
- Basic project structure for a lofigui Go application
- Integration with Go's standard `net/http` server
- Simple print output to web page
- MVC architecture pattern in Go

## Structure

```
01_hello_world/
├── main.go           # Main application (Model + Controller)
├── templates/
│   └── hello.html    # View (Go template)
├── go.mod            # Go module definition
└── readme.md         # This file
```

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
task go-example-01
```

Then open your browser to: http://localhost:1340

## How It Works

### 1. Model (`model()` function)

The model contains your business logic:

```go
func model() {
    lofigui.Print("Hello world from Go!")
}
```

This outputs "Hello world" to the buffer.

### 2. Controller (`Controller` struct and handler)

The controller manages state and routing:

```go
func (ctrl *Controller) handleRoot(w http.ResponseWriter, r *http.Request) {
    lofigui.Reset()  // Clear previous output
    model()          // Generate new output

    data := struct {
        Results string
    }{
        Results: lofigui.Buffer(),
    }

    ctrl.templates.ExecuteTemplate(w, "hello.html", data)
}
```

### 3. View (`templates/hello.html`)

The view renders the buffered HTML:

```html
<div class="box">
    {{.Results}}
</div>
```

The `Results` variable contains the accumulated HTML from lofigui.

## API Examples

### Print

```go
lofigui.Print("Hello world")                    // Paragraph
lofigui.Print("Inline", lofigui.WithEnd(""))   // Inline
lofigui.Printf("Value: %d", 42)                 // Formatted
```

### Markdown

```go
lofigui.Markdown("# Heading\n\nThis is **bold** text")
```

### HTML

```go
lofigui.HTML("<div class='notification'>Custom HTML</div>")
```

### Table

```go
data := [][]string{
    {"Alice", "30", "Engineer"},
    {"Bob", "25", "Designer"},
}
lofigui.Table(data, lofigui.WithHeader([]string{"Name", "Age", "Role"}))
```

## Next Steps

- Try adding more `lofigui.Print()` calls to the model
- Add markdown output with `lofigui.Markdown()`
- Add a table with `lofigui.Table()`
- Add hyperlinks to make it interactive

See example 02 for a more complex example with charts.

## Comparison with Python Version

| Feature | Python | Go |
|---------|--------|-----|
| Syntax | `lg.print()` | `lofigui.Print()` |
| Type Safety | Runtime | Compile-time |
| Performance | Interpreted | Compiled (faster) |
| Concurrency | asyncio | Goroutines (built-in) |
| Deployment | Needs Python | Single binary |

## Building for Production

```bash
# Build standalone binary
go build -o hello-world main.go

# Run the binary
./hello-world
```

The binary is self-contained and can be deployed anywhere!
