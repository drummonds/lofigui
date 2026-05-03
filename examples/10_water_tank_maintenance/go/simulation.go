package main

import (
	"context"
	"fmt"
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

// MaintenanceSnapshot holds a point-in-time copy of maintenance state.
type MaintenanceSnapshot struct {
	Type     string   // "" | "pump" | "valve"
	Progress float64  // 0.0–100.0
	Status   string   // "" | "running" | "completed" | "cancelled"
	Log      []string // last 5 step messages
}

// Simulation holds the water tank state with diagnostic tracking and maintenance.
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

	// Maintenance state
	maintType     string // "" | "pump" | "valve"
	maintProgress float64
	maintStatus   string // "" | "running" | "completed" | "cancelled"
	maintCancel   context.CancelFunc
	maintLog      []string
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
		// Suppress auto-on during pump maintenance
		if s.maintType != "pump" || s.maintStatus != "running" {
			s.pumpOn = true
			s.floatTrips++
		}
	}

	// Track pump on/off transitions caused by float switch
	now := time.Now()
	if prevPump && !s.pumpOn {
		if !s.pumpLastOn.IsZero() {
			s.pumpOnTime += now.Sub(s.pumpLastOn)
			s.pumpLastOn = time.Time{}
		}
	} else if !prevPump && s.pumpOn {
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
	if s.tickCount%4 == 0 {
		s.levelHistory = append(s.levelHistory, LevelEntry{
			Tick:  s.tickCount,
			Level: s.tankLevel,
		})
		if len(s.levelHistory) > 60 {
			s.levelHistory = s.levelHistory[len(s.levelHistory)-60:]
		}
	}
}

// TogglePump toggles the pump state and tracks cycles.
// Returns early if pump maintenance is running.
func (s *Simulation) TogglePump() {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Lockout during pump maintenance
	if s.maintType == "pump" && s.maintStatus == "running" {
		return
	}

	now := time.Now()
	if s.pumpOn {
		if !s.pumpLastOn.IsZero() {
			s.pumpOnTime += now.Sub(s.pumpLastOn)
			s.pumpLastOn = time.Time{}
		}
	} else {
		s.pumpCycles++
		if s.running {
			s.pumpLastOn = now
		}
	}
	s.pumpOn = !s.pumpOn
}

// ToggleValve toggles the valve state and tracks cycles.
// Returns early if valve maintenance is running.
func (s *Simulation) ToggleValve() {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Lockout during valve maintenance
	if s.maintType == "valve" && s.maintStatus == "running" {
		return
	}

	now := time.Now()
	if s.valveOpen {
		if !s.valveLastOpen.IsZero() {
			s.valveOnTime += now.Sub(s.valveLastOpen)
			s.valveLastOpen = time.Time{}
		}
	} else {
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
	now := time.Now()
	if s.pumpOn && !s.pumpLastOn.IsZero() {
		snap.PumpOnTime += now.Sub(s.pumpLastOn)
	}
	if s.valveOpen && !s.valveLastOpen.IsZero() {
		snap.ValveOnTime += now.Sub(s.valveLastOpen)
	}
	snap.History = make([]LevelEntry, len(s.levelHistory))
	copy(snap.History, s.levelHistory)
	return snap
}

// MaintenanceStatus returns a thread-safe snapshot of maintenance state.
func (s *Simulation) MaintenanceStatus() MaintenanceSnapshot {
	s.mu.Lock()
	defer s.mu.Unlock()
	log := make([]string, len(s.maintLog))
	copy(log, s.maintLog)
	return MaintenanceSnapshot{
		Type:     s.maintType,
		Progress: s.maintProgress,
		Status:   s.maintStatus,
		Log:      log,
	}
}

// StartMaintenance begins a maintenance operation. Returns error if one is already running.
func (s *Simulation) StartMaintenance(maintType string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.maintStatus == "running" {
		return fmt.Errorf("maintenance already running: %s", s.maintType)
	}

	s.maintType = maintType
	s.maintProgress = 0
	s.maintStatus = "running"
	s.maintLog = nil

	// Disable equipment under maintenance
	now := time.Now()
	if maintType == "pump" {
		s.disablePumpLocked(now)
	} else if maintType == "valve" {
		s.closeValveLocked(now)
	}

	ctx, cancel := context.WithCancel(context.Background())
	s.maintCancel = cancel

	go s.runMaintenance(ctx, maintType)
	return nil
}

// CancelMaintenance cancels the running maintenance operation.
func (s *Simulation) CancelMaintenance() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.maintStatus != "running" {
		return
	}
	if s.maintCancel != nil {
		s.maintCancel()
	}
}

// ClearMaintenance resets maintenance state after completed/cancelled.
func (s *Simulation) ClearMaintenance() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.maintStatus == "running" {
		return // can't clear while running
	}
	s.maintType = ""
	s.maintProgress = 0
	s.maintStatus = ""
	s.maintCancel = nil
	s.maintLog = nil
}

// disablePumpLocked turns off the pump. Must be called with lock held.
func (s *Simulation) disablePumpLocked(now time.Time) {
	if s.pumpOn {
		if !s.pumpLastOn.IsZero() {
			s.pumpOnTime += now.Sub(s.pumpLastOn)
			s.pumpLastOn = time.Time{}
		}
		s.pumpOn = false
	}
}

// closeValveLocked closes the valve. Must be called with lock held.
func (s *Simulation) closeValveLocked(now time.Time) {
	if s.valveOpen {
		if !s.valveLastOpen.IsZero() {
			s.valveOnTime += now.Sub(s.valveLastOpen)
			s.valveLastOpen = time.Time{}
		}
		s.valveOpen = false
	}
}

// runMaintenance executes the maintenance steps in a goroutine.
func (s *Simulation) runMaintenance(ctx context.Context, maintType string) {
	var steps []string
	if maintType == "pump" {
		steps = []string{
			"Shutting down pump",
			"Draining pump casing",
			"Inspecting impeller",
			"Checking shaft seal",
			"Lubricating bearings",
			"Testing motor windings",
			"Reassembling pump",
			"Priming pump casing",
		}
	} else {
		steps = []string{
			"Closing valve",
			"Inspecting valve seat",
			"Checking actuator",
			"Replacing gaskets",
			"Testing valve travel",
		}
	}

	for i, step := range steps {
		select {
		case <-ctx.Done():
			s.mu.Lock()
			s.maintStatus = "cancelled"
			s.appendLog("Maintenance cancelled")
			s.mu.Unlock()
			return
		case <-time.After(1 * time.Second):
			s.mu.Lock()
			s.maintProgress = float64(i+1) / float64(len(steps)) * 100
			s.appendLog(step)
			s.mu.Unlock()
		}
	}

	s.mu.Lock()
	s.maintProgress = 100
	s.maintStatus = "completed"
	s.appendLog("Maintenance complete")
	s.mu.Unlock()
}

// appendLog adds a message to the maintenance log, keeping last 5 entries.
// Must be called with lock held.
func (s *Simulation) appendLog(msg string) {
	s.maintLog = append(s.maintLog, msg)
	if len(s.maintLog) > 5 {
		s.maintLog = s.maintLog[len(s.maintLog)-5:]
	}
}

// Snapshot returns a watertank.State suitable for the schematic renderer,
// including a maintenance overlay when a maintenance op is running.
func (s *Simulation) Snapshot() watertank.State {
	s.mu.Lock()
	defer s.mu.Unlock()
	state := watertank.State{
		Level:     s.tankLevel,
		PumpOn:    s.pumpOn,
		ValveOpen: s.valveOpen,
		Running:   s.running,
		PumpHref:  "/pump",
		ValveHref: "/valve",
	}
	if s.maintStatus == "running" {
		state.MaintType = s.maintType
		state.MaintProgress = s.maintProgress
	}
	return state
}
