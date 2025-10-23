# Go Controller Guide

The lofigui Controller provides extensible logic for managing application state, routing, and template rendering in Go applications. This guide shows how to customize it for your needs.

## Quick Start

### Option 1: Using ControllerConfig (Recommended)

The most flexible approach using the configuration struct:

```go
package main

import (
    "log"
    "net/http"
    "time"
    
    "github.com/drummonds/lofigui"
)

func model(ctrl *lofigui.Controller) {
    lofigui.Print("Processing...")
    for i := 0; i < 5; i++ {
        time.Sleep(1 * time.Second)
        lofigui.Printf("Step %d", i)
    }
    lofigui.Print("Done!")
    ctrl.EndAction()
}

func main() {
    // Create controller with custom settings
    ctrl, err := lofigui.NewController(lofigui.ControllerConfig{
        TemplatePath: "../templates/hello.html", // Custom location!
        RefreshTime:  1,                          // Refresh every 1 second
        DisplayURL:   "/display",                 // Where to show results
    })
    if err != nil {
        log.Fatal(err)
    }

    // Setup routes using helper methods
    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        ctrl.HandleRoot(w, r, model, true)
    })

    http.HandleFunc("/display", func(w http.ResponseWriter, r *http.Request) {
        ctrl.HandleDisplay(w, r, nil)
    })

    http.HandleFunc("/favicon.ico", lofigui.ServeFavicon)

    log.Fatal(http.ListenAndServe(":1340", nil))
}
```

### Option 2: Using NewControllerFromDir (Simpler)

A convenience function for the common case of directory + filename:

```go
func main() {
    // Create controller from directory and filename
    ctrl, err := lofigui.NewControllerFromDir("../templates", "hello.html", 1)
    if err != nil {
        log.Fatal(err)
    }

    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        ctrl.HandleRoot(w, r, model, true)
    })

    // Can use ServeHTTP directly as http.Handler!
    http.HandleFunc("/display", ctrl.ServeHTTP)

    http.HandleFunc("/favicon.ico", lofigui.ServeFavicon)

    log.Fatal(http.ListenAndServe(":1340", nil))
}
```

## Controller Customization

### Custom Template Directory

The template directory can be anywhere, not just a default location:

```go
// Relative paths
ctrl, _ := lofigui.NewController(lofigui.ControllerConfig{
    TemplatePath: "../templates/hello.html",
})

// Absolute paths
ctrl, _ := lofigui.NewController(lofigui.ControllerConfig{
    TemplatePath: "/opt/myapp/templates/hello.html",
})

// Different subdirectory
ctrl, _ := lofigui.NewController(lofigui.ControllerConfig{
    TemplatePath: "custom_templates/page.html",
})
```

### Custom Refresh Time

Change how often the page auto-refreshes during actions:

```go
// Refresh every 2 seconds
ctrl, _ := lofigui.NewController(lofigui.ControllerConfig{
    TemplatePath: "templates/hello.html",
    RefreshTime:  2,
})

// Or change it later
ctrl.SetRefreshTime(5)
```

### Custom Display URL

Customize where the root endpoint redirects to:

```go
ctrl, _ := lofigui.NewController(lofigui.ControllerConfig{
    TemplatePath: "templates/hello.html",
    DisplayURL:   "/status", // Custom URL
})
```

### Custom Context

Use a separate buffer context instead of the global one:

```go
customContext := lofigui.NewContext()

ctrl, _ := lofigui.NewController(lofigui.ControllerConfig{
    TemplatePath: "templates/hello.html",
    Context:      customContext,
})

// Now this controller uses its own isolated buffer
```

## Controller API Reference

### ControllerConfig

```go
type ControllerConfig struct {
    TemplatePath string    // Path to template file (required)
    RefreshTime  int       // Seconds between refresh (default: 1)
    DisplayURL   string    // URL to redirect to (default: "/display")
    Context      *Context  // Custom context (default: global context)
}
```

### Constructor Functions

```go
// Full configuration
func NewController(config ControllerConfig) (*Controller, error)

// Convenience function for directory + filename
func NewControllerFromDir(templateDir, templateName string, refreshTime int) (*Controller, error)
```

### Controller Methods

#### Action Management

```go
// Start an action and enable auto-refresh
func (ctrl *Controller) StartAction()

// Stop the action and disable auto-refresh
func (ctrl *Controller) EndAction()

// Check if an action is currently running
func (ctrl *Controller) IsActionRunning() bool

// Update the refresh time
func (ctrl *Controller) SetRefreshTime(seconds int)
```

#### Template Context

```go
// Generate template context with current state
func (ctrl *Controller) StateDict(r *http.Request) pongo2.Context
```

Returns a `pongo2.Context` containing:
- `request`: The HTTP request object
- `results`: Buffer content from Print/Markdown calls
- `refresh`: Meta tag for auto-refresh (if action is running)

#### Route Helpers

```go
// Helper for root endpoint - starts action and redirects
func (ctrl *Controller) HandleRoot(
    w http.ResponseWriter,
    r *http.Request,
    modelFunc func(*Controller),
    resetBuffer bool,
)

// Helper for display endpoint - shows progress
func (ctrl *Controller) HandleDisplay(
    w http.ResponseWriter,
    r *http.Request,
    extraContext pongo2.Context,
)

// Implements http.Handler interface (calls HandleDisplay with nil context)
func (ctrl *Controller) ServeHTTP(w http.ResponseWriter, r *http.Request)
```

#### Advanced Methods

```go
// Render template with custom context
func (ctrl *Controller) RenderTemplate(w http.ResponseWriter, context pongo2.Context) error

// Get the underlying pongo2 template
func (ctrl *Controller) GetTemplate() *pongo2.Template

// Reload template from disk (useful during development)
func (ctrl *Controller) ReloadTemplate(templatePath string) error
```

## Advanced Patterns

### Custom Route Handlers

You can build custom routing patterns:

```go
http.HandleFunc("/start/", func(w http.ResponseWriter, r *http.Request) {
    taskID := r.URL.Path[len("/start/"):]
    
    lofigui.Reset()
    ctrl.StartAction()
    
    go func() {
        lofigui.Printf("Running task %s", taskID)
        // Task logic here
        ctrl.EndAction()
    }()
    
    http.Redirect(w, r, "/status/"+taskID, http.StatusSeeOther)
})
```

### Extra Template Context

Pass additional variables to your template:

```go
http.HandleFunc("/display", func(w http.ResponseWriter, r *http.Request) {
    extra := pongo2.Context{
        "title": "My Custom Title",
        "user":  "John Doe",
    }
    ctrl.HandleDisplay(w, r, extra)
})
```

### Multiple Controllers

Use different controllers for different pages:

```go
mainCtrl, _ := lofigui.NewController(lofigui.ControllerConfig{
    TemplatePath: "templates/main.html",
})

adminCtrl, _ := lofigui.NewController(lofigui.ControllerConfig{
    TemplatePath: "templates/admin.html",
    DisplayURL:   "/admin/status",
})

http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
    mainCtrl.HandleRoot(w, r, mainModel, true)
})

http.HandleFunc("/admin", func(w http.ResponseWriter, r *http.Request) {
    adminCtrl.HandleRoot(w, r, adminModel, true)
})
```

### Custom Template Rendering

For advanced use cases, render templates directly:

```go
http.HandleFunc("/custom", func(w http.ResponseWriter, r *http.Request) {
    context := pongo2.Context{
        "data":    myData,
        "results": lofigui.Buffer(),
    }
    
    if err := ctrl.RenderTemplate(w, context); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
    }
})
```

### Development Mode with Template Reloading

Reload templates on each request during development:

```go
http.HandleFunc("/display", func(w http.ResponseWriter, r *http.Request) {
    // Reload template (only do this in development!)
    if err := ctrl.ReloadTemplate("../templates/hello.html"); err != nil {
        log.Printf("Template reload error: %v", err)
    }
    ctrl.HandleDisplay(w, r, nil)
})
```

## Integration with Web Frameworks

### Standard net/http (shown above)

The controller works seamlessly with Go's standard library.

### With Gorilla Mux

```go
import "github.com/gorilla/mux"

func main() {
    ctrl, _ := lofigui.NewController(lofigui.ControllerConfig{
        TemplatePath: "templates/hello.html",
    })

    r := mux.NewRouter()
    r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        ctrl.HandleRoot(w, r, model, true)
    })
    r.HandleFunc("/display", ctrl.ServeHTTP)
    r.HandleFunc("/favicon.ico", lofigui.ServeFavicon)

    http.ListenAndServe(":1340", r)
}
```

### With Chi

```go
import "github.com/go-chi/chi/v5"

func main() {
    ctrl, _ := lofigui.NewController(lofigui.ControllerConfig{
        TemplatePath: "templates/hello.html",
    })

    r := chi.NewRouter()
    r.Get("/", func(w http.ResponseWriter, r *http.Request) {
        ctrl.HandleRoot(w, r, model, true)
    })
    r.Get("/display", ctrl.ServeHTTP)
    r.Get("/favicon.ico", lofigui.ServeFavicon)

    http.ListenAndServe(":1340", r)
}
```

## Migration from Old Pattern

### Before (controller in main.go)

```go
type Controller struct {
    template      *pongo2.Template
    actionRunning bool
}

func NewController() *Controller {
    tmpl, err := pongo2.FromFile("../templates/hello.html")
    if err != nil {
        log.Fatal(err)
    }
    return &Controller{template: tmpl}
}

func (ctrl *Controller) handleDisplay(w http.ResponseWriter, r *http.Request) {
    data := ctrl.StateDict(r)
    if err := ctrl.template.ExecuteWriter(data, w); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
    }
}
```

### After (using library controller)

```go
ctrl, err := lofigui.NewController(lofigui.ControllerConfig{
    TemplatePath: "../templates/hello.html",
})
if err != nil {
    log.Fatal(err)
}

http.HandleFunc("/display", func(w http.ResponseWriter, r *http.Request) {
    ctrl.HandleDisplay(w, r, nil)
})
```

Much cleaner! The controller logic is now in the library where it belongs.

## Template Requirements

Your templates should include the `{{refresh|safe}}` tag for auto-refresh:

```html
<!doctype html>
<html>
<head>
    <title>My App</title>
    {{refresh | safe}}
</head>
<body>
    <h1>Status</h1>
    {{ results | safe }}
</body>
</html>
```

The controller will automatically populate:
- `refresh`: Auto-refresh meta tag (when action is running)
- `results`: Buffered output from Print/Markdown calls
- `request`: The HTTP request object

## Error Handling

Always check errors when creating controllers:

```go
ctrl, err := lofigui.NewController(lofigui.ControllerConfig{
    TemplatePath: "../templates/hello.html",
})
if err != nil {
    log.Fatalf("Failed to create controller: %v", err)
}
```

Common errors:
- Template file not found
- Template syntax errors
- Invalid template path

## Best Practices

1. **Template Paths**: Use relative paths from your executable location
2. **Error Handling**: Always check controller creation errors
3. **Buffer Management**: Call `lofigui.Reset()` at the start of actions to clear previous output
4. **Action Cleanup**: Always call `ctrl.EndAction()` when your model function completes
5. **Development**: Use `ctrl.ReloadTemplate()` during development for faster iteration
6. **Production**: Preload templates once at startup for better performance

## Complete Example

See `examples/01_hello_world/go/main.go` for a complete working example that demonstrates all the key features.
