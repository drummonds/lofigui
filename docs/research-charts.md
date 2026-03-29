# Research: Charts

Charting options for lofigui — Go SVG vs JavaScript libraries.

## The problem

Charts are the next big gap in lofigui. The current approach (example 02) uses Go libraries to produce SVG server-side. The results are functional but poor — limited axis formatting, weak label placement, no interactivity, and mediocre visual quality compared to what users expect from modern dashboards.

This was explored in the [gobank chart comparison](https://drummonds.github.io/gobank/research/chart-comparison.html), which tested three Go SVG renderers against financial data.

## Go SVG libraries tested (gobank)

| Library               | Strengths                               | Weaknesses                                                                                                      |
| --------------------- | --------------------------------------- | --------------------------------------------------------------------------------------------------------------- |
| **Hand-rolled SVG**   | Full control, no dependencies           | Enormous effort for basic features (axis scaling, labels, legends). Every chart type is a fresh implementation. |
| **go-analyze/charts** | Good chart variety, reasonable defaults | Axis formatting issues (decimal values where percentages expected), limited customisation                       |
| **margaid**           | Clean line charts                       | Limited chart types, sparse documentation                                                                       |

The go-chart library (used in lofigui example 02) has similar issues — it works for simple bar/line charts but struggles with axis formatting, date handling, and multi-series layouts.

**Core issue**: Go's charting ecosystem is immature compared to JavaScript's. The libraries exist but produce output that looks 10 years behind what a JS library produces with the same data and less code.

## The JavaScript charting landscape

If we accept a JS dependency for charts (as we accepted HTMX for interactivity), the question is: which library, and does it drag in framework complexity?

| Library             | Size             | Framework needed? | CDN single-file? | SVG output?      | Notes                                                                                              |
| ------------------- | ---------------- | ----------------- | ---------------- | ---------------- | -------------------------------------------------------------------------------------------------- |
| **Chart.js**        | ~65KB            | No                | Yes              | Canvas (not SVG) | Simple API, good defaults, huge community. Canvas means no CSS styling of chart elements.          |
| **D3.js**           | ~90KB            | No                | Yes              | Yes (native SVG) | The gold standard. Total control over every pixel. Steep learning curve but unmatched flexibility. |
| **Observable Plot** | ~50KB (needs D3) | No                | Yes              | Yes (SVG)        | D3's "high-level" layer. Concise API, good defaults, less boilerplate than raw D3.                 |
| **Plotly.js**       | ~1MB             | No                | Yes              | SVG + Canvas     | Feature-rich but heavy. Built on D3. Good for scientific/financial charts.                         |
| **Apache ECharts**  | ~400KB           | No                | Yes              | Canvas + SVG     | Rich interactive charts. Heavy but well-documented.                                                |
| **Frappe Charts**   | ~17KB            | No                | Yes              | SVG              | Lightweight, GitHub-inspired aesthetics. Limited chart types.                                      |
| **uPlot**           | ~35KB            | No                | Yes              | Canvas           | Extremely fast for time-series. Minimal but performant.                                            |
| **Vega-Lite**       | ~400KB           | No                | Yes              | SVG + Canvas     | Declarative grammar-of-graphics. JSON spec, no imperative code.                                    |

## What to avoid

The key constraint is: **no React, no Svelte, no Angular, no build step**. Libraries that require a framework or a bundler are out:

- **Recharts** — React-only
- **Victory** — React-only
- **Nivo** — React-only
- **SvelteKit charts** — Svelte-only

The lofigui pattern is: server renders HTML, browser displays it. A chart library must work with a `<script>` tag and a `<div>` target, nothing more.

## Leading candidates for lofigui

**D3.js** is the most interesting option. It:

- Produces native SVG (inspectable, stylable with CSS, printable)
- Works from a single CDN `<script>` tag
- Has no framework dependency
- Is the foundation most other libraries build on
- Supports every chart type imaginable

The downside is D3's verbosity — a simple line chart is ~30 lines of JS. But lofigui could generate the D3 JavaScript server-side (Go `fmt.Sprintf` with data injected into a template), keeping the complexity on the server while the browser just executes the rendering.

**Observable Plot** is D3's high-level API, worth considering if D3 feels too low-level. Same SVG output, much less code.

**Chart.js** is the simplest option if SVG isn't required. Canvas output means less flexibility but the API is very approachable.

## How it would work in lofigui

The pattern would mirror HTMX — a `<script>` tag in the template, with the Go server injecting data:

```go
// Server-side: Go generates the chart div + script
lofigui.HTML(fmt.Sprintf(`
<div id="chart-%s"></div>
<script>
  // D3 or Chart.js code here, with data from Go
  const data = %s;
  // ... render chart into #chart-%s
</script>
`, chartID, jsonData, chartID))
```

This keeps the Go code in control of _what_ to chart. The JS library handles _how_ to render it. No npm, no build step, no framework.

## Potential example: water tank charts

A natural next example would chart the water tank simulation data — tank level and pump/valve activity over time. This would:

- Demonstrate time-series charting with real simulation data
- Show how to pass Go data to a JS charting library
- Build on the existing water tank examples (07-10)
- Provide a reference pattern for charting in lofigui apps

The chart would update via HTMX (like example 09) — the fragment endpoint returns fresh chart HTML with updated data on each poll.
