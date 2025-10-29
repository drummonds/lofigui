package lofigui

import (
	"sync"
	"testing"
	"time"
)

// TestAppControllerCanBeSetAndRetrieved tests that a controller can be set and retrieved
func TestAppControllerCanBeSetAndRetrieved(t *testing.T) {
	app := NewApp()

	// Initially nil
	if app.GetController() != nil {
		t.Error("Expected nil controller initially")
	}

	// Create and set controller
	ctrl, err := NewController(ControllerConfig{
		TemplatePath: "examples/01_hello_world/templates/hello.html",
	})
	if err != nil {
		t.Fatalf("Failed to create controller: %v", err)
	}

	app.SetController(ctrl)

	// Should be the same controller
	if app.GetController() != ctrl {
		t.Error("Expected controller to be set")
	}
}

// TestAppControllerCanBeCleared tests that a controller can be set to nil
func TestAppControllerCanBeCleared(t *testing.T) {
	app := NewApp()

	ctrl, err := NewController(ControllerConfig{
		TemplatePath: "examples/01_hello_world/templates/hello.html",
	})
	if err != nil {
		t.Fatalf("Failed to create controller: %v", err)
	}

	app.SetController(ctrl)
	app.SetController(nil)

	if app.GetController() != nil {
		t.Error("Expected nil controller after clearing")
	}
}

// TestAppControllerReplacementStopsRunningAction tests that replacing a controller stops running actions
func TestAppControllerReplacementStopsRunningAction(t *testing.T) {
	app := NewApp()

	ctrl1, err := NewController(ControllerConfig{
		TemplatePath: "examples/01_hello_world/templates/hello.html",
	})
	if err != nil {
		t.Fatalf("Failed to create controller: %v", err)
	}

	app.SetController(ctrl1)
	app.StartAction()

	if !app.IsActionRunning() {
		t.Error("Expected action to be running")
	}

	// Replace with new controller
	ctrl2, err := NewController(ControllerConfig{
		TemplatePath: "examples/01_hello_world/templates/hello.html",
	})
	if err != nil {
		t.Fatalf("Failed to create controller: %v", err)
	}

	app.SetController(ctrl2)

	// Action should be stopped (app-level state)
	if app.IsActionRunning() {
		t.Error("Expected action to be stopped after controller replacement")
	}

	// New controller should be set
	if app.GetController() != ctrl2 {
		t.Error("Expected new controller to be set")
	}
}

// TestAppMultipleControllerReplacements tests multiple sequential replacements
func TestAppMultipleControllerReplacements(t *testing.T) {
	app := NewApp()

	for i := 0; i < 3; i++ {
		ctrl, err := NewController(ControllerConfig{
			TemplatePath: "examples/01_hello_world/templates/hello.html",
		})
		if err != nil {
			t.Fatalf("Failed to create controller: %v", err)
		}

		app.SetController(ctrl)
		app.StartAction()

		if !app.IsActionRunning() {
			t.Errorf("Iteration %d: Expected action to be running", i)
		}
	}

	// Should still be working fine
	if !app.IsActionRunning() {
		t.Error("Expected action to still be running after multiple replacements")
	}
}

// TestAppControllerInInit tests creating an app with a controller in NewAppWithController
func TestAppControllerInInit(t *testing.T) {
	ctrl, err := NewController(ControllerConfig{
		TemplatePath: "examples/01_hello_world/templates/hello.html",
	})
	if err != nil {
		t.Fatalf("Failed to create controller: %v", err)
	}

	app := NewAppWithController(ctrl)

	if app.GetController() != ctrl {
		t.Error("Expected controller to be set from init")
	}
}

// TestAppControllerNoneToController tests transitioning from nil to a controller
func TestAppControllerNoneToController(t *testing.T) {
	app := NewApp()

	if app.GetController() != nil {
		t.Error("Expected nil controller initially")
	}

	ctrl, err := NewController(ControllerConfig{
		TemplatePath: "examples/01_hello_world/templates/hello.html",
	})
	if err != nil {
		t.Fatalf("Failed to create controller: %v", err)
	}

	app.SetController(ctrl)

	if app.GetController() != ctrl {
		t.Error("Expected controller to be set")
	}
}

// TestAppControllerToNoneStopsAction tests that setting to nil stops running actions
func TestAppControllerToNoneStopsAction(t *testing.T) {
	app := NewApp()

	ctrl, err := NewController(ControllerConfig{
		TemplatePath: "examples/01_hello_world/templates/hello.html",
	})
	if err != nil {
		t.Fatalf("Failed to create controller: %v", err)
	}

	app.SetController(ctrl)
	app.StartAction()

	if !app.IsActionRunning() {
		t.Error("Expected action to be running")
	}

	// Clear controller
	app.SetController(nil)

	// Action should be stopped (app-level state)
	if app.IsActionRunning() {
		t.Error("Expected action to be stopped after clearing controller")
	}
}

// TestAppThreadSafety tests that the App is thread-safe
func TestAppThreadSafety(t *testing.T) {
	app := NewApp()

	var wg sync.WaitGroup
	numGoroutines := 10

	// Multiple goroutines trying to set controllers concurrently
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			ctrl, err := NewController(ControllerConfig{
				TemplatePath: "examples/01_hello_world/templates/hello.html",
			})
			if err != nil {
				t.Errorf("Failed to create controller: %v", err)
				return
			}

			app.SetController(ctrl)
			time.Sleep(10 * time.Millisecond)
			_ = app.GetController()
		}()
	}

	wg.Wait()

	// Should still have a valid controller
	if app.GetController() == nil {
		t.Error("Expected a controller to be set after concurrent operations")
	}
}

// TestAppMethodsWithNoController tests that methods handle nil controller gracefully
func TestAppMethodsWithNoController(t *testing.T) {
	app := NewApp()

	// These should not panic even with no controller
	app.StartAction()

	// Action state is managed at app level, so it should work without controller
	if !app.IsActionRunning() {
		t.Error("Expected IsActionRunning to return true after StartAction")
	}

	app.EndAction()

	if app.IsActionRunning() {
		t.Error("Expected IsActionRunning to return false after EndAction")
	}
}

// TestAppStartActionManagesState tests that StartAction manages app-level state
func TestAppStartActionManagesState(t *testing.T) {
	app := NewApp()

	ctrl, err := NewController(ControllerConfig{
		TemplatePath: "examples/01_hello_world/templates/hello.html",
	})
	if err != nil {
		t.Fatalf("Failed to create controller: %v", err)
	}

	app.SetController(ctrl)

	if app.IsActionRunning() {
		t.Error("Expected action not to be running initially")
	}

	app.StartAction()

	if !app.IsActionRunning() {
		t.Error("Expected action to be running after StartAction")
	}
}

// TestAppEndActionManagesState tests that EndAction manages app-level state
func TestAppEndActionManagesState(t *testing.T) {
	app := NewApp()

	ctrl, err := NewController(ControllerConfig{
		TemplatePath: "examples/01_hello_world/templates/hello.html",
	})
	if err != nil {
		t.Fatalf("Failed to create controller: %v", err)
	}

	app.SetController(ctrl)
	app.StartAction()

	if !app.IsActionRunning() {
		t.Error("Expected action to be running")
	}

	app.EndAction()

	if app.IsActionRunning() {
		t.Error("Expected action to be stopped after EndAction")
	}
}

// TestAppSetControllerIsIdempotent tests that setting the same controller again doesn't stop the action
func TestAppSetControllerIsIdempotent(t *testing.T) {
	app := NewApp()

	ctrl, err := NewController(ControllerConfig{
		TemplatePath: "examples/01_hello_world/templates/hello.html",
	})
	if err != nil {
		t.Fatalf("Failed to create controller: %v", err)
	}

	// Set controller
	app.SetController(ctrl)

	if app.GetController() != ctrl {
		t.Error("Expected controller to be set")
	}

	// Start an action (app-level)
	app.StartAction()

	if !app.IsActionRunning() {
		t.Error("Expected action to be running")
	}

	// Set the same controller again - should NOT stop action (idempotent)
	app.SetController(ctrl)

	// Action should still be running
	if !app.IsActionRunning() {
		t.Error("Expected action to still be running after setting same controller (idempotent)")
	}

	// Controller should still be set
	if app.GetController() != ctrl {
		t.Error("Expected same controller to still be set")
	}
}
