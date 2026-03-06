# Example 11: Water Tank with Persistent Storage

WASM frontend + Go API server architecture. The WASM binary runs the simulation in-browser while a Go HTTP server provides a JSON storage API backed by local filesystem.

## Architecture

```
Browser (WASM)                    Go API Server (:1350)
┌──────────────────┐             ┌──────────────────┐
│ simulation.go    │  fetch()    │ main.go           │
│ main_wasm.go     │ ──────────>│ storage.go        │
│ (renders UI)     │  JSON API   │ (serves WASM +   │
│                  │ <──────────│  reads/writes JSON)│
└──────────────────┘             └────────┬─────────┘
                                          │
                                   data/state.json
                                   data/logs.json
                                   data/diagnostics.json
```

## Running

```bash
task go-example:11
```

Then open http://localhost:1350

## What's persisted

- **state.json** — tank level, pump/valve state, diagnostic counters
- **logs.json** — maintenance log entries (last 100)
- **diagnostics.json** — periodic diagnostic snapshots (last 500)

State auto-saves every 5 seconds while running, and immediately on user actions (start/stop, pump/valve toggle). Uses `navigator.sendBeacon` on page unload for best-effort save.

## Data directory

State files are stored in `examples/11_water_tank_storage/go/data/` (gitignored). Delete this directory to reset all state.
