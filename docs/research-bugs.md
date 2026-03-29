# Research: Bugs

Known bugs, interaction analysis, and root causes.

## Bugs Found

### BUG 1 (CRITICAL): Go `app.HandleDisplay()` never injects polling/refresh state

**File**: `app.go`

`ctrl.HandleDisplay()` calls `ctrl.StateDict(r)` which only returns `request` and `results`. The `refresh`, `polling`, `version`, `controller_name`, and `poll_count` template variables are NEVER populated.

**Impact**: In the Go async pattern (example 01), `{{ refresh | safe }}` in the template always renders as empty. The page never auto-refreshes. The entire polling mechanism is non-functional in Go.

**Fix**: `app.HandleDisplay()` should use `app.StateDict()` to build the full context, then pass it to `ctrl.HandleDisplay()` as extra context, or render the template directly.

### BUG 2 (CRITICAL): Go `app.StateDict()` deadlocks

**File**: `app.go`

`StateDict()` acquires `app.mu.Lock()` then calls `app.ControllerName()` which tries `app.mu.RLock()`. `sync.RWMutex` is NOT re-entrant — a goroutine holding a write lock that attempts a read lock on the same mutex deadlocks permanently.

**Impact**: Any call to `app.StateDict()` hangs forever. Currently masked by BUG 1 (HandleDisplay doesn't call StateDict), but fixing BUG 1 naively triggers this deadlock.

**Fix**: Inline the controller name lookup inside `StateDict()` (use `app.controller` directly since lock is already held), or refactor to avoid nested locking.

### BUG 3 (CRITICAL): No cancellation mechanism for running goroutines/tasks

**Files**: `app.go` (Go), `examples/01_hello_world/python/hello.py` (Python)

When the user triggers the root endpoint while a model is already running, a NEW goroutine/task is launched — the old one keeps running. There is no `context.Context`, channel, flag, or any other cancellation mechanism. The old goroutine continues writing to the shared buffer and eventually calls `EndAction()`, **prematurely stopping the polling for the NEW action**.

**Scenario**:

```
T=0:  User hits /  -> goroutine A starts, StartAction()
T=2:  User hits /  -> goroutine B starts, StartAction() (resets state)
T=5:  goroutine A finishes -> calls EndAction() -> STOPS POLLING
T=5:  goroutine B is still running but polling is now off
T=8:  goroutine B finishes -> calls EndAction() (no-op, already stopped)
```

**Fix**: Model function should accept `context.Context`. `HandleRoot` should cancel the previous context before starting a new one.

### BUG 4 (SIGNIFICANT): Python `reset()` doesn't clear the asyncio queue

**File**: `lofigui/context.py`

`print()` puts messages into `ctx.queue`. `buffer()` drains the queue into the buffer string. `reset()` clears the buffer string but leaves the queue intact — messages from the previous run survive the reset.

**Fix**: `reset()` should drain and discard the queue before clearing the buffer.

### BUG 5 (SIGNIFICANT): Example 05 (demo_app.py) is completely broken

**File**: `examples/05_demo_app/demo_app.py`

Uses Controller API methods that were removed during refactoring: `is_action_running()`, `get_refresh()`, `start_action()`, `end_action()`, and a `template_path` constructor arg.

**Impact**: Crashes with `AttributeError` on any process-related endpoint.

### BUG 6 (MODERATE): Go `HandleRoot` race between reset and goroutine writes

**File**: `app.go`

Between `ctrl.context.Reset()` and the new goroutine starting, an old goroutine can write to the buffer. These writes appear in the new action's output.

**Impact**: Stale output from previous runs can leak into new runs.

### BUG 7 (MODERATE): Go `StateDict` doesn't defer unlock

**File**: `app.go`

`app.mu.Lock()` is not deferred. If any code between Lock and Unlock panics, the mutex is never released, permanently deadlocking all future operations.

**Fix**: Use `defer app.mu.Unlock()` immediately after `app.mu.Lock()`.

### BUG 8 (MINOR): Python mutable default arguments

**File**: `lofigui/app.py`

`state_dict(extra: dict = {})` and `template_response(extra: dict = {})` use mutable default arguments.

**Fix**: Use `extra: dict = None` and `if extra is None: extra = {}`.

### BUG 9 (MINOR): Python `start_demo_process` blocks async event loop

**File**: `examples/05_demo_app/demo_app.py`

Uses `time.sleep()` in an async handler, blocking all request processing.

### BUG 10 (MINOR): Go Context uses sync.Mutex instead of sync.RWMutex

**File**: `lofigui.go`

`Buffer()` is a read-only operation but takes an exclusive lock, blocking concurrent readers unnecessarily.

---

## Bug Interaction Analysis

Bugs 1, 2, and 3 interact in a particularly insidious way:

- **BUG 1** means Go's `app.HandleDisplay()` bypasses `app.StateDict()`, so the auto-refresh meta tag is never rendered. This masks **BUG 2** (the deadlock in `StateDict`).
- If you fix **BUG 1** by making `HandleDisplay` call `app.StateDict()`, you immediately hit **BUG 2** (deadlock).
- If you fix **BUG 2** AND **BUG 1**, the auto-refresh now works, but then **BUG 3** becomes much more visible — old goroutines calling `EndAction()` prematurely kill the new action's polling.

The correct fix order is: BUG 2 -> BUG 1 -> BUG 3.

Similarly, **BUG 4** (Python queue not cleared on reset) and **BUG 3** (no cancellation) interact: even if you add cancellation to stop old tasks from calling `EndAction()`, messages from old tasks still leak through the queue.

---

## Root Cause: Missing Cancellation Architecture

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
    // else: stale call from superseded goroutine -- ignore
}
```

A fuller fix would use `context.Context` for cancellation signaling.
