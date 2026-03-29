# Research: Technical

Architecture overview, polling mechanism, and task scheduling.

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
- The Python path works correctly: `template_response()` -> `state_dict()` -> injects refresh meta tag.

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
User -> GET / -> app.HandleRoot(w, r, model, true)
  1. RLock -> read ctrl, displayURL -> RUnlock
  2. ctrl.context.Reset()
  3. app.StartAction()          <- sets actionRunning=true, polling=true
  4. go modelFunc(app)          <- goroutine launched, no cancellation mechanism
  5. Write redirect HTML to /display

Background goroutine:
  model(app) {
      lofigui.Print("...")
      time.Sleep(...)
      app.EndAction()           <- sets actionRunning=false, polling=false
  }

User -> GET /display -> app.HandleDisplay(w, r)
  1. RLock -> read ctrl -> RUnlock
  2. ctrl.HandleDisplay(w, r, nil)  <- uses ctrl.StateDict (NOT app.StateDict)
```

### Python Flow (async pattern)

```
User -> GET / -> root(background_tasks)
  1. lg.reset()
  2. background_tasks.add_task(model)
  3. app.start_action()         <- sets _action_running=True, poll=True
  4. Return redirect HTML to /display

Background task:
  model() {
      lg.print("...")
      sleep(...)
      app.end_action()          <- sets _action_running=False, poll=False
  }

User -> GET /display -> display(request)
  1. app.template_response(request, "hello.html")
     -> app.state_dict(request)  <- includes refresh meta tag if poll=True
     -> templates.TemplateResponse(...)
```
