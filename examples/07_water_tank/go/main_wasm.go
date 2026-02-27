//go:build js && wasm

package main

import (
	"fmt"
	"syscall/js"

	"github.com/drummonds/lofigui"
)

var sim = &Simulation{pumpOn: true}

func goRender(this js.Value, args []js.Value) any {
	lofigui.Reset()
	lofigui.HTML(sim.buildSVG())

	sim.mu.Lock()
	level := sim.tankLevel
	pump := sim.pumpOn
	valve := sim.valveOpen
	sim.mu.Unlock()

	// Status tags
	pumpTag := `<span class="tag is-light">Pump Off</span>`
	if pump {
		pumpTag = `<span class="tag is-success">Pump On</span>`
	}
	valveTag := `<span class="tag is-light">Valve Closed</span>`
	if valve {
		valveTag = `<span class="tag is-success">Valve Open</span>`
	}
	floatTag := `<span class="tag is-light">Float: OK</span>`
	if level >= 95 {
		floatTag = `<span class="tag is-danger">Float: HIGH</span>`
	} else if level <= 5 {
		floatTag = `<span class="tag is-warning">Float: LOW</span>`
	}

	lofigui.HTML(fmt.Sprintf(`<div class="field is-grouped is-grouped-multiline mb-4">
  <div class="control">%s</div>
  <div class="control">%s</div>
  <div class="control">%s</div>
</div>`, pumpTag, valveTag, floatTag))

	return js.ValueOf(lofigui.Buffer())
}

func goStart(this js.Value, args []js.Value) any {
	sim.Start()
	return nil
}

func goStop(this js.Value, args []js.Value) any {
	sim.Stop()
	return nil
}

func goTogglePump(this js.Value, args []js.Value) any {
	sim.TogglePump()
	return nil
}

func goToggleValve(this js.Value, args []js.Value) any {
	sim.ToggleValve()
	return nil
}

func goIsRunning(this js.Value, args []js.Value) any {
	return js.ValueOf(sim.IsRunning())
}

func main() {
	js.Global().Set("goRender", js.FuncOf(goRender))
	js.Global().Set("goStart", js.FuncOf(goStart))
	js.Global().Set("goStop", js.FuncOf(goStop))
	js.Global().Set("goTogglePump", js.FuncOf(goTogglePump))
	js.Global().Set("goToggleValve", js.FuncOf(goToggleValve))
	js.Global().Set("goIsRunning", js.FuncOf(goIsRunning))

	js.Global().Call("wasmReady")

	<-make(chan struct{})
}
