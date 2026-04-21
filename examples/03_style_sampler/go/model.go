package main

import (
	"embed"
	"fmt"
	"sync"

	"codeberg.org/hum3/lofigui"
)

// templateFS ships every page layout with the binary — usable both from the
// HTTP server (main.go) and from the WASM build (main_wasm.go), which has no
// filesystem access.
//
//go:embed templates
var templateFS embed.FS

// templateNames are the layouts served at "/" (home) and "/style/*" (variants).
// The shared table means main.go and main_wasm.go register the same set.
var templateNames = []string{
	"home.html",
	"style_scrolling.html",
	"style_fixed.html",
	"style_three_panel_nav.html",
	"style_three_panel_controls.html",
	"style_fullwidth.html",
}

// loadControllers parses every layout against base.html via lofigui's
// Go-template inheritance loader. Called once at startup from both builds.
func loadControllers() map[string]*lofigui.Controller {
	controllers := make(map[string]*lofigui.Controller, len(templateNames))
	for _, name := range templateNames {
		ctrl, err := lofigui.NewControllerFromFS(templateFS, "templates", name)
		if err != nil {
			panic(fmt.Sprintf("template %s: %v", name, err))
		}
		controllers[name] = ctrl
	}
	return controllers
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
