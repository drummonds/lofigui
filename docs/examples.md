# Examples — Structure and Standards

How examples are organised, documented, and built. This document serves both as user-facing documentation and as a reference for generating or refreshing example pages.

## Interactivity spectrum

Examples map to the [interactivity spectrum](research-philosophy.html#the-interactivity-spectrum) from the philosophy document. Each example lives at a specific level:

| Level | Approach | Examples |
|-------|----------|----------|
| 1 | Static | (docs site itself) |
| 2 | Scrolling / printing output | 02 (Output Showcase) |
| 3 | Polling (whole page refresh) | 01 (Hello World), 07 (Water Tank), 08 (Multi-Page) |
| 4 | Static + forms (CRUD) | 06 (Notes CRUD) |
| 5 | HTMX partial updates | 09 (Water Tank HTMX), 10 (Maintenance), 12 (Batch Yield) |
| 6 | WASM (browser-only) | 03 (Style Sampler) |
| 7 | WASM + API server | 11 (Water Tank Storage) |

The progression is deliberate — each level adds one concept. Users stop at the level of complexity their project needs.

## File structure per example

Each example has two locations: source code in `examples/` and rendered documentation in `docs/`.

```
examples/NN_name/
  go/                     # Go source code
    main.go               # Server entry point
    model.go              # Business logic (shared with WASM)
    main_wasm.go          # WASM entry point (if applicable)
    go.mod
  python/                 # Python source (where applicable)
  templates/
    index.html            # WASM demo page (standalone, clean)
    app.js                # WASM JavaScript glue
    hello.html            # Server-side template (if applicable)
  index.md                 # Documentation source (markdown)
  README.md               # Brief description for browsing on Codeberg

docs/NN_name/
  index.html              # Rendered from index.md — NEVER overwritten by build
  demo.html               # Copied from templates/index.html (WASM demo)
  app.js                  # Copied from templates/app.js
  main.wasm               # Compiled WASM binary
  wasm_exec.js            # Go WASM runtime
  *.svg                   # Screenshots captured via url2svg
```

Key rule: `docs:build-wasm` copies WASM assets and `demo.html` but **never overwrites `index.html`**. The documentation page and the demo page are separate concerns.

## Documentation page standard (index.md)

Each example's `index.md` is the single source of truth for its documentation. It is rendered to `docs/NN_name/index.html` by `task-plus md2html`. The rendered page uses shared CSS for consistent styling.

### Required sections

Every example page follows this structure:

```
# NN — Example Name

One-paragraph introduction. What does this example demonstrate?
Why would you start from this example?

[Screenshots: polling state + complete state]
[Buttons: Launch Demo | Source on Codeberg]

---

## The model — your application logic

Annotated walkthrough of model.go. Explain what each Print/Sleep/EndAction
does and why. Include the code inline with annotations below it.

---

## The server — wiring it up

Annotated walkthrough of main.go. Show how the model is connected to HTTP.

---

## How it works

Brief explanation of the request flow. Link to technical.html for internals
if the example has one.

---

## WASM: running in the browser (if applicable)

How the same model runs in WebAssembly. Show main_wasm.go.
Build command: GOOS=js GOARCH=wasm go build -o main.wasm .

---

[Footer: Back to docs | Source on Codeberg | pkg.go.dev]
```

### Styling conventions

The documentation page uses these custom elements (defined in shared CSS):

**Annotations** — blue left-border callout boxes for explanations:

```html
<div class="annotation">
  <strong>Key concept</strong> — explanation of what's happening and why.
</div>
```

**Code blocks** — standard markdown fenced blocks for source code.

**Screenshots** — captured via `url2svg` during `docs:capture:NN` task:

```html
<div class="columns">
  <div class="column is-5">
    <figure class="image screenshot">
      <img src="../NN_polling.svg" alt="During polling">
      <figcaption>During polling — partial output</figcaption>
    </figure>
  </div>
</div>
```

### CSS classes

These classes should be available in the rendered HTML (via md2html template or inline `<style>`):

```css
.annotation {
  border-left: 3px solid #3273dc;
  background: #f0f4ff;
  padding: 0.75em 1em;
  margin: 0.75em 0;
  border-radius: 0 4px 4px 0;
  font-size: 0.9em;
}
.annotation strong { color: #3273dc; }
.screenshot {
  border: 1px solid #dbdbdb;
  border-radius: 4px;
  box-shadow: 0 2px 6px rgba(0,0,0,0.1);
  overflow: hidden;
}
.screenshot img { display: block; }
```

## WASM demo page standard (templates/index.html)

The demo page is a clean standalone page for the WASM build. It is **not** documentation — it is a runnable demo.

### Required elements

- Bulma 1.0.4 CSS from CDN
- Navbar with example title + status tag (Ready/Running) + Cancel button
- Output `<div>` with loading indicator
- Start button
- Loads `wasm_exec.js` + `app.js`

### app.js pattern

All examples should use the same JavaScript pattern:

```javascript
const outputDiv = document.getElementById('output');
const statusTag = document.getElementById('status-tag');
const startBtn = document.getElementById('startBtn');
const cancelBtn = document.getElementById('cancel-btn');

let renderInterval = null;

function render() {
    if (typeof goRender === 'function') {
        outputDiv.innerHTML = goRender();
        updateStatus();
    }
}

function updateStatus() {
    const running = typeof goIsRunning === 'function' && goIsRunning();
    statusTag.textContent = running ? 'Running' : 'Ready';
    statusTag.className = running ? 'tag is-warning' : 'tag is-success';
    startBtn.disabled = running;
    cancelBtn.style.display = running ? 'inline-flex' : 'none';
    if (!running && renderInterval) {
        clearInterval(renderInterval);
        renderInterval = null;
    }
}

function start() {
    goStart();
    renderInterval = setInterval(render, 500);
    render();
}

window.wasmReady = function() {
    startBtn.disabled = false;
    outputDiv.innerHTML = '<p>Click Start to run.</p>';
};

async function loadWASM() {
    const go = new Go();
    const result = await WebAssembly.instantiateStreaming(
        fetch('main.wasm'), go.importObject
    );
    go.run(result.instance);
}

startBtn.addEventListener('click', start);
cancelBtn.addEventListener('click', function() {
    if (typeof goCancel === 'function') { goCancel(); render(); }
});

loadWASM();
```

Examples with extra controls (07–12) extend this base pattern with additional event listeners.

## Build process

### docs:build (markdown → HTML)

Renders all `index.md` files to `docs/NN_name/index.html`:

```yaml
docs:build:
  cmds:
    # ... existing md2html commands ...
    # Example documentation pages:
    - task-plus md2html --file examples/01_hello_world/index.md --dst docs/01_hello_world
    - task-plus md2html --file examples/02_svg_graph/index.md --dst docs/02_svg_graph
    # etc.
```

### docs:build-wasm (compile + copy assets)

Builds WASM binaries and copies demo assets. Must **not** copy `index.html`:

```yaml
# Per example:
- mkdir -p docs/NN_name
- cd examples/NN_name/go && GOOS=js GOARCH=wasm go build -o main.wasm .
- cp wasm_exec.js docs/NN_name/
- cp examples/NN_name/go/main.wasm docs/NN_name/
- cp examples/NN_name/templates/index.html docs/NN_name/demo.html  # NOT index.html
- cp examples/NN_name/templates/app.js docs/NN_name/
```

### docs:capture:NN (screenshots)

Captures screenshots using `url2svg` with `LOFIGUI_HOLD=1`:

```yaml
docs:capture:NN:
  cmds:
    - |
      LOFIGUI_HOLD=1 go run . &
      # wait for server, capture polling state, wait, capture complete state
      url2svg --url http://localhost:1340 -o docs/NN_polling.svg
      sleep 5
      url2svg --url http://localhost:1340 -o docs/NN_complete.svg
```

## Example list

| # | Name | Level | Pattern | Key concepts |
|---|------|-------|---------|-------------|
| 01 | Hello World | 3 — Polling | Async + refresh | `App.Handle`, `Print`, `Sleep`, auto-refresh, cancel |
| 02 | Output Showcase | 2 — Scrolling | Async + refresh | All output types: `Print`, `Markdown`, `HTML`, `Table`, SVG charts |
| 03 | Style Sampler | 6 — WASM | Browser-only | Template inheritance, `NewControllerFromFS`, multiple page layouts |
| 06 | Notes CRUD | 4 — Forms | CRUD + redirect | Form POST, `StateDict`, `RenderTemplate`, redirect-after-POST |
| 07 | Water Tank | 3 — Polling | Dashboard | Generated SVG, simulation goroutine, clickable `<a>` links |
| 08 | Water Tank Multi | 3 — Polling | Multi-page | Multiple routes, `LayoutNavbar`, HTTP Refresh per-page |
| 09 | Water Tank HTMX | 5 — HTMX | Partial updates | `hx-get`/`hx-trigger`, fragment endpoints, `renderAndCapture` mutex |
| 10 | Water Tank Maintenance | 5 — HTMX | Background ops | `context.Context` cancellation, progress, equipment lockout |
| 11 | Water Tank Storage | 7 — WASM+API | Persistent state | WASM frontend + Go API server, SeaweedFS storage |
| 12 | Batch Yield | 5 — HTMX | Cooperative scheduling | `lofigui.Yield()` for long computations |

## Choosing a starting point

- **Status page / report** → 01 (polling) or 02 (scrolling output)
- **CRUD / forms** → 06
- **Real-time dashboard** → 08 (HTTP Refresh) or 09 (HTMX)
- **Background tasks with progress** → 10
- **Browser-only (no server)** → 03 (WASM)
- **All output types demo** → 02

## Refreshing example documentation

When regenerating or updating an example's `index.md`, follow these steps:

1. Read the example's Go source code (main.go, model.go)
2. Read this document for the standard format
3. Read the existing `index.md` if one exists — preserve any hand-written exposition
4. Ensure screenshots exist (run `task docs:capture:NN` if needed)
5. Write index.md following the required sections above
6. Run `task docs:build` to render
7. Verify with `tp pages`

The documentation should explain *why*, not just *what*. Each annotation should help a reader understand the design decision, not merely restate the code.
