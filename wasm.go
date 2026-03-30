//go:build js && wasm

package lofigui

import "syscall/js"

// RunWASM is the simplest way to run a model function in the browser via
// WebAssembly. It exports three functions to JavaScript:
//
//   - goStart():    starts the model (ignored if already running)
//   - goRender():   returns the current buffer HTML
//   - goIsRunning(): returns whether the model is still running
//
// Then it calls wasmReady() and blocks forever.
//
// For apps that need custom JS exports (extra buttons, multiple render
// functions, etc.), write main_wasm.go by hand instead.
func RunWASM(model func(*App)) {
	app := NewApp()

	js.Global().Set("goStart", js.FuncOf(func(this js.Value, args []js.Value) any {
		if app.IsActionRunning() {
			return nil
		}
		app.RunModel(model)
		return nil
	}))

	js.Global().Set("goRender", js.FuncOf(func(this js.Value, args []js.Value) any {
		return js.ValueOf(Buffer())
	}))

	js.Global().Set("goIsRunning", js.FuncOf(func(this js.Value, args []js.Value) any {
		return js.ValueOf(app.IsActionRunning())
	}))

	js.Global().Call("wasmReady")
	<-make(chan struct{})
}
