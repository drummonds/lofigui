<style>
.annotation { border-left: 3px solid #3273dc; background: #f0f4ff; padding: 0.75em 1em; margin: 0.75em 0; border-radius: 0 4px 4px 0; font-size: 0.9em; }
.annotation strong { color: #3273dc; }
.screenshot { border: 1px solid #dbdbdb; border-radius: 4px; box-shadow: 0 2px 6px rgba(0,0,0,0.1); overflow: hidden; }
.screenshot img { display: block; width: 100%; height: auto; }
</style>

# 07 — Water Tank

A real-time SCADA-style dashboard with a simulated water tank. The model is a background goroutine that ticks the level every 500 ms; the page polls every second to redraw the schematic. Pump and valve are clickable on the SVG and via buttons.

**[Interactivity level](../research-philosophy.html#the-interactivity-spectrum):** 3 — Polling (whole page refresh)
**[State scope](../research-philosophy.html#the-state-dimension):** Global (server build — one shared simulation) / Individual (WASM build — each browser runs its own tank)

<div class="buttons">
<a href="demo.html" class="button is-primary">Launch Demo</a>
<a target="_blank" href="https://codeberg.org/hum3/lofigui/src/branch/main/examples/07_water_tank" class="button is-light">Source on Codeberg</a>
</div>

<div class="columns is-vcentered">
<div class="column is-5">
<figure class="image screenshot">
<img src="../07_stopped.png" alt="Stopped — tank at 0%, simulation idle">
<figcaption class="has-text-centered has-text-grey is-size-7 mt-1">Stopped (level = 0%)</figcaption>
</figure>
</div>
<div class="column is-narrow has-text-centered">
<span style="font-size: 2rem; color: #999;">&rarr;</span>
</div>
<div class="column is-5">
<figure class="image screenshot">
<img src="../07_running.png" alt="Running — pump filling, level 33%">
<figcaption class="has-text-centered has-text-grey is-size-7 mt-1">Running (pump filling)</figcaption>
</figure>
</div>
</div>

<div class="annotation">
<strong>What changed between the two shots:</strong> the user pressed <em>Start Simulation</em>. The pump turns the inlet pipe blue (flow), the level climbs ~6%/s, and the navbar badge flips from <em>Stopped</em> to <em>Running</em>. The 95% / 5% dashed marks are the float-switch trip points; the pump auto-cuts at 95% and auto-restarts at 5%.
</div>

---

## The simulation

`Simulation` is a tiny state machine: tank level, pump on/off, valve open/closed, plus a tick goroutine started from `Start()`:

```go
go func() {
    ticker := time.NewTicker(500 * time.Millisecond)
    for {
        select {
        case <-ctx.Done(): return
        case <-ticker.C:    s.tick()
        }
    }
}()
```

Each tick adds 3 if the pump is on, subtracts 1 if the valve is open, and trips the float switches at 5% and 95%. `Stop()` cancels the context; the goroutine exits cleanly.

---

## Schematic — shared widget

The SVG is rendered by [`watertank.Render(state)`](https://codeberg.org/hum3/lofigui/src/branch/main/widgets/watertank), shared across examples 07–11. The example's only schematic-related code is `Snapshot()`, which copies the simulation state into a `watertank.State`:

```go
func (s *Simulation) Snapshot() watertank.State {
    s.mu.Lock(); defer s.mu.Unlock()
    return watertank.State{
        Level:     s.tankLevel,
        PumpOn:    s.pumpOn,
        ValveOpen: s.valveOpen,
        Running:   s.running,
        PumpHref:  "/pump",
        ValveHref: "/valve",
    }
}
```

The `<a href>` wrapping pump/valve in the SVG makes the schematic itself clickable — clicking the pump POSTs to `/pump` and toggles it.

---

## Dual build: server and WASM

`main.go` and `main_wasm.go` share `simulation.go`. The server build uses `app.HandleRoot`/`HandleDisplay` for polling; the WASM build calls into the same `Simulation` via a thin JS bridge.

```go
//go:build !(js && wasm)
func main() {
    sim := &Simulation{pumpOn: true}
    app := lofigui.NewApp()
    // ... register /, /start, /stop, /pump, /valve ...
    log.Fatal(http.ListenAndServe(":1347", nil))
}
```

[Source on Codeberg](https://codeberg.org/hum3/lofigui/src/branch/main/examples/07_water_tank)

---

## Running

```bash
task go-example:07         # server on :1347
task docs:capture:07       # capture the two screenshots above
```
