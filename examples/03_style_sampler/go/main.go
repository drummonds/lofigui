//go:build !(js && wasm)

package main

import (
	"fmt"
	"html/template"
	"net/http"

	"codeberg.org/hum3/lofigui"
)

// pathToTemplate maps HTTP URL paths to the layout template that serves them.
// The same map is mirrored in templates/app.js so the WASM build routes
// in-page links without reloading.
var pathToTemplate = map[string]string{
	"/":                          "home.html",
	"/style/scrolling":           "style_scrolling.html",
	"/style/fixed":               "style_fixed.html",
	"/style/three-panel-nav":     "style_three_panel_nav.html",
	"/style/three-panel-controls": "style_three_panel_controls.html",
	"/style/fullwidth":           "style_fullwidth.html",
}

func main() {
	controllers := loadControllers()

	for path, name := range pathToTemplate {
		tpl := name
		http.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
			if path == "/" && r.URL.Path != "/" {
				http.NotFound(w, r)
				return
			}
			controllers[tpl].RenderTemplate(w, lofigui.TemplateContext{
				"results":      template.HTML(sampleOutput()),
				"current_path": r.URL.Path,
			})
		})
	}

	http.HandleFunc("/favicon.ico", lofigui.ServeFavicon)
	http.HandleFunc("/assets/bulma.min.css", lofigui.ServeBulma)

	fmt.Println("Style Sampler running at http://localhost:1340")
	http.ListenAndServe(":1340", nil)
}
