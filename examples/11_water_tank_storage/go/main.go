//go:build !(js && wasm)

package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

func main() {
	// Determine data directory relative to the executable's working dir
	dataDir := filepath.Join("data")
	store, err := NewStorage(dataDir)
	if err != nil {
		log.Fatalf("Failed to create storage: %v", err)
	}

	templateDir := "templates"

	// Determine wasm_exec.js path from Go toolchain
	goRoot := os.Getenv("GOROOT")
	if goRoot == "" {
		// Try to find it
		goRoot = "/usr/local/go"
	}
	wasmExecPath := filepath.Join(goRoot, "misc", "wasm", "wasm_exec.js")

	// Static files
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		http.ServeFile(w, r, filepath.Join(templateDir, "index.html"))
	})

	http.HandleFunc("/app.js", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filepath.Join(templateDir, "app.js"))
	})

	http.HandleFunc("/main.wasm", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "main.wasm")
	})

	http.HandleFunc("/wasm_exec.js", func(w http.ResponseWriter, r *http.Request) {
		// Try local copy first, fall back to GOROOT
		if _, err := os.Stat("wasm_exec.js"); err == nil {
			http.ServeFile(w, r, "wasm_exec.js")
			return
		}
		http.ServeFile(w, r, wasmExecPath)
	})

	// JSON API: state
	http.HandleFunc("/api/state", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			st, err := store.LoadState()
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			if st == nil {
				w.Header().Set("Content-Type", "application/json")
				w.Write([]byte("null"))
				return
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(st)

		case "POST":
			body, err := io.ReadAll(r.Body)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			var st SimState
			if err := json.Unmarshal(body, &st); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			if err := store.SaveState(st); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusNoContent)

		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// JSON API: logs
	http.HandleFunc("/api/logs", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			logs, err := store.LoadLogs()
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			if logs == nil {
				logs = []LogEntry{}
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(logs)

		case "POST":
			body, err := io.ReadAll(r.Body)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			var entry LogEntry
			if err := json.Unmarshal(body, &entry); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			if err := store.AppendLog(entry); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusNoContent)

		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// JSON API: diagnostics
	http.HandleFunc("/api/diagnostics", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			recs, err := store.LoadDiagnostics()
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			if recs == nil {
				recs = []DiagRecord{}
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(recs)

		case "POST":
			body, err := io.ReadAll(r.Body)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			var rec DiagRecord
			if err := json.Unmarshal(body, &rec); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			if err := store.SaveDiagnostic(rec); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusNoContent)

		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	addr := ":1350"
	log.Printf("Starting Water Tank Storage on http://localhost%s", addr)
	log.Printf("Data directory: %s", dataDir)
	log.Fatal(http.ListenAndServe(addr, nil))
}
