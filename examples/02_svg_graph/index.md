<style>
.annotation { border-left: 3px solid #3273dc; background: #f0f4ff; padding: 0.75em 1em; margin: 0.75em 0; border-radius: 0 4px 4px 0; font-size: 0.9em; }
.annotation strong { color: #3273dc; }
.screenshot { border: 1px solid #dbdbdb; border-radius: 4px; box-shadow: 0 2px 6px rgba(0,0,0,0.1); overflow: hidden; }
.screenshot img { display: block; width: 100%; height: auto; }
.demo-output { background: #fff; border: 1px solid #dbdbdb; border-radius: 4px; padding: 1em; }
.demo-code pre { margin: 0 !important; }
</style>

# 02 — Output Showcase

A scrolling demonstration of every output type lofigui supports. The model runs in the background while the browser auto-refreshes to show new content appearing section by section.

**Interactivity level:** 1 — Teletype (same async polling pattern as example 01)

<div class="buttons">
<a href="wasm_demo/" class="button is-primary">Launch Demo</a>
<a target="_blank" href="https://codeberg.org/hum3/lofigui/src/branch/main/examples/02_svg_graph" class="button is-light">Source on Codeberg</a>
</div>

<div class="columns is-vcentered">
<div class="column is-5">
<figure class="image screenshot">
<img src="../02_polling.svg" alt="During polling — showcase partway through">
<figcaption class="has-text-centered has-text-grey is-size-7 mt-1">During polling</figcaption>
</figure>
</div>
<div class="column is-narrow has-text-centered">
<span style="font-size: 2rem; color: #999;">&rarr;</span>
</div>
<div class="column is-5">
<figure class="image screenshot">
<img src="../02_complete.svg" alt="After completion — every output type rendered">
<figcaption class="has-text-centered has-text-grey is-size-7 mt-1">Complete</figcaption>
</figure>
</div>
</div>

---

## Output Gallery

Each panel shows the Go code on the left and the rendered result on the right.

### Print() and Printf()

<div class="columns">
<div class="column is-6 demo-code">
<pre><code>lofigui.Print("Hello world.")
lofigui.Printf("π ≈ %.4f", math.Pi)</code></pre>
</div>
<div class="column is-6">
<div class="demo-output">
<p>Hello world.</p>
<p>π ≈ 3.1416</p>
</div>
</div>
</div>

### Print() with options

<div class="columns">
<div class="column is-6 demo-code">
<pre><code>lofigui.Print("Inline", lofigui.WithEnd(""))
lofigui.Print("text", lofigui.WithEnd(""))
lofigui.Print("joined", lofigui.WithEnd(""))</code></pre>
</div>
<div class="column is-6">
<div class="demo-output">
&nbsp;Inline&nbsp;&nbsp;text&nbsp;&nbsp;joined&nbsp;
</div>
</div>
</div>

<div class="columns">
<div class="column is-6 demo-code">
<pre><code>lofigui.Print("&lt;b&gt;Bold&lt;/b&gt; via &lt;code&gt;WithEscape(false)&lt;/code&gt;",
    lofigui.WithEscape(false))</code></pre>
</div>
<div class="column is-6">
<div class="demo-output">
<p><b>Bold</b> via <code>WithEscape(false)</code></p>
</div>
</div>
</div>

### Markdown()

<div class="columns">
<div class="column is-6 demo-code">
<pre><code>lofigui.Markdown("## Heading 2")
lofigui.Markdown("**Bold**, *italic*, ~~struck~~, `code`")
lofigui.Markdown("[A link](https://example.com)")</code></pre>
</div>
<div class="column is-6">
<div class="demo-output">
<h2>Heading 2</h2>
<p><strong>Bold</strong>, <em>italic</em>, <del>struck</del>, <code>code</code></p>
<p><a href="https://example.com">A link</a></p>
</div>
</div>
</div>

<div class="columns">
<div class="column is-6 demo-code">
<pre><code>lofigui.Markdown("- Alpha\n- Beta\n  - Nested")
lofigui.Markdown("1. First\n2. Second\n3. Third")</code></pre>
</div>
<div class="column is-6">
<div class="demo-output">
<ul><li>Alpha</li><li>Beta<ul><li>Nested</li></ul></li></ul>
<ol><li>First</li><li>Second</li><li>Third</li></ol>
</div>
</div>
</div>

<div class="columns">
<div class="column is-6 demo-code">
<pre><code>lofigui.Markdown("&gt; To be or not to be\n&gt;\n&gt; — *Shakespeare*")</code></pre>
</div>
<div class="column is-6">
<div class="demo-output">
<blockquote><p>To be or not to be</p><p>— <em>Shakespeare</em></p></blockquote>
</div>
</div>
</div>

### HTML() — inline elements

<div class="columns">
<div class="column is-6 demo-code">
<pre><code>lofigui.HTML(`&lt;p&gt;E = mc&lt;sup&gt;2&lt;/sup&gt;&lt;/p&gt;`)
lofigui.HTML(`&lt;p&gt;H&lt;sub&gt;2&lt;/sub&gt;O&lt;/p&gt;`)
lofigui.HTML(`&lt;p&gt;&lt;mark&gt;highlighted&lt;/mark&gt; text&lt;/p&gt;`)
lofigui.HTML(`&lt;p&gt;&lt;small&gt;small&lt;/small&gt;, &lt;ins&gt;ins&lt;/ins&gt;, &lt;del&gt;del&lt;/del&gt;&lt;/p&gt;`)</code></pre>
</div>
<div class="column is-6">
<div class="demo-output">
<p>E = mc<sup>2</sup></p>
<p>H<sub>2</sub>O</p>
<p><mark>highlighted</mark> text</p>
<p><small>small</small>, <ins>ins</ins>, <del>del</del></p>
</div>
</div>
</div>

### Table()

<div class="columns">
<div class="column is-6 demo-code">
<pre><code>lofigui.Table(
    [][]string{
        {"Go", "1.25", "Systems, web"},
        {"Python", "3.13", "Data science"},
        {"Rust", "1.86", "Embedded"},
    },
    lofigui.WithHeader([]string{
        "Language", "Version", "Use Cases",
    }),
)</code></pre>
</div>
<div class="column is-6">
<div class="demo-output">
<table class="table is-striped is-hoverable"><thead><tr><th>Language</th><th>Version</th><th>Use Cases</th></tr></thead><tbody><tr><td>Go</td><td>1.25</td><td>Systems, web</td></tr><tr><td>Python</td><td>3.13</td><td>Data science</td></tr><tr><td>Rust</td><td>1.86</td><td>Embedded</td></tr></tbody></table>
</div>
</div>
</div>

### HTML() — Bulma components

<div class="columns">
<div class="column is-6 demo-code">
<pre><code>lofigui.HTML(`&lt;div class="notification is-info is-light"&gt;
  An &lt;strong&gt;info&lt;/strong&gt; notification.
&lt;/div&gt;`)</code></pre>
</div>
<div class="column is-6">
<div class="demo-output">
<div class="notification is-info is-light">An <strong>info</strong> notification.</div>
</div>
</div>
</div>

<div class="columns">
<div class="column is-6 demo-code">
<pre><code>lofigui.HTML(`&lt;div class="tags"&gt;
  &lt;span class="tag is-primary"&gt;Primary&lt;/span&gt;
  &lt;span class="tag is-success"&gt;Success&lt;/span&gt;
  &lt;span class="tag is-warning"&gt;Warning&lt;/span&gt;
  &lt;span class="tag is-danger"&gt;Danger&lt;/span&gt;
&lt;/div&gt;`)</code></pre>
</div>
<div class="column is-6">
<div class="demo-output">
<div class="tags">
<span class="tag is-primary">Primary</span>
<span class="tag is-success">Success</span>
<span class="tag is-warning">Warning</span>
<span class="tag is-danger">Danger</span>
</div>
</div>
</div>
</div>

<div class="columns">
<div class="column is-6 demo-code">
<pre><code>lofigui.HTML(`&lt;progress class="progress is-primary"
  value="65" max="100"&gt;65%&lt;/progress&gt;`)</code></pre>
</div>
<div class="column is-6">
<div class="demo-output">
<progress class="progress is-primary" value="65" max="100">65%</progress>
</div>
</div>
</div>

### HTML() — HTML5 elements

<div class="columns">
<div class="column is-6 demo-code">
<pre><code>lofigui.HTML(`&lt;details&gt;
  &lt;summary&gt;&lt;strong&gt;Click to expand&lt;/strong&gt;&lt;/summary&gt;
  &lt;p&gt;Hidden content revealed on click.&lt;/p&gt;
&lt;/details&gt;`)</code></pre>
</div>
<div class="column is-6">
<div class="demo-output">
<details>
<summary><strong>Click to expand</strong></summary>
<p>Hidden content revealed on click.</p>
</details>
</div>
</div>
</div>

### HTML() — inline SVG

<div class="columns">
<div class="column is-6 demo-code">
<pre><code>lofigui.HTML(`&lt;svg width="220" height="60"&gt;
  &lt;rect x="5" y="5" width="50" height="50"
    rx="8" fill="#3298dc"/&gt;
  &lt;circle cx="90" cy="30" r="25"
    fill="#48c774"/&gt;
  &lt;polygon points="140,5 165,55 115,55"
    fill="#f14668"/&gt;
&lt;/svg&gt;`)</code></pre>
</div>
<div class="column is-6">
<div class="demo-output">
<svg width="220" height="60" xmlns="http://www.w3.org/2000/svg">
<rect x="5" y="5" width="50" height="50" rx="8" fill="#3298dc"/>
<circle cx="90" cy="30" r="25" fill="#48c774"/>
<polygon points="140,5 165,55 115,55" fill="#f14668"/>
</svg>
</div>
</div>
</div>

---

## The model — structure

The model delegates to section functions, each ending with `app.Sleep()`:

```go
func model(app *lofigui.App) {
    lofigui.Markdown("# Output Showcase")
    lofigui.Print("Walking through every output type lofigui supports.")
    app.Sleep(1 * time.Second)

    sectionHeadings(app)
    sectionTextFormatting(app)
    sectionPrintVariations(app)
    // ... 8 more sections ...

    lofigui.Print("Showcase complete.")
}
```

<div class="annotation">
<strong>Scrolling output</strong> — because the buffer is append-only, each refresh shows everything printed so far. The user sees the page grow as new sections appear — like a terminal session rendered as styled HTML.
</div>

[model.go source on Codeberg](https://codeberg.org/hum3/lofigui/src/branch/main/examples/02_svg_graph/go/model.go)

---

## SVG charts — hand-rolled mocks vs a real library

`charts.go` ships two hand-rolled SVG helpers — bar and pie — so the example has zero charting dependencies:

```go
lofigui.HTML(barChartSVG(values, labels, "Title"))
lofigui.HTML(pieChartSVG(slices, "Title"))
```

<div class="annotation">
<strong>These are demo mocks.</strong> They illustrate that any SVG string can be streamed into the buffer via <code>lofigui.HTML()</code>, but they skip the things a real charting library handles — proper axes, tick spacing, legends, theming, time-series scales. For production use, reach for <a href="https://codeberg.org/hum3/gogal">gogal</a>, which the live line chart in <code>sectionStaticCharts</code> already uses:
</div>

```go
lineChart := gogal.NewLineChart(
    gogal.WithTitle("Growth Trend"),
    gogal.WithSize(400, 150),
    gogal.WithGrid(true),
    gogal.WithSmooth(true),
)
lineChart.AddXY("Growth", xValues, values)
svg, _ := lineChart.RenderString()
lofigui.HTML(svg)
```

The live-updating chart prints 5 versions of a growing bar chart, one per second:

```go
for _, n := range []int{2, 4, 6, 8, 10} {
    lofigui.HTML(barChartSVG(fib[:n], labels[:n],
        fmt.Sprintf("Fibonacci (n=%d)", n)))
    app.Sleep(1 * time.Second)
}
```

<div class="annotation">
<strong>Updating by re-printing</strong> — each iteration appends a new chart below the previous one. The user sees the chart "grow" through the sequence. No DOM manipulation, no JavaScript — just print and refresh.
</div>

[charts.go source on Codeberg](https://codeberg.org/hum3/lofigui/src/branch/main/examples/02_svg_graph/go/charts.go)

---

## The server

Identical to example 01 — `Run` handles everything:

```go
func main() {
    app := lofigui.NewApp()
    app.Version = "Output Showcase v1.0"
    app.Run(":1340", model)
}
```

[main.go source on Codeberg](https://codeberg.org/hum3/lofigui/src/branch/main/examples/02_svg_graph/go/main.go)

---

## Running

```bash
task go-example:02        # Run the server
# Open http://localhost:1340 — output scrolls for ~16 seconds
```

---

## WASM

The same model compiles to WebAssembly with no changes — and, like example 01, no on-disk templates. `app.RunWASM()` uses lofigui's built-in default WASM template (served via service worker):

```go
//go:build js && wasm

package main

import "codeberg.org/hum3/lofigui"

func main() {
    app := lofigui.NewApp()
    app.Version = "Output Showcase v1.0"
    app.RunWASM(model)
}
```

<div class="buttons">
<a href="wasm_demo/" class="button is-primary is-small">Go WASM Demo</a>
</div>

[main_wasm.go source on Codeberg](https://codeberg.org/hum3/lofigui/src/branch/main/examples/02_svg_graph/go/main_wasm.go)
