package main

import (
	"fmt"
	"math"
	"strings"
	"time"

	"codeberg.org/hum3/gogal"
	"codeberg.org/hum3/lofigui"
)

// model demonstrates every output type lofigui supports.
// It runs as a scrolling presentation — each section appears after a pause.
func model(app *lofigui.App) {
	lofigui.Markdown("# Output Showcase")
	lofigui.Print("Walking through every output type lofigui supports.")
	app.Sleep(1 * time.Second)

	sectionHeadings(app)
	sectionTextFormatting(app)
	sectionPrintVariations(app)
	sectionLists(app)
	sectionBlockquotesAndCode(app)
	sectionTables(app)
	sectionBulmaComponents(app)
	sectionHTML5Elements(app)
	sectionSVGShapes(app)
	sectionStaticCharts(app)
	sectionLiveChart(app)

	lofigui.Markdown("---")
	lofigui.Print("Showcase complete.")
}

func sectionHeadings(app *lofigui.App) {
	lofigui.Markdown("---\n## Heading Levels")
	for i := 1; i <= 6; i++ {
		lofigui.Markdown(fmt.Sprintf("%s Heading %d", strings.Repeat("#", i), i))
	}
	app.Sleep(1 * time.Second)
}

func sectionTextFormatting(app *lofigui.App) {
	lofigui.Markdown("---\n## Text Formatting")
	lofigui.Markdown("**Bold**, *italic*, ~~strikethrough~~, and `inline code`.")
	lofigui.Markdown("[A hyperlink](https://example.com) rendered via Markdown.")

	lofigui.Markdown("### Superscript & Subscript")
	lofigui.HTML(`<p>Einstein: E = mc<sup>2</sup></p>`)
	lofigui.HTML(`<p>Water: H<sub>2</sub>O</p>`)
	lofigui.HTML(`<p>Quadratic: ax<sup>2</sup> + bx + c = 0</p>`)

	lofigui.Markdown("### Other Inline HTML")
	lofigui.HTML(`<p>The <mark>highlighted text</mark> stands out.</p>`)
	lofigui.HTML(`<p><abbr title="HyperText Markup Language">HTML</abbr> is the language of the web.</p>`)
	lofigui.HTML(`<p><small>Small text</small>, <ins>inserted</ins>, and <del>deleted</del>.</p>`)
	app.Sleep(1 * time.Second)
}

func sectionPrintVariations(app *lofigui.App) {
	lofigui.Markdown("---\n## Print Variations")
	lofigui.Print("Standard paragraph via Print()")
	lofigui.Printf("Formatted via Printf(): π ≈ %.6f", math.Pi)
	lofigui.Print("Inline", lofigui.WithEnd(""))
	lofigui.Print("text", lofigui.WithEnd(""))
	lofigui.Print("joined with WithEnd(\"\")", lofigui.WithEnd(""))
	lofigui.HTML("<br>") // line break after inline sequence
	lofigui.Print("<b>Raw HTML</b> via <code>WithEscape(false)</code>", lofigui.WithEscape(false))
	app.Sleep(1 * time.Second)
}

func sectionLists(app *lofigui.App) {
	lofigui.Markdown("---\n## Lists")
	lofigui.Markdown("Unordered:\n\n- Alpha\n- Beta\n- Gamma\n  - Nested item\n  - Another nested")
	lofigui.Markdown("Ordered:\n\n1. First\n2. Second\n3. Third")
	app.Sleep(1 * time.Second)
}

func sectionBlockquotesAndCode(app *lofigui.App) {
	lofigui.Markdown("---\n## Blockquotes & Code Blocks")
	lofigui.Markdown("> To be or not to be, that is the question.\n>\n> — *William Shakespeare*")
	lofigui.Markdown("```go\nfunc fibonacci(n int) int {\n    if n <= 1 {\n        return n\n    }\n    return fibonacci(n-1) + fibonacci(n-2)\n}\n```")
	app.Sleep(1 * time.Second)
}

func sectionTables(app *lofigui.App) {
	lofigui.Markdown("---\n## Tables")
	lofigui.Table(
		[][]string{
			{"Go", "1.25", "Systems, web, CLI"},
			{"Python", "3.13", "Data science, scripting"},
			{"Rust", "1.86", "Systems, embedded"},
			{"JavaScript", "ES2024", "Web, full-stack"},
		},
		lofigui.WithHeader([]string{"Language", "Version", "Primary Use Cases"}),
	)
	app.Sleep(1 * time.Second)
}

func sectionBulmaComponents(app *lofigui.App) {
	lofigui.Markdown("---\n## Bulma Components")

	// Notifications
	lofigui.HTML(`<div class="notification is-info is-light">This is an <strong>info</strong> notification.</div>`)
	lofigui.HTML(`<div class="notification is-success is-light">This is a <strong>success</strong> notification.</div>`)
	lofigui.HTML(`<div class="notification is-warning is-light">This is a <strong>warning</strong> notification.</div>`)
	lofigui.HTML(`<div class="notification is-danger is-light">This is a <strong>danger</strong> notification.</div>`)

	// Tags
	lofigui.HTML(`<div class="tags">
  <span class="tag is-primary">Primary</span>
  <span class="tag is-link">Link</span>
  <span class="tag is-info">Info</span>
  <span class="tag is-success">Success</span>
  <span class="tag is-warning">Warning</span>
  <span class="tag is-danger">Danger</span>
  <span class="tag is-dark">Dark</span>
</div>`)

	// Progress bars
	lofigui.Markdown("### Progress Bars")
	lofigui.HTML(`<progress class="progress is-primary" value="25" max="100">25%</progress>`)
	lofigui.HTML(`<progress class="progress is-info" value="50" max="100">50%</progress>`)
	lofigui.HTML(`<progress class="progress is-success" value="75" max="100">75%</progress>`)

	// Box
	lofigui.HTML(`<div class="box"><p>Content inside a <strong>Bulma box</strong> — a simple container with shadow and padding.</p></div>`)
	app.Sleep(1 * time.Second)
}

func sectionHTML5Elements(app *lofigui.App) {
	lofigui.Markdown("---\n## HTML5 Elements")

	// Details/Summary
	lofigui.HTML(`<details>
  <summary><strong>Collapsible section</strong> — click to expand</summary>
  <div class="content ml-4 mt-2">
    <p>Hidden content revealed on click.</p>
    <p>Uses the HTML5 <code>&lt;details&gt;</code> and <code>&lt;summary&gt;</code> elements.</p>
  </div>
</details>`)

	// Definition list
	lofigui.Markdown("### Definition List")
	lofigui.HTML(`<dl>
  <dt><strong>lofigui</strong></dt>
  <dd>A lightweight web-UI framework with a print-like interface.</dd>
  <dt><strong>Bulma</strong></dt>
  <dd>A modern CSS framework based on Flexbox.</dd>
  <dt><strong>pongo2</strong></dt>
  <dd>A Django-syntax template engine for Go.</dd>
</dl>`)

	// Horizontal rule demonstration
	lofigui.Markdown("### Horizontal Rule")
	lofigui.Print("Above the rule.")
	lofigui.Markdown("---")
	lofigui.Print("Below the rule.")
	app.Sleep(1 * time.Second)
}

func sectionSVGShapes(app *lofigui.App) {
	lofigui.Markdown("---\n## Inline SVG Shapes")
	lofigui.HTML(`<svg width="420" height="100" xmlns="http://www.w3.org/2000/svg">
  <rect x="10" y="10" width="80" height="80" rx="10" fill="#3298dc"/>
  <circle cx="150" cy="50" r="40" fill="#48c774"/>
  <ellipse cx="250" cy="50" rx="50" ry="30" fill="#ffdd57"/>
  <polygon points="350,10 390,90 310,90" fill="#f14668"/>
</svg>`)
	app.Sleep(1 * time.Second)
}

func sectionStaticCharts(app *lofigui.App) {
	lofigui.Markdown("---\n## SVG Charts")

	lofigui.HTML(pieChartSVG(
		[]pieSlice{
			{"Go", 35, "#3298dc"},
			{"Python", 28, "#48c774"},
			{"Rust", 18, "#f14668"},
			{"JS", 12, "#ffdd57"},
			{"Other", 7, "#b86bff"},
		},
		"Language Popularity",
	))

	// Line chart via gogal (replaces hand-built sparkline)
	values := []float64{3, 7, 5, 12, 8, 15, 11, 20, 17, 25, 22, 30, 28, 35, 32, 40}
	xValues := make([]float64, len(values))
	for i := range values {
		xValues[i] = float64(i + 1)
	}
	lineChart := gogal.NewLineChart(
		gogal.WithTitle("Growth Trend"),
		gogal.WithSize(400, 150),
		gogal.WithGrid(true),
		gogal.WithSmooth(true),
	)
	lineChart.AddXY("Growth", xValues, values)
	svg, _ := lineChart.RenderString()
	lofigui.HTML(svg)
	app.Sleep(1 * time.Second)
}

func sectionLiveChart(app *lofigui.App) {
	lofigui.Markdown("---\n## Live Updating Chart")
	lofigui.Print("Fibonacci bar chart — grows one bar per second:")

	fib := []float64{1, 1, 2, 3, 5, 8, 13, 21, 34, 55}
	labels := []string{"F1", "F2", "F3", "F4", "F5", "F6", "F7", "F8", "F9", "F10"}

	// Show 5 stages: 2, 4, 6, 8, 10 bars
	for _, n := range []int{2, 4, 6, 8, 10} {
		lofigui.HTML(barChartSVG(fib[:n], labels[:n],
			fmt.Sprintf("Fibonacci (n=%d)", n)))
		app.Sleep(1 * time.Second)
	}
}
