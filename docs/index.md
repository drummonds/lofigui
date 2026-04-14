# lofigui — Documentation

<div class="columns">
<div class="column is-3">

<aside class="menu">
<p class="menu-label">Pages</p>
<ul class="menu-list">
<li><a href="README.html">lofigui README</a></li>
<li><a href="ROADMAP.html">Roadmap</a></li>
<li>
<a href="RESEARCH.html">Research</a>
<ul>
<li><a href="research-philosophy.html">Philosophy</a></li>
<li><a href="research-charts.html">Charts</a></li>
<li><a href="research-layouts.html">Page Layouts</a></li>
<li><a href="research-technical.html">Technical</a></li>
<li><a href="research-bugs.html">Bugs</a></li>
<li><a href="research-python.html">Python</a></li>
<li><a href="research-wasm-service-workers.html">WASM Service Workers</a></li>
</ul>
</li>
<li><a href="CHANGELOG.html">Changelog</a></li>
<li><a href="recipes.html">Recipes</a></li>
<li><a href="examples.html">Examples Guide</a></li>
</ul>
</aside>

</div>
<div class="column is-9">

A minimalist Go library for creating simple web-based GUIs. Some parts available as Python.

## Examples

| # | Name | Description |
|---|------|-------------|
| [01](01_hello_world/) | Hello World | Async with polling |
| [01a](01a_hello_world_explicit/) | Hello World Explicit | Explicit handlers + service worker WASM |
| [02](02_svg_graph/) | Output Showcase | All output types: Print, Markdown, HTML, Table, SVG |
| [03](03_style_sampler/) | Style Sampler | WASM with template inheritance |
| [06](06_notes.svg) | Notes CRUD | Form POST handlers |
| [07](07_water_tank/) | Water Tank | SVG dashboard with simulation |
| [08](08_water_tank_multi/) | Water Tank Multi | Multi-page with HTTP Refresh |
| [09](09_water_tank_htmx/) | Water Tank HTMX | Partial updates with HTMX |
| [10](10_water_tank_maintenance/) | Water Tank Maintenance | Background ops with progress |
| [11](11_water_tank_storage/) | Water Tank Storage | WASM + API server with persistent state |
| [12](12_batch_yield/) | Batch Yield | Cooperative scheduling with `lofigui.Yield()` |

## Links

- [pkg.go.dev](https://pkg.go.dev/codeberg.org/hum3/lofigui)

<!-- auto:links -->
<!-- /auto:links -->

</div>
</div>
