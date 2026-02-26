package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/drummonds/lofigui"
)

// Simulation holds the water tank state.
type Simulation struct {
	mu        sync.Mutex
	running   bool
	cancel    context.CancelFunc
	tankLevel float64 // 0.0–100.0
	pumpOn    bool
	valveOpen bool
}

// Start begins the simulation tick loop.
func (s *Simulation) Start(app *lofigui.App) {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return
	}
	ctx, cancel := context.WithCancel(context.Background())
	s.cancel = cancel
	s.running = true
	s.mu.Unlock()

	app.StartAction()

	go func() {
		ticker := time.NewTicker(500 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				s.tick()
			}
		}
	}()
}

// Stop halts the simulation.
func (s *Simulation) Stop(app *lofigui.App) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.running {
		return
	}
	s.running = false
	if s.cancel != nil {
		s.cancel()
		s.cancel = nil
	}
	app.EndAction()
}

// tick updates tank level once.
func (s *Simulation) tick() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.pumpOn {
		s.tankLevel += 3.0
	}
	if s.valveOpen {
		s.tankLevel -= 1.0
	}

	// Float switch: auto-off at 95%, auto-on at 5%
	if s.tankLevel >= 95.0 {
		s.pumpOn = false
	}
	if s.tankLevel <= 5.0 {
		s.pumpOn = true
	}

	// Clamp
	if s.tankLevel < 0 {
		s.tankLevel = 0
	}
	if s.tankLevel > 100 {
		s.tankLevel = 100
	}
}

// TogglePump toggles the pump state.
func (s *Simulation) TogglePump() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.pumpOn = !s.pumpOn
}

// ToggleValve toggles the valve state.
func (s *Simulation) ToggleValve() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.valveOpen = !s.valveOpen
}

// render writes the current state into the lofigui buffer.
func (s *Simulation) render() {
	s.mu.Lock()
	level := s.tankLevel
	pump := s.pumpOn
	valve := s.valveOpen
	running := s.running
	s.mu.Unlock()

	// Tank level bar
	progressClass := "is-info"
	if level > 80 {
		progressClass = "is-danger"
	} else if level > 60 {
		progressClass = "is-warning"
	}

	lofigui.HTML(fmt.Sprintf(`<div class="box">
  <h3 class="title is-4">Tank Level: %.1f%%</h3>
  <progress class="progress is-large %s" value="%.0f" max="100">%.0f%%</progress>
</div>`, level, progressClass, level, level))

	// Status tags
	pumpTag := `<span class="tag is-light">Pump Off</span>`
	if pump {
		pumpTag = `<span class="tag is-success">Pump On</span>`
	}
	valveTag := `<span class="tag is-light">Valve Closed</span>`
	if valve {
		valveTag = `<span class="tag is-success">Valve Open</span>`
	}

	floatStatus := ""
	if level >= 95 {
		floatStatus = `<span class="tag is-danger">Float: HIGH</span>`
	} else if level <= 5 {
		floatStatus = `<span class="tag is-warning">Float: LOW</span>`
	} else {
		floatStatus = `<span class="tag is-light">Float: OK</span>`
	}

	lofigui.HTML(fmt.Sprintf(`<div class="field is-grouped is-grouped-multiline mb-4">
  <div class="control">%s</div>
  <div class="control">%s</div>
  <div class="control">%s</div>
</div>`, pumpTag, valveTag, floatStatus))

	// Controls
	var startStopBtn string
	if running {
		startStopBtn = `<form action="/stop" method="post" style="display:inline"><button class="button is-danger" type="submit">Stop Simulation</button></form>`
	} else {
		startStopBtn = `<form action="/start" method="post" style="display:inline"><button class="button is-success" type="submit">Start Simulation</button></form>`
	}

	pumpLabel := "Pump On"
	pumpClass := "is-info"
	if pump {
		pumpLabel = "Pump Off"
		pumpClass = "is-info is-light"
	}
	pumpBtn := fmt.Sprintf(`<form action="/pump" method="post" style="display:inline"><button class="button %s" type="submit">%s</button></form>`, pumpClass, pumpLabel)

	valveLabel := "Open Valve"
	valveClass := "is-info"
	if valve {
		valveLabel = "Close Valve"
		valveClass = "is-info is-light"
	}
	valveBtn := fmt.Sprintf(`<form action="/valve" method="post" style="display:inline"><button class="button %s" type="submit">%s</button></form>`, valveClass, valveLabel)

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
		sim.Start(app)
		http.Redirect(w, r, "/", http.StatusSeeOther)
	})

	// POST /stop
	http.HandleFunc("/stop", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}
		sim.Stop(app)
		http.Redirect(w, r, "/", http.StatusSeeOther)
	})

	// POST /pump
	http.HandleFunc("/pump", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}
		sim.TogglePump()
		http.Redirect(w, r, "/", http.StatusSeeOther)
	})

	// POST /valve
	http.HandleFunc("/valve", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}
		sim.ToggleValve()
		http.Redirect(w, r, "/", http.StatusSeeOther)
	})

	http.HandleFunc("/favicon.ico", lofigui.ServeFavicon)

	addr := ":1347"
	log.Printf("Starting Water Tank SCADA on http://localhost%s", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}
