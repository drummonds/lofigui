# Example 07 — Water Tank SCADA Simulation

SCADA-style water tank simulation with an inline SVG P&ID schematic and real-time polling.

## Run (server mode)

```bash
task go-example:07
# Open http://localhost:1347
```

## Run (WASM in browser)

```bash
task build-pages
go run ./cmd/serve -dir docs -port 8000
# Open http://localhost:8000/07_water_tank/
```

## How it works

Hybrid pattern combining background goroutine polling (like example 01) with POST-based controls (like example 06). The buffer is re-rendered from simulation state on every GET — appropriate for a live dashboard.

- **Pump** fills tank at ~6%/s
- **Valve** drains at ~2%/s
- **Float switch** auto-off pump at 95%, auto-on at 5%
- SVG schematic updates colour to reflect pump/valve/level state

Server mode uses `lofigui.LayoutNavbar` — no custom template file needed. WASM mode uses a standalone HTML page with the same Bulma styling.

## File structure

- `go/simulation.go` — Shared simulation logic (no build tags)
- `go/main.go` — Server mode (`//go:build !(js && wasm)`)
- `go/main_wasm.go` — WASM mode (`//go:build js && wasm`)
- `templates/index.html` — WASM HTML page
- `templates/app.js` — WASM JavaScript loader

## SVG symbols

The P&ID symbols (centrifugal pump, gate valve) are inspired by
[FUXA-SVG-Widgets](https://github.com/frangoteam/FUXA-SVG-Widgets) (MIT license).
