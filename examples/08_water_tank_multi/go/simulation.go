package main

import (
	"context"
	"sync"
	"time"

	"codeberg.org/hum3/lofigui/widgets/watertank"
)

// LevelEntry records a tank level at a point in time.
type LevelEntry struct {
	Tick  int
	Level float64
}

// DiagSnapshot holds a point-in-time copy of diagnostic counters.
type DiagSnapshot struct {
	PumpCycles  int
	PumpOnTime  time.Duration
	ValveCycles int
	ValveOnTime time.Duration
	FloatTrips  int
	TickCount   int
	History     []LevelEntry
}

// Simulation holds the water tank state with diagnostic tracking.
type Simulation struct {
	mu        sync.Mutex
	running   bool
	cancel    context.CancelFunc
	tankLevel float64 // 0.0–100.0
	pumpOn    bool
	valveOpen bool

	// Diagnostic counters
	pumpCycles    int
	pumpOnTime    time.Duration
	pumpLastOn    time.Time
	valveCycles   int
	valveOnTime   time.Duration
	valveLastOpen time.Time
	floatTrips    int
	levelHistory  []LevelEntry
	tickCount     int
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
	// Record initial state transitions
	now := time.Now()
	if s.pumpOn {
		s.pumpLastOn = now
	}
	if s.valveOpen {
		s.valveLastOpen = now
	}
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
	// Accumulate any open durations
	now := time.Now()
	if s.pumpOn && !s.pumpLastOn.IsZero() {
		s.pumpOnTime += now.Sub(s.pumpLastOn)
		s.pumpLastOn = time.Time{}
	}
	if s.valveOpen && !s.valveLastOpen.IsZero() {
		s.valveOnTime += now.Sub(s.valveLastOpen)
		s.valveLastOpen = time.Time{}
	}
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

// tick updates tank level once and tracks diagnostics.
func (s *Simulation) tick() {
	s.mu.Lock()
	defer s.mu.Unlock()

	prevPump := s.pumpOn

	if s.pumpOn {
		s.tankLevel += 3.0
	}
	if s.valveOpen {
		s.tankLevel -= 1.0
	}

	// Float switch: auto-off at 95%, auto-on at 5%
	if s.tankLevel >= 95.0 && s.pumpOn {
		s.pumpOn = false
		s.floatTrips++
	}
	if s.tankLevel <= 5.0 && !s.pumpOn {
		s.pumpOn = true
		s.floatTrips++
	}

	// Track pump on/off transitions caused by float switch
	now := time.Now()
	if prevPump && !s.pumpOn {
		// Pump turned off
		if !s.pumpLastOn.IsZero() {
			s.pumpOnTime += now.Sub(s.pumpLastOn)
			s.pumpLastOn = time.Time{}
		}
	} else if !prevPump && s.pumpOn {
		// Pump turned on
		s.pumpCycles++
		s.pumpLastOn = now
	}

	// Clamp
	if s.tankLevel < 0 {
		s.tankLevel = 0
	}
	if s.tankLevel > 100 {
		s.tankLevel = 100
	}

	s.tickCount++
	// Record history every 4 ticks (~2 seconds)
	if s.tickCount%4 == 0 {
		s.levelHistory = append(s.levelHistory, LevelEntry{
			Tick:  s.tickCount,
			Level: s.tankLevel,
		})
		// Keep last 60 entries
		if len(s.levelHistory) > 60 {
			s.levelHistory = s.levelHistory[len(s.levelHistory)-60:]
		}
	}
}

// TogglePump toggles the pump state and tracks cycles.
func (s *Simulation) TogglePump() {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now()
	if s.pumpOn {
		// Turning off
		if !s.pumpLastOn.IsZero() {
			s.pumpOnTime += now.Sub(s.pumpLastOn)
			s.pumpLastOn = time.Time{}
		}
	} else {
		// Turning on
		s.pumpCycles++
		if s.running {
			s.pumpLastOn = now
		}
	}
	s.pumpOn = !s.pumpOn
}

// ToggleValve toggles the valve state and tracks cycles.
func (s *Simulation) ToggleValve() {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now()
	if s.valveOpen {
		// Closing
		if !s.valveLastOpen.IsZero() {
			s.valveOnTime += now.Sub(s.valveLastOpen)
			s.valveLastOpen = time.Time{}
		}
	} else {
		// Opening
		s.valveCycles++
		if s.running {
			s.valveLastOpen = now
		}
	}
	s.valveOpen = !s.valveOpen
}

// Diagnostics returns a snapshot of diagnostic counters.
func (s *Simulation) Diagnostics() DiagSnapshot {
	s.mu.Lock()
	defer s.mu.Unlock()

	snap := DiagSnapshot{
		PumpCycles:  s.pumpCycles,
		PumpOnTime:  s.pumpOnTime,
		ValveCycles: s.valveCycles,
		ValveOnTime: s.valveOnTime,
		FloatTrips:  s.floatTrips,
		TickCount:   s.tickCount,
	}
	// Add live durations if currently on
	now := time.Now()
	if s.pumpOn && !s.pumpLastOn.IsZero() {
		snap.PumpOnTime += now.Sub(s.pumpLastOn)
	}
	if s.valveOpen && !s.valveLastOpen.IsZero() {
		snap.ValveOnTime += now.Sub(s.valveLastOpen)
	}
	// Copy history
	snap.History = make([]LevelEntry, len(s.levelHistory))
	copy(snap.History, s.levelHistory)
	return snap
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
