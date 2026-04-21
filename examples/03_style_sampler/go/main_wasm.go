//go:build js && wasm

package main

import (
	"html/template"
	"syscall/js"

	"codeberg.org/hum3/lofigui"
)

var controllers = loadControllers()

// goRenderPage is the JS-facing entry point: it renders a full page HTML
// string for the given layout. templates/app.js calls it whenever the user
// clicks a style button or an intercepted in-page link.
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

	html, err := ctrl.RenderToString(lofigui.TemplateContext{
		"results":      template.HTML(sampleOutput()),
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
