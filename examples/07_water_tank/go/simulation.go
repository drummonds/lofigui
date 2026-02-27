package main

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"
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
func (s *Simulation) Start() {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return
	}
	ctx, cancel := context.WithCancel(context.Background())
	s.cancel = cancel
	s.running = true
	s.mu.Unlock()

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
func (s *Simulation) Stop() {
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
}

// IsRunning returns whether the simulation is running.
func (s *Simulation) IsRunning() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.running
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

// buildSVG generates a P&ID-style schematic of the water tank system.
// Symbols inspired by FUXA-SVG-Widgets (MIT): https://github.com/frangoteam/FUXA-SVG-Widgets
func (s *Simulation) buildSVG() string {
	s.mu.Lock()
	level := s.tankLevel
	pump := s.pumpOn
	valve := s.valveOpen
	running := s.running
	s.mu.Unlock()

	var b strings.Builder

	// Colours
	pumpFill := "#dbdbdb"
	if pump {
		pumpFill = "#48c78e"
	}
	valveFill := "#dbdbdb"
	if valve {
		valveFill = "#48c78e"
	}
	waterColor := "#3e8ed0"
	if level > 80 {
		waterColor = "#f14668"
	} else if level > 60 {
		waterColor = "#ffe08a"
	}
	pipeInColor := "#dbdbdb"
	if pump && running {
		pipeInColor = "#3e8ed0"
	}
	pipeOutColor := "#dbdbdb"
	if valve && level > 0 {
		pipeOutColor = "#3e8ed0"
	}

	// Tank geometry
	const (
		tankX = 270.0
		tankY = 40.0
		tankW = 200.0
		tankH = 300.0
	)
	waterH := tankH * level / 100
	waterY := tankY + tankH - waterH
	highY := tankY + tankH*0.05 // 95% level position
	lowY := tankY + tankH*0.95  // 5% level position

	// SVG header
	b.WriteString(`<svg viewBox="0 0 740 380" xmlns="http://www.w3.org/2000/svg" style="max-width:740px;width:100%;height:auto">`)
	b.WriteString(`<style>text{font-family:Arial,Helvetica,sans-serif}</style>`)

	// --- Pipes (drawn first, behind tank) ---

	// Supply pipe: left edge to pump inlet
	fmt.Fprintf(&b, `<rect x="0" y="193" width="45" height="14" rx="1" fill="%s" stroke="#363636" stroke-width="1"/>`, pipeInColor)

	// Discharge pipe: pump to tank left wall (extends slightly under tank border)
	fmt.Fprintf(&b, `<rect x="120" y="193" width="155" height="14" rx="1" fill="%s" stroke="#363636" stroke-width="1"/>`, pipeInColor)
	// Flow arrow on inlet pipe
	if pump && running {
		b.WriteString(`<polygon points="200,196 210,200 200,204" fill="#fff" opacity="0.6"/>`)
	}

	// Outlet pipe: tank right wall to valve (extends slightly under tank border)
	fmt.Fprintf(&b, `<rect x="465" y="193" width="105" height="14" rx="1" fill="%s" stroke="#363636" stroke-width="1"/>`, pipeOutColor)

	// Drain pipe: valve to outlet
	fmt.Fprintf(&b, `<rect x="650" y="193" width="65" height="14" rx="1" fill="%s" stroke="#363636" stroke-width="1"/>`, pipeOutColor)
	// Drain arrowhead
	b.WriteString(`<polygon points="712,196 720,200 712,204" fill="#363636"/>`)

	// --- Tank ---

	// Tank body
	fmt.Fprintf(&b, `<rect x="%.0f" y="%.0f" width="%.0f" height="%.0f" rx="6" fill="#f5f5f5" stroke="#363636" stroke-width="3"/>`,
		tankX, tankY, tankW, tankH)

	// Water fill
	if waterH > 0.5 {
		fmt.Fprintf(&b, `<rect x="%.0f" y="%.1f" width="%.0f" height="%.1f" fill="%s" opacity="0.7" rx="3"/>`,
			tankX+3, waterY, tankW-6, waterH, waterColor)
	}

	// Float switch high mark (95%)
	fmt.Fprintf(&b, `<line x1="%.0f" y1="%.1f" x2="%.0f" y2="%.1f" stroke="#f14668" stroke-width="2" stroke-dasharray="4,2"/>`,
		tankX-6, highY, tankX+6, highY)
	fmt.Fprintf(&b, `<text x="%.0f" y="%.1f" text-anchor="end" font-size="10" fill="#f14668">95%%</text>`,
		tankX-9, highY+4)

	// Float switch low mark (5%)
	fmt.Fprintf(&b, `<line x1="%.0f" y1="%.1f" x2="%.0f" y2="%.1f" stroke="#b5890a" stroke-width="2" stroke-dasharray="4,2"/>`,
		tankX-6, lowY, tankX+6, lowY)
	fmt.Fprintf(&b, `<text x="%.0f" y="%.1f" text-anchor="end" font-size="10" fill="#b5890a">5%%</text>`,
		tankX-9, lowY+4)

	// Level text
	fmt.Fprintf(&b, `<text x="%.0f" y="195" text-anchor="middle" font-size="32" font-weight="bold" fill="#363636">%.1f%%</text>`,
		tankX+tankW/2, level)
	fmt.Fprintf(&b, `<text x="%.0f" y="220" text-anchor="middle" font-size="13" fill="#4a4a4a">TANK</text>`,
		tankX+tankW/2)

	// --- Pump: ISA centrifugal pump (circle + internal triangle), clickable ---

	b.WriteString(`<a href="/pump" style="cursor:pointer">`)
	fmt.Fprintf(&b, `<circle cx="80" cy="200" r="40" fill="%s" stroke="#363636" stroke-width="2.5"/>`, pumpFill)
	// Triangle inside circle (discharge direction indicator, pointing right)
	b.WriteString(`<polygon points="65,182 65,218 100,200" fill="none" stroke="#363636" stroke-width="2"/>`)
	b.WriteString(`<text x="80" y="258" text-anchor="middle" font-size="13" font-weight="bold" fill="#363636">PUMP</text>`)
	pumpLabel := "OFF"
	if pump {
		pumpLabel = "ON"
	}
	fmt.Fprintf(&b, `<text x="80" y="274" text-anchor="middle" font-size="11" fill="#4a4a4a">%s</text>`, pumpLabel)
	b.WriteString(`</a>`)

	// --- Valve: ISA gate valve (bowtie), clickable ---

	b.WriteString(`<a href="/valve" style="cursor:pointer">`)
	fmt.Fprintf(&b, `<polygon points="570,175 610,200 570,225" fill="%s" stroke="#363636" stroke-width="2"/>`, valveFill)
	fmt.Fprintf(&b, `<polygon points="650,175 610,200 650,225" fill="%s" stroke="#363636" stroke-width="2"/>`, valveFill)
	b.WriteString(`<text x="610" y="248" text-anchor="middle" font-size="13" font-weight="bold" fill="#363636">VALVE</text>`)
	valveLabel := "CLOSED"
	if valve {
		valveLabel = "OPEN"
	}
	fmt.Fprintf(&b, `<text x="610" y="264" text-anchor="middle" font-size="11" fill="#4a4a4a">%s</text>`, valveLabel)
	b.WriteString(`</a>`)

	// Flow arrow on outlet pipe
	if valve && level > 0 {
		b.WriteString(`<polygon points="670,196 680,200 670,204" fill="#fff" opacity="0.6"/>`)
	}

	b.WriteString(`</svg>`)
	return b.String()
}
