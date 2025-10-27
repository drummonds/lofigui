package lofigui

import (
	"fmt"
	"net/http"
	"path/filepath"

	"github.com/flosch/pongo2/v6"
)

// Controller manages application state and routing for lofigui apps.
//
// The Controller provides extensible logic for:
//   - Managing action state (running/stopped)
//   - Auto-refresh polling during actions
//   - Template rendering with state
//   - Customizable template directories and paths
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
//	    TemplatePath:  "../templates/hello.html",
//	    RefreshTime:   2,
//	    DisplayURL:    "/display",
//	})
type Controller struct {
	template      *pongo2.Template
	actionRunning bool
	polling       bool
	PollCount     int
	refreshTime   int    // seconds between refresh
	displayURL    string // URL to redirect to for display
	context       *Context
}

// ControllerConfig holds configuration for creating a Controller.
type ControllerConfig struct {
	// TemplatePath is the path to the template file (required).
	// Can be absolute or relative. Examples:
	//   - "../templates/hello.html"
	//   - "/opt/myapp/templates/custom.html"
	//   - "templates/page.html"
	TemplatePath string

	// RefreshTime is the number of seconds between auto-refresh when action is running.
	// Default: 1 second
	RefreshTime int

	// DisplayURL is the URL to redirect to after starting an action.
	// Default: "/display"
	DisplayURL string

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
	if config.RefreshTime <= 0 {
		config.RefreshTime = 1
	}
	if config.DisplayURL == "" {
		config.DisplayURL = "/display"
	}
	if config.Context == nil {
		config.Context = defaultContext
	}

	return &Controller{
		template:      tmpl,
		actionRunning: false,
		refreshTime:   config.RefreshTime,
		displayURL:    config.DisplayURL,
		context:       config.Context,
	}, nil
}

// NewControllerFromDir creates a new Controller by loading a template from a directory.
//
// This is a convenience function that constructs the full template path.
//
// Example:
//
//	ctrl, err := lofigui.NewControllerFromDir("../templates", "hello.html", 1)
func NewControllerFromDir(templateDir, templateName string, refreshTime int) (*Controller, error) {
	templatePath := filepath.Join(templateDir, templateName)
	return NewController(ControllerConfig{
		TemplatePath: templatePath,
		RefreshTime:  refreshTime,
	})
}

// StartAction starts an action and enables auto-refresh polling.
func (ctrl *Controller) StartAction() {
	ctrl.actionRunning = true
	ctrl.polling = true
	ctrl.PollCount = 0
}

// EndAction stops the action and disables auto-refresh.
func (ctrl *Controller) EndAction() {
	ctrl.actionRunning = false
	ctrl.polling = false
}

// IsActionRunning returns whether an action is currently running.
func (ctrl *Controller) IsActionRunning() bool {
	return ctrl.actionRunning
}

// SetRefreshTime updates the refresh time (in seconds) for auto-refresh.
func (ctrl *Controller) SetRefreshTime(seconds int) {
	ctrl.refreshTime = seconds
}

// StateDict generates a template context dictionary with current state.
//
// Returns a pongo2.Context containing:
//   - request: The HTTP request object
//   - results: Buffer content from Print/Markdown calls
//   - refresh: Meta tag for auto-refresh (if action is running)
//
// You can merge additional context by using pongo2.Context.Update().
func (ctrl *Controller) StateDict(r *http.Request) pongo2.Context {
	ctx := pongo2.Context{
		"request": r,
		"results": ctrl.context.Buffer(),
	}

	if ctrl.IsActionRunning() {
		ctx["polling"] = "Running"
		ctrl.PollCount += 1
		ctx["refresh"] = fmt.Sprintf(
			`<meta http-equiv="Refresh" content="%d; URL=%s"/>`,
			ctrl.refreshTime,
			ctrl.displayURL,
		)
	} else {
		ctx["refresh"] = ""
		ctrl.PollCount = 0
		ctx["polling"] = "Stopped"
	}
	ctx["poll_count"] = ctrl.PollCount

	return ctx
}

// HandleRoot is a helper for the root endpoint that starts an action.
//
// This function:
//  1. Resets the buffer (if resetBuffer is true)
//  2. Starts the action
//  3. Launches the model function in a goroutine
//  4. Returns HTML that redirects to the display page
//
// Example:
//
//	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
//	    ctrl.HandleRoot(w, r, model, true)
//	})
func (ctrl *Controller) HandleRoot(w http.ResponseWriter, r *http.Request, modelFunc func(*Controller), resetBuffer bool) {
	if resetBuffer {
		ctrl.context.Reset()
	}

	ctrl.StartAction()
	go modelFunc(ctrl)

	w.Header().Set("Content-Type", "text/html")
	fmt.Fprintf(w, `<head><meta http-equiv="Refresh" content="0; URL=%s"/></head>`, ctrl.displayURL)
}

// HandleDisplay is a helper for the display endpoint that shows action progress.
//
// This function:
//  1. Generates state dict with current buffer and refresh status
//  2. Renders the template with the state
//
// You can pass additional context via extraContext which will be merged into the state.
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
