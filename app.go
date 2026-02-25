package lofigui

import (
	"context"
	"fmt"
	"net/http"
	"sync"

	"github.com/flosch/pongo2/v6"
)

// App provides a wrapper around a Controller with safe controller replacement.
//
// The app manages:
//   - Action state (running/stopped)
//   - Auto-refresh polling during actions
//   - Version information
//   - Controller lifecycle and composition
//
// When replacing a controller, App ensures that any running action is safely
// stopped before the new controller is set. This prevents dangling goroutines
// and ensures clean state transitions.
//
// Example usage:
//
//	app := lofigui.NewApp()
//	ctrl, err := lofigui.NewController(lofigui.ControllerConfig{
//	    TemplatePath: "templates/page.html",
//	})
//	if err != nil {
//	    log.Fatal(err)
//	}
//	app.SetController(ctrl)
type App struct {
	controller    *Controller
	Version       string // Version/name of the application
	actionRunning bool   // Whether an action is currently running (singleton active model)
	polling       bool   // Whether auto-refresh polling is enabled
	PollCount     int    // Number of polling cycles
	refreshTime   int    // Seconds between refresh when polling
	displayURL    string // URL to redirect to for display
	cancelFunc    context.CancelFunc
	mu            sync.RWMutex
}

// NewApp creates a new App with no controller.
func NewApp() *App {
	return &App{
		Version:     "Lofigui",
		refreshTime: 1,
		displayURL:  "/display",
	}
}

// NewAppWithController creates a new App with the given controller.
func NewAppWithController(ctrl *Controller) *App {
	app := &App{
		Version:     "Lofigui",
		refreshTime: 1,
		displayURL:  "/display",
	}
	app.SetController(ctrl)
	return app
}

// GetController returns the current controller.
// Returns nil if no controller is set.
func (app *App) GetController() *Controller {
	app.mu.RLock()
	defer app.mu.RUnlock()
	return app.controller
}

// SetController sets a new controller with safe cleanup of the existing controller.
//
// If there's an existing controller with a running action, this will safely
// stop the action before replacing with the new controller. This prevents
// dangling goroutines and ensures clean state transitions.
//
// The cleanup logic is defensive - it handles controllers that may not have
// standard methods implemented and silently ignores any errors during cleanup.
//
// This method is idempotent - if the same controller is being set again,
// no cleanup is performed and the running action continues.
//
// Args:
//   - ctrl: The new controller to set (can be nil to clear)
func (app *App) SetController(ctrl *Controller) {
	app.mu.Lock()
	defer app.mu.Unlock()

	// If setting the same controller, do nothing (idempotent)
	if app.controller == ctrl {
		return
	}

	// If there's an existing controller and action is running, stop it
	// We already have the lock, so directly access and modify the state
	// (no need to call methods that would try to acquire locks)
	if app.controller != nil && app.actionRunning {
		app.actionRunning = false
		app.polling = false
		if app.cancelFunc != nil {
			app.cancelFunc()
			app.cancelFunc = nil
		}
	}

	// Set the new controller
	app.controller = ctrl
}

// StartAction starts an action and enables auto-refresh polling.
// This implements the singleton active model concept - only one action
// can be running at a time across the entire app.
//
// If a previous action is still running, its context is cancelled before
// starting the new one. Returns a context that will be cancelled when
// the action is stopped or replaced.
func (app *App) StartAction() context.Context {
	app.mu.Lock()
	defer app.mu.Unlock()

	// Cancel any previous action
	if app.cancelFunc != nil {
		app.cancelFunc()
	}

	app.actionRunning = true
	app.polling = true
	app.PollCount = 0

	ctx, cancel := context.WithCancel(context.Background())
	app.cancelFunc = cancel
	return ctx
}

// EndAction stops the action and disables auto-refresh polling.
// Also cancels the context returned by StartAction.
func (app *App) EndAction() {
	app.mu.Lock()
	defer app.mu.Unlock()

	app.actionRunning = false
	app.polling = false
	if app.cancelFunc != nil {
		app.cancelFunc()
		app.cancelFunc = nil
	}
}

// IsActionRunning returns whether an action is currently running.
// This checks the app-level state (singleton active model).
func (app *App) IsActionRunning() bool {
	app.mu.RLock()
	defer app.mu.RUnlock()

	return app.actionRunning
}

// SetRefreshTime sets the refresh time in seconds for auto-refresh polling.
func (app *App) SetRefreshTime(seconds int) {
	app.mu.Lock()
	defer app.mu.Unlock()

	app.refreshTime = seconds
}

// SetDisplayURL sets the URL to redirect to for displaying results.
func (app *App) SetDisplayURL(url string) {
	app.mu.Lock()
	defer app.mu.Unlock()

	app.displayURL = url
}

// HandleRoot is a helper for the root endpoint that starts an action.
//
// This function:
//  1. Resets the buffer (if resetBuffer is true)
//  2. Starts the action (app-level state) and gets a cancellable context
//  3. Launches the model function in a goroutine with the context
//  4. Returns HTML that redirects to the display page
//
// The model function receives a context.Context that is cancelled when the action
// is stopped (EndAction) or replaced by a new action (another HandleRoot call).
//
// Example:
//
//	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
//	    app.HandleRoot(w, r, model, true)
//	})
func (app *App) HandleRoot(w http.ResponseWriter, r *http.Request, modelFunc func(context.Context, *App), resetBuffer bool) {
	app.mu.RLock()
	ctrl := app.controller
	displayURL := app.displayURL
	app.mu.RUnlock()

	if ctrl == nil {
		http.Error(w, "No controller set", http.StatusInternalServerError)
		return
	}

	if resetBuffer {
		ctrl.context.Reset()
	}

	ctx := app.StartAction()
	go modelFunc(ctx, app)

	w.Header().Set("Content-Type", "text/html")
	if _, err := fmt.Fprintf(w, `<head><meta http-equiv="Refresh" content="0; URL=%s"/></head>`, displayURL); err != nil {
		http.Error(w, "Failed to write response", http.StatusInternalServerError)
	}
}

// HandleDisplay renders the template with full app state (including polling/refresh).
// Returns an error if no controller is set.
func (app *App) HandleDisplay(w http.ResponseWriter, r *http.Request) {
	app.mu.RLock()
	ctrl := app.controller
	app.mu.RUnlock()

	if ctrl == nil {
		http.Error(w, "No controller set", http.StatusInternalServerError)
		return
	}

	data := app.StateDict(r, nil)
	if err := ctrl.RenderTemplate(w, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// ControllerName returns the name of the current controller.
// Returns "Lofigui no controller" if no controller is set.
func (app *App) ControllerName() string {
	app.mu.RLock()
	defer app.mu.RUnlock()

	if app.controller != nil {
		return app.controller.Name
	}
	return "Lofigui no controller"
}

// StateDict generates a template context dictionary with app and controller state merged.
//
// This method provides centralized state management by combining:
//   - App-level state (version, controller name, polling status)
//   - Controller-level state (buffer content)
//   - Extra context passed by the caller
//
// Returns a pongo2.Context containing:
//   - request: The HTTP request object
//   - version: Application version string
//   - controller_name: Name of the active controller
//   - results: Buffer content from Print/Markdown calls
//   - polling: "Running" or "Stopped" (app-level singleton state)
//   - poll_count: Number of refresh cycles (app-level)
//   - refresh: Meta tag for auto-refresh (if action is running)
//   - Any additional keys from extraContext
//
// Example:
//
//	func (app *App) HandleCustomDisplay(w http.ResponseWriter, r *http.Request) {
//	    extra := pongo2.Context{"title": "My Page"}
//	    data := app.StateDict(r, extra)
//	    // Use data for template rendering
//	}
func (app *App) StateDict(r *http.Request, extraContext pongo2.Context) pongo2.Context {
	app.mu.Lock()
	defer app.mu.Unlock()

	ctrl := app.controller

	// Get buffer content from controller
	var buffer string
	if ctrl != nil {
		buffer = ctrl.context.Buffer()
	}

	// Inline controller name lookup to avoid nested lock
	controllerName := "Lofigui no controller"
	if ctrl != nil {
		controllerName = ctrl.Name
	}

	// Build context with app-level state (singleton active model)
	ctx := pongo2.Context{
		"request":         r,
		"version":         app.Version,
		"controller_name": controllerName,
		"results":         buffer,
	}

	// Add polling state from app (singleton active model concept)
	if app.polling {
		ctx["polling"] = "Running"
		app.PollCount++
		ctx["refresh"] = fmt.Sprintf(
			`<meta http-equiv="Refresh" content="%d; URL=%s"/>`,
			app.refreshTime,
			app.displayURL,
		)
	} else {
		ctx["refresh"] = ""
		app.PollCount = 0
		ctx["polling"] = "Stopped"
	}
	ctx["poll_count"] = app.PollCount

	// Merge extra context last so it can override anything
	if extraContext != nil {
		ctx.Update(extraContext)
	}

	return ctx
}
