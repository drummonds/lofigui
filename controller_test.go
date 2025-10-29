package lofigui

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
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
		if ctrl.Name == "" {
			t.Error("Expected controller to have a default name")
		}
	})

	t.Run("CustomConfig", func(t *testing.T) {
		ctrl, err := NewController(ControllerConfig{
			TemplatePath: templatePath,
			Name:         "Custom Controller",
		})
		if err != nil {
			t.Fatalf("NewController failed: %v", err)
		}
		if ctrl.Name != "Custom Controller" {
			t.Errorf("Expected Name='Custom Controller', got %s", ctrl.Name)
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

	ctrl, err := NewControllerFromDir(tmpDir, "test.html")
	if err != nil {
		t.Fatalf("NewControllerFromDir failed: %v", err)
	}
	if ctrl == nil {
		t.Fatal("Expected non-nil controller")
	}
}

// NOTE: TestActionManagement and TestSetRefreshTime have been removed.
// Action state management has moved to the App level.
// See app_test.go for tests of app.StartAction(), app.EndAction(), etc.

// TestStateDict tests state dictionary generation (controller-level only)
func TestStateDict(t *testing.T) {
	tmpDir := t.TempDir()
	templatePath := filepath.Join(tmpDir, "test.html")
	if err := os.WriteFile(templatePath, []byte("<html></html>"), 0644); err != nil {
		t.Fatal(err)
	}

	ctrl, _ := NewController(ControllerConfig{
		TemplatePath: templatePath,
	})

	req := httptest.NewRequest("GET", "/display", nil)

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

	// Controller-level StateDict no longer includes refresh/polling info
	// Those are now in App.StateDict
	if _, hasRefresh := state["refresh"]; hasRefresh {
		t.Error("Expected no refresh in controller StateDict (moved to App)")
	}
}

// NOTE: TestHandleRoot has been removed.
// HandleRoot has moved to the App level.
// See app_test.go or examples for tests of app.HandleRoot().

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
