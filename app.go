package lofigui

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/flosch/pongo2/v6"
)

// ErrCancelled is returned by [App.ListenAndServe] when the server shuts down
// because the user cancelled the running action (via [App.HandleCancel]),
// as opposed to the model completing normally.
var ErrCancelled = errors.New("lofigui: cancelled by user")

// defaultTemplate is the built-in template used when no controller is set.
// defaultTemplate is the built-in template used when no controller is set.
const defaultTemplate = `<!DOCTYPE html>
<html>
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>{{ version }}</title>
  <link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/bulma@1.0.4/css/bulma.min.css">
</head>
<body>
  <nav class="navbar is-primary" role="navigation">
    <div class="navbar-brand">
      <span class="navbar-item has-text-weight-bold">{{ version }}</span>
    </div>
    <div class="navbar-end">
      <div class="navbar-item">
        {% if polling == "Running" %}
        <span class="tag is-warning">Running</span>
        <a href="/cancel" class="tag is-danger is-light ml-1">Cancel</a>
        {% else %}
        <span class="tag is-success">Done</span>
        {% endif %}
      </div>
    </div>
  </nav>
  <section class="section">
    <div class="container content">
      {{ results | safe }}
    </div>
  </section>
</body>
</html>`

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
	hold          bool   // Keep server running after model completes (for screenshots)
	cancelFunc    context.CancelFunc
	actionCtx     context.Context // context for the current action
	server        *http.Server    // set by app.ListenAndServe for graceful shutdown
	done          chan struct{}   // closed by Handle when model completes
	cancelled     bool            // true when shutdown was triggered by cancel, not normal completion
	mu            sync.RWMutex
}

// NewApp creates a new App with no controller.
//
// If the LOFIGUI_HOLD environment variable is set (any non-empty value),
// the server will not shut down after the model completes. This is useful
// for taking screenshots with tools like url2svg.
func NewApp() *App {
	return &App{
		Version:     "Lofigui",
		refreshTime: 1,
		displayURL:  "/display",
		hold:        os.Getenv("LOFIGUI_HOLD") != "",
	}
}

// NewAppWithController creates a new App with the given controller.
func NewAppWithController(ctrl *Controller) *App {
	app := &App{
		Version:     "Lofigui",
		refreshTime: 1,
		displayURL:  "/display",
		hold:        os.Getenv("LOFIGUI_HOLD") != "",
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

// ensureController lazily creates a default controller if none is set.
// Must be called with app.mu held (at least read lock).
// Returns the controller (possibly newly created) or an error.
func (app *App) ensureController() (*Controller, error) {
	if app.controller != nil {
		return app.controller, nil
	}
	ctrl, err := NewController(ControllerConfig{
		TemplateString: defaultTemplate,
		Name:           "Default Controller",
	})
	if err != nil {
		return nil, err
	}
	app.controller = ctrl
	return ctrl, nil
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
	app.actionCtx = ctx
	defaultContext.setContext(ctx)
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
	app.actionCtx = nil
	defaultContext.clearContext()
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
//  3. Launches the model function in a goroutine
//  4. Returns HTML that redirects to the display page
//
// The model function receives the App for calling Sleep, EndAction, etc.
// Cancellation is transparent: if the action is cancelled (by a new HandleRoot
// call or EndAction), output functions and Sleep panic with an internal sentinel
// that is recovered here. The server continues running.
//
// For advanced use, call app.Context() inside the model to get the raw
// context.Context for passing to database calls, HTTP clients, etc.
//
// Example:
//
//	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
//	    app.HandleRoot(w, r, model, true)
//	})
func (app *App) HandleRoot(w http.ResponseWriter, r *http.Request, modelFunc func(*App), resetBuffer bool) {
	app.mu.Lock()
	ctrl, err := app.ensureController()
	displayURL := app.displayURL
	app.mu.Unlock()

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if resetBuffer {
		ctrl.context.Reset()
	}

	app.StartAction()
	go func() {
		defer func() {
			if r := recover(); r != nil {
				if _, ok := r.(cancelledError); ok {
					return // cancelled — goroutine exits cleanly
				}
				panic(r) // re-panic real errors
			}
		}()
		modelFunc(app)
	}()

	http.Redirect(w, r, displayURL, http.StatusSeeOther)
}

// Handle is a single-endpoint handler that starts the model on the first
// request and renders the current state on subsequent polling requests.
//
// On each request:
//   - If no action is running and buffer is empty: starts the model in a
//     background goroutine and renders the page with auto-refresh.
//   - If an action is running: renders the current buffer (polling continues).
//   - If no action is running but buffer has content: renders the final output
//     with no refresh (model completed).
//
// When the model goroutine returns normally, EndAction is called automatically
// to stop polling. The model does not need to call EndAction itself.
//
// The buffer acts as state: empty means ready to start, non-empty after
// completion means done. For restart support, use HandleRoot and HandleDisplay.
//
// Example:
//
//	http.HandleFunc("/", app.Handle(model))
func (app *App) Handle(modelFunc func(*App)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		app.mu.Lock()
		ctrl, err := app.ensureController()
		running := app.actionRunning
		bufferEmpty := ctrl != nil && ctrl.context.Buffer() == ""
		app.mu.Unlock()

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if !running && bufferEmpty {
			app.StartAction()
			go func() {
				defer func() {
					if r := recover(); r != nil {
						if _, ok := r.(cancelledError); ok {
							return
						}
						panic(r)
					}
				}()
				modelFunc(app)
				app.flush()
			}()
		}

		app.WriteRefreshHeader(w)
		data := app.StateDict(r, nil)
		if err := ctrl.RenderTemplate(w, data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

// flush stops polling and, after a grace period for the browser to pick up
// the final render, signals the server to shut down. Called automatically
// by Handle when the model goroutine returns.
//
// For explicit use in a model, call Flush() — same behaviour but public.
func (app *App) flush() {
	app.EndAction()
	app.mu.RLock()
	hold := app.hold
	app.mu.RUnlock()
	if hold {
		return // keep server running for screenshots
	}
	// Grace period: the browser has a pending refresh from the last response's
	// Refresh header. Give it time to arrive and render the final output.
	time.Sleep(2 * time.Second)
	app.signalDone()
}

// Flush stops polling and signals the server to shut down after a grace
// period for the browser to display the final output.
func (app *App) Flush() {
	app.flush()
}

// RunModel resets the buffer, starts the action, and runs the model function
// in a background goroutine. When the model returns, EndAction is called
// automatically. Cancellation panics from Sleep are recovered cleanly.
//
// If a previous action is running, it is cancelled before the new one starts.
// Use IsActionRunning to guard against unwanted restarts.
//
// This is the WASM-friendly equivalent of Handle — same lifecycle, no HTTP.
func (app *App) RunModel(modelFunc func(*App)) {
	Reset()
	app.StartAction()
	go func() {
		defer func() {
			if r := recover(); r != nil {
				if _, ok := r.(cancelledError); ok {
					return
				}
				panic(r)
			}
		}()
		modelFunc(app)
		app.EndAction()
	}()
}

// Run is the simplest way to serve a model over HTTP. It registers the model
// on "/", a cancel handler on "/cancel", and starts the server with graceful
// shutdown. When the model completes, the server exits.
//
// This is the HTTP equivalent of [RunWASM] — one call does everything.
// For custom routes, multiple endpoints, or long-running servers, use
// [App.Handle], [App.HandleCancel], and [App.ListenAndServe] directly.
//
// Example:
//
//	app := lofigui.NewApp()
//	app.Run(":1340", model)
func (app *App) Run(addr string, modelFunc func(*App)) {
	http.HandleFunc("/", app.Handle(modelFunc))
	http.HandleFunc("/cancel", app.HandleCancel("/"))
	if err := app.ListenAndServe(addr, nil); err != nil {
		if errors.Is(err, ErrCancelled) {
			log.Printf("%v", err)
			os.Exit(1)
		}
		log.Fatal(err)
	}
}

// Sleep pauses for the given duration, or until the action is cancelled.
// If cancelled, Sleep panics with the internal sentinel value, which
// HandleRoot's recover wrapper catches to terminate the goroutine cleanly.
func (app *App) Sleep(d time.Duration) {
	ctx := app.Context()
	select {
	case <-ctx.Done():
		panic(errCancelled)
	case <-time.After(d):
	}
}

// Context returns the context for the currently running action.
// Use this to pass context to database calls, HTTP clients, etc.
// Returns context.Background() if no action is running.
func (app *App) Context() context.Context {
	app.mu.RLock()
	defer app.mu.RUnlock()
	if app.actionCtx != nil {
		return app.actionCtx
	}
	return context.Background()
}

// HandleDisplay renders the template with full app state (including polling/refresh).
// Sets the HTTP Refresh header when polling is active, so the browser reloads
// the current page (not a hardcoded URL). Returns an error if no controller is set.
func (app *App) HandleDisplay(w http.ResponseWriter, r *http.Request) {
	app.mu.Lock()
	ctrl, err := app.ensureController()
	app.mu.Unlock()

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	app.WriteRefreshHeader(w)
	data := app.StateDict(r, nil)
	if err := ctrl.RenderTemplate(w, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// HandleCancel returns an http.HandlerFunc that cancels the running action
// and redirects to the given URL. The model goroutine terminates cleanly
// via the panic-recover mechanism, and the buffer retains its partial output.
//
// When using [App.ListenAndServe], HandleCancel also triggers graceful server
// shutdown after a 2-second grace period. Unlike normal model completion,
// [App.ListenAndServe] returns [ErrCancelled] so the caller can distinguish
// cancel from success and exit with a non-zero status.
//
// For long-running servers that should keep running after cancel, write a
// custom handler that calls [App.EndAction] directly instead.
//
// Example:
//
//	http.HandleFunc("/cancel", app.HandleCancel("/"))
func (app *App) HandleCancel(redirectURL string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		app.EndAction()
		app.mu.Lock()
		app.cancelled = true
		app.mu.Unlock()
		http.Redirect(w, r, redirectURL, http.StatusSeeOther)
		go func() {
			app.mu.RLock()
			hold := app.hold
			app.mu.RUnlock()
			if hold {
				return
			}
			time.Sleep(2 * time.Second)
			app.signalDone()
		}()
	}
}

// WriteRefreshHeader sets the HTTP Refresh header when polling is active.
// This causes the browser to reload the current page after refreshTime seconds,
// which works correctly for multi-page apps (unlike a meta refresh with a hardcoded URL).
func (app *App) WriteRefreshHeader(w http.ResponseWriter) {
	app.mu.RLock()
	polling := app.polling
	refreshTime := app.refreshTime
	app.mu.RUnlock()
	if polling {
		w.Header().Set("Refresh", fmt.Sprintf("%d", refreshTime))
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
//   - refresh: Always empty string (refresh now uses HTTP header, kept for template compat)
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
	// Refresh is now handled via HTTP Refresh header (see WriteRefreshHeader),
	// so the template variable is always empty for backward compatibility.
	ctx["refresh"] = ""
	if app.polling {
		ctx["polling"] = "Running"
		app.PollCount++
	} else {
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
