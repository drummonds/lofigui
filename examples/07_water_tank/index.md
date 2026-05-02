# 07 — Water Tank

A real-time SVG dashboard with a simulated water tank. Combines polling with POST controls and generated SVG schematics. Also compiles to WASM.

**[Interactivity level](../research-philosophy.html#the-interactivity-spectrum):** 3 — Polling (whole page refresh)
**[State scope](../research-philosophy.html#the-state-dimension):** Global (server build — one shared simulation) / Individual (WASM build — each browser runs its own tank)

<!-- TODO: Add screenshots, annotated code walkthrough -->

---

## The simulation

<!-- TODO: Background goroutine, tick rate, SVG generation -->

---

## SVG schematic

<!-- TODO: buildSVG(), clickable <a> links, P&ID style -->

---

## Dual build: server and WASM

<!-- TODO: main.go vs main_wasm.go, shared model -->

---

## Links

- [Launch Demo](demo.html)
- [Source on Codeberg](https://codeberg.org/hum3/lofigui/src/branch/main/examples/07_water_tank)
