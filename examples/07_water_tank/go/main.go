//go:build !(js && wasm)

package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/drummonds/lofigui"
)

// render writes the SVG schematic and controls into the lofigui buffer.
func (s *Simulation) render() {
	lofigui.HTML(s.buildSVG())

	s.mu.Lock()
	level := s.tankLevel
	pump := s.pumpOn
	valve := s.valveOpen
	running := s.running
	s.mu.Unlock()

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
}

func main() {
	sim := &Simulation{pumpOn: true}

	app := lofigui.NewApp()
	app.Version = "Water Tank SCADA v1.0"
	app.SetRefreshTime(1)
	app.SetDisplayURL("/")

	ctrl, err := lofigui.NewControllerWithLayout(lofigui.LayoutNavbar, "Water Tank")
	if err != nil {
		log.Fatalf("Failed to create controller: %v", err)
	}
	app.SetController(ctrl)

	// GET / — render current state
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		lofigui.Reset()
		sim.render()
		app.HandleDisplay(w, r)
	})

	// POST /start
	http.HandleFunc("/start", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}
		sim.Start()
		app.StartAction()
		http.Redirect(w, r, "/", http.StatusSeeOther)
	})

	// POST /stop
	http.HandleFunc("/stop", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}
		sim.Stop()
		app.EndAction()
		http.Redirect(w, r, "/", http.StatusSeeOther)
	})

	// GET|POST /pump — toggle pump (GET from SVG click, POST from button)
	http.HandleFunc("/pump", func(w http.ResponseWriter, r *http.Request) {
		sim.TogglePump()
		http.Redirect(w, r, "/", http.StatusSeeOther)
	})

	// GET|POST /valve — toggle valve (GET from SVG click, POST from button)
	http.HandleFunc("/valve", func(w http.ResponseWriter, r *http.Request) {
		sim.ToggleValve()
		http.Redirect(w, r, "/", http.StatusSeeOther)
	})

	http.HandleFunc("/favicon.ico", lofigui.ServeFavicon)

	addr := ":1347"
	log.Printf("Starting Water Tank SCADA on http://localhost%s", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}
