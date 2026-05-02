# 09 — Water Tank HTMX

Replaces full-page HTTP Refresh polling with HTMX partial updates. Only the results div is swapped — forms, buttons, and inputs remain functional during live updates.

**[Interactivity level](../research-philosophy.html#the-interactivity-spectrum):** 5 — HTMX partial updates
**[State scope](../research-philosophy.html#the-state-dimension):** Global (server) / Individual (WASM build)

<!-- TODO: Add screenshots, annotated code walkthrough -->

---

## Why HTMX?

<!-- TODO: The polling limitation — forms reset, clicks interrupted -->

---

## Fragment endpoints

<!-- TODO: hx-get, hx-trigger, renderAndCapture mutex -->

---

## Controller without App

<!-- TODO: No App needed, ctrl.RenderTemplate directly -->

---

## Links

- [Launch Demo](demo.html)
- [Source on Codeberg](https://codeberg.org/hum3/lofigui/src/branch/main/examples/09_water_tank_htmx)
