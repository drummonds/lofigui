package lofigui

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestNewController tests controller creation with various configurations
func TestNewController(t *testing.T) {
	// Create a temporary template file
	tmpDir := t.TempDir()
	templatePath := filepath.Join(tmpDir, "test.html")
	templateContent := `<!DOCTYPE html>
<html>
<head>{{refresh|safe}}</head>
<body>{{results|safe}}</body>
</html>`

	if err := os.WriteFile(templatePath, []byte(templateContent), 0644); err != nil {
		t.Fatalf("Failed to create test template: %v", err)
	}

	t.Run("BasicConfig", func(t *testing.T) {
		ctrl, err := NewController(ControllerConfig{
			TemplatePath: templatePath,
		})
		if err != nil {
			t.Fatalf("NewController failed: %v", err)
		}
		if ctrl == nil {
			t.Fatal("Expected non-nil controller")
		}
		if ctrl.refreshTime != 1 {
			t.Errorf("Expected default refreshTime=1, got %d", ctrl.refreshTime)
		}
		if ctrl.displayURL != "/display" {
			t.Errorf("Expected default displayURL=/display, got %s", ctrl.displayURL)
		}
	})

	t.Run("CustomConfig", func(t *testing.T) {
		ctrl, err := NewController(ControllerConfig{
			TemplatePath: templatePath,
			RefreshTime:  2,
			DisplayURL:   "/status",
		})
		if err != nil {
			t.Fatalf("NewController failed: %v", err)
		}
		if ctrl.refreshTime != 2 {
			t.Errorf("Expected refreshTime=2, got %d", ctrl.refreshTime)
		}
		if ctrl.displayURL != "/status" {
			t.Errorf("Expected displayURL=/status, got %s", ctrl.displayURL)
		}
	})

	t.Run("MissingTemplate", func(t *testing.T) {
		_, err := NewController(ControllerConfig{
			TemplatePath: "/nonexistent/template.html",
		})
		if err == nil {
			t.Fatal("Expected error for nonexistent template")
		}
	})

	t.Run("EmptyTemplatePath", func(t *testing.T) {
		_, err := NewController(ControllerConfig{})
		if err == nil {
			t.Fatal("Expected error for empty TemplatePath")
		}
	})
}

// TestNewControllerFromDir tests the convenience constructor
func TestNewControllerFromDir(t *testing.T) {
	tmpDir := t.TempDir()
	templatePath := filepath.Join(tmpDir, "test.html")
	templateContent := `<html><body>{{results|safe}}</body></html>`

	if err := os.WriteFile(templatePath, []byte(templateContent), 0644); err != nil {
		t.Fatalf("Failed to create test template: %v", err)
	}

	ctrl, err := NewControllerFromDir(tmpDir, "test.html", 1)
	if err != nil {
		t.Fatalf("NewControllerFromDir failed: %v", err)
	}
	if ctrl == nil {
		t.Fatal("Expected non-nil controller")
	}
}

// TestActionManagement tests action state management
func TestActionManagement(t *testing.T) {
	tmpDir := t.TempDir()
	templatePath := filepath.Join(tmpDir, "test.html")
	if err := os.WriteFile(templatePath, []byte("<html></html>"), 0644); err != nil {
		t.Fatal(err)
	}

	ctrl, err := NewController(ControllerConfig{
		TemplatePath: templatePath,
	})
	if err != nil {
		t.Fatal(err)
	}

	// Initially not running
	if ctrl.IsActionRunning() {
		t.Error("Expected action not running initially")
	}

	// Start action
	ctrl.StartAction()
	if !ctrl.IsActionRunning() {
		t.Error("Expected action to be running after StartAction")
	}

	// End action
	ctrl.EndAction()
	if ctrl.IsActionRunning() {
		t.Error("Expected action not running after EndAction")
	}
}

// TestSetRefreshTime tests refresh time updates
func TestSetRefreshTime(t *testing.T) {
	tmpDir := t.TempDir()
	templatePath := filepath.Join(tmpDir, "test.html")
	if err := os.WriteFile(templatePath, []byte("<html></html>"), 0644); err != nil {
		t.Fatal(err)
	}

	ctrl, _ := NewController(ControllerConfig{
		TemplatePath: templatePath,
		RefreshTime:  1,
	})

	ctrl.SetRefreshTime(5)
	if ctrl.refreshTime != 5 {
		t.Errorf("Expected refreshTime=5, got %d", ctrl.refreshTime)
	}
}

// TestStateDict tests state dictionary generation
func TestStateDict(t *testing.T) {
	tmpDir := t.TempDir()
	templatePath := filepath.Join(tmpDir, "test.html")
	if err := os.WriteFile(templatePath, []byte("<html></html>"), 0644); err != nil {
		t.Fatal(err)
	}

	ctrl, _ := NewController(ControllerConfig{
		TemplatePath: templatePath,
		RefreshTime:  1,
		DisplayURL:   "/display",
	})

	req := httptest.NewRequest("GET", "/display", nil)

	t.Run("ActionNotRunning", func(t *testing.T) {
		Reset()
		Print("Test content")

		state := ctrl.StateDict(req)

		if state["request"] != req {
			t.Error("Expected request in state dict")
		}

		results := state["results"].(string)
		if !strings.Contains(results, "Test content") {
			t.Error("Expected results to contain test content")
		}

		refresh := state["refresh"].(string)
		if refresh != "" {
			t.Errorf("Expected empty refresh when not running, got: %s", refresh)
		}
	})

	t.Run("ActionRunning", func(t *testing.T) {
		ctrl.StartAction()

		state := ctrl.StateDict(req)

		refresh := state["refresh"].(string)
		if !strings.Contains(refresh, "meta http-equiv") {
			t.Error("Expected refresh meta tag when action is running")
		}
		if !strings.Contains(refresh, "/display") {
			t.Error("Expected refresh to contain display URL")
		}

		ctrl.EndAction()
	})
}

// TestHandleRoot tests the root handler helper
func TestHandleRoot(t *testing.T) {
	tmpDir := t.TempDir()
	templatePath := filepath.Join(tmpDir, "test.html")
	if err := os.WriteFile(templatePath, []byte("<html></html>"), 0644); err != nil {
		t.Fatal(err)
	}

	ctrl, _ := NewController(ControllerConfig{
		TemplatePath: templatePath,
		DisplayURL:   "/display",
	})

	modelCalled := false
	model := func(c *Controller) {
		modelCalled = true
		c.EndAction()
	}

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	ctrl.HandleRoot(w, req, model, true)

	// Check response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	body := w.Body.String()
	if !strings.Contains(body, "meta http-equiv") {
		t.Error("Expected redirect meta tag in response")
	}
	if !strings.Contains(body, "/display") {
		t.Error("Expected redirect to /display")
	}

	// Check action started
	if !ctrl.IsActionRunning() {
		t.Error("Expected action to be running")
	}

	// Wait for model to execute
	time.Sleep(50 * time.Millisecond)
	if !modelCalled {
		t.Error("Expected model function to be called")
	}
}

// TestHandleDisplay tests the display handler helper
func TestHandleDisplay(t *testing.T) {
	tmpDir := t.TempDir()
	templatePath := filepath.Join(tmpDir, "test.html")
	templateContent := `<html><body>{{results|safe}}</body></html>`
	if err := os.WriteFile(templatePath, []byte(templateContent), 0644); err != nil {
		t.Fatal(err)
	}

	ctrl, _ := NewController(ControllerConfig{
		TemplatePath: templatePath,
	})

	Reset()
	Print("Test output")

	req := httptest.NewRequest("GET", "/display", nil)
	w := httptest.NewRecorder()

	ctrl.HandleDisplay(w, req, nil)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	body := w.Body.String()
	if !strings.Contains(body, "Test output") {
		t.Error("Expected body to contain test output")
	}
}

// TestHandleDisplayWithExtraContext tests display with additional context
func TestHandleDisplayWithExtraContext(t *testing.T) {
	tmpDir := t.TempDir()
	templatePath := filepath.Join(tmpDir, "test.html")
	templateContent := `<html><body>{{title}} - {{results|safe}}</body></html>`
	if err := os.WriteFile(templatePath, []byte(templateContent), 0644); err != nil {
		t.Fatal(err)
	}

	ctrl, _ := NewController(ControllerConfig{
		TemplatePath: templatePath,
	})

	Reset()
	Print("Content")

	req := httptest.NewRequest("GET", "/display", nil)
	w := httptest.NewRecorder()

	extra := map[string]interface{}{
		"title": "Custom Title",
	}

	ctrl.HandleDisplay(w, req, extra)

	body := w.Body.String()
	if !strings.Contains(body, "Custom Title") {
		t.Error("Expected body to contain custom title from extra context")
	}
}

// TestServeHTTP tests the http.Handler interface implementation
func TestServeHTTP(t *testing.T) {
	tmpDir := t.TempDir()
	templatePath := filepath.Join(tmpDir, "test.html")
	templateContent := `<html><body>{{results|safe}}</body></html>`
	if err := os.WriteFile(templatePath, []byte(templateContent), 0644); err != nil {
		t.Fatal(err)
	}

	ctrl, _ := NewController(ControllerConfig{
		TemplatePath: templatePath,
	})

	Reset()
	Print("ServeHTTP test")

	req := httptest.NewRequest("GET", "/display", nil)
	w := httptest.NewRecorder()

	ctrl.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	body := w.Body.String()
	if !strings.Contains(body, "ServeHTTP test") {
		t.Error("Expected body to contain test content")
	}
}

// TestCustomContext tests using a custom context instead of global
func TestCustomContext(t *testing.T) {
	tmpDir := t.TempDir()
	templatePath := filepath.Join(tmpDir, "test.html")
	templateContent := `<html><body>{{results|safe}}</body></html>`
	if err := os.WriteFile(templatePath, []byte(templateContent), 0644); err != nil {
		t.Fatal(err)
	}

	customCtx := NewContext()
	ctrl, _ := NewController(ControllerConfig{
		TemplatePath: templatePath,
		Context:      customCtx,
	})

	// Add to custom context
	customCtx.Print("Custom context content")

	// Add to global context
	Reset()
	Print("Global context content")

	req := httptest.NewRequest("GET", "/display", nil)
	w := httptest.NewRecorder()

	ctrl.HandleDisplay(w, req, nil)

	body := w.Body.String()

	// Should contain custom context content
	if !strings.Contains(body, "Custom context content") {
		t.Error("Expected body to contain custom context content")
	}

	// Should NOT contain global context content
	if strings.Contains(body, "Global context content") {
		t.Error("Expected body to not contain global context content")
	}
}
