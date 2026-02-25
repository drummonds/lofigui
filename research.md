# Lofigui Deep Research Report

## 1. Codebase Overview

lofigui is a lightweight web-UI framework with dual Python and Go implementations. It provides a print-like interface for building server-side rendered HTML applications.

### Architecture

The framework has three layers:

1. **Buffer Layer** (`lofigui.go` / `context.py`, `print.py`, `markdown.py`) — Global mutable buffer that accumulates HTML fragments via `Print()`, `Markdown()`, `HTML()`, `Table()` calls. Python uses an `asyncio.Queue` that drains into a string buffer; Go uses a `strings.Builder` directly.

2. **Controller Layer** (`controller.go` / `controller.py`) — Wraps a template engine (pongo2 in Go, Jinja2 in Python). Renders templates with the buffer content injected as `{{ results | safe }}`.

3. **App Layer** (`app.go` / `app.py`) — Manages controller lifecycle, action state (running/stopped), and auto-refresh polling. Implements the "singleton active model" concept — only one background task should run at a time.

### Execution Patterns

| Pattern | Flow | Examples |
|---|---|---|
| **Async + Polling** | `GET /` resets buffer, starts model in goroutine/background task, redirects to `/display`. `/display` includes `<meta http-equiv="Refresh">` tag while polling is active. | 01 (Go+Python) |
| **Synchronous** | Model runs inline, `EndAction()` called immediately, redirects to `/display`. | 02 (Go+Python) |
| **WASM** | Go compiled to WebAssembly, no server-side app. | 03, 04 |
| **CRUD** | Form POSTs modify state, redirect to `GET /`. Uses `ctrl.StateDict()` + `ctrl.RenderTemplate()` directly. | 06 (Go+Python) |

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

| Line | Call | Status |
|------|------|--------|
| 136 | `current_controller.is_action_running()` | Method doesn't exist on Controller |
| 141 | `current_controller.get_refresh()` | Method doesn't exist on Controller |
| 181 | `Controller(template_path="templates/process.html")` | Constructor doesn't accept `template_path` |
| 186 | `current_controller.start_action(refresh_time=2)` | Method doesn't exist on Controller |
| 207 | `current_controller.end_action()` | Method doesn't exist on Controller |

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
