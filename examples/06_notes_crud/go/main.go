package main

import (
	"fmt"
	"log"
	"net/http"
	"sort"
	"strconv"

	"github.com/drummonds/lofigui"
)

// Simple in-memory notes database
var notesDB = map[int]string{
	1: "First note - Welcome to the notes CRUD example!",
	2: "Second note - Add, edit, and delete notes.",
	3: "Third note - All data is stored in memory.",
}
var nextID = 4

// listNotes displays all notes in a table
func listNotes() {
	lofigui.Print("<h2>Notes Database</h2>")

	if len(notesDB) == 0 {
		lofigui.Print("<p>No notes in database.</p>")
		return
	}

	// Create sorted list of IDs
	ids := make([]int, 0, len(notesDB))
	for id := range notesDB {
		ids = append(ids, id)
	}
	sort.Ints(ids)

	// Create table data
	tableData := [][]string{}
	for _, id := range ids {
		noteText := notesDB[id]
		// Truncate long notes for display
		displayText := noteText
		if len(noteText) > 50 {
			displayText = noteText[:50] + "..."
		}
		tableData = append(tableData, []string{fmt.Sprintf("%d", id), displayText})
	}

	lofigui.Table(tableData, lofigui.WithHeader([]string{"ID", "Note"}))
	lofigui.Print(fmt.Sprintf("<p>Total notes: %d</p>", len(notesDB)))
}

// createNote creates a new note
func createNote(noteText string) {
	notesDB[nextID] = noteText
	lofigui.Print(fmt.Sprintf("<p class='notification is-success'>Created note #%d: %s</p>", nextID, noteText))
	nextID++
}

// readNote reads a specific note
func readNote(noteID int) {
	if text, exists := notesDB[noteID]; exists {
		lofigui.Print(fmt.Sprintf("<p><strong>Note #%d:</strong> %s</p>", noteID, text))
	} else {
		lofigui.Print(fmt.Sprintf("<p class='notification is-danger'>Note #%d not found.</p>", noteID))
	}
}

// updateNote updates an existing note
func updateNote(noteID int, newText string) {
	if oldText, exists := notesDB[noteID]; exists {
		notesDB[noteID] = newText
		lofigui.Print(fmt.Sprintf("<p class='notification is-info'>Updated note #%d</p>", noteID))
		lofigui.Print(fmt.Sprintf("<p>Old: %s</p>", oldText))
		lofigui.Print(fmt.Sprintf("<p>New: %s</p>", newText))
	} else {
		lofigui.Print(fmt.Sprintf("<p class='notification is-danger'>Note #%d not found.</p>", noteID))
	}
}

// deleteNote deletes a note
func deleteNote(noteID int) {
	if text, exists := notesDB[noteID]; exists {
		delete(notesDB, noteID)
		lofigui.Print(fmt.Sprintf("<p class='notification is-warning'>Deleted note #%d: %s</p>", noteID, text))
	} else {
		lofigui.Print(fmt.Sprintf("<p class='notification is-danger'>Note #%d not found.</p>", noteID))
	}
}

func main() {
	// Create an App which provides safe controller management
	app := lofigui.NewApp()
	app.Version = "Notes CRUD v1.0"

	// Create controller with custom template directory
	ctrl, err := lofigui.NewController(lofigui.ControllerConfig{
		Name:         "Notes CRUD Controller",
		TemplatePath: "../templates/notes.html",
	})
	if err != nil {
		log.Fatalf("Failed to create controller: %v", err)
	}

	app.SetController(ctrl)

	// Root endpoint - display notes interface
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		lofigui.Reset()
		listNotes()

		// Add form for creating new notes
		lofigui.Markdown(`
		<div class="box">
			<h3 class="title is-4">Create New Note</h3>
			<form action="/create" method="post">
				<div class="field">
					<div class="control">
						<input class="input" type="text" name="note_text" placeholder="Enter note text" required>
					</div>
				</div>
				<div class="field">
					<div class="control">
						<button class="button is-primary" type="submit">Create Note</button>
					</div>
				</div>
			</form>
		</div>
		`)

		// Add forms for other CRUD operations
		lofigui.Markdown(`
		<div class="columns">
			<div class="column">
				<div class="box">
					<h3 class="title is-4">Read Note</h3>
					<form action="/read" method="post">
						<div class="field has-addons">
							<div class="control is-expanded">
								<input class="input" type="number" name="note_id" placeholder="Note ID" required>
							</div>
							<div class="control">
								<button class="button is-info" type="submit">Read</button>
							</div>
						</div>
					</form>
				</div>
			</div>

			<div class="column">
				<div class="box">
					<h3 class="title is-4">Update Note</h3>
					<form action="/update" method="post">
						<div class="field">
							<div class="control">
								<input class="input" type="number" name="note_id" placeholder="Note ID" required>
							</div>
						</div>
						<div class="field has-addons">
							<div class="control is-expanded">
								<input class="input" type="text" name="new_text" placeholder="New text" required>
							</div>
							<div class="control">
								<button class="button is-warning" type="submit">Update</button>
							</div>
						</div>
					</form>
				</div>
			</div>

			<div class="column">
				<div class="box">
					<h3 class="title is-4">Delete Note</h3>
					<form action="/delete" method="post">
						<div class="field has-addons">
							<div class="control is-expanded">
								<input class="input" type="number" name="note_id" placeholder="Note ID" required>
							</div>
							<div class="control">
								<button class="button is-danger" type="submit">Delete</button>
							</div>
						</div>
					</form>
				</div>
			</div>
		</div>
		`)

		context := ctrl.StateDict(r)
		context["content"] = lofigui.Buffer()
		ctrl.RenderTemplate(w, context)
	})

	// Create endpoint
	http.HandleFunc("/create", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}

		r.ParseForm()
		noteText := r.FormValue("note_text")
		if noteText != "" {
			createNote(noteText)
		}
		http.Redirect(w, r, "/", http.StatusSeeOther)
	})

	// Read endpoint
	http.HandleFunc("/read", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}

		r.ParseForm()
		noteID, err := strconv.Atoi(r.FormValue("note_id"))
		if err == nil {
			lofigui.Reset()
			readNote(noteID)
			listNotes()
		}
		http.Redirect(w, r, "/", http.StatusSeeOther)
	})

	// Update endpoint
	http.HandleFunc("/update", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}

		r.ParseForm()
		noteID, err := strconv.Atoi(r.FormValue("note_id"))
		newText := r.FormValue("new_text")
		if err == nil && newText != "" {
			lofigui.Reset()
			updateNote(noteID, newText)
			listNotes()
		}
		http.Redirect(w, r, "/", http.StatusSeeOther)
	})

	// Delete endpoint
	http.HandleFunc("/delete", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}

		r.ParseForm()
		noteID, err := strconv.Atoi(r.FormValue("note_id"))
		if err == nil {
			lofigui.Reset()
			deleteNote(noteID)
			listNotes()
		}
		http.Redirect(w, r, "/", http.StatusSeeOther)
	})

	// Favicon endpoint
	http.HandleFunc("/favicon.ico", lofigui.ServeFavicon)

	addr := ":1346"
	log.Printf("Starting Notes CRUD server on http://localhost%s", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}
