//go:build js && wasm

package main

import (
	"syscall/js"

	"codeberg.org/hum3/lofigui"
)

var app = lofigui.NewApp()

func goStart(this js.Value, args []js.Value) any {
	if app.IsActionRunning() {
		return nil
	}
	app.RunModel(model)
	return nil
}

func goRender(this js.Value, args []js.Value) any {
	return js.ValueOf(lofigui.Buffer())
}

func goIsRunning(this js.Value, args []js.Value) any {
	return js.ValueOf(app.IsActionRunning())
}

func main() {
	js.Global().Set("goStart", js.FuncOf(goStart))
	js.Global().Set("goRender", js.FuncOf(goRender))
	js.Global().Set("goIsRunning", js.FuncOf(goIsRunning))
	js.Global().Call("wasmReady")
	<-make(chan struct{})
}
