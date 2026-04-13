package lofigui

import (
	"bytes"
	"fmt"
	"html/template"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
)

// TemplateContext is the data passed to templates during rendering.
// It replaces pongo2.Context as a simple map of string keys to values.
type TemplateContext map[string]interface{}

// Update merges all key-value pairs from other into this context.
// Keys in other overwrite existing keys.
func (ctx TemplateContext) Update(other TemplateContext) {
	for k, v := range other {
		ctx[k] = v
	}
}

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
	Name         string // Name of the controller
	template     *template.Template
	templateName string // which named template to execute
	context      *Context
}

// ControllerConfig holds configuration for creating a Controller.
type ControllerConfig struct {
	// Name is the display name for the controller.
	// Default: "Lofigui Controller"
	Name string

	// TemplatePath is the path to the template file.
	// Can be absolute or relative. Examples:
	//   - "../templates/hello.html"
	//   - "/opt/myapp/templates/custom.html"
	//   - "templates/page.html"
	// Either TemplatePath or TemplateString must be provided.
	TemplatePath string

	// TemplateString is the template content as a string.
	// Use this for embedded templates (e.g. via Go's embed package).
	// Either TemplatePath or TemplateString must be provided.
	TemplateString string

	// Context is an optional custom Context for buffer management.
	// If nil, uses the default global context.
	Context *Context
}

// NewController creates a new Controller with the given configuration.
//
// Either TemplatePath or TemplateString must be provided. If both are set,
// TemplateString takes precedence.
//
// Returns an error if the template cannot be loaded or parsed.
//
// Example:
//
//	// From file:
//	ctrl, err := lofigui.NewController(lofigui.ControllerConfig{
//	    TemplatePath: "../templates/hello.html",
//	})
//
//	// From embedded string:
//	//go:embed templates/hello.html
//	var helloTemplate string
//	ctrl, err := lofigui.NewController(lofigui.ControllerConfig{
//	    TemplateString: helloTemplate,
//	})
func NewController(config ControllerConfig) (*Controller, error) {
	var tmpl *template.Template
	var tplName string
	var err error

	switch {
	case config.TemplateString != "":
		tplName = "inline"
		tmpl, err = template.New(tplName).Parse(config.TemplateString)
		if err != nil {
			return nil, fmt.Errorf("failed to parse template string: %w", err)
		}
	case config.TemplatePath != "":
		tplName = filepath.Base(config.TemplatePath)
		tmpl, err = parseWithBase(config.TemplatePath)
		if err != nil {
			return nil, fmt.Errorf("failed to load template from %s: %w", config.TemplatePath, err)
		}
	default:
		return nil, fmt.Errorf("either TemplatePath or TemplateString is required")
	}

	// Set defaults
	if config.Name == "" {
		config.Name = "Lofigui Controller"
	}
	if config.Context == nil {
		config.Context = defaultContext
	}

	return &Controller{
		Name:         config.Name,
		template:     tmpl,
		templateName: tplName,
		context:      config.Context,
	}, nil
}

// parseWithBase loads a template file and, if a base.html exists in the same
// directory, parses both so that Go template inheritance ({{block}}/{{define}})
// works. Returns the template set with the child template's base name as root.
func parseWithBase(templatePath string) (*template.Template, error) {
	tplName := filepath.Base(templatePath)
	dir := filepath.Dir(templatePath)
	basePath := filepath.Join(dir, "base.html")

	if tplName != "base.html" {
		if _, err := os.Stat(basePath); err == nil {
			// Parse base first, then child overrides its blocks
			return template.New("base.html").ParseFiles(basePath, templatePath)
		}
	}
	return template.New(tplName).ParseFiles(templatePath)
}

// NewControllerFromDir creates a new Controller by loading a template from a directory.
//
// If base.html exists in the directory alongside the requested template,
// both are parsed together to support Go template inheritance ({{block}}/{{define}}).
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

// NewControllerFromString creates a new Controller from a template string.
//
// This is a convenience function for embedded templates.
//
// Example:
//
//	//go:embed templates/hello.html
//	var helloTemplate string
//	ctrl, err := lofigui.NewControllerFromString(helloTemplate)
func NewControllerFromString(templateString string) (*Controller, error) {
	return NewController(ControllerConfig{
		TemplateString: templateString,
	})
}

// NewControllerFromFS creates a Controller from an fs.FS (e.g. embed.FS).
//
// This enables template inheritance ({{block}}/{{define}}) to work with embedded
// filesystems, including in WASM builds where there is no local filesystem.
// If base.html exists in the directory, it is parsed alongside the named template.
//
// Example:
//
//	//go:embed templates
//	var templateFS embed.FS
//
//	ctrl, err := lofigui.NewControllerFromFS(templateFS, "templates", "page.html")
func NewControllerFromFS(fsys fs.FS, dir string, templateName string) (*Controller, error) {
	subFS, err := fs.Sub(fsys, dir)
	if err != nil {
		return nil, fmt.Errorf("failed to open subdirectory %s: %w", dir, err)
	}

	// Determine which files to parse: base.html + child (if base exists)
	tplName := templateName
	files := []string{templateName}

	if templateName != "base.html" {
		if _, err := fs.Stat(subFS, "base.html"); err == nil {
			files = []string{"base.html", templateName}
			tplName = "base.html"
		}
	}

	tmpl, err := template.New(tplName).ParseFS(subFS, files...)
	if err != nil {
		return nil, fmt.Errorf("failed to load template %s from FS: %w", templateName, err)
	}
	return &Controller{
		Name:         "Lofigui Controller",
		template:     tmpl,
		templateName: tplName,
		context:      defaultContext,
	}, nil
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
// Returns a TemplateContext containing:
//   - request: The HTTP request object
//   - results: Buffer content from Print/Markdown calls (as template.HTML)
//
// You can merge additional context by using TemplateContext.Update().
func (ctrl *Controller) StateDict(r *http.Request) TemplateContext {
	ctx := TemplateContext{
		"request": r,
		"results": template.HTML(ctrl.context.Buffer()),
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
//	    extra := lofigui.TemplateContext{"title": "My Page"}
//	    ctrl.HandleDisplay(w, r, extra)
//	})
func (ctrl *Controller) HandleDisplay(w http.ResponseWriter, r *http.Request, extraContext TemplateContext) {
	data := ctrl.StateDict(r)

	// Merge extra context if provided
	if extraContext != nil {
		data.Update(extraContext)
	}

	// Render template
	if err := ctrl.template.ExecuteTemplate(w, ctrl.templateName, data); err != nil {
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
func (ctrl *Controller) RenderTemplate(w http.ResponseWriter, context TemplateContext) error {
	return ctrl.template.ExecuteTemplate(w, ctrl.templateName, context)
}

// RenderToString renders the controller's template to a string.
// This is useful for WASM builds where there is no http.ResponseWriter.
func (ctrl *Controller) RenderToString(context TemplateContext) (string, error) {
	var buf bytes.Buffer
	if err := ctrl.template.ExecuteTemplate(&buf, ctrl.templateName, context); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// GetTemplate returns the underlying html/template.Template.
// This allows advanced users to work directly with the template if needed.
func (ctrl *Controller) GetTemplate() *template.Template {
	return ctrl.template
}

// ReloadTemplate reloads the template from the original path.
// This is useful during development when templates change.
func (ctrl *Controller) ReloadTemplate(templatePath string) error {
	tmpl, err := parseWithBase(templatePath)
	if err != nil {
		return fmt.Errorf("failed to reload template: %w", err)
	}
	ctrl.template = tmpl
	ctrl.templateName = filepath.Base(templatePath)
	return nil
}
