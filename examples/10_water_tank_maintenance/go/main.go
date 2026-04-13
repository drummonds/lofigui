//go:build !(js && wasm)

package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"sync"
	"time"

	"codeberg.org/hum3/lofigui"
)

const htmxLayout = `<!DOCTYPE html>
<html>
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>{{.controller_name}}</title>
  <link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/bulma@1.0.4/css/bulma.min.css">
  <script src="https://unpkg.com/htmx.org@2.0.4"></script>
</head>
<body>
  <nav class="navbar is-primary" role="navigation" aria-label="main navigation">
    <div class="navbar-brand">
      <span class="navbar-item has-text-weight-bold">{{.controller_name}}</span>
    </div>
    <div class="navbar-end">
      <div class="navbar-item">
        <span class="tag {{if eq .sim_status "Running"}}is-warning{{else}}is-success{{end}}">{{.sim_status}}</span>
      </div>
    </div>
  </nav>
  <section class="section">
    <div class="container">
      <div id="results" hx-get="{{.fragment_url}}" hx-trigger="every 1s" hx-swap="innerHTML">
        {{.results}}
      </div>
    </div>
  </section>
  <footer class="footer">
    <div class="content has-text-centered">
      <p>{{.version}}</p>
    </div>
  </footer>
</body>
</html>`

var renderMu sync.Mutex

func renderAndCapture(fn func()) string {
	renderMu.Lock()
	defer renderMu.Unlock()
	lofigui.Reset()
	fn()
	return lofigui.Buffer()
}

// renderSchematic writes the SVG schematic, controls, and maintenance UI into the buffer.
func renderSchematic(sim *Simulation) {
	lofigui.HTML(sim.buildSVG())

	sim.mu.Lock()
	level := sim.tankLevel
	pump := sim.pumpOn
	valve := sim.valveOpen
	running := sim.running
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
      <form action="/maintenance/cancel" method="post"><button class="button is-danger is-small" type="submit">Cancel</button></form>
    </div>
  </div>
</div>`, capitalize(maint.Type), lastLog(maint.Log), maint.Progress, maint.Progress))
	} else if maint.Status == "completed" {
		lofigui.HTML(fmt.Sprintf(`<div class="notification is-success">
  <p class="has-text-weight-bold">%s Maintenance Complete</p>
  <form action="/maintenance/clear" method="post" class="mt-2"><button class="button is-small" type="submit">Dismiss</button></form>
</div>`, capitalize(maint.Type)))
	} else if maint.Status == "cancelled" {
		lofigui.HTML(fmt.Sprintf(`<div class="notification is-info">
  <p class="has-text-weight-bold">%s Maintenance Cancelled</p>
  <form action="/maintenance/clear" method="post" class="mt-2"><button class="button is-small" type="submit">Dismiss</button></form>
</div>`, capitalize(maint.Type)))
	}

	// Controls
	var startStopBtn string
	if running {
		startStopBtn = `<form action="/stop" method="post" style="display:inline"><button class="button is-danger" type="submit">Stop Simulation</button></form>`
	} else {
		startStopBtn = `<form action="/start" method="post" style="display:inline"><button class="button is-success" type="submit">Start Simulation</button></form>`
	}

	pumpBtnLabel := "Pump On"
	pumpBtnClass := "is-info"
	if pump {
		pumpBtnLabel = "Pump Off"
		pumpBtnClass = "is-info is-light"
	}
	pumpDisabled := ""
	if maint.Type == "pump" && maint.Status == "running" {
		pumpDisabled = " disabled"
	}
	pumpBtn := fmt.Sprintf(`<form action="/pump" method="post" style="display:inline"><button class="button %s" type="submit"%s>%s</button></form>`, pumpBtnClass, pumpDisabled, pumpBtnLabel)

	valveBtnLabel := "Open Valve"
	valveBtnClass := "is-info"
	if valve {
		valveBtnLabel = "Close Valve"
		valveBtnClass = "is-info is-light"
	}
	valveDisabled := ""
	if maint.Type == "valve" && maint.Status == "running" {
		valveDisabled = " disabled"
	}
	valveBtn := fmt.Sprintf(`<form action="/valve" method="post" style="display:inline"><button class="button %s" type="submit"%s>%s</button></form>`, valveBtnClass, valveDisabled, valveBtnLabel)

	lofigui.HTML(fmt.Sprintf(`<div class="buttons">%s %s %s</div>`, startStopBtn, pumpBtn, valveBtn))

	// Maintenance action buttons (only when idle)
	if maint.Status == "" {
		lofigui.HTML(`<div class="buttons">
  <form action="/maintenance/pump" method="post" style="display:inline"><button class="button is-warning is-outlined" type="submit">Pump Maintenance</button></form>
  <form action="/maintenance/valve" method="post" style="display:inline"><button class="button is-warning is-outlined" type="submit">Valve Inspection</button></form>
</div>`)
	}

	// Nav link
	lofigui.HTML(`<a href="/diagnostics" class="button is-small is-link is-outlined">View Diagnostics</a>`)
}

// renderDiagnostics writes diagnostic info into the lofigui buffer.
func renderDiagnostics(sim *Simulation) {
	diag := sim.Diagnostics()
	maint := sim.MaintenanceStatus()

	lofigui.HTML(`<h2 class="title is-4">Diagnostics</h2>`)

	// Summary table
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

	// Level history
	if len(diag.History) > 0 {
		lofigui.HTML(`<h3 class="title is-5">Level History</h3>`)
		lofigui.HTML(`<table class="table is-bordered is-striped is-narrow"><thead><tr><th>Tick</th><th>Level (%)</th></tr></thead><tbody>`)
		for _, entry := range diag.History {
			lofigui.HTML(fmt.Sprintf(`<tr><td>%d</td><td>%.1f</td></tr>`, entry.Tick, entry.Level))
		}
		lofigui.HTML(`</tbody></table>`)
	}

	// Nav link
	lofigui.HTML(`<a href="/" class="button is-small is-link is-outlined">Back to Schematic</a>`)
}

func simStatus(sim *Simulation) string {
	if sim.IsRunning() {
		return "Running"
	}
	return "Stopped"
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
	sim := &Simulation{pumpOn: true}

	ctrl, err := lofigui.NewController(lofigui.ControllerConfig{
		TemplateString: htmxLayout,
		Name:           "Water Tank Maintenance",
	})
	if err != nil {
		log.Fatalf("Failed to create controller: %v", err)
	}

	version := "Water Tank Maintenance v1.0"

	// GET / — full page with schematic
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		if r.Method != "GET" {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		content := renderAndCapture(func() { renderSchematic(sim) })
		ctrl.RenderTemplate(w, lofigui.TemplateContext{
			"controller_name": ctrl.Name,
			"version":         version,
			"results":         template.HTML(content),
			"fragment_url":    "/fragment",
			"sim_status":      simStatus(sim),
		})
	})

	// GET /diagnostics — full page with diagnostics
	http.HandleFunc("/diagnostics", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		content := renderAndCapture(func() { renderDiagnostics(sim) })
		ctrl.RenderTemplate(w, lofigui.TemplateContext{
			"controller_name": ctrl.Name,
			"version":         version,
			"results":         template.HTML(content),
			"fragment_url":    "/fragment/diagnostics",
			"sim_status":      simStatus(sim),
		})
	})

	// GET /fragment — HTML fragment: schematic only
	http.HandleFunc("/fragment", func(w http.ResponseWriter, r *http.Request) {
		content := renderAndCapture(func() { renderSchematic(sim) })
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprint(w, content)
	})

	// GET /fragment/diagnostics — HTML fragment: diagnostics only
	http.HandleFunc("/fragment/diagnostics", func(w http.ResponseWriter, r *http.Request) {
		content := renderAndCapture(func() { renderDiagnostics(sim) })
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprint(w, content)
	})

	// POST /start
	http.HandleFunc("/start", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}
		sim.Start()
		http.Redirect(w, r, "/", http.StatusSeeOther)
	})

	// POST /stop
	http.HandleFunc("/stop", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}
		sim.Stop()
		http.Redirect(w, r, "/", http.StatusSeeOther)
	})

	// GET|POST /pump — toggle pump
	http.HandleFunc("/pump", func(w http.ResponseWriter, r *http.Request) {
		sim.TogglePump()
		http.Redirect(w, r, "/", http.StatusSeeOther)
	})

	// GET|POST /valve — toggle valve
	http.HandleFunc("/valve", func(w http.ResponseWriter, r *http.Request) {
		sim.ToggleValve()
		http.Redirect(w, r, "/", http.StatusSeeOther)
	})

	// POST /maintenance/pump — start pump maintenance
	http.HandleFunc("/maintenance/pump", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}
		sim.StartMaintenance("pump")
		http.Redirect(w, r, "/", http.StatusSeeOther)
	})

	// POST /maintenance/valve — start valve maintenance
	http.HandleFunc("/maintenance/valve", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}
		sim.StartMaintenance("valve")
		http.Redirect(w, r, "/", http.StatusSeeOther)
	})

	// POST /maintenance/cancel — cancel running maintenance
	http.HandleFunc("/maintenance/cancel", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}
		sim.CancelMaintenance()
		http.Redirect(w, r, "/", http.StatusSeeOther)
	})

	// POST /maintenance/clear — dismiss completed/cancelled maintenance
	http.HandleFunc("/maintenance/clear", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}
		sim.ClearMaintenance()
		http.Redirect(w, r, "/", http.StatusSeeOther)
	})

	http.HandleFunc("/favicon.ico", lofigui.ServeFavicon)

	addr := ":1349"
	log.Printf("Starting Water Tank Maintenance on http://localhost%s", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}
