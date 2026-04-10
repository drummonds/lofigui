# Lofigui Research

Design philosophy, trade-offs, and technical deep-dives.

## Research Pages

| Page | Description |
|------|-------------|
| [Philosophy](research-philosophy.html) | Design philosophy, the interactivity spectrum, and where lofigui sits |
| [Charts](research-charts.html) | Charting options — Go SVG vs JavaScript libraries |
| [Layouts](research-layouts.html) | Page layout complexity progression — single page to three-panel |
| [Technical](research-technical.html) | Architecture overview, polling mechanism, task scheduling |
| [Bugs](research-bugs.html) | Known bugs, interaction analysis, and root causes |
| [Python](research-python.html) | Python implementation — API reference, installation, development |
| [WASM Service Workers](research-wasm-service-workers.html) | Go/TinyGo WASM in service workers — routing, HTMX, static hosting |

## Overview

lofigui is for **single-process, small-audience tools**. The sweet spot: 1-10 users, one real object (a machine, a simulation, a long-running process) with a few pages showing different views of it.

Web applications sit on a spectrum from static sites to full SPAs. lofigui covers levels 2-4 (static+forms through HTMX partial updates). See [Philosophy](research-philosophy.html) for the full interactivity spectrum.

### Layout progression

| Layout | Built-in | Use case |
|--------|----------|----------|
| Single page | `LayoutSingle` | Quick tools, WASM apps |
| Header (fixed) | `LayoutNavbar` | Dashboards, monitoring |
| Three-panel | `LayoutThreePanel` | Multi-page apps, CRUD |

See [Layouts](research-layouts.html) for details and examples.
