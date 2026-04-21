package main

import (
	"embed"
	"fmt"
	"html/template"
	"net/http"
	"sync"

	"codeberg.org/hum3/lofigui"
)

// templateFS ships every page layout with the binary — usable both from the
// HTTP server (main.go) and from the WASM build (main_wasm.go), which has no
// filesystem access.
//
//go:embed templates
var templateFS embed.FS

// pathToTemplate maps URL paths to the layout template that serves them.
// Used by the shared mux builder below, so server and WASM register the
// same set of routes.
var pathToTemplate = map[string]string{
	"/":                           "home.html",
	"/style/scrolling":            "style_scrolling.html",
	"/style/fixed":                "style_fixed.html",
	"/style/three-panel-nav":      "style_three_panel_nav.html",
	"/style/three-panel-controls": "style_three_panel_controls.html",
	"/style/fullwidth":            "style_fullwidth.html",
}

// loadControllers parses every layout against base.html via lofigui's
// Go-template inheritance loader. Called once at startup from both builds.
func loadControllers() map[string]*lofigui.Controller {
	controllers := make(map[string]*lofigui.Controller, len(pathToTemplate))
	for _, name := range pathToTemplate {
		if _, seen := controllers[name]; seen {
			continue
		}
		ctrl, err := lofigui.NewControllerFromFS(templateFS, "templates", name)
		if err != nil {
			panic(fmt.Sprintf("template %s: %v", name, err))
		}
		controllers[name] = ctrl
	}
	return controllers
}

// buildMux returns the full ServeMux the example runs on. Both main.go
// (net/http) and main_wasm.go (go-wasm-http-server) hand the same mux to
// their respective servers, so the routing and rendering code is identical
// across the two targets.
//
// basePrefix is rendered into <base href="..."> so relative links in the
// templates resolve correctly whether the app is hosted at the site root
// ("/") or under a service-worker scope ("/03_style_sampler/wasm_demo/").
func buildMux(basePrefix string) *http.ServeMux {
	controllers := loadControllers()
	mux := http.NewServeMux()

	for p, name := range pathToTemplate {
		tpl := name
		pattern := "GET " + p
		if p == "/" {
			pattern = "GET /{$}" // exact-match "/", not a prefix
		}
		mux.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
			controllers[tpl].RenderTemplate(w, lofigui.TemplateContext{
				"results":      template.HTML(sampleOutput()),
				"current_path": r.URL.Path,
				"base":         basePrefix,
			})
		})
	}

	mux.HandleFunc("GET /favicon.ico", lofigui.ServeFavicon)
	mux.HandleFunc("GET /assets/bulma.min.css", lofigui.ServeBulma)
	return mux
}

var renderMu sync.Mutex

func renderContent(fn func()) string {
	renderMu.Lock()
	defer renderMu.Unlock()
	lofigui.Reset()
	fn()
	return lofigui.Buffer()
}

// sampleOutput generates teletype content shown in every style demo.
func sampleOutput() string {
	return renderContent(func() {
		lofigui.Markdown("## Teletype Output")
		lofigui.Print("This is sample teletype output.")
		lofigui.Print("Each line appears as the model runs.")
		lofigui.Print("Like a CLI printing to continuous paper.")
		lofigui.Markdown("---")
		lofigui.Table(
			[][]string{
				{"Level 1", "Teletype", "Pure output, no interaction"},
				{"Level 2", "Teletype+ web", "Templates, navbars, forms"},
				{"Level 3", "Polling", "Whole page refresh"},
				{"Level 4", "HTMX", "Partial updates"},
			},
			lofigui.WithHeader([]string{"Level", "Name", "Description"}),
		)
		lofigui.Print("Done.")
	})
}
