//go:build !(js && wasm)

package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
)

// Storage provides JSON file-based persistence with atomic writes.
type Storage struct {
	mu      sync.Mutex
	dataDir string
}

// NewStorage creates a Storage backed by the given directory.
func NewStorage(dataDir string) (*Storage, error) {
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, err
	}
	return &Storage{dataDir: dataDir}, nil
}

func (s *Storage) path(name string) string {
	return filepath.Join(s.dataDir, name)
}

// atomicWrite writes data to a temp file then renames for crash safety.
func (s *Storage) atomicWrite(name string, data []byte) error {
	p := s.path(name)
	tmp := p + ".tmp"
	if err := os.WriteFile(tmp, data, 0644); err != nil {
		return err
	}
	return os.Rename(tmp, p)
}

func (s *Storage) readFile(name string) ([]byte, error) {
	data, err := os.ReadFile(s.path(name))
	if os.IsNotExist(err) {
		return nil, nil
	}
	return data, err
}

// LoadState reads the simulation state from disk. Returns nil if not found.
func (s *Storage) LoadState() (*SimState, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := s.readFile("state.json")
	if err != nil || data == nil {
		return nil, err
	}
	var st SimState
	if err := json.Unmarshal(data, &st); err != nil {
		return nil, err
	}
	return &st, nil
}

// SaveState writes the simulation state to disk atomically.
func (s *Storage) SaveState(st SimState) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := json.MarshalIndent(st, "", "  ")
	if err != nil {
		return err
	}
	return s.atomicWrite("state.json", data)
}

// LoadLogs reads log entries from disk. Returns empty slice if not found.
func (s *Storage) LoadLogs() ([]LogEntry, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := s.readFile("logs.json")
	if err != nil || data == nil {
		return nil, err
	}
	var logs []LogEntry
	if err := json.Unmarshal(data, &logs); err != nil {
		return nil, err
	}
	return logs, nil
}

// AppendLog adds a log entry and writes back to disk.
func (s *Storage) AppendLog(entry LogEntry) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	var logs []LogEntry
	data, err := s.readFile("logs.json")
	if err != nil {
		return err
	}
	if data != nil {
		if err := json.Unmarshal(data, &logs); err != nil {
			return err
		}
	}
	logs = append(logs, entry)
	// Keep last 100 entries
	if len(logs) > 100 {
		logs = logs[len(logs)-100:]
	}

	out, err := json.MarshalIndent(logs, "", "  ")
	if err != nil {
		return err
	}
	return s.atomicWrite("logs.json", out)
}

// LoadDiagnostics reads diagnostic records from disk. Returns empty slice if not found.
func (s *Storage) LoadDiagnostics() ([]DiagRecord, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := s.readFile("diagnostics.json")
	if err != nil || data == nil {
		return nil, err
	}
	var recs []DiagRecord
	if err := json.Unmarshal(data, &recs); err != nil {
		return nil, err
	}
	return recs, nil
}

// SaveDiagnostic appends a diagnostic record and writes back to disk.
func (s *Storage) SaveDiagnostic(rec DiagRecord) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	var recs []DiagRecord
	data, err := s.readFile("diagnostics.json")
	if err != nil {
		return err
	}
	if data != nil {
		if err := json.Unmarshal(data, &recs); err != nil {
			return err
		}
	}
	recs = append(recs, rec)
	// Keep last 500 records
	if len(recs) > 500 {
		recs = recs[len(recs)-500:]
	}

	out, err := json.MarshalIndent(recs, "", "  ")
	if err != nil {
		return err
	}
	return s.atomicWrite("diagnostics.json", out)
}
