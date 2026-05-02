# 05 — Demo App (Python only)

Multi-page Python application demonstrating **Jinja2 template inheritance** with lofigui. Every page extends a single `base.html` (navbar, two-column layout, status panel, footer) and overrides only the blocks it needs. Also wires the lofigui `Controller` to a long-running process so the page auto-refreshes while work is in flight.

> **Python-only.** No Go or WASM build for this example — Jinja2 inheritance is the lesson. The Go-template equivalent (`html/template` `{{block}}`/`{{define}}`) lives in [example 03 Style Sampler](../03_style_sampler/), which also has a WASM build.

See [`examples/05_demo_app/index.md`](./index.md) for the rendered documentation page (screenshots, annotated walkthrough, block reference).

## Quick start

```bash
task example-05            # aliases: task py05, task demo
# Server: http://localhost:8050
```

That runs (in `examples/05_demo_app/python/`):

```bash
uv sync --no-install-project
uv run --no-project python demo_app.py
```

`pyproject.toml` pins lofigui to a relative editable install (`../../../`) so the demo always uses the in-tree library.

## Pages

| Path | Template | What it shows |
|------|----------|---------------|
| `/` | `home.html` | Feature cards, intro text |
| `/data` | `data.html` | Two `lg.table` snapshots side-by-side |
| `/charts` | `charts.html` | Markdown + code samples for chart integration |
| `/process` | `process.html` | Form to start a long-running demo job |
| `/display` | `display.html` | Final `lg.buffer()` after the job completes |
| `/about` | `about.html` | Static info |

## Project structure

```
05_demo_app/
├── README.md                # This file
├── index.md                 # Rendered docs (screenshots, walkthrough)
└── python/
    ├── demo_app.py          # FastAPI app
    ├── pyproject.toml       # uv-managed deps; lofigui pinned to ../../../
    ├── uv.lock
    └── templates/
        ├── base.html        # Owns navbar, columns, status panel, footer
        ├── home.html        # extends base
        ├── data.html        # extends base
        ├── charts.html      # extends base
        ├── process.html     # extends base
        ├── display.html     # extends base
        └── about.html       # extends base
```

## License

Part of the lofigui project; shares the same license.
