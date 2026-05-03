package main

import (
	"context"
	"sync"
	"time"

	"codeberg.org/hum3/lofigui/widgets/watertank"
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

// Snapshot returns a watertank.State suitable for the schematic renderer.
func (s *Simulation) Snapshot() watertank.State {
	s.mu.Lock()
	defer s.mu.Unlock()
	return watertank.State{
		Level:     s.tankLevel,
		PumpOn:    s.pumpOn,
		ValveOpen: s.valveOpen,
		Running:   s.running,
		PumpHref:  "/pump",
		ValveHref: "/valve",
	}
}
