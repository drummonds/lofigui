<style>
.annotation { border-left: 3px solid #3273dc; background: #f0f4ff; padding: 0.75em 1em; margin: 0.75em 0; border-radius: 0 4px 4px 0; font-size: 0.9em; }
.annotation strong { color: #3273dc; }
.screenshot { border: 1px solid #dbdbdb; border-radius: 4px; box-shadow: 0 2px 6px rgba(0,0,0,0.1); overflow: hidden; }
.screenshot img { display: block; width: 100%; height: auto; }
.python-only { border-left: 4px solid #3776ab; background: #eef6fc; padding: 1em 1.25em; margin: 1em 0; border-radius: 0 4px 4px 0; }
.python-only strong { color: #3776ab; }
</style>

# 05 — Demo App

Multi-page Python application built around **Jinja2 template inheritance**. Every page extends a single `base.html` that owns the navbar, layout columns, and status panel; child templates only fill in the page-specific blocks. The example also wires the lofigui `Controller` to a long-running background process so the page auto-refreshes while work is in flight.

<div class="python-only">
<strong>Python-only example.</strong> There is no Go or WASM build of 05 — Jinja2's <code>{% extends %}</code> / <code>{% block %}</code> inheritance is the lesson, and the equivalent Go-template story (using <code>html/template</code>'s <code>{{block}}</code> / <code>{{define}}</code>) lives in <a href="../03_style_sampler/">example 03</a>. If you came here looking for a Go starting point, jump to 03 instead.
</div>

**[Interactivity level](../research-philosophy.html#the-interactivity-spectrum):** 3 — Polling (whole-page refresh while a process runs)
**[State scope](../research-philosophy.html#the-state-dimension):** Global (single FastAPI process — every browser sees the same accumulated buffer)

<div class="buttons">
<a target="_blank" href="https://codeberg.org/hum3/lofigui/src/branch/main/examples/05_demo_app" class="button is-light">Source on Codeberg</a>
</div>

<div class="columns is-vcentered">
<div class="column is-6">
<figure class="image screenshot">
<img src="../05_home.svg" alt="Home page — feature cards in main column, status panel on the right">
<figcaption class="has-text-centered has-text-grey is-size-7 mt-1">Home — <code>home.html</code> fills the <code>mainpanel</code> block with feature cards</figcaption>
</figure>
</div>
<div class="column is-6">
<figure class="image screenshot">
<img src="../05_data.svg" alt="Data Tables page — same navbar, table content in main column">
<figcaption class="has-text-centered has-text-grey is-size-7 mt-1">Data Tables — <code>data.html</code> fills the same block with rendered tables</figcaption>
</figure>
</div>
</div>

<p class="has-text-grey is-size-7 has-text-centered">Same <code>base.html</code> on both pages — only the <code>mainpanel</code> block changes. The navbar, columns, status panel, and footer are inherited unchanged.</p>

---

## How lofigui helps here

Five things from the library do real work in this example:

1. **`lg.markdown` / `lg.table` / `lg.print` accumulate HTML** in the lofigui buffer; routes flush the buffer into a Jinja2 context variable (`table_html`, `chart_content`, `process_output`) and render the template.
2. **`App.template_response(request, name, extra)`** locates `templates/<name>` (Jinja2 `FileSystemLoader`), injects lofigui state (`refresh`, `polling`, `version`, `name`), merges `extra`, and returns an `HTMLResponse` — the only call needed in each route handler.
3. **`Controller.start_action(refresh_time)` / `end_action()`** drive the polling lifecycle. While the action is running, `template_response` emits a `<meta http-equiv="Refresh" ...>` tag into the rendered page, so the browser reloads on a timer.
4. **Built-in `/assets/bulma.min.css`** is served by `App` next to your routes — `base.html` links to it directly with no CDN round-trip.
5. **`lg.get_favicon_response()`** returns a ready-made favicon so the navbar's logo slot doesn't 404.

Everything else — Jinja2, FastAPI, Bulma — is plain ecosystem code.

---

## Template inheritance — the point of this example

`base.html` declares every layout element exactly once and exposes named blocks for child templates to override:

```html
<!-- base.html (abridged) -->
<head>
  {% block title %}<title>Lofigui Demo App</title>{% endblock %}
  {{refresh | safe}}
  <link rel="stylesheet" href="/assets/bulma.min.css"/>
  <style>{% block extra_css %}{% endblock %}</style>
</head>
<body>
  <nav class="navbar is-primary is-fixed-top">…shared navbar…</nav>

  <div class="main-content"><div class="container">
    {% block content %}
      <div class="columns">
        <div class="column is-8"><div class="box">
          {% block mainpanel %}<div id="content">Main content</div>{% endblock %}
        </div></div>
        <div class="column is-4"><div class="box">
          {% block statuspanel %}
            <p><strong>Status:</strong> {{status | default('Ready')}}</p>
            <p><strong>Polling:</strong> {{polling | default('Stopped')}}</p>
          {% endblock %}
        </div></div>
      </div>
    {% endblock %}
  </div></div>

  <footer class="footer">…shared footer…</footer>
</body>
```

Child pages declare only what they change. `home.html` extends the base and overrides just the title and the main-column content:

```html
<!-- home.html -->
{% extends "base.html" %}

{% block title %}<title>Home - Lofigui Demo App</title>{% endblock %}

{% block mainpanel %}
  <h1 class="title">Welcome to Lofigui Demo</h1>
  <div class="columns is-multiline">
    <div class="column is-6"><div class="box">…feature card…</div></div>
    …
  </div>
{% endblock %}
```

`data.html` overrides the **same `mainpanel` block** with table HTML produced by `lg.table` — the navbar and status panel come through unchanged from `base.html`:

```html
<!-- data.html -->
{% extends "base.html" %}
{% block title %}<title>Data Tables - Lofigui Demo App</title>{% endblock %}
{% block mainpanel %}
  <h2 class="title">Data Tables Demo</h2>
  <div class="content">{{table_html | safe}}</div>
  <hr>
  <div class="content">{{comparison_table | safe}}</div>
{% endblock %}
```

<div class="annotation">
<strong>Block override granularity.</strong> Each child template overrides only the blocks it cares about. The navbar lives in <code>base.html</code> and is inherited by every page; adding a new nav item is a one-file change. The two-column <code>content</code> block, the <code>statuspanel</code>, the footer, and the burger-menu JS are all shared the same way. Compare this to copying a navbar into seven separate templates.
</div>

### Available blocks

| Block | Where | Purpose |
|-------|-------|---------|
| `title` | `<head>` | Per-page `<title>` |
| `extra_css` | `<head>` | Per-page CSS additions |
| `extra_js` | `<script>` (end of body) | Per-page JS additions |
| `content` | Main area | Replace the entire two-column layout (rare) |
| `mainpanel` | Left column | Page-specific content (most common) |
| `statuspanel` | Right column | Override the default Status / Polling display |

---

## The routes — turning lofigui buffers into template context

Each route resets the lofigui buffer, prints page-specific output, snapshots the buffer with `lg.buffer()`, and passes that snapshot as a context variable to `template_response`:

```python
import lofigui as lf
from lofigui import App

app_instance = App(template_dir="templates")
app_instance.controller = Controller()
fastapi_app = FastAPI(title="Lofigui Demo")

@fastapi_app.get("/data", response_class=HTMLResponse)
async def data_tables(request: Request):
    lf.reset()
    lf.markdown("### Employee Data")
    lf.table(employee_data, header=["Name", "Role", "Skills", "Experience"])
    table_html = lf.buffer()                 # snapshot one section

    lf.reset()
    lf.markdown("### Framework Comparison")
    lf.table(comparison_data, header=[…])
    comparison_table = lf.buffer()           # snapshot the second

    return app_instance.template_response(request, "data.html", {
        "table_html": table_html,
        "comparison_table": comparison_table,
        "status": "Ready",
        "polling": "Stopped",
    })
```

<div class="annotation">
<strong>Why <code>reset</code> + <code>buffer</code> twice.</strong> The lofigui buffer is process-global. Each <code>reset()</code> clears it; each <code>buffer()</code> returns a snapshot of what has been printed since. Pulling two snapshots in a single handler lets a single template render two independent regions (the employee table and the comparison table) without merging them. The same idea generalises to any number of named regions.
</div>

---

## Process management — `Controller.start_action` / `end_action`

The `/start_demo_process` route runs a long simulated job and shows the auto-refresh polling pattern in action:

```python
@fastapi_app.post("/start_demo_process")
async def start_demo_process(duration: int = Form(10)):
    app_instance.start_action(refresh_time=2)        # turn on auto-refresh
    lf.reset()
    lf.markdown("## Demo Process Started")
    duration = max(1, min(duration, 60))
    for i in range(duration):
        lf.print(f"Step {i + 1}/{duration} - Progress: {int((i + 1) / duration * 100)}%")
        await asyncio.sleep(1)
    lf.markdown("### Process Complete!")
    app_instance.end_action()                        # turn off auto-refresh
    return RedirectResponse(url="/display", status_code=303)
```

While the action is running, `template_response` injects a `<meta http-equiv="Refresh" content="2; URL=…">` tag into every rendered page, so the browser reloads every two seconds and re-reads the live `lf.buffer()`. Once `end_action()` returns, the meta tag is omitted and the page stops reloading.

<div class="annotation">
<strong>Cancel from anywhere.</strong> The shared navbar carries a Cancel button that POSTs to <code>/stop</code>, which calls <code>end_action()</code>. Because every page extends <code>base.html</code>, the Cancel button is present on every screen — the user can interrupt the process from the Home, Data, or Charts page without losing context.
</div>

---

## Routes at a glance

| Method + Path | Template | Purpose |
|---------------|----------|---------|
| `GET /` | `home.html` | Feature overview, links into the rest of the demo |
| `GET /data` | `data.html` | Two `lg.table` snapshots rendered side-by-side |
| `GET /charts` | `charts.html` | Markdown + code blocks demonstrating chart integration points |
| `GET /process` | `process.html` | Form + status; redirects to `/display` when running |
| `POST /start_demo_process` | (redirect) | Kicks off the simulated job |
| `POST /stop` | (redirect) | Cancels a running job |
| `GET /display` | `display.html` | Renders the final `lf.buffer()` after the job completes |
| `GET /about` | `about.html` | Static info page (still inherits the shared layout) |

---

## Running it

```bash
task example-05            # or: task py05, task demo
# Server listens on http://localhost:8050
```

The `example-05` task lives in `Taskfile.yml` and runs:

```bash
cd examples/05_demo_app/python
uv sync --no-install-project
uv run --no-project python demo_app.py
```

`pyproject.toml` pins lofigui to a relative editable install (`../../../`) so the demo always uses the in-tree library.

---

## Where to go next

- **The same inheritance idea in Go** — [03 Style Sampler](../03_style_sampler/) does five layouts on top of one `base.html` using Go's `html/template` `{{block}}` / `{{define}}`. Same idea, different engine, plus a WASM build.
- **A simpler polling example** — [01 Hello World](../01_hello_world/) is the minimum: one route, one template, one model.
- **Forms / CRUD** — [06 Notes CRUD](../06_notes_crud/) shows the redirect-after-POST pattern with no polling at all.

---

## Links

- [Source on Codeberg](https://codeberg.org/hum3/lofigui/src/branch/main/examples/05_demo_app)
