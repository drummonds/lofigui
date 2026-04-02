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

---

## 4. Handle (Single-Endpoint Pattern)

`Handle(model)` returns an `http.HandlerFunc` that manages the full lifecycle on a single URL. It uses buffer state as a signal:

```
GET /  (buffer empty, not running)   → start model in goroutine, render with Refresh header
GET /  (running)                     → render current buffer (polling continues)
GET /  (not running, buffer has content) → render final output, no refresh
```

### State machine

| Buffer  | Running | Action |
|---------|---------|--------|
| empty   | false   | Start model, render with polling |
| any     | true    | Render current output, keep polling |
| content | false   | Render final output, stop polling |

When the model goroutine returns normally, `flush()` is called which:
1. Calls `EndAction()` — stops polling, cancels context
2. If `LOFIGUI_HOLD` is **not** set: waits 2 seconds (grace period for browser to pick up final render), then signals server shutdown via `signalDone()`
3. If `LOFIGUI_HOLD` **is** set: returns immediately — server stays alive

### Restart behaviour

With `Handle`, the model only starts when the buffer is empty **and** no action is running. After the model completes, the buffer retains content, so subsequent requests just render the final output — no automatic restart. For restart support, use `HandleRoot` + `HandleDisplay` instead (the model can include a link to `HandleRoot`'s URL).

---

## 5. HOLD Mode

Setting `LOFIGUI_HOLD=1` keeps the server running after the model completes. This is designed for screenshot capture with tools like `url2svg`.

### How it works

```go
// In NewApp():
hold: os.Getenv("LOFIGUI_HOLD") != "",

// In flush():
func (app *App) flush() {
    app.EndAction()
    if app.hold {
        return  // keep server running
    }
    time.Sleep(2 * time.Second)  // grace period
    app.signalDone()             // trigger server shutdown
}
```

### Usage for screenshot capture

```bash
# Start server — stays alive after model completes
LOFIGUI_HOLD=1 go run .

# In another terminal, capture with url2svg
url2svg --url http://localhost:1340 -o screenshot.svg
```

The `docs:capture:*` Taskfile tasks automate this: start server in background, trigger the model, capture at timed intervals, then kill the server.

---

## 6. Cancellation Flow

Cancellation is **transparent** — model code does not need explicit cancel handling. The framework uses a panic-recover mechanism with an internal sentinel type.

### Trigger

- **Server**: `HandleCancel(redirectURL)` calls `EndAction()`, sets `cancelled=true`, redirects, and triggers graceful shutdown
- **WASM**: `goCancel()` calls `EndAction()`
- **Restart**: `StartAction()` cancels the previous action's context before starting a new one

### Mechanism

```
EndAction() / StartAction()
  → cancels context via cancelFunc()
  → next Print(), Sleep(), or Yield() call checks context
  → panics with cancelledError{} sentinel
  → Handle/HandleRoot's recover wrapper catches it
  → goroutine exits cleanly
```

### Detail

1. `EndAction()` calls `cancelFunc()`, setting the context to done
2. `defaultContext` stores the action context; `checkCancelled()` checks it
3. `Print()`, `Sleep()`, `Yield()` all call `checkCancelled()` — if the context is done, they `panic(errCancelled)`
4. The `go func()` in `Handle`/`HandleRoot` has `defer recover()` that catches `cancelledError` and returns silently
5. The buffer retains whatever was printed before cancellation — this becomes the final output

### Server shutdown after cancel

When using `app.ListenAndServe`, `HandleCancel` triggers the same graceful shutdown as normal model completion:

1. `EndAction()` cancels the context and stops polling
2. `cancelled` flag is set on the app
3. The HTTP redirect is sent to the client
4. After a 2-second grace period, `signalDone()` triggers `srv.Shutdown()`
5. `ListenAndServe` returns `ErrCancelled` (not `nil`) — allowing the caller to exit with a non-zero status

```go
if err := app.ListenAndServe(":1340", nil); err != nil {
    if errors.Is(err, lofigui.ErrCancelled) {
        log.Println(err)
        os.Exit(1)
    }
    log.Fatal(err)
}
```

Normal completion returns `nil` (exit 0). Cancel returns `ErrCancelled` (exit 1).

### Overriding cancel behaviour

For long-running servers that should keep running after cancel (e.g. HTMX apps, multi-page dashboards), write a custom handler instead of using `HandleCancel`:

```go
http.HandleFunc("/cancel", func(w http.ResponseWriter, r *http.Request) {
    app.EndAction()
    http.Redirect(w, r, "/", http.StatusSeeOther)
})
```

This calls `EndAction()` (stops the model) without triggering server shutdown.

### Buffer state after cancel

After cancellation, the buffer has partial content and `actionRunning` is false. With `Handle`, this means subsequent requests render the partial output as the "done" state (no restart, no polling).

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
