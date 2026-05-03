<style>
.annotation { border-left: 3px solid #3273dc; background: #f0f4ff; padding: 0.75em 1em; margin: 0.75em 0; border-radius: 0 4px 4px 0; font-size: 0.9em; }
.annotation strong { color: #3273dc; }
.screenshot { border: 1px solid #dbdbdb; border-radius: 4px; box-shadow: 0 2px 6px rgba(0,0,0,0.1); overflow: hidden; }
.screenshot img { display: block; width: 100%; height: auto; }
</style>

# 09 — Water Tank HTMX

The same multi-page tank as [08](../08_water_tank_multi/), upgraded from full-page Refresh polling to **HTMX partial updates**. Only the results div is swapped each second — forms, buttons, and any in-flight click are not interrupted. No `App` is needed; the example talks to a `Controller` directly.

**[Interactivity level](../research-philosophy.html#the-interactivity-spectrum):** 5 — HTMX partial updates
**[State scope](../research-philosophy.html#the-state-dimension):** Global (server build) / Individual (WASM build)

<div class="buttons">
<a href="demo.html" class="button is-primary">Launch Demo</a>
<a target="_blank" href="https://codeberg.org/hum3/lofigui/src/branch/main/examples/09_water_tank_htmx" class="button is-light">Source on Codeberg</a>
</div>

<div class="columns is-vcentered">
<div class="column is-5">
<figure class="image screenshot">
<img src="../09_schematic.png" alt="Schematic page with HTMX partial updates">
<figcaption class="has-text-centered has-text-grey is-size-7 mt-1">/ — schematic (HTMX-polled)</figcaption>
</figure>
</div>
<div class="column is-narrow has-text-centered">
<span style="font-size: 2rem; color: #999;">&rarr;</span>
</div>
<div class="column is-5">
<figure class="image screenshot">
<img src="../09_diagnostics.png" alt="Diagnostics page with HTMX partial updates">
<figcaption class="has-text-centered has-text-grey is-size-7 mt-1">/diagnostics — counters & history</figcaption>
</figure>
</div>
</div>

<div class="annotation">
<strong>What HTMX buys here:</strong> in 08, the whole page reloads every second — focus is lost, forms are reset, an in-flight POST may race the next refresh. In 09, only the results div is swapped via <code>hx-get="/fragment" hx-trigger="every 1s"</code>. The page chrome (navbar, buttons, forms) is stable; the schematic updates underneath.
</div>

---

## The page + fragment pair

Each route has two endpoints: a full page that mounts the HTMX wiring, and a fragment endpoint that returns just the inner HTML.

```
GET /                     → full page, hx-get="/fragment"
GET /fragment             → just the schematic HTML

GET /diagnostics          → full page, hx-get="/fragment/diagnostics"
GET /fragment/diagnostics → just the diagnostics HTML
```

```html
<div id="results"
     hx-get="/fragment"
     hx-trigger="every 1s"
     hx-swap="innerHTML">
  {{.results}}
</div>
```

The first render fills `{{.results}}` from the controller; HTMX takes over from there.

---

## Concurrency: `renderAndCapture`

`lofigui.Print`, `Reset`, and `Buffer` all touch a shared global. Multiple HTMX clients hitting `/fragment` concurrently would interleave their writes. A small mutex around reset+render+capture keeps each request's output isolated:

```go
var renderMu sync.Mutex

func renderAndCapture(fn func()) string {
    renderMu.Lock(); defer renderMu.Unlock()
    lofigui.Reset()
    fn()
    return lofigui.Buffer()
}
```

Each fragment endpoint calls `renderAndCapture(func() { renderSchematic(sim) })` and writes the captured string to the response.

[Source on Codeberg](https://codeberg.org/hum3/lofigui/src/branch/main/examples/09_water_tank_htmx)

---

## Running

```bash
task go-example:09         # server on :1349
task docs:capture:09       # capture the two screenshots above
```
