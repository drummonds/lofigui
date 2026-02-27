//go:build js && wasm

package main

import (
	"fmt"
	"syscall/js"
	"time"

	"github.com/drummonds/lofigui"
)

var sim = &Simulation{pumpOn: true}

func goRenderSchematic(this js.Value, args []js.Value) any {
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

func goRenderDiagnostics(this js.Value, args []js.Value) any {
	lofigui.Reset()
	diag := sim.Diagnostics()

	lofigui.HTML(`<h2 class="title is-4">Diagnostics</h2>`)

	lofigui.HTML(`<table class="table is-bordered is-striped is-narrow">`)
	lofigui.HTML(`<thead><tr><th>Metric</th><th>Value</th></tr></thead><tbody>`)
	lofigui.HTML(fmt.Sprintf(`<tr><td>Pump Cycles</td><td>%d</td></tr>`, diag.PumpCycles))
	lofigui.HTML(fmt.Sprintf(`<tr><td>Pump On Time</td><td>%s</td></tr>`, diag.PumpOnTime.Truncate(time.Second)))
	lofigui.HTML(fmt.Sprintf(`<tr><td>Valve Cycles</td><td>%d</td></tr>`, diag.ValveCycles))
	lofigui.HTML(fmt.Sprintf(`<tr><td>Valve On Time</td><td>%s</td></tr>`, diag.ValveOnTime.Truncate(time.Second)))
	lofigui.HTML(fmt.Sprintf(`<tr><td>Float Switch Trips</td><td>%d</td></tr>`, diag.FloatTrips))
	lofigui.HTML(fmt.Sprintf(`<tr><td>Simulation Ticks</td><td>%d</td></tr>`, diag.TickCount))
	lofigui.HTML(`</tbody></table>`)

	if len(diag.History) > 0 {
		lofigui.HTML(`<h3 class="title is-5">Level History</h3>`)
		lofigui.HTML(`<table class="table is-bordered is-striped is-narrow"><thead><tr><th>Tick</th><th>Level (%)</th></tr></thead><tbody>`)
		for _, entry := range diag.History {
			lofigui.HTML(fmt.Sprintf(`<tr><td>%d</td><td>%.1f</td></tr>`, entry.Tick, entry.Level))
		}
		lofigui.HTML(`</tbody></table>`)
	}

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
	js.Global().Set("goRenderSchematic", js.FuncOf(goRenderSchematic))
	js.Global().Set("goRenderDiagnostics", js.FuncOf(goRenderDiagnostics))
	js.Global().Set("goStart", js.FuncOf(goStart))
	js.Global().Set("goStop", js.FuncOf(goStop))
	js.Global().Set("goTogglePump", js.FuncOf(goTogglePump))
	js.Global().Set("goToggleValve", js.FuncOf(goToggleValve))
	js.Global().Set("goIsRunning", js.FuncOf(goIsRunning))

	js.Global().Call("wasmReady")

	<-make(chan struct{})
}
