<style>
.annotation { border-left: 3px solid #3273dc; background: #f0f4ff; padding: 0.75em 1em; margin: 0.75em 0; border-radius: 0 4px 4px 0; font-size: 0.9em; }
.annotation strong { color: #3273dc; }
.screenshot { border: 1px solid #dbdbdb; border-radius: 4px; box-shadow: 0 2px 6px rgba(0,0,0,0.1); overflow: hidden; }
.screenshot img { display: block; width: 100%; height: auto; }
</style>

# 08 — Water Tank Multi-Page

Same simulation as [07](../07_water_tank/), but with a second page (`/diagnostics`) that displays accumulated counters and a level history table. Both pages are independently polled by HTTP `Refresh` headers, so `/diagnostics` keeps refreshing itself rather than ping-ponging back to the schematic.

**[Interactivity level](../research-philosophy.html#the-interactivity-spectrum):** 3 — Polling (whole page refresh, per-page)
**[State scope](../research-philosophy.html#the-state-dimension):** Global (server build) / Individual (WASM build)

<div class="buttons">
<a href="demo.html" class="button is-primary">Launch Demo</a>
<a target="_blank" href="https://codeberg.org/hum3/lofigui/src/branch/main/examples/08_water_tank_multi" class="button is-light">Source on Codeberg</a>
</div>

<div class="columns is-vcentered">
<div class="column is-5">
<figure class="image screenshot">
<img src="../08_schematic.png" alt="Schematic page — pump filling, level 30%">
<figcaption class="has-text-centered has-text-grey is-size-7 mt-1">/ — schematic</figcaption>
</figure>
</div>
<div class="column is-narrow has-text-centered">
<span style="font-size: 2rem; color: #999;">&rarr;</span>
</div>
<div class="column is-5">
<figure class="image screenshot">
<img src="../08_diagnostics.png" alt="Diagnostics page — counters and level history table">
<figcaption class="has-text-centered has-text-grey is-size-7 mt-1">/diagnostics — counters & history</figcaption>
</figure>
</div>
</div>

<div class="annotation">
<strong>Same simulation, two views:</strong> the schematic page polls <code>/</code>, the diagnostics page polls <code>/diagnostics</code>. Both views read from the same shared <code>Simulation</code>. The schematic uses the shared <a href="https://codeberg.org/hum3/lofigui/src/branch/main/widgets/watertank"><code>watertank</code></a> widget; the diagnostics view is rendered straight to the buffer via <code>lofigui.Print</code>/<code>Table</code>.
</div>

---

## Diagnostics counters

`Simulation` accumulates pump/valve cycles, on-time durations, and a rolling level history. `Diagnostics()` returns a `DiagSnapshot` for the diagnostics page to render:

```go
type DiagSnapshot struct {
    PumpCycles  int
    PumpOnTime  time.Duration
    ValveCycles int
    ValveOnTime time.Duration
    FloatTrips  int
    TickCount   int
    History     []LevelEntry
}
```

The level history is kept down-sampled (one entry every N ticks) so the table stays a manageable size.

---

## Multi-page refresh

The Refresh header reloads whichever URL the browser was on, so each page polls itself:

```
GET /             → renderSchematic() → app.HandleDisplay(w, r)
GET /diagnostics  → renderDiagnostics() → app.HandleDisplay(w, r)
```

`app.HandleDisplay` writes a Refresh header pointing at `r.URL.Path`. Hitting `/diagnostics` keeps refreshing `/diagnostics`; hitting `/` keeps refreshing `/`. Add a third page and it just works.

[Source on Codeberg](https://codeberg.org/hum3/lofigui/src/branch/main/examples/08_water_tank_multi)

---

## Running

```bash
task go-example:08         # server on :1348
task docs:capture:08       # capture the two screenshots above
```
