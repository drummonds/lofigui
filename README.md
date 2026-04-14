# lofigui

<!-- auto:version -->Latest: v0.17.34<!-- /auto:version -->

**Lofi GUI** - A minimalist Go library for creating really simple web-based GUIs for CLI tools and small projects. It provides a print-like interface for building lightweight web UIs with minimal complexity.

The application is where you have a single real object (e.g. machine or long-running process) which then has a number of pages around it to show various aspects of it.

[![Go Version](https://img.shields.io/badge/go-1.21+-blue.svg)](https://golang.org/dl/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

## Overview

`lofigui` provides a print-like interface for building lightweight web applications with minimal complexity. Perfect for:

- Creating quick GUIs for command-line tools
- Internal tools for small teams (1-10 users)
- Single-process or single-object front-ends
- Rapid prototyping without JavaScript overhead

## Key Features

- **Simple API**: Print-like interface (`Print()`, `Markdown()`, `HTML()`, `Table()`)
- **No JavaScript**: Pure HTML/CSS using the Bulma framework
- **Single binary**: Deploy as one executable, no dependencies
- **WebAssembly**: Same code runs in the browser via WASM
- **Built-in layouts**: `LayoutSingle`, `LayoutNavbar`, `LayoutThreePanel`

## Quick Start

```go
package main

import (
    "net/http"
    "codeberg.org/hum3/lofigui"
)

func model(app *lofigui.App) {
    lofigui.Print("Hello world")
    app.EndAction()
}

func main() {
    ctrl, _ := lofigui.NewControllerWithLayout(lofigui.LayoutNavbar, "My App")
    app := lofigui.NewApp()
    app.SetController(ctrl)
    app.SetDisplayURL("/display")

    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        app.HandleRoot(w, r, model, true)
    })
    http.HandleFunc("/display", func(w http.ResponseWriter, r *http.Request) {
        app.HandleDisplay(w, r)
    })
    http.HandleFunc("/favicon.ico", lofigui.ServeFavicon)
    http.ListenAndServe(":8080", nil)
}
```

## Element

Your project is essentially a web site. To make design simple you completely refresh pages so no code for partial refreshes. To make things dynamic it has to be asynchronous, using goroutines for background processing.

Like a normal terminal program you essentially just print things to a screen but now have the ability to print enriched objects.

### Model View Controller architecture

All I really want to do is to write the model. The controller and view (in the browser and templating system) are a necessary evil. The controller includes the routing and web server. The controller is split between the app (single instance) and a model-specific controller. The view is the HTML templating (pongo2) and the browser.

### Buffer

In order to decouple the display from the output and be able to refresh, you need to buffer the output. It is more efficient to buffer the output in the browser but more complicated. Moving the buffer to the server simplifies the software but requires you to refresh the whole page. lofigui relies on hyperlinks to perform updates. HTMX plays nicely here for improving interactivity (examples 09-10).

## Installation

```bash
go get codeberg.org/hum3/lofigui
```

## Examples & Documentation

See the **[documentation site](https://h3-lofigui.statichost.page/)** for interactive examples (including WASM demos), research notes, and the roadmap.

To run examples locally, use [Task](https://taskfile.dev/):

```bash
task --list              # Show all available tasks
task go-example:09       # Run any Go example by number
```

## Python

A Python implementation also exists with the same API using FastAPI and Jinja2. See [Python Notes](https://h3-lofigui.statichost.page/research-python.html) for installation, API reference, and development instructions.

## Roadmap

See [ROADMAP.md](ROADMAP.md) for planned features and future direction.

## Contributing

Contributions are welcome! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

## License

MIT License - see [LICENSE](LICENSE) file for details.

## Author

Humphrey Drummond - [hum3@drummond.info](mailto:hum3@drummond.info)

## Links

<!-- auto:links -->
| | |
|---|---|
| Documentation | https://h3-lofigui.statichost.page/ |
| PyPI | https://pypi.org/project/lofigui/ |
| Source (Codeberg) | https://codeberg.org/hum3/lofigui |
| Mirror (GitHub) | https://github.com/drummonds/lofigui |
<!-- /auto:links -->
