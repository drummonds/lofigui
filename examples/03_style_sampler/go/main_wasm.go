//go:build js && wasm

package main

import (
	"embed"
	"html/template"
	"syscall/js"

	"codeberg.org/hum3/lofigui"
)

// templates/ is copied into go/ by build.sh before WASM compilation.
// go:embed cannot traverse parent directories, so this is the workaround.
//
//go:embed templates
var templateFS embed.FS

var controllers map[string]*lofigui.Controller

func init() {
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
		ctrl, err := lofigui.NewControllerFromFS(templateFS, "templates", name)
		if err != nil {
			panic("template " + name + ": " + err.Error())
		}
		controllers[name] = ctrl
	}
}

// goRenderPage renders a full page HTML string for the given style.
// Called from JavaScript to switch between styles without a server round-trip.
func goRenderPage(this js.Value, args []js.Value) any {
	templateName := "home.html"
	currentPath := "/"
	if len(args) > 0 {
		templateName = args[0].String()
	}
	if len(args) > 1 {
		currentPath = args[1].String()
	}

	ctrl, ok := controllers[templateName]
	if !ok {
		return js.ValueOf("<p>Unknown style: " + templateName + "</p>")
	}

	content := sampleOutput()

	html, err := ctrl.RenderToString(lofigui.TemplateContext{
		"results":      template.HTML(content),
		"current_path": currentPath,
	})
	if err != nil {
		return js.ValueOf("<p>Render error: " + err.Error() + "</p>")
	}
	return js.ValueOf(html)
}

func main() {
	js.Global().Set("goRenderPage", js.FuncOf(goRenderPage))
	js.Global().Call("wasmReady")
	<-make(chan struct{})
}
