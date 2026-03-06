//go:build js && wasm

package main

import (
	"encoding/json"
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

	maint := sim.MaintenanceStatus()

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

	// Maintenance notification
	if maint.Status == "running" {
		lofigui.HTML(fmt.Sprintf(`<div class="notification is-warning">
  <div class="columns is-vcentered">
    <div class="column">
      <p class="has-text-weight-bold">%s Maintenance In Progress</p>
      <p class="is-size-7">%s</p>
      <progress class="progress is-warning" value="%.0f" max="100">%.0f%%</progress>
    </div>
    <div class="column is-narrow">
      <button class="button is-danger is-small" onclick="goCancelMaintenance(); render();">Cancel</button>
    </div>
  </div>
</div>`, capitalize(maint.Type), lastLog(maint.Log), maint.Progress, maint.Progress))
	} else if maint.Status == "completed" {
		lofigui.HTML(fmt.Sprintf(`<div class="notification is-success">
  <p class="has-text-weight-bold">%s Maintenance Complete</p>
  <button class="button is-small mt-2" onclick="goClearMaintenance(); render();">Dismiss</button>
</div>`, capitalize(maint.Type)))
	} else if maint.Status == "cancelled" {
		lofigui.HTML(fmt.Sprintf(`<div class="notification is-info">
  <p class="has-text-weight-bold">%s Maintenance Cancelled</p>
  <button class="button is-small mt-2" onclick="goClearMaintenance(); render();">Dismiss</button>
</div>`, capitalize(maint.Type)))
	}

	// Maintenance action buttons (only when idle)
	if maint.Status == "" {
		lofigui.HTML(`<div class="buttons" id="maint-buttons">
  <button class="button is-warning is-outlined" onclick="goStartMaintenance('pump'); render();">Pump Maintenance</button>
  <button class="button is-warning is-outlined" onclick="goStartMaintenance('valve'); render();">Valve Inspection</button>
</div>`)
	}

	return js.ValueOf(lofigui.Buffer())
}

func goRenderDiagnostics(this js.Value, args []js.Value) any {
	lofigui.Reset()
	diag := sim.Diagnostics()
	maint := sim.MaintenanceStatus()

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

	// Maintenance section
	lofigui.HTML(`<h3 class="title is-5">Maintenance</h3>`)
	if maint.Status == "" {
		lofigui.HTML(`<p class="has-text-grey">No maintenance activity</p>`)
	} else {
		lofigui.HTML(`<table class="table is-bordered is-striped is-narrow">`)
		lofigui.HTML(`<thead><tr><th>Field</th><th>Value</th></tr></thead><tbody>`)
		lofigui.HTML(fmt.Sprintf(`<tr><td>Type</td><td>%s</td></tr>`, capitalize(maint.Type)))
		lofigui.HTML(fmt.Sprintf(`<tr><td>Status</td><td>%s</td></tr>`, maint.Status))
		lofigui.HTML(fmt.Sprintf(`<tr><td>Progress</td><td>%.0f%%</td></tr>`, maint.Progress))
		lofigui.HTML(`</tbody></table>`)
		if len(maint.Log) > 0 {
			lofigui.HTML(`<h4 class="title is-6">Maintenance Log</h4><ul>`)
			for _, entry := range maint.Log {
				lofigui.HTML(fmt.Sprintf(`<li>%s</li>`, entry))
			}
			lofigui.HTML(`</ul>`)
		}
	}

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

func goStartMaintenance(this js.Value, args []js.Value) any {
	if len(args) < 1 {
		return nil
	}
	sim.StartMaintenance(args[0].String())
	return nil
}

func goCancelMaintenance(this js.Value, args []js.Value) any {
	sim.CancelMaintenance()
	return nil
}

func goClearMaintenance(this js.Value, args []js.Value) any {
	sim.ClearMaintenance()
	return nil
}

func goMaintenanceStatus(this js.Value, args []js.Value) any {
	maint := sim.MaintenanceStatus()
	obj := js.Global().Get("Object").New()
	obj.Set("type", maint.Type)
	obj.Set("progress", maint.Progress)
	obj.Set("status", maint.Status)
	logArr := js.Global().Get("Array").New(len(maint.Log))
	for i, entry := range maint.Log {
		logArr.SetIndex(i, entry)
	}
	obj.Set("log", logArr)
	return obj
}

// goExportState returns the simulation state as a JSON string.
func goExportState(this js.Value, args []js.Value) any {
	st := sim.ExportState()
	data, err := json.Marshal(st)
	if err != nil {
		return js.Null()
	}
	return js.ValueOf(string(data))
}

// goImportState restores simulation state from a JSON string.
func goImportState(this js.Value, args []js.Value) any {
	if len(args) < 1 {
		return nil
	}
	var st SimState
	if err := json.Unmarshal([]byte(args[0].String()), &st); err != nil {
		return nil
	}
	sim.ImportState(st)
	return nil
}

// goExportDiagnostics returns a diagnostic record as a JSON string.
func goExportDiagnostics(this js.Value, args []js.Value) any {
	rec := sim.ExportDiagRecord()
	data, err := json.Marshal(rec)
	if err != nil {
		return js.Null()
	}
	return js.ValueOf(string(data))
}

func capitalize(s string) string {
	if s == "" {
		return s
	}
	return fmt.Sprintf("%c%s", s[0]-32, s[1:])
}

func lastLog(log []string) string {
	if len(log) == 0 {
		return ""
	}
	return log[len(log)-1]
}

func main() {
	js.Global().Set("goRenderSchematic", js.FuncOf(goRenderSchematic))
	js.Global().Set("goRenderDiagnostics", js.FuncOf(goRenderDiagnostics))
	js.Global().Set("goStart", js.FuncOf(goStart))
	js.Global().Set("goStop", js.FuncOf(goStop))
	js.Global().Set("goTogglePump", js.FuncOf(goTogglePump))
	js.Global().Set("goToggleValve", js.FuncOf(goToggleValve))
	js.Global().Set("goIsRunning", js.FuncOf(goIsRunning))
	js.Global().Set("goStartMaintenance", js.FuncOf(goStartMaintenance))
	js.Global().Set("goCancelMaintenance", js.FuncOf(goCancelMaintenance))
	js.Global().Set("goClearMaintenance", js.FuncOf(goClearMaintenance))
	js.Global().Set("goMaintenanceStatus", js.FuncOf(goMaintenanceStatus))
	js.Global().Set("goExportState", js.FuncOf(goExportState))
	js.Global().Set("goImportState", js.FuncOf(goImportState))
	js.Global().Set("goExportDiagnostics", js.FuncOf(goExportDiagnostics))

	js.Global().Call("wasmReady")

	<-make(chan struct{})
}
