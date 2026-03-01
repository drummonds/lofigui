//go:build !(js && wasm)

package main

import (
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/drummonds/lofigui"
	"github.com/flosch/pongo2/v6"
)

const htmxLayout = `<!DOCTYPE html>
<html>
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>{{ controller_name }}</title>
  <link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/bulma@1.0.4/css/bulma.min.css">
  <script src="https://unpkg.com/htmx.org@2.0.4"></script>
</head>
<body>
  <nav class="navbar is-primary" role="navigation" aria-label="main navigation">
    <div class="navbar-brand">
      <span class="navbar-item has-text-weight-bold">{{ controller_name }}</span>
    </div>
    <div class="navbar-end">
      <div class="navbar-item">
        <span class="tag {% if sim_status == "Running" %}is-warning{% else %}is-success{% endif %}">{{ sim_status }}</span>
      </div>
    </div>
  </nav>
  <section class="section">
    <div class="container">
      <div id="results" hx-get="{{ fragment_url }}" hx-trigger="every 1s" hx-swap="innerHTML">
        {{ results | safe }}
      </div>
    </div>
  </section>
  <footer class="footer">
    <div class="content has-text-centered">
      <p>{{ version }}</p>
    </div>
  </footer>
</body>
</html>`

var renderMu sync.Mutex

// renderAndCapture runs fn (which calls lofigui output functions) under a lock,
// captures the buffer content, and returns it. This ensures concurrent HTTP
// requests don't interleave buffer writes.
func renderAndCapture(fn func()) string {
	renderMu.Lock()
	defer renderMu.Unlock()
	lofigui.Reset()
	fn()
	return lofigui.Buffer()
}

// renderSchematic writes the SVG schematic and controls into the lofigui buffer.
func renderSchematic(sim *Simulation) {
	lofigui.HTML(sim.buildSVG())

	sim.mu.Lock()
	level := sim.tankLevel
	pump := sim.pumpOn
	valve := sim.valveOpen
	running := sim.running
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
	pumpBtn := fmt.Sprintf(`<form action="/pump" method="post" style="display:inline"><button class="button %s" type="submit">%s</button></form>`, pumpBtnClass, pumpBtnLabel)

	valveBtnLabel := "Open Valve"
	valveBtnClass := "is-info"
	if valve {
		valveBtnLabel = "Close Valve"
		valveBtnClass = "is-info is-light"
	}
	valveBtn := fmt.Sprintf(`<form action="/valve" method="post" style="display:inline"><button class="button %s" type="submit">%s</button></form>`, valveBtnClass, valveBtnLabel)

	lofigui.HTML(fmt.Sprintf(`<div class="buttons">%s %s %s</div>`, startStopBtn, pumpBtn, valveBtn))

	// Nav link
	lofigui.HTML(`<a href="/diagnostics" class="button is-small is-link is-outlined">View Diagnostics</a>`)
}

// renderDiagnostics writes diagnostic info into the lofigui buffer.
func renderDiagnostics(sim *Simulation) {
	diag := sim.Diagnostics()

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

func main() {
	sim := &Simulation{pumpOn: true}

	ctrl, err := lofigui.NewController(lofigui.ControllerConfig{
		TemplateString: htmxLayout,
		Name:           "Water Tank HTMX",
	})
	if err != nil {
		log.Fatalf("Failed to create controller: %v", err)
	}

	version := "Water Tank HTMX v1.0"

	// GET / — full page with schematic, HTMX polls /fragment for updates
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
		ctrl.RenderTemplate(w, pongo2.Context{
			"controller_name": ctrl.Name,
			"version":         version,
			"results":         content,
			"fragment_url":    "/fragment",
			"sim_status":      simStatus(sim),
		})
	})

	// GET /diagnostics — full page with diagnostics, HTMX polls /fragment/diagnostics
	http.HandleFunc("/diagnostics", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		content := renderAndCapture(func() { renderDiagnostics(sim) })
		ctrl.RenderTemplate(w, pongo2.Context{
			"controller_name": ctrl.Name,
			"version":         version,
			"results":         content,
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

	http.HandleFunc("/favicon.ico", lofigui.ServeFavicon)

	addr := ":1349"
	log.Printf("Starting Water Tank HTMX on http://localhost%s", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}
