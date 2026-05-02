<style>
.annotation { border-left: 3px solid #3273dc; background: #f0f4ff; padding: 0.75em 1em; margin: 0.75em 0; border-radius: 0 4px 4px 0; font-size: 0.9em; }
.annotation strong { color: #3273dc; }
.screenshot { border: 1px solid #dbdbdb; border-radius: 4px; box-shadow: 0 2px 6px rgba(0,0,0,0.1); overflow: hidden; }
.screenshot img { display: block; width: 100%; height: auto; }
</style>

# 06 — Notes CRUD

Smallest interesting CRUD app: a numeric-keyed in-memory map of notes, a **master / detail** UI with per-row Read / Edit / Delete buttons, and the **Post / Redirect / Get** pattern. One Go codebase, **two deployment targets** — a real HTTP server (`main.go`) and a browser-only WASM build that runs the same `*http.ServeMux` inside a service worker (`main_wasm.go`). A Python implementation alongside the Go one shows the same shape with FastAPI.

Each POST handler mutates the notes map, stashes a one-shot **flash message** describing what it just did, and redirects with `303 See Other` — to `./` (back to the list) for create / delete, or to `../{id}` (back to the detail page) for update. The next GET consumes the flash, prepends it as a Bulma notification, and renders the requested view.

**[Interactivity level](../research-philosophy.html#the-interactivity-spectrum):** 4 — Static + forms (CRUD pattern, no polling)
**[State scope](../research-philosophy.html#the-state-dimension):** Global (server build — one shared notes map) / Individual (WASM build — each browser's SW has its own seeded notes)

<div class="buttons">
<a href="wasm_demo/" class="button is-primary">Launch Demo</a>
<a target="_blank" href="https://codeberg.org/hum3/lofigui/src/branch/main/examples/06_notes_crud" class="button is-light">Source on Codeberg</a>
</div>

<div class="columns is-vcentered">
<div class="column is-5">
<figure class="image screenshot">
<img src="../06_initial.svg" alt="Initial state — three seeded notes, each row has Read / Edit / Delete buttons, empty Create form below">
<figcaption class="has-text-centered has-text-grey is-size-7 mt-1">Initial — three seeded notes, per-row action buttons, Create form</figcaption>
</figure>
</div>
<div class="column is-narrow has-text-centered">
<span style="font-size: 2rem; color: #999;">&rarr;</span>
</div>
<div class="column is-5">
<figure class="image screenshot">
<img src="../06_populated.svg" alt="After the curl-driven CRUD sequence — flash notification at the top, two new notes, note 2 updated, note 3 deleted">
<figcaption class="has-text-centered has-text-grey is-size-7 mt-1">After a curl-driven CRUD sequence — flash + updated table</figcaption>
</figure>
</div>
</div>

<figure class="image screenshot mt-4" style="max-width: 60%; margin: 1.5em auto">
<img src="../06_detail.svg" alt="Detail view of note 1 — full text in a box, Back / Edit / Delete buttons">
<figcaption class="has-text-centered has-text-grey is-size-7 mt-1">Detail view (the "Read" button target) — full text, no truncation</figcaption>
</figure>

<p class="has-text-grey is-size-7 has-text-centered">All three captures are produced by <a href="https://codeberg.org/hum3/lofigui/src/branch/main/Taskfile.yml"><code>task docs:capture:06</code></a>, which drives the server with a sequence of <code>curl</code> POSTs. The capture asserts every POST returns <code>303 See Other</code>, the validation cases (oversized text, non-existent ID) flash the right error, and the rendered SVGs contain the expected note text — the screenshots double as an integration test.</p>

---

## How lofigui helps here

Six things from the library are doing real work in this example:

1. **`lofigui.Reset` + `lofigui.Buffer`** — the lofigui buffer is a process-global accumulator. Every handler resets it, prints the page-specific content, and reads `Buffer()` into the template's `{{.content}}` slot.
2. **`lofigui.HTML`** — writes raw HTML into the buffer (use this for chrome and notification markup; use `html.EscapeString` separately on user-supplied text). The original version of this example used `lofigui.Print("<h2>…</h2>")` and ended up with literal angle brackets on screen — `Print` escapes by default; `HTML` does not.
3. **`lofigui.NewControllerFromFS`** — parses the embedded `notes.html` template once at startup (read from `//go:embed templates`, the only filesystem a WASM build can reach).
4. **`Controller.RenderTemplate`** — writes to any `http.ResponseWriter`, so the same call works for `net/http.ListenAndServe` and `go-wasm-http-server` inside a service worker.
5. **`Controller.StateDict`** — pre-fills the template context with `version`, `name`, `polling`, `refresh` (no-op here since this example doesn't poll) so the handler only adds CRUD-specific keys (`content`, `base`).
6. **`lofigui.ServeFavicon` + `lofigui.ServeBulma`** — registered alongside the CRUD routes. Bulma loads from local `/assets/bulma.min.css` on both server and WASM builds — no CDN round-trip.

Everything else — the redirect-after-POST cycle, the form parsing, the in-memory map — is plain Go standard library.

---

## The route table

| Method + Path | View | Redirects to |
|---------------|------|--------------|
| `GET /` | Master list — table of notes with per-row Read / Edit / Delete | — |
| `GET /notes/{id}` | Detail page (the "Read" button target) | — |
| `GET /notes/{id}/edit` | Edit form pre-filled with the current text | — |
| `POST /create` | (form on master) | `./` (list) |
| `POST /notes/{id}/update` | (form on edit page) | `../{id}` (detail page, so the user sees what they saved) |
| `POST /notes/{id}/delete` | (form on master + detail) | `../../` (list) |
| `GET /favicon.ico` | (`lofigui.ServeFavicon`) | — |
| `GET /assets/bulma.min.css` | (`lofigui.ServeBulma`) | — |

<div class="annotation">
<strong>Master / detail, not one-page-with-ID-fields.</strong> An earlier version had four side-by-side forms on the master page where the user typed the note ID by hand. Now each row carries its own buttons that already know the ID — no copy-pasting numbers. Read becomes a real navigation to a detail page (full text, no truncation); Edit lands on a pre-filled form; Delete is a one-button POST. The example demonstrates the simplest CRUD UX you'd actually ship.
</div>

---

## The model — `model.go`

`model.go` is shared between the server and WASM builds. It owns the notes map, the CRUD operations, the flash channel, the view renderers, and the `*http.ServeMux` builder.

### State + flash

```go
const MaxNoteSize = 4 << 10 // 4 KiB cap on note text

//go:embed templates
var templateFS embed.FS

var (
    mu       sync.Mutex
    notesDB  = seedNotes()
    nextID   = 4
    flashMsg string // one-shot notification carried across PRG redirects
)

func seedNotes() map[int]string {
    return map[int]string{
        1: "First note - Welcome to the notes CRUD example!",
        2: "Second note - Add, edit, and delete notes.",
        3: "Third note - All data is stored in memory.",
    }
}
```

<div class="annotation">
<strong>State resets on restart.</strong> Package-level initializers run once per process: a Go server restart wipes all user-created notes and rebuilds the seeded three. The WASM build's "process" is the lifetime of the service worker — clearing it means unregistering the SW (the <code>demo.html</code> recovery stub does that, then redirects back to the entry point).
</div>

<div class="annotation">
<strong>The flash variable bridges the redirect.</strong> POST handlers mutate the map and write a single HTML fragment into <code>flashMsg</code>; the redirect tells the browser to GET something; the GET handler calls <code>consumeFlash()</code> which atomically reads-and-clears the variable and prepends it to the buffer. That is how the user sees "Created note #4" / "Updated note #2" / "Deleted note #3" exactly once even though the work happened during a different request.
</div>

### CRUD operations validate and flash

```go
func createNote(text string) (int, error) {
    if text == "" { return 0, fmt.Errorf("note text cannot be empty") }
    if len(text) > MaxNoteSize {
        return 0, fmt.Errorf("note is %d bytes; maximum is %d", len(text), MaxNoteSize)
    }
    mu.Lock()
    id := nextID
    notesDB[id] = text
    nextID++
    mu.Unlock()
    setFlash(fmt.Sprintf(`<div class="notification is-success">Created note #%d: %s</div>`,
        id, html.EscapeString(text)))
    return id, nil
}

func updateNote(id int, newText string) error { /* …same shape, returns error if id missing… */ }
func deleteNoteByID(id int) error              { /* …flashes is-warning on success, is-danger if missing… */ }
```

<div class="annotation">
<strong>Why <code>html.EscapeString</code>, not <code>template.HTMLEscapeString</code>?</strong> Both work. <code>html.EscapeString</code> is the canonical choice for escaping into raw HTML strings; <code>template.HTMLEscapeString</code> is a thin wrapper. The user-supplied text in the flash and the table is the only thing that needs escaping — everything else in the page is generated by the example itself.
</div>

### View renderers — three pages, one template

Three view functions print into the lofigui buffer; the single `notes.html` template just slots `{{.content}}` into the page chrome:

```go
func renderListView()         // master: table + per-row buttons + Create form
func renderDetailView(id int) // single note, Back / Edit / Delete buttons
func renderEditView(id int)   // form pre-filled with current text + Save / Cancel
```

The master view's per-row buttons use base-relative URLs so they resolve correctly under both `<base href="/">` (server) and `<base href="/06_notes_crud/wasm_demo/">` (WASM):

```html
<a href="notes/4" class="button is-info">Read</a>
<a href="notes/4/edit" class="button is-warning">Edit</a>
<form action="notes/4/delete" method="post" style="display:inline">
  <button class="button is-danger" type="submit">Delete</button>
</form>
```

### `buildMux` — every route in one place

```go
func buildMux(basePrefix string) *http.ServeMux {
    ctrl, _ := lofigui.NewControllerFromFS(templateFS, "templates", "notes.html")
    mux := http.NewServeMux()

    render := func(w http.ResponseWriter, r *http.Request) {
        ctx := ctrl.StateDict(r)
        ctx["content"] = template.HTML(lofigui.Buffer())
        ctx["base"]    = basePrefix       // → <base href="…">
        ctrl.RenderTemplate(w, ctx)
    }
    prepBuf := func() { lofigui.Reset(); if msg := consumeFlash(); msg != "" { lofigui.HTML(msg) } }

    mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
        prepBuf(); renderListView(); render(w, r)
    })
    mux.HandleFunc("GET /notes/{id}", func(w http.ResponseWriter, r *http.Request) {
        prepBuf(); renderDetailView(idOf(r)); render(w, r)
    })
    mux.HandleFunc("GET /notes/{id}/edit", func(w http.ResponseWriter, r *http.Request) {
        prepBuf(); renderEditView(idOf(r)); render(w, r)
    })

    mux.HandleFunc("POST /create", func(w http.ResponseWriter, r *http.Request) {
        _ = r.ParseForm()
        if _, err := createNote(r.FormValue("note_text")); err != nil { setErrFlash(err) }
        http.Redirect(w, r, basePrefix, http.StatusSeeOther)                       // → list
    })
    mux.HandleFunc("POST /notes/{id}/update", func(w http.ResponseWriter, r *http.Request) {
        _ = r.ParseForm()
        id := idOf(r)
        if err := updateNote(id, r.FormValue("new_text")); err != nil { setErrFlash(err) }
        http.Redirect(w, r, fmt.Sprintf("%snotes/%d", basePrefix, id), http.StatusSeeOther) // → detail page
    })
    mux.HandleFunc("POST /notes/{id}/delete", func(w http.ResponseWriter, r *http.Request) {
        if err := deleteNoteByID(idOf(r)); err != nil { setErrFlash(err) }
        http.Redirect(w, r, basePrefix, http.StatusSeeOther)                       // → list
    })

    mux.HandleFunc("GET /favicon.ico",          lofigui.ServeFavicon)
    mux.HandleFunc("GET /assets/bulma.min.css", lofigui.ServeBulma)
    return mux
}
```

<div class="annotation">
<strong>Why absolute redirects from <code>basePrefix</code>?</strong> <code>net/http.Redirect</code> rewrites a relative <code>Location</code> against <code>r.URL.Path</code> <em>before</em> writing the header. Under WASM, <code>go-wasm-http-server</code> has already <code>StripPrefix</code>'d the SW scope from <code>r.URL.Path</code> by the time the handler runs — so a redirect to <code>"./"</code> from <code>POST /create</code> resolves against <code>/create</code>, lands on <code>/</code>, and the browser navigates out of the SW scope onto the host root (where the CRUD handlers don't exist). Building the redirect URL from <code>basePrefix</code> (which is <code>"/"</code> on the server and <code>"/06_notes_crud/wasm_demo/"</code> under WASM) sidesteps the rewrite — Go leaves absolute paths alone — and keeps every redirect inside the scope.
</div>

---

## The template — `notes.html`

```html
<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <base href="{{.base}}">
  <title>Notes CRUD - Lofigui Example</title>
  <link rel="stylesheet" href="/assets/bulma.min.css">
</head>
<body>
  <div class="container">
    <h1 class="title">Notes CRUD Example</h1>
    <p class="subtitle">Simple database operations with lofigui</p>
    <hr>
    <div class="content">{{.content}}</div>
    <hr>
    <footer><a href="">Refresh</a></footer>
  </div>
</body>
</html>
```

`{{.content}}` is the lofigui buffer — flash + view-specific HTML. `<a href="">Refresh</a>` resolves against `<base>` — empty href means "the page itself," which in WASM means the home page inside the SW scope (not the host root).

`/assets/bulma.min.css` keeps its leading slash because absolute URLs ignore `<base>` — the SW scope rewrites `/assets/...` to its local copy, and the server registers `lofigui.ServeBulma` at the same path.

---

## The server — `main.go`

```go
//go:build !(js && wasm)

package main

import (
    "fmt"
    "net/http"
)

func main() {
    fmt.Println("Notes CRUD running at http://localhost:1340")
    http.ListenAndServe(":1340", buildMux("/"))
}
```

Three lines. Base prefix `"/"` because the server hosts the app at the site root.

---

## The WASM entry point — `main_wasm.go`

```go
//go:build js && wasm

package main

import (
    "strings"

    "codeberg.org/hum3/lofigui"
    wasmhttp "github.com/nlepage/go-wasm-http-server/v2"
)

func main() {
    base := strings.TrimSuffix(lofigui.WASMScopePath(), "/") + "/"
    if _, err := wasmhttp.Serve(buildMux(base)); err != nil { panic(err) }
    select {} // keep the Go runtime alive to service SW fetches
}
```

`lofigui.WASMScopePath()` reads the SW scope from `go-wasm-http-server`'s JS bridge (e.g. `/06_notes_crud/wasm_demo/`). Normalising to a single trailing slash gives a clean `<base href>`. The handlers on the mux don't know they're running in a service worker — `wasmhttp.Serve` translates `fetch` events into `*http.Request`s and forwards them to Go.

<div class="annotation">
<strong>No JS bridge for CRUD.</strong> Forms POST normally; the SW intercepts the <code>fetch</code>, hands it to Go, the handler returns 303, the browser follows the redirect, the SW forwards the GET, Go renders. The JavaScript layer is only ever the SW shim — there's no <code>syscall/js</code> exposed to the page.
</div>

---

## Same code, two targets

| Aspect | Server (`main.go`) | WASM (`main_wasm.go`) |
|--------|--------------------|------------------------|
| Build tag | `!(js && wasm)` | `js && wasm` |
| App setup | `buildMux("/")` | `buildMux(scopePath)` |
| Serving runtime | `http.ListenAndServe(":1340", mux)` | `wasmhttp.Serve(mux)` + `select{}` |
| Request shape | `http.Request` / `http.ResponseWriter` | `http.Request` / `http.ResponseWriter` |
| Form POST | HTTP request → Go handler | `fetch` intercepted by SW → Go handler |
| `<base href>` | `"/"` | `"/06_notes_crud/wasm_demo/"` |
| Redirect targets | `basePrefix + …` → `/…` | `basePrefix + …` → `/06_notes_crud/wasm_demo/…` |
| State lifetime | until process exits | until SW unregistered |

`buildMux`, `model.go`, and `notes.html` are the same source for both builds. The entry-point files exist purely because `syscall/js` (transitively imported by `go-wasm-http-server`) doesn't link on non-WASM targets, and `http.ListenAndServe` is a runtime no-op under WASM.

---

## The capture *is* the integration test

`task docs:capture:06` does both jobs in one shell block: it depends on `clean-ports` (so a leftover server from a previous run doesn't block port 1340), runs the CRUD sequence, and asserts both the HTTP behaviour and the rendered SVG content:

```bash
check_303() {
  code=$(curl -sf -o /dev/null -w '%{http_code}' "$@")
  [ "$code" = "303" ] || { echo "FAIL: expected 303, got $code"; exit 1; }
}

# Two creates + delete seed #3 + update seed #2 (last action — its flash shows on the populated screenshot)
check_303 -X POST -d 'note_text=Buy milk and eggs'                                               http://localhost:1340/create
check_303 -X POST -d 'note_text=Pay electricity bill'                                            http://localhost:1340/create
check_303 -X POST                                                                                 http://localhost:1340/notes/3/delete
check_303 -X POST --data-urlencode 'new_text=Add, edit, and delete notes — and now WASM!'         http://localhost:1340/notes/2/update

url2svg --url http://localhost:1340/         -o docs/06_populated.svg     # master, "Updated note #2" flash
url2svg --url http://localhost:1340/notes/1  -o docs/06_detail.svg        # detail page

# 4 KiB cap is enforced
big=$(printf 'x%.0s' $(seq 1 4097))
check_303 -X POST --data-urlencode "note_text=${big}" http://localhost:1340/create
curl -sf http://localhost:1340/ | grep -q "maximum is 4096" || { echo "FAIL: oversize flash missing"; exit 1; }

# Non-existent ID is a no-op + flash
check_303 -X POST http://localhost:1340/notes/9999/delete
curl -sf http://localhost:1340/ | grep -q "note #9999 not found" || { echo "FAIL: not-found flash missing"; exit 1; }

# Final SVG content asserts
for needle in "Buy milk and eggs" "Pay electricity bill" "now WASM" "Updated note #2"; do
  grep -q "$needle" docs/06_populated.svg || { echo "FAIL: '$needle' missing"; exit 1; }
done
grep -q "Welcome to the notes CRUD example" docs/06_detail.svg || { echo "FAIL: detail text missing"; exit 1; }
```

<div class="annotation">
<strong>Authenticity over hand-crafting.</strong> The screenshots aren't posed — they are the actual server's output after a real CRUD sequence. Every <code>check_303</code> call asserts the POST handler returned <code>303 See Other</code>; if any handler regresses (returns <code>200</code>, <code>500</code>, etc.) the capture task fails. The <code>grep</code> assertions on the final SVGs verify that the create/update/delete operations actually mutated the visible state and that the validation flashes fired. Adding new CRUD operations means adding to this sequence — the documentation, the screenshots, and the test grow together.
</div>

After the sequence:

- Note #1 — original, full text visible on the detail page (the screenshot)
- Note #2 — updated (text now ends "…and now WASM!")
- Note #3 — deleted (gone from the table, total count down by one)
- Note #4 — newly created ("Buy milk and eggs")
- Note #5 — newly created ("Pay electricity bill")
- Plus a flash: **Updated note #2** at the top of the master view

---

## Running it

```bash
task go-example:06       # Go server  → http://localhost:1340
task example-06          # Python      → http://localhost:1340

task docs:capture:06     # capture all three SVGs and run the integration test
```

---

## Where to go next

- **Add HTMX** — [09 Water Tank HTMX](../09_water_tank_htmx/) drops the redirect entirely and uses HTMX `hx-post` to swap a fragment in place. The flash-message variable disappears.
- **Multiple pages** — [03 Style Sampler](../03_style_sampler/) shows the same "shared mux + WASM" story with multiple routes and template inheritance.
- **Persistence** — [11 Water Tank Storage](../11_water_tank_storage/) shows a WASM frontend pointed at a separate Go API server using SeaweedFS; the in-memory map here is what graduates into a real backend in 11.

---

## Links

- [Launch Demo](wasm_demo/)
- [Source on Codeberg](https://codeberg.org/hum3/lofigui/src/branch/main/examples/06_notes_crud)
