"""
Notes CRUD Example using lofigui.

This example demonstrates:
1. Create, Read, Update, Delete operations on a simple notes database
2. In-memory database (dict with numeric keys)
3. Interactive web interface for managing notes
4. Controller pattern with custom template
"""

import lofigui as lg
from typing import Dict

from fastapi import Request, BackgroundTasks, Form
from fastapi.responses import HTMLResponse, RedirectResponse
import uvicorn


# Simple in-memory notes database: {id: note_text}
notes_db: Dict[int, str] = {
    1: "First note - Welcome to the notes CRUD example!",
    2: "Second note - Add, edit, and delete notes.",
    3: "Third note - All data is stored in memory.",
}
next_id = 4


# Create controller with custom template directory
controller = lg.Controller()

# Use create_app which automatically includes favicon route
app = lg.create_app(template_dir="../templates", controller=controller)


def list_notes():
    """Display all notes in a table"""
    lg.print("<h2>Notes Database</h2>")

    if not notes_db:
        lg.print("<p>No notes in database.</p>")
        return

    # Create table data
    table_data = []
    for note_id in sorted(notes_db.keys()):
        note_text = notes_db[note_id]
        # Truncate long notes for display
        display_text = note_text[:50] + "..." if len(note_text) > 50 else note_text
        table_data.append([str(note_id), display_text])

    lg.table(table_data, header=["ID", "Note"])

    lg.print(f"<p>Total notes: {len(notes_db)}</p>")


def create_note(note_text: str):
    """Create a new note"""
    global next_id
    notes_db[next_id] = note_text
    lg.print(f"<p class='notification is-success'>Created note #{next_id}: {note_text}</p>")
    next_id += 1


def read_note(note_id: int):
    """Read a specific note"""
    if note_id in notes_db:
        lg.print(f"<p><strong>Note #{note_id}:</strong> {notes_db[note_id]}</p>")
    else:
        lg.print(f"<p class='notification is-danger'>Note #{note_id} not found.</p>")


def update_note(note_id: int, new_text: str):
    """Update an existing note"""
    if note_id in notes_db:
        old_text = notes_db[note_id]
        notes_db[note_id] = new_text
        lg.print(f"<p class='notification is-info'>Updated note #{note_id}</p>")
        lg.print(f"<p>Old: {old_text}</p>")
        lg.print(f"<p>New: {new_text}</p>")
    else:
        lg.print(f"<p class='notification is-danger'>Note #{note_id} not found.</p>")


def delete_note(note_id: int):
    """Delete a note"""
    if note_id in notes_db:
        deleted_text = notes_db.pop(note_id)
        lg.print(f"<p class='notification is-warning'>Deleted note #{note_id}: {deleted_text}</p>")
    else:
        lg.print(f"<p class='notification is-danger'>Note #{note_id} not found.</p>")


@app.get("/", response_class=HTMLResponse)
async def root(request: Request):
    """Root endpoint - display the notes interface"""
    lg.reset()
    list_notes()

    # Add form for creating new notes
    lg.markdown("""
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
    """)

    # Add forms for other CRUD operations
    lg.markdown("""
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
    """)

    return app.template_response(request, "notes.html")


@app.post("/create", response_class=HTMLResponse)
async def create(note_text: str = Form(...)):
    """Create a new note"""
    lg.reset()
    create_note(note_text)
    list_notes()
    return RedirectResponse(url="/", status_code=303)


@app.post("/read", response_class=HTMLResponse)
async def read(note_id: int = Form(...)):
    """Read a specific note"""
    lg.reset()
    read_note(note_id)
    list_notes()
    return RedirectResponse(url="/", status_code=303)


@app.post("/update", response_class=HTMLResponse)
async def update(note_id: int = Form(...), new_text: str = Form(...)):
    """Update a note"""
    lg.reset()
    update_note(note_id, new_text)
    list_notes()
    return RedirectResponse(url="/", status_code=303)


@app.post("/delete", response_class=HTMLResponse)
async def delete(note_id: int = Form(...)):
    """Delete a note"""
    lg.reset()
    delete_note(note_id)
    list_notes()
    return RedirectResponse(url="/", status_code=303)


if __name__ == "__main__":
    uvicorn.run("notes:app", host="127.0.0.1", port=1346, reload=True)
