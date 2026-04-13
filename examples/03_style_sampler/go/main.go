//go:build !(js && wasm)

package main

import (
	"fmt"
	"html/template"
	"net/http"

	"codeberg.org/hum3/lofigui"
)

var controllers map[string]*lofigui.Controller

func handleStyle(w http.ResponseWriter, r *http.Request, templateName string) {
	ctrl := controllers[templateName]
	content := sampleOutput()
	ctrl.RenderTemplate(w, lofigui.TemplateContext{
		"results":      template.HTML(content),
		"current_path": r.URL.Path,
	})
}

func main() {
	tplDir := "../templates"
	controllers = make(map[string]*lofigui.Controller)

	templates := []string{
		"home.html",
		"style_scrolling.html",
		"style_fixed.html",
		"style_three_panel_nav.html",
		"style_three_panel_controls.html",
		"style_fullwidth.html",
	}

	for _, name := range templates {
		ctrl, err := lofigui.NewControllerFromDir(tplDir, name)
		if err != nil {
			panic(fmt.Sprintf("template %s: %v", name, err))
		}
		controllers[name] = ctrl
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		handleStyle(w, r, "home.html")
	})
	http.HandleFunc("/style/scrolling", func(w http.ResponseWriter, r *http.Request) {
		handleStyle(w, r, "style_scrolling.html")
	})
	http.HandleFunc("/style/fixed", func(w http.ResponseWriter, r *http.Request) {
		handleStyle(w, r, "style_fixed.html")
	})
	http.HandleFunc("/style/three-panel-nav", func(w http.ResponseWriter, r *http.Request) {
		handleStyle(w, r, "style_three_panel_nav.html")
	})
	http.HandleFunc("/style/three-panel-controls", func(w http.ResponseWriter, r *http.Request) {
		handleStyle(w, r, "style_three_panel_controls.html")
	})
	http.HandleFunc("/style/fullwidth", func(w http.ResponseWriter, r *http.Request) {
		handleStyle(w, r, "style_fullwidth.html")
	})

	http.HandleFunc("/favicon.ico", lofigui.ServeFavicon)

	fmt.Println("Style Sampler running at http://localhost:1340")
	http.ListenAndServe(":1340", nil)
}
