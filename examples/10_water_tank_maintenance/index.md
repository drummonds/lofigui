<style>
.annotation { border-left: 3px solid #3273dc; background: #f0f4ff; padding: 0.75em 1em; margin: 0.75em 0; border-radius: 0 4px 4px 0; font-size: 0.9em; }
.annotation strong { color: #3273dc; }
.screenshot { border: 1px solid #dbdbdb; border-radius: 4px; box-shadow: 0 2px 6px rgba(0,0,0,0.1); overflow: hidden; }
.screenshot img { display: block; width: 100%; height: auto; }
</style>

# 10 — Water Tank Maintenance

Extends [09](../09_water_tank_htmx/) with long-running background maintenance goroutines on the pump and valve. Each maintenance op runs N steps × 1 s, updating progress under a mutex; HTMX picks up the progress on its existing 1 s poll. Cancellation is via `context.Context`. While maintenance is running, the affected equipment is locked out — buttons disabled, float-switch suppressed, an orange dashed ring drawn on the schematic.

**[Interactivity level](../research-philosophy.html#the-interactivity-spectrum):** 5 — HTMX partial updates (with background ops)
**[State scope](../research-philosophy.html#the-state-dimension):** Global (server build) / Individual (WASM build — each browser runs its own maintenance ops)

<div class="buttons">
<a href="demo.html" class="button is-primary">Launch Demo</a>
<a target="_blank" href="https://codeberg.org/hum3/lofigui/src/branch/main/examples/10_water_tank_maintenance" class="button is-light">Source on Codeberg</a>
</div>

<div class="columns is-vcentered">
<div class="column is-5">
<figure class="image screenshot">
<img src="../10_idle.png" alt="Idle — simulation running, no maintenance">
<figcaption class="has-text-centered has-text-grey is-size-7 mt-1">Idle — sim running, no maintenance</figcaption>
</figure>
</div>
<div class="column is-narrow has-text-centered">
<span style="font-size: 2rem; color: #999;">&rarr;</span>
</div>
<div class="column is-5">
<figure class="image screenshot">
<img src="../10_maintenance.png" alt="Pump maintenance running — orange ring + progress bar">
<figcaption class="has-text-centered has-text-grey is-size-7 mt-1">Pump maintenance in progress</figcaption>
</figure>
</div>
</div>

<div class="annotation">
<strong>What changed:</strong> the user clicked <em>Pump Maintenance</em>. A goroutine kicks off a multi-step procedure ("Inspecting impeller", …) that emits progress every second; the orange dashed ring on the pump and the "MAINT 38%" caption come from the shared <a href="https://codeberg.org/hum3/lofigui/src/branch/main/widgets/watertank">watertank</a> widget reading <code>State.MaintType</code> / <code>State.MaintProgress</code>. The <em>Pump On</em> button is disabled (lockout); the float-switch stops auto-toggling the pump too.
</div>

---

## The maintenance goroutine

`runMaintenance` walks a list of steps, sleeping 1 s between each. At every step it records progress and a log line under the simulation mutex, then `select`s on cancellation:

```go
func (s *Simulation) runMaintenance(ctx context.Context, kind string) {
    steps := stepsFor(kind)
    for i, step := range steps {
        select {
        case <-ctx.Done():
            s.mu.Lock()
            s.maintStatus = "cancelled"
            s.mu.Unlock()
            return
        case <-time.After(1 * time.Second):
            s.mu.Lock()
            s.maintProgress = float64(i+1) / float64(len(steps)) * 100
            s.appendLog(step)
            s.mu.Unlock()
        }
    }
    // … set status = "completed"
}
```

`StartMaintenance` rejects a second concurrent op with an error; `CancelMaintenance` calls the context's `cancel()`, which the `select` above picks up.

---

## Equipment lockout

`TogglePump`/`ToggleValve` early-return when their equipment is under maintenance. The float-switch logic is suppressed for the same equipment — water can rise past 95% during a pump maintenance window without the pump auto-shutting (because it's already off and locked out).

The schematic reflects all this without adding any rendering code in the example: `Snapshot()` just sets `MaintType`/`MaintProgress` on the `watertank.State`, and the widget draws the orange ring and `MAINT n%` label.

---

## Cancellation, end to end

Three independent `context.Context`s are at play:

- The HTTP request context (`r.Context()`) — auto-cancelled when the client disconnects.
- The simulation tick goroutine — cancelled by `Stop()`.
- The maintenance goroutine — cancelled by `CancelMaintenance()` or by stopping the simulation.

All three terminate cleanly via `select { case <-ctx.Done(): ... }`. There's no shared "I'm done" flag; the context tree is the single source of truth.

[Source on Codeberg](https://codeberg.org/hum3/lofigui/src/branch/main/examples/10_water_tank_maintenance)

---

## Running

```bash
task go-example:10         # server on :1349
task docs:capture:10       # capture the two screenshots above
```
