package main

import (
	"embed"
	"fmt"
	"html"
	"html/template"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"

	"codeberg.org/hum3/lofigui"
)

//go:embed templates
var templateFS embed.FS

// MaxNoteSize caps each note's text in bytes. Enforced server-side on POST
// and mirrored as the textarea's maxlength attribute on the form.
const MaxNoteSize = 4 << 10 // 4 KiB

// Each process starts with the same three seeded notes — package-level
// initializers run once per process, so a Go server restart wipes all
// user-created state. The WASM build's "process" is the lifetime of the
// service worker; clearing it means unregistering the SW (via demo.html).
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

func setFlash(h string) {
	mu.Lock()
	flashMsg = h
	mu.Unlock()
}

func consumeFlash() string {
	mu.Lock()
	m := flashMsg
	flashMsg = ""
	mu.Unlock()
	return m
}

// CRUD ops — pure state mutators that also stash a flash. The flash is the
// only piece of user-visible feedback that survives the redirect.

func createNote(text string) (int, error) {
	if text == "" {
		return 0, fmt.Errorf("note text cannot be empty")
	}
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

func updateNote(id int, newText string) error {
	if newText == "" {
		return fmt.Errorf("note text cannot be empty")
	}
	if len(newText) > MaxNoteSize {
		return fmt.Errorf("note is %d bytes; maximum is %d", len(newText), MaxNoteSize)
	}
	mu.Lock()
	old, ok := notesDB[id]
	if ok {
		notesDB[id] = newText
	}
	mu.Unlock()
	if !ok {
		return fmt.Errorf("note #%d not found", id)
	}
	setFlash(fmt.Sprintf(
		`<div class="notification is-warning"><strong>Updated note #%d</strong><br>Old: %s<br>New: %s</div>`,
		id, html.EscapeString(old), html.EscapeString(newText),
	))
	return nil
}

func deleteNoteByID(id int) error {
	mu.Lock()
	text, ok := notesDB[id]
	if ok {
		delete(notesDB, id)
	}
	mu.Unlock()
	if !ok {
		return fmt.Errorf("note #%d not found", id)
	}
	setFlash(fmt.Sprintf(`<div class="notification is-warning">Deleted note #%d: %s</div>`,
		id, html.EscapeString(text)))
	return nil
}

// Views — each renders into the lofigui buffer. The single template
// (notes.html) just slots {{.content}} into the page chrome.

var createFormHTML = fmt.Sprintf(`
<div class="box">
  <h3 class="title is-4">Create New Note</h3>
  <form action="create" method="post">
    <div class="field">
      <label class="label">Note text <span class="has-text-grey is-size-7">(max %[1]d characters)</span></label>
      <div class="control">
        <textarea class="textarea" name="note_text" maxlength="%[1]d" rows="3" placeholder="Enter note text" required></textarea>
      </div>
    </div>
    <div class="field"><div class="control">
      <button class="button is-primary" type="submit">Create Note</button>
    </div></div>
  </form>
</div>`, MaxNoteSize)

func renderListView() {
	mu.Lock()
	ids := make([]int, 0, len(notesDB))
	for id := range notesDB {
		ids = append(ids, id)
	}
	sort.Ints(ids)
	type row struct {
		id      int
		preview string
	}
	rows := make([]row, 0, len(ids))
	for _, id := range ids {
		text := notesDB[id]
		display := text
		if len(text) > 60 {
			display = text[:60] + "…"
		}
		rows = append(rows, row{id, display})
	}
	total := len(notesDB)
	mu.Unlock()

	lofigui.HTML(`<h2 class="title is-3">Notes Database</h2>`)

	if total == 0 {
		lofigui.HTML(`<p>No notes yet. Use the form below to create one.</p>`)
		lofigui.HTML(createFormHTML)
		return
	}

	var sb strings.Builder
	sb.WriteString(`<table class="table is-fullwidth is-striped"><thead><tr>`)
	sb.WriteString(`<th style="width: 4em">ID</th><th>Note</th><th style="width: 18em">Actions</th>`)
	sb.WriteString(`</tr></thead><tbody>`)
	for _, r := range rows {
		fmt.Fprintf(&sb, `<tr>
<td>%d</td>
<td>%s</td>
<td><div class="buttons are-small">
<a href="notes/%d" class="button is-info">Read</a>
<a href="notes/%d/edit" class="button is-warning">Edit</a>
<form action="notes/%d/delete" method="post" style="display:inline">
<button class="button is-danger" type="submit">Delete</button>
</form>
</div></td>
</tr>`, r.id, html.EscapeString(r.preview), r.id, r.id, r.id)
	}
	sb.WriteString(`</tbody></table>`)
	lofigui.HTML(sb.String())
	lofigui.HTML(fmt.Sprintf(`<p class="has-text-grey is-size-7 mb-4">Total notes: %d</p>`, total))
	lofigui.HTML(createFormHTML)
}

func renderDetailView(id int) {
	mu.Lock()
	text, ok := notesDB[id]
	mu.Unlock()
	if !ok {
		lofigui.HTML(fmt.Sprintf(
			`<div class="notification is-danger">Note #%d not found. <a href="">Back to list</a></div>`, id))
		return
	}
	lofigui.HTML(fmt.Sprintf(`
<h2 class="title is-3">Note #%d</h2>
<div class="box"><pre style="white-space: pre-wrap; font-family: inherit; background: transparent; padding: 0">%s</pre></div>
<div class="buttons">
  <a href="" class="button">&larr; Back to list</a>
  <a href="notes/%d/edit" class="button is-warning">Edit</a>
  <form action="notes/%d/delete" method="post" style="display:inline">
    <button class="button is-danger" type="submit">Delete</button>
  </form>
</div>`, id, html.EscapeString(text), id, id))
}

func renderEditView(id int) {
	mu.Lock()
	text, ok := notesDB[id]
	mu.Unlock()
	if !ok {
		lofigui.HTML(fmt.Sprintf(
			`<div class="notification is-danger">Note #%d not found. <a href="">Back to list</a></div>`, id))
		return
	}
	lofigui.HTML(fmt.Sprintf(`
<h2 class="title is-3">Edit Note #%d</h2>
<form action="notes/%d/update" method="post">
  <div class="field">
    <label class="label">Note text <span class="has-text-grey is-size-7">(max %d characters)</span></label>
    <div class="control">
      <textarea class="textarea" name="new_text" maxlength="%d" rows="6" required>%s</textarea>
    </div>
  </div>
  <div class="buttons">
    <button class="button is-primary" type="submit">Save</button>
    <a href="notes/%d" class="button">Cancel</a>
  </div>
</form>`, id, id, MaxNoteSize, MaxNoteSize, html.EscapeString(text), id))
}

// buildMux wires the routes for both server and WASM builds. basePrefix is
// rendered into <base href="..."> so relative URLs in the page resolve
// against the right root (site root for the server, the SW scope for WASM).
//
// POST handlers redirect with absolute paths built from basePrefix rather
// than relative paths like "./" or "../". Why: net/http's Redirect
// absolutifies relative URLs against r.URL.Path *before* writing the
// Location header, and go-wasm-http-server has already stripped basePrefix
// from r.URL.Path by the time the handler runs. Under WASM this turns "./"
// into "/" and the browser navigates out of the SW scope. Sending the
// fully-qualified path keeps every redirect inside the scope.
func buildMux(basePrefix string) *http.ServeMux {
	ctrl, err := lofigui.NewControllerFromFS(templateFS, "templates", "notes.html")
	if err != nil {
		panic(fmt.Sprintf("template notes.html: %v", err))
	}
	mux := http.NewServeMux()

	render := func(w http.ResponseWriter, r *http.Request) {
		ctx := ctrl.StateDict(r)
		ctx["content"] = template.HTML(lofigui.Buffer())
		ctx["base"] = basePrefix
		ctrl.RenderTemplate(w, ctx)
	}
	prepBuf := func() {
		lofigui.Reset()
		if msg := consumeFlash(); msg != "" {
			lofigui.HTML(msg)
		}
	}
	pathID := func(r *http.Request) (int, bool) {
		id, err := strconv.Atoi(r.PathValue("id"))
		return id, err == nil
	}

	// GET / — master list with per-row Read/Edit/Delete buttons + create form
	mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		prepBuf()
		renderListView()
		render(w, r)
	})

	// GET /notes/{id} — detail page (the "Read" button target)
	mux.HandleFunc("GET /notes/{id}", func(w http.ResponseWriter, r *http.Request) {
		prepBuf()
		if id, ok := pathID(r); ok {
			renderDetailView(id)
		} else {
			lofigui.HTML(`<div class="notification is-danger">Invalid note ID. <a href="">Back to list</a></div>`)
		}
		render(w, r)
	})

	// GET /notes/{id}/edit — edit form pre-filled with current text
	mux.HandleFunc("GET /notes/{id}/edit", func(w http.ResponseWriter, r *http.Request) {
		prepBuf()
		if id, ok := pathID(r); ok {
			renderEditView(id)
		} else {
			lofigui.HTML(`<div class="notification is-danger">Invalid note ID. <a href="">Back to list</a></div>`)
		}
		render(w, r)
	})

	// POST /create — redirect to basePrefix lands on GET /
	mux.HandleFunc("POST /create", func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		if _, err := createNote(r.FormValue("note_text")); err != nil {
			setFlash(fmt.Sprintf(`<div class="notification is-danger">%s</div>`, html.EscapeString(err.Error())))
		}
		http.Redirect(w, r, basePrefix, http.StatusSeeOther)
	})

	// POST /notes/{id}/update — redirect to the detail page so the user sees
	// the updated text after saving.
	mux.HandleFunc("POST /notes/{id}/update", func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		id, ok := pathID(r)
		if !ok {
			http.Redirect(w, r, basePrefix, http.StatusSeeOther)
			return
		}
		if err := updateNote(id, r.FormValue("new_text")); err != nil {
			setFlash(fmt.Sprintf(`<div class="notification is-danger">%s</div>`, html.EscapeString(err.Error())))
		}
		http.Redirect(w, r, fmt.Sprintf("%snotes/%d", basePrefix, id), http.StatusSeeOther)
	})

	// POST /notes/{id}/delete — redirect to basePrefix lands back at the list
	mux.HandleFunc("POST /notes/{id}/delete", func(w http.ResponseWriter, r *http.Request) {
		if id, ok := pathID(r); ok {
			if err := deleteNoteByID(id); err != nil {
				setFlash(fmt.Sprintf(`<div class="notification is-danger">%s</div>`, html.EscapeString(err.Error())))
			}
		}
		http.Redirect(w, r, basePrefix, http.StatusSeeOther)
	})

	mux.HandleFunc("GET /favicon.ico", lofigui.ServeFavicon)
	mux.HandleFunc("GET /assets/bulma.min.css", lofigui.ServeBulma)
	return mux
}
