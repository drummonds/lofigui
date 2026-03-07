# Lofigui Research

Design philosophy, trade-offs, and technical deep-dives.

## Philosophy

### The original vision: no CSS, no JavaScript

lofigui started from a simple premise: what if a web UI was just `print()` statements rendered as plain HTML? No CSS framework, no JavaScript, no build step. The reasons:

1. **Simplicity** — Every dependency is a thing to learn, update, and debug. Plain HTML is the lowest common denominator. A developer who can write `print("hello")` can build a UI.

2. **Deployment** — A single binary (Go) or a minimal Python package that serves HTML over HTTP. No node_modules, no bundler, no static asset pipeline. Copy the binary to a server and run it. This matters especially for gokrazy deployments and internal tools where infrastructure is minimal.

3. **Understandability** — "View Source" shows exactly what the server sent. There is no client-side rendering, no virtual DOM diffing, no hydration step. The browser does what browsers were built to do: render HTML.

### The Bulma compromise

Plain HTML is functional but ugly. For internal tools used daily, aesthetics matter enough to justify a CSS framework. Bulma was chosen because:

- It is CSS-only — no JavaScript runtime
- It is a single CDN link — no build step
- It makes tables, forms, and layout look professional with class names alone

This is the first trade-off: we accepted a CDN dependency for better-looking output. The framework still works without Bulma (plain HTML renders fine), but the examples and defaults assume it.

### Removing JavaScript: precedent and practice

The UK Government Digital Service [removed jQuery from GOV.UK](https://insidegovuk.blog.gov.uk/2022/08/11/how-and-why-we-removed-jquery-from-gov-uk/) in 2022 — a site serving millions of users. Their reasoning: fewer bytes, fewer failure modes, better accessibility. If GOV.UK can serve a nation without jQuery, an internal tool can certainly manage without React.

lofigui takes this further. The base framework uses zero JavaScript. The browser's native capabilities — HTML rendering, form submission, HTTP Refresh — handle everything in examples 01-08.

### Where JavaScript creeps back in

Two features introduce JavaScript, both deliberately:

**WASM** (examples 03, 04, 07, 08) — Go compiled to WebAssembly requires a small JS loader (`wasm_exec.js`). This is the price of running the same Go code in the browser without a server. The JS is boilerplate glue, not application logic.

**HTMX** (examples 09, 10) — A single `<script>` tag that adds `hx-get` and `hx-trigger` attributes to HTML elements. HTMX exists because full-page HTTP Refresh polling has a real usability problem: if you are trying to enter information in a form or click a button, the page refresh interrupts you. The input loses focus, the form resets, the click never registers. For display-only dashboards, polling is fine. For anything interactive, it is maddening.

HTMX solves this by updating only the parts of the page that change, leaving forms and buttons untouched. It is the minimum JavaScript needed to make multi-page dynamic sites usable.

### The JavaScript budget

The position is not "no JavaScript ever" but "justify every byte":

| Layer                 | JS?          | Justification                                 |
| --------------------- | ------------ | --------------------------------------------- |
| Base (examples 01-08) | None         | Full-page refresh is sufficient               |
| HTMX (examples 09-10) | ~14KB        | Partial updates make interactive pages usable |
| WASM (examples 03-04) | ~16KB loader | Enables server-free deployment                |

No bundler, no npm, no build step. Each JS dependency is a single file loaded from a CDN or embedded.

### Print as interface

The fundamental insight is that `print()` is the most natural programming interface. Every developer learns it first. lofigui preserves that — you print things, they appear on a web page. The abstraction cost is near zero.

### Progressive complexity

The examples are ordered deliberately:

1. **Print and poll** (01) — the simplest useful pattern
2. **Synchronous render** (02) — when you don't need async
3. **WASM** (03, 04) — same code, no server
4. **CRUD** (06) — forms and state
5. **Real-time dashboards** (07-09) — SVG, multi-page, HTMX
6. **Background operations** (10) — goroutines, cancellation, progress

Each step adds one concept. You stop at the level of complexity your project needs.

### The Bulma lesson applied more broadly

godocs originally started with very simple hand-written CSS. Over time it grew complex and inconsistent — and it was actually _smaller_ to switch to Bulma for a more consistent result. The same principle may apply to charts and other areas: a focused, well-chosen dependency can be simpler than a DIY approach that accumulates complexity over time.

### Where does lofigui sit?

lofigui is for **single-process, small-audience tools**. The sweet spot: 1-10 users, one real object (a machine, a simulation, a long-running process) with a few pages showing different views of it. It is not competing with React or even Streamlit — it is competing with "I'll just use the terminal" or "I'll write a quick bash CGI script".

---

## Charts

### The problem

Charts are the next big gap in lofigui. The current approach (example 02) uses Go libraries to produce SVG server-side. The results are functional but poor — limited axis formatting, weak label placement, no interactivity, and mediocre visual quality compared to what users expect from modern dashboards.

This was explored in the [gobank chart comparison](https://drummonds.github.io/gobank/research/chart-comparison.html), which tested three Go SVG renderers against financial data:

### Go SVG libraries tested (gobank)

| Library               | Strengths                               | Weaknesses                                                                                                      |
| --------------------- | --------------------------------------- | --------------------------------------------------------------------------------------------------------------- |
| **Hand-rolled SVG**   | Full control, no dependencies           | Enormous effort for basic features (axis scaling, labels, legends). Every chart type is a fresh implementation. |
| **go-analyze/charts** | Good chart variety, reasonable defaults | Axis formatting issues (decimal values where percentages expected), limited customisation                       |
| **margaid**           | Clean line charts                       | Limited chart types, sparse documentation                                                                       |

The go-chart library (used in lofigui example 02) has similar issues — it works for simple bar/line charts but struggles with axis formatting, date handling, and multi-series layouts.

**Core issue**: Go's charting ecosystem is immature compared to JavaScript's. The libraries exist but produce output that looks 10 years behind what a JS library produces with the same data and less code.

### The JavaScript charting landscape

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

### What to avoid

The key constraint is: **no React, no Svelte, no Angular, no build step**. Libraries that require a framework or a bundler are out:

- **Recharts** — React-only
- **Victory** — React-only
- **Nivo** — React-only
- **SvelteKit charts** — Svelte-only

The lofigui pattern is: server renders HTML, browser displays it. A chart library must work with a `<script>` tag and a `<div>` target, nothing more.

### Leading candidates for lofigui

**D3.js** is the most interesting option. It:

- Produces native SVG (inspectable, stylable with CSS, printable)
- Works from a single CDN `<script>` tag
- Has no framework dependency
- Is the foundation most other libraries build on
- Supports every chart type imaginable

The downside is D3's verbosity — a simple line chart is ~30 lines of JS. But lofigui could generate the D3 JavaScript server-side (Go `fmt.Sprintf` with data injected into a template), keeping the complexity on the server while the browser just executes the rendering.

**Observable Plot** is D3's high-level API, worth considering if D3 feels too low-level. Same SVG output, much less code.

**Chart.js** is the simplest option if SVG isn't required. Canvas output means less flexibility but the API is very approachable.

### How it would work in lofigui

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

### Potential example 12: water tank charts

A natural next example would chart the water tank simulation data — tank level and pump/valve activity over time. This would:

- Demonstrate time-series charting with real simulation data
- Show how to pass Go data to a JS charting library
- Build on the existing water tank examples (07-10)
- Provide a reference pattern for charting in lofigui apps

The chart would update via HTMX (like example 09) — the fragment endpoint returns fresh chart HTML with updated data on each poll.

---

# Technical Research

## 1. Codebase Overview

lofigui is a lightweight web-UI framework with dual Python and Go implementations. It provides a print-like interface for building server-side rendered HTML applications.

### Architecture

The framework has three layers:

1. **Buffer Layer** (`lofigui.go` / `context.py`, `print.py`, `markdown.py`) — Global mutable buffer that accumulates HTML fragments via `Print()`, `Markdown()`, `HTML()`, `Table()` calls. Python uses an `asyncio.Queue` that drains into a string buffer; Go uses a `strings.Builder` directly.

2. **Controller Layer** (`controller.go` / `controller.py`) — Wraps a template engine (pongo2 in Go, Jinja2 in Python). Renders templates with the buffer content injected as `{{ results | safe }}`.

3. **App Layer** (`app.go` / `app.py`) — Manages controller lifecycle, action state (running/stopped), and auto-refresh polling. Implements the "singleton active model" concept — only one background task should run at a time.

### Execution Patterns

| Pattern             | Flow                                                                                                                                                                      | Examples       |
| ------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | -------------- |
| **Async + Polling** | `GET /` resets buffer, starts model in goroutine/background task, redirects to `/display`. `/display` includes `<meta http-equiv="Refresh">` tag while polling is active. | 01 (Go+Python) |
| **Synchronous**     | Model runs inline, `EndAction()` called immediately, redirects to `/display`.                                                                                             | 02 (Go+Python) |
| **WASM**            | Go compiled to WebAssembly, no server-side app.                                                                                                                           | 03, 04         |
| **CRUD**            | Form POSTs modify state, redirect to `GET /`. Uses `ctrl.StateDict()` + `ctrl.RenderTemplate()` directly.                                                                 | 06 (Go+Python) |

---

## 2. Notification System (Polling/Auto-Refresh)

The "notification system" is a server-side polling mechanism using HTML `<meta http-equiv="Refresh">` tags. There is no WebSocket, SSE, or JavaScript-based notification — it's pure HTTP redirect-based polling.

### How It Works

1. **Start**: `app.StartAction()` sets `actionRunning=true`, `polling=true`, `PollCount=0`.
2. **Poll Cycle**: On each `/display` request, `app.StateDict()` (Python) checks `self.poll`:
   - If `True`: injects `<meta http-equiv="Refresh" content="N">` into the `refresh` template variable. Increments `poll_count`.
   - If `False`: sets `refresh=""`, resets `poll_count=0`.
3. **Stop**: `app.EndAction()` sets `actionRunning=false`, `polling=false`. Next `/display` request renders without the refresh tag — polling stops.
4. **Template**: The template must include `{{ refresh | safe }}` in the `<head>` for auto-refresh to work.

### Python Implementation Details

- `App.state_dict()` builds the full context dict including `refresh`, `polling`, `poll_count`.
- `App.template_response()` calls `state_dict()` then renders via Jinja2.
- The Python path works correctly: `template_response()` → `state_dict()` → injects refresh meta tag.

### Go Implementation Details

- `App.StateDict()` is the method that should build the full context with polling info.
- `Controller.StateDict()` only returns `request` and `results` (no polling info).
- `App.HandleDisplay()` delegates to `ctrl.HandleDisplay()` which only uses `ctrl.StateDict()`.

### Startup Bounce (Python only)

The Python `App` has a `startup` flag. On the first call to `template_response()`, if the URL path isn't `/`, it redirects to `/`. This prevents a stale `/display` page from showing when the server restarts. The `startup_bounce_count` limit of 3 is effectively unreachable because `startup` is set to `False` on the first invocation.

---

## 3. Task Scheduling Flow

### Go Flow (async pattern)

```
User → GET / → app.HandleRoot(w, r, model, true)
  1. RLock → read ctrl, displayURL → RUnlock
  2. ctrl.context.Reset()
  3. app.StartAction()          ← sets actionRunning=true, polling=true
  4. go modelFunc(app)          ← goroutine launched, no cancellation mechanism
  5. Write redirect HTML to /display

Background goroutine:
  model(app) {
      lofigui.Print("...")
      time.Sleep(...)
      app.EndAction()           ← sets actionRunning=false, polling=false
  }

User → GET /display → app.HandleDisplay(w, r)
  1. RLock → read ctrl → RUnlock
  2. ctrl.HandleDisplay(w, r, nil)  ← uses ctrl.StateDict (NOT app.StateDict)
```

### Python Flow (async pattern)

```
User → GET / → root(background_tasks)
  1. lg.reset()
  2. background_tasks.add_task(model)
  3. app.start_action()         ← sets _action_running=True, poll=True
  4. Return redirect HTML to /display

Background task:
  model() {
      lg.print("...")
      sleep(...)
      app.end_action()          ← sets _action_running=False, poll=False
  }

User → GET /display → display(request)
  1. app.template_response(request, "hello.html")
     → app.state_dict(request)  ← includes refresh meta tag if poll=True
     → templates.TemplateResponse(...)
```

---

## 4. Bugs Found

### BUG 1 (CRITICAL): Go `app.HandleDisplay()` never injects polling/refresh state

**File**: `app.go:192-203`

```go
func (app *App) HandleDisplay(w http.ResponseWriter, r *http.Request) {
    // ...
    ctrl.HandleDisplay(w, r, nil)  // ← delegates to controller, NOT app
}
```

`ctrl.HandleDisplay()` calls `ctrl.StateDict(r)` which only returns `request` and `results`. The `refresh`, `polling`, `version`, `controller_name`, and `poll_count` template variables are NEVER populated.

**Impact**: In the Go async pattern (example 01), `{{ refresh | safe }}` in the template always renders as empty. The page never auto-refreshes. The user must manually refresh to see progress updates. The entire polling mechanism is non-functional in Go.

**Fix**: `app.HandleDisplay()` should use `app.StateDict()` to build the full context, then pass it to `ctrl.HandleDisplay()` as extra context, or render the template directly.

### BUG 2 (CRITICAL): Go `app.StateDict()` deadlocks

**File**: `app.go:241-283`

```go
func (app *App) StateDict(r *http.Request, extraContext pongo2.Context) pongo2.Context {
    app.mu.Lock()    // ← acquires write lock (line 242)
    // ...
    ctx := pongo2.Context{
        // ...
        "controller_name": app.ControllerName(),  // ← line 257
    }
```

`ControllerName()` at `app.go:207-215`:

```go
func (app *App) ControllerName() string {
    app.mu.RLock()      // ← tries to acquire read lock while write lock is held
    defer app.mu.RUnlock()
    // ...
}
```

`sync.RWMutex` is NOT re-entrant. A goroutine holding a write lock that attempts to acquire a read lock on the same mutex will deadlock permanently.

**Impact**: Any call to `app.StateDict()` will hang forever. Currently this is masked by BUG 1 (HandleDisplay doesn't call StateDict), but if BUG 1 is fixed naively by calling `app.StateDict()`, the program will deadlock.

**Fix**: Either inline the controller name lookup inside `StateDict()` (use `app.controller` directly since lock is already held), or refactor to avoid nested locking.

### BUG 3 (CRITICAL): No cancellation mechanism for running goroutines/tasks

**Files**: `app.go:166-188` (Go), `examples/01_hello_world/python/hello.py:41-51` (Python)

When the user triggers the root endpoint while a model is already running:

**Go**:

```go
func (app *App) HandleRoot(...) {
    // ...
    app.StartAction()       // resets polling state for the "new" action
    go modelFunc(app)       // launches NEW goroutine — old one is still running!
    // ...
}
```

**Python**:

```python
async def root(background_tasks: BackgroundTasks):
    lg.reset()
    background_tasks.add_task(model)  # schedules NEW task — old one still running!
    app.start_action()
```

There is no `context.Context`, channel, flag, or any other cancellation mechanism. The old goroutine/task:

1. Continues running indefinitely
2. Continues writing to the shared buffer (which was just reset)
3. Eventually calls `EndAction()`, **prematurely stopping the polling for the NEW action**

**This is the "tasks that should have been cancelled" bug.** The old task's `EndAction()` call terminates the polling that the new task needs.

**Scenario**:

```
T=0:  User hits /  → goroutine A starts, StartAction()
T=2:  User hits /  → goroutine B starts, StartAction() (resets state)
T=5:  goroutine A finishes → calls EndAction() → STOPS POLLING
T=5:  goroutine B is still running but polling is now off
T=8:  goroutine B finishes → calls EndAction() (no-op, already stopped)
```

Result: From T=5 to T=8, goroutine B is running but the page has stopped auto-refreshing.

**Fix**: The model function signature should accept a `context.Context` (Go) or a cancellation flag. `HandleRoot` should cancel the previous context before starting a new one. Model functions must check for cancellation in their loops.

### BUG 4 (SIGNIFICANT): Python `reset()` doesn't clear the asyncio queue

**File**: `lofigui/context.py:103-121`

```python
def reset(ctx=None):
    if ctx is None:
        ctx = _ctx
    ctx.buffer = ""  # clears buffer string only — queue is untouched!
```

The `print()` function puts messages into `ctx.queue`. `buffer()` drains the queue into the buffer string. `reset()` clears the buffer string but leaves the queue intact.

**Scenario**:

```
1. lg.print("msg1")     → queue: ["<p>msg1</p>"]
2. lg.reset()            → buffer="" but queue still has ["<p>msg1</p>"]
3. lg.print("msg2")     → queue: ["<p>msg1</p>", "<p>msg2</p>"]
4. lg.buffer()           → drains queue → buffer="<p>msg1</p><p>msg2</p>"
```

`msg1` survives the reset and appears in the output.

**Impact**: When the user restarts an action (hits `/` again), messages from the previous run that were queued but not yet drained will persist into the new run's output.

**Fix**: `reset()` should drain and discard the queue before clearing the buffer:

```python
def reset(ctx=None):
    if ctx is None:
        ctx = _ctx
    # Drain and discard the queue
    while not ctx.queue.empty():
        try:
            ctx.queue.get_nowait()
            ctx.queue.task_done()
        except asyncio.QueueEmpty:
            break
    ctx.buffer = ""
```

### BUG 5 (SIGNIFICANT): Example 05 (demo_app.py) is completely broken

**File**: `examples/05_demo_app/demo_app.py`

The demo app uses Controller API methods that were removed during the refactoring that moved action management to the App level:

| Line | Call                                                 | Status                                     |
| ---- | ---------------------------------------------------- | ------------------------------------------ |
| 136  | `current_controller.is_action_running()`             | Method doesn't exist on Controller         |
| 141  | `current_controller.get_refresh()`                   | Method doesn't exist on Controller         |
| 181  | `Controller(template_path="templates/process.html")` | Constructor doesn't accept `template_path` |
| 186  | `current_controller.start_action(refresh_time=2)`    | Method doesn't exist on Controller         |
| 207  | `current_controller.end_action()`                    | Method doesn't exist on Controller         |

**Impact**: The demo app crashes with `AttributeError` on any process-related endpoint.

### BUG 6 (MODERATE): Go `HandleRoot` race between reset and goroutine writes

**File**: `app.go:166-188`

```go
func (app *App) HandleRoot(...) {
    app.mu.RLock()
    ctrl := app.controller
    app.mu.RUnlock()
    // ← window: controller could be replaced here

    if resetBuffer {
        ctrl.context.Reset()     // reset buffer
    }
    // ← window: old goroutine could write here between reset and new goroutine start

    app.StartAction()
    go modelFunc(app)            // new goroutine starts writing
```

Between `ctrl.context.Reset()` and the new goroutine starting, an old goroutine (from a previous `HandleRoot` call) can write to the buffer. These writes will appear in the new action's output.

**Impact**: Stale output from previous runs can leak into new runs.

### BUG 7 (MODERATE): Go `StateDict` doesn't defer unlock

**File**: `app.go:242-283`

```go
func (app *App) StateDict(...) pongo2.Context {
    app.mu.Lock()
    // ... 30+ lines of code ...
    app.mu.Unlock()    // ← not deferred
```

If any code between Lock and Unlock panics (e.g., `ctrl.context.Buffer()` on a nil controller context), the mutex is never released, permanently deadlocking all future operations.

**Fix**: Use `defer app.mu.Unlock()` immediately after `app.mu.Lock()`.

### BUG 8 (MINOR): Python mutable default arguments

**File**: `lofigui/app.py:85, 111`

```python
def state_dict(self, request: Request, extra: dict = {}) -> dict:
def template_response(self, request: Request, templateName: str, extra: dict = {}) -> HTMLResponse:
```

Mutable default arguments are shared across calls. While the code doesn't mutate `extra` directly in all paths, `self.controller.state_dict(extra=extra)` at line 105 passes the same default dict object through, which could be mutated by controller implementations.

**Fix**: Use `extra: dict = None` and `if extra is None: extra = {}`.

### BUG 9 (MINOR): Python `start_demo_process` blocks async event loop

**File**: `examples/05_demo_app/demo_app.py:197-200`

```python
async def start_demo_process(duration: int = Form(10)):
    for i in range(duration):
        time.sleep(1)  # blocks entire event loop
```

Using `time.sleep()` in an async handler blocks all request processing. The auto-refresh requests can't be served while the process runs, defeating the purpose of polling.

### BUG 10 (MINOR): Go Context uses sync.Mutex instead of sync.RWMutex

**File**: `lofigui.go:17-19`

```go
type Context struct {
    buffer        strings.Builder
    mu            sync.Mutex    // ← should be sync.RWMutex for concurrent reads
}
```

`Buffer()` is a read-only operation but takes an exclusive lock, blocking concurrent readers unnecessarily.

---

## 5. Bug Interaction Analysis

Bugs 1, 2, and 3 interact in a particularly insidious way:

- **BUG 1** means Go's `app.HandleDisplay()` bypasses `app.StateDict()`, so the auto-refresh meta tag is never rendered. This masks **BUG 2** (the deadlock in `StateDict`).
- If you fix **BUG 1** by making `HandleDisplay` call `app.StateDict()`, you immediately hit **BUG 2** (deadlock).
- If you fix **BUG 2** AND **BUG 1**, the auto-refresh now works, but then **BUG 3** becomes much more visible — old goroutines calling `EndAction()` prematurely kill the new action's polling.

The correct fix order is: BUG 2 → BUG 1 → BUG 3.

Similarly, **BUG 4** (Python queue not cleared on reset) and **BUG 3** (no cancellation) interact: even if you add cancellation to stop old tasks from calling `EndAction()`, messages from old tasks still leak through the queue.

---

## 6. Root Cause: Missing Cancellation Architecture

The fundamental design issue is that the framework has a "singleton active model" concept (only one action at a time) but no mechanism to enforce it:

1. **No cancellation propagation**: Go goroutines and Python background tasks have no way to learn they've been superseded.
2. **No action ID/generation counter**: There's no way to distinguish which action's `EndAction()` call is current.
3. **Global mutable state**: The buffer and action flags are global, shared across all requests and goroutines.

A minimal fix would add a generation counter:

```go
type App struct {
    // ...
    actionGeneration uint64
}

func (app *App) StartAction() uint64 {
    app.mu.Lock()
    defer app.mu.Unlock()
    app.actionGeneration++
    app.actionRunning = true
    app.polling = true
    return app.actionGeneration
}

func (app *App) EndAction(generation uint64) {
    app.mu.Lock()
    defer app.mu.Unlock()
    if generation == app.actionGeneration {
        app.actionRunning = false
        app.polling = false
    }
    // else: stale call from superseded goroutine — ignore
}
```

A fuller fix would use `context.Context` for cancellation signaling.
