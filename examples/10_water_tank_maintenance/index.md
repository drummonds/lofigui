# 10 — Water Tank Maintenance

Extends HTMX pattern with long-running background goroutines for equipment maintenance. Demonstrates progress tracking, context cancellation, and equipment lockout.

**[Interactivity level](../research-philosophy.html#the-interactivity-spectrum):** 5 — HTMX partial updates
**[State scope](../research-philosophy.html#the-state-dimension):** Global (server) / Individual (WASM build — maintenance goroutines run per-browser)

<!-- TODO: Add screenshots, annotated code walkthrough -->

---

## Background operations

<!-- TODO: runMaintenance goroutine, progress under mutex -->

---

## Context cancellation

<!-- TODO: context.Context, cancel(), select on ctx.Done() -->

---

## Equipment lockout

<!-- TODO: Suppress controls during maintenance -->

---

## Links

- [Launch Demo](demo.html)
- [Source on Codeberg](https://codeberg.org/hum3/lofigui/src/branch/main/examples/10_water_tank_maintenance)
