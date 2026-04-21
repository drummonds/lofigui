package lofigui

import (
	"context"
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"runtime/debug"
	"strings"
	"sync"
	"time"
)

// ErrCancelled is returned by [App.ListenAndServe] when the server shuts down
// because the user cancelled the running action (via [App.HandleCancel]),
// as opposed to the model completing normally.
var ErrCancelled = errors.New("lofigui: cancelled by user")

// defaultBuildDate returns a short UTC timestamp for App.BuildDate.
//
// Preference order:
//  1. Go build metadata vcs.time (the HEAD commit's timestamp), with a
//     "+dirty" suffix if vcs.modified=true — present whenever the binary
//     was built inside a git checkout with -buildvcs=true (the default).
//  2. time.Now() at process start — for builds without VCS info (e.g.
//     "go run" without a repo, or -buildvcs=false).
//
// Users who want an exact build time can override by setting App.BuildDate
// after NewApp, e.g. from a package var populated via
// -ldflags="-X 'main.buildDate=<timestamp>'".
func defaultBuildDate() string {
	if info, ok := debug.ReadBuildInfo(); ok {
		var vcsTime, modified string
		for _, s := range info.Settings {
			switch s.Key {
			case "vcs.time":
				vcsTime = s.Value
			case "vcs.modified":
				modified = s.Value
			}
		}
		if vcsTime != "" {
			if t, err := time.Parse(time.RFC3339, vcsTime); err == nil {
				out := t.UTC().Format("2006-01-02 15:04Z")
				if modified == "true" {
					out += " +dirty"
				}
				return out
			}
		}
	}
	return time.Now().UTC().Format("2006-01-02 15:04Z")
}

// defaultTemplate is the built-in template used when no controller is set.
const defaultTemplate = `<!DOCTYPE html>
<html>
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>{{.version}}</title>
  <link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/bulma@1.0.4/css/bulma.min.css">
</head>
<body>
  <nav class="navbar is-primary" role="navigation">
    <div class="navbar-brand">
      <span class="navbar-item has-text-weight-bold">{{.version}}</span>
    </div>
    <div class="navbar-end">
      <div class="navbar-item">
        {{if eq .polling "Running"}}
        <span class="tag is-warning">Running</span>
        <a href="/cancel" class="tag is-danger is-light ml-1">Cancel</a>
        {{else}}
        <span class="tag is-success">Done</span>
        {{end}}
      </div>
    </div>
  </nav>
  <section class="section">
    <div class="container content">
      {{.results}}
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
	controller *Controller
	Version    string // Version/name of the application
	// BuildDate is an optional short string shown in the default template
	// next to Version (e.g. "2026-04-18 09:42Z"). NewApp defaults it via
	// [defaultBuildDate] — the VCS commit timestamp if available, otherwise
	// the current UTC time. Override for an exact build timestamp via
	// -ldflags="-X main.buildDate=..." and then `app.BuildDate = buildDate`.
	BuildDate     string
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
		BuildDate:   defaultBuildDate(),
		refreshTime: 1,
		displayURL:  "/display",
		hold:        os.Getenv("LOFIGUI_HOLD") != "",
	}
}

// NewAppWithController creates a new App with the given controller.
func NewAppWithController(ctrl *Controller) *App {
	app := &App{
		Version:     "Lofigui",
		BuildDate:   defaultBuildDate(),
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
	http.HandleFunc("/assets/bulma.min.css", ServeBulma)
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

// StatusControls returns a navbar HTML fragment showing Running/Stopped status
// with Cancel/Start links. It is intended to be rendered into a template's
// navbar (e.g. as {{.status}}), giving users a consistent way to drive the
// model lifecycle from any page.
//
// basePrefix is the URL prefix the app is mounted under — "/" for a server
// build, or the service-worker scope path for a WASM build (e.g.
// "/03_style_sampler/wasm_demo/"). A missing or trailing-slash-less prefix is
// normalised to end with a single "/".
//
// The rendered links target "<basePrefix>start" and "<basePrefix>cancel";
// wire those routes with [App.RegisterLifecycle].
func (app *App) StatusControls(basePrefix string) template.HTML {
	app.mu.RLock()
	running := app.actionRunning
	app.mu.RUnlock()

	prefix := strings.TrimSuffix(basePrefix, "/") + "/"
	if running {
		return template.HTML(fmt.Sprintf(
			`<span class="tag is-warning">Running</span>`+
				`<a href="%scancel" class="tag is-danger is-light ml-1">Cancel</a>`,
			prefix,
		))
	}
	return template.HTML(fmt.Sprintf(
		`<span class="tag is-success">Stopped</span>`+
			`<a href="%sstart" class="tag is-primary is-light ml-1">Start</a>`,
		prefix,
	))
}

// RegisterLifecycle wires "GET /start" and "GET /cancel" on mux so the UI
// rendered by [App.StatusControls] can drive the model.
//
//   - /start  — if no action is running, resets the buffer and launches
//     modelFunc in a goroutine (via [App.RunModel]), then redirects to the
//     Referer page (or basePrefix if the Referer is missing).
//   - /cancel — calls [App.EndAction] to stop the running model and redirects
//     the same way. Unlike [App.HandleCancel], this does NOT shut the server
//     down, so it is suitable for long-running multi-page apps.
//
// basePrefix must match the prefix passed to [App.StatusControls] (e.g. "/"
// for a server build, or the service-worker scope for a WASM build). It is
// used as the fallback redirect target when no Referer is available, so
// lifecycle links stay inside the SW scope.
func (app *App) RegisterLifecycle(mux *http.ServeMux, modelFunc func(*App), basePrefix string) {
	fallback := strings.TrimSuffix(basePrefix, "/") + "/"
	mux.HandleFunc("GET /start", func(w http.ResponseWriter, r *http.Request) {
		if !app.IsActionRunning() {
			app.RunModel(modelFunc)
		}
		lifecycleRedirect(w, r, fallback)
	})
	mux.HandleFunc("GET /cancel", func(w http.ResponseWriter, r *http.Request) {
		app.EndAction()
		lifecycleRedirect(w, r, fallback)
	})
}

// lifecycleRedirect redirects to the request's Referer path, falling back to
// the given URL when no Referer is present. Only the Referer's path+query is
// used, so a cross-origin Referer can at worst steer navigation within the
// app's own origin.
func lifecycleRedirect(w http.ResponseWriter, r *http.Request, fallback string) {
	dest := fallback
	if ref := r.Header.Get("Referer"); ref != "" {
		if u, err := r.URL.Parse(ref); err == nil && u.Path != "" {
			dest = u.RequestURI()
		}
	}
	http.Redirect(w, r, dest, http.StatusSeeOther)
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
// Returns a TemplateContext containing:
//   - request: The HTTP request object
//   - version: Application version string
//   - controller_name: Name of the active controller
//   - results: Buffer content from Print/Markdown calls (as template.HTML)
//   - polling: "Running" or "Stopped" (app-level singleton state)
//   - poll_count: Number of refresh cycles (app-level)
//   - refresh: Always empty template.HTML (refresh now uses HTTP header, kept for template compat)
//   - Any additional keys from extraContext
//
// Example:
//
//	func (app *App) HandleCustomDisplay(w http.ResponseWriter, r *http.Request) {
//	    extra := lofigui.TemplateContext{"title": "My Page"}
//	    data := app.StateDict(r, extra)
//	    // Use data for template rendering
//	}
func (app *App) StateDict(r *http.Request, extraContext TemplateContext) TemplateContext {
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
	ctx := TemplateContext{
		"request":         r,
		"version":         app.Version,
		"build_date":      app.BuildDate,
		"controller_name": controllerName,
		"results":         template.HTML(buffer),
	}

	// Add polling state from app (singleton active model concept)
	// Refresh is now handled via HTTP Refresh header (see WriteRefreshHeader),
	// so the template variable is always empty for backward compatibility.
	ctx["refresh"] = template.HTML("")
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
