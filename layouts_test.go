package lofigui

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestLayoutSingleRenders(t *testing.T) {
	ctrl, err := NewControllerWithLayout(LayoutSingle, "Test Single")
	if err != nil {
		t.Fatal(err)
	}

	app := NewAppWithController(ctrl)
	Print("Hello Layout")

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	app.HandleDisplay(w, req)

	body := w.Body.String()
	if !strings.Contains(body, "bulma@1.0.4") {
		t.Error("Expected Bulma CDN link")
	}
	if !strings.Contains(body, "Hello Layout") {
		t.Error("Expected content in output")
	}
	Reset()
}

func TestLayoutNavbarRenders(t *testing.T) {
	ctrl, err := NewControllerWithLayout(LayoutNavbar, "Test Navbar")
	if err != nil {
		t.Fatal(err)
	}

	app := NewAppWithController(ctrl)
	app.Version = "v1.0"
	Print("Navbar Content")

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	app.HandleDisplay(w, req)

	body := w.Body.String()
	if !strings.Contains(body, "<nav") {
		t.Error("Expected nav element")
	}
	if !strings.Contains(body, "Test Navbar") {
		t.Error("Expected controller name in navbar")
	}
	if !strings.Contains(body, "Stopped") {
		t.Error("Expected Stopped status tag")
	}
	if !strings.Contains(body, "v1.0") {
		t.Error("Expected version in footer")
	}
	if !strings.Contains(body, "Navbar Content") {
		t.Error("Expected content in output")
	}
	Reset()
}

func TestLayoutThreePanelRenders(t *testing.T) {
	ctrl, err := NewControllerWithLayout(LayoutThreePanel, "Test Panel")
	if err != nil {
		t.Fatal(err)
	}

	app := NewAppWithController(ctrl)
	Print("Main Content")

	extra := map[string]interface{}{"sidebar": "<p>Sidebar Menu</p>"}
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	data := app.StateDict(req, extra)
	if err := ctrl.RenderTemplate(w, data); err != nil {
		t.Fatal(err)
	}

	body := w.Body.String()
	if !strings.Contains(body, "Sidebar Menu") {
		t.Error("Expected sidebar content")
	}
	if !strings.Contains(body, "Main Content") {
		t.Error("Expected main content")
	}
	if !strings.Contains(body, "is-3") {
		t.Error("Expected sidebar column class")
	}
	Reset()
}

func TestNewControllerWithLayout(t *testing.T) {
	ctrl, err := NewControllerWithLayout(LayoutSingle, "Custom Name")
	if err != nil {
		t.Fatal(err)
	}
	if ctrl.Name != "Custom Name" {
		t.Errorf("Expected name 'Custom Name', got %q", ctrl.Name)
	}
}
