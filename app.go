package lofigui

import (
	"net/http"
	"sync"
)

// App provides a wrapper around a Controller with safe controller replacement.
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
	mu         sync.RWMutex
}

// NewApp creates a new App with no controller.
func NewApp() *App {
	return &App{}
}

// NewAppWithController creates a new App with the given controller.
func NewAppWithController(ctrl *Controller) *App {
	app := &App{}
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

	// If there's an existing controller, try to clean it up
	if app.controller != nil {
		// Safely check if action is running and try to end it
		// We wrap this in a defer/recover to handle any panics during cleanup
		func() {
			defer func() {
				// Silently ignore any panics during cleanup
				// We're replacing the controller anyway
				_ = recover()
			}()

			// Try to stop running action
			if app.controller.IsActionRunning() {
				app.controller.EndAction()
			}
		}()
	}

	// Set the new controller
	app.controller = ctrl
}

// StartAction starts an action on the current controller.
// Does nothing if no controller is set.
func (app *App) StartAction() {
	app.mu.RLock()
	defer app.mu.RUnlock()

	if app.controller != nil {
		app.controller.StartAction()
	}
}

// EndAction stops the action on the current controller.
// Does nothing if no controller is set.
func (app *App) EndAction() {
	app.mu.RLock()
	defer app.mu.RUnlock()

	if app.controller != nil {
		app.controller.EndAction()
	}
}

// IsActionRunning returns whether an action is currently running.
// Returns false if no controller is set.
func (app *App) IsActionRunning() bool {
	app.mu.RLock()
	defer app.mu.RUnlock()

	if app.controller != nil {
		return app.controller.IsActionRunning()
	}
	return false
}

// HandleRoot is a helper that delegates to the controller's HandleRoot.
// Panics if no controller is set.
func (app *App) HandleRoot(w http.ResponseWriter, r *http.Request, modelFunc func(*Controller), resetBuffer bool) {
	app.mu.RLock()
	ctrl := app.controller
	app.mu.RUnlock()

	if ctrl == nil {
		http.Error(w, "No controller set", http.StatusInternalServerError)
		return
	}

	ctrl.HandleRoot(w, r, modelFunc, resetBuffer)
}

// HandleDisplay is a helper that delegates to the controller's HandleDisplay.
// Returns an error if no controller is set.
func (app *App) HandleDisplay(w http.ResponseWriter, r *http.Request) {
	app.mu.RLock()
	ctrl := app.controller
	app.mu.RUnlock()

	if ctrl == nil {
		http.Error(w, "No controller set", http.StatusInternalServerError)
		return
	}

	ctrl.HandleDisplay(w, r, nil)
}
