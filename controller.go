package lofigui

import (
	"fmt"
	"net/http"
	"path/filepath"

	"github.com/flosch/pongo2/v6"
)

// Controller manages template rendering and buffer content for lofigui apps.
//
// The Controller provides:
//   - Template rendering with state
//   - Access to the output buffer
//   - Customizable template directories and paths
//
// NOTE: Action state management (polling, refresh) is now handled at the App level
// to implement the singleton active model concept. Use App methods for action control.
//
// Example usage:
//
//	// Basic usage with defaults
//	ctrl, err := lofigui.NewController(lofigui.ControllerConfig{
//	    TemplatePath: "../templates/hello.html",
//	})
//
//	// With custom settings
//	ctrl, err := lofigui.NewController(lofigui.ControllerConfig{
//	    TemplatePath: "../templates/hello.html",
//	    Name:         "My Custom Controller",
//	})
type Controller struct {
	Name     string // Name of the controller
	template *pongo2.Template
	context  *Context
}

// ControllerConfig holds configuration for creating a Controller.
type ControllerConfig struct {
	// Name is the display name for the controller.
	// Default: "Lofigui Controller"
	Name string

	// TemplatePath is the path to the template file (required).
	// Can be absolute or relative. Examples:
	//   - "../templates/hello.html"
	//   - "/opt/myapp/templates/custom.html"
	//   - "templates/page.html"
	TemplatePath string

	// Context is an optional custom Context for buffer management.
	// If nil, uses the default global context.
	Context *Context
}

// NewController creates a new Controller with the given configuration.
//
// Returns an error if the template file cannot be loaded.
//
// Example:
//
//	ctrl, err := lofigui.NewController(lofigui.ControllerConfig{
//	    TemplatePath: "../templates/hello.html",
//	    RefreshTime:  1,
//	})
//	if err != nil {
//	    log.Fatal(err)
//	}
func NewController(config ControllerConfig) (*Controller, error) {
	if config.TemplatePath == "" {
		return nil, fmt.Errorf("TemplatePath is required")
	}

	// Load template
	tmpl, err := pongo2.FromFile(config.TemplatePath)
	if err != nil {
		return nil, fmt.Errorf("failed to load template from %s: %w", config.TemplatePath, err)
	}

	// Set defaults
	if config.Name == "" {
		config.Name = "Lofigui Controller"
	}
	if config.Context == nil {
		config.Context = defaultContext
	}

	return &Controller{
		Name:     config.Name,
		template: tmpl,
		context:  config.Context,
	}, nil
}

// NewControllerFromDir creates a new Controller by loading a template from a directory.
//
// This is a convenience function that constructs the full template path.
//
// Example:
//
//	ctrl, err := lofigui.NewControllerFromDir("../templates", "hello.html")
func NewControllerFromDir(templateDir, templateName string) (*Controller, error) {
	templatePath := filepath.Join(templateDir, templateName)
	return NewController(ControllerConfig{
		TemplatePath: templatePath,
	})
}

// NOTE: Action state management (StartAction, EndAction, IsActionRunning)
// has been moved to the App level to implement the singleton active model concept.
// Use app.StartAction(), app.EndAction(), and app.IsActionRunning() instead.

// StateDict generates a template context dictionary with controller state.
//
// NOTE: This method now only provides basic controller state (request, buffer).
// Polling state and action management are now handled at the App level.
// Use app.StateDict() for complete state including polling status.
//
// Returns a pongo2.Context containing:
//   - request: The HTTP request object
//   - results: Buffer content from Print/Markdown calls
//
// You can merge additional context by using pongo2.Context.Update().
func (ctrl *Controller) StateDict(r *http.Request) pongo2.Context {
	ctx := pongo2.Context{
		"request": r,
		"results": ctrl.context.Buffer(),
	}

	return ctx
}

// NOTE: HandleRoot has been moved to the App level to implement the singleton
// active model concept. Use app.HandleRoot() instead.

// HandleDisplay renders the template with the provided context.
//
// NOTE: This method now only handles template rendering. For complete state
// management including polling status, use app.HandleDisplay() or app.StateDict().
//
// This function:
//  1. Generates basic state dict with buffer content
//  2. Merges extra context if provided
//  3. Renders the template
//
// Example:
//
//	http.HandleFunc("/display", func(w http.ResponseWriter, r *http.Request) {
//	    ctrl.HandleDisplay(w, r, nil)
//	})
//
//	// With extra context
//	http.HandleFunc("/display", func(w http.ResponseWriter, r *http.Request) {
//	    extra := pongo2.Context{"title": "My Page"}
//	    ctrl.HandleDisplay(w, r, extra)
//	})
func (ctrl *Controller) HandleDisplay(w http.ResponseWriter, r *http.Request, extraContext pongo2.Context) {
	data := ctrl.StateDict(r)

	// Merge extra context if provided
	if extraContext != nil {
		data.Update(extraContext)
	}

	// Render template
	if err := ctrl.template.ExecuteWriter(data, w); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// ServeHTTP allows Controller to be used as an http.Handler.
// It serves the display page by default.
func (ctrl *Controller) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctrl.HandleDisplay(w, r, nil)
}

// RenderTemplate renders the controller's template with custom context.
// This is useful for one-off custom rendering.
func (ctrl *Controller) RenderTemplate(w http.ResponseWriter, context pongo2.Context) error {
	return ctrl.template.ExecuteWriter(context, w)
}

// GetTemplate returns the underlying pongo2 template.
// This allows advanced users to work directly with the template if needed.
func (ctrl *Controller) GetTemplate() *pongo2.Template {
	return ctrl.template
}

// ReloadTemplate reloads the template from the original path.
// This is useful during development when templates change.
func (ctrl *Controller) ReloadTemplate(templatePath string) error {
	tmpl, err := pongo2.FromFile(templatePath)
	if err != nil {
		return fmt.Errorf("failed to reload template: %w", err)
	}
	ctrl.template = tmpl
	return nil
}
