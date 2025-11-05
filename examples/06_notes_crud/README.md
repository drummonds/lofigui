# Notes CRUD Example

A simple demonstration of Create, Read, Update, Delete (CRUD) operations using lofigui.

## Features

- **Simple Database**: In-memory dictionary/map with numeric IDs as keys and text strings as values
- **CRUD Operations**: Full Create, Read, Update, Delete functionality
- **Interactive Interface**: Web-based forms for all operations
- **Data Table Display**: Shows all notes in a formatted table
- **Both Languages**: Implementations in Python and Go

## Database Structure

The database is as simple as possible:
- **Keys**: Numeric IDs (1, 2, 3, ...)
- **Values**: Text strings (the note content)
- **Storage**: In-memory (no persistence)

## Running the Example

### Python Version

```bash
cd examples/06_notes_crud/python
uv pip install -e .
python notes.py
```

Or from the root directory:
```bash
task py06
```

The application will start on http://localhost:1346

### Go Version

```bash
cd examples/06_notes_crud/go
go run main.go
```

Or from the root directory:
```bash
task go06
```

The application will start on http://localhost:1346

## Using the Application

The interface provides four main operations:

### 1. Create - Add New Notes
- Enter text in the "Create New Note" form
- Click "Create Note"
- A new note is added with an auto-incremented ID

### 2. Read - View a Specific Note
- Enter a note ID in the "Read Note" form
- Click "Read"
- The note content is displayed

### 3. Update - Modify Existing Notes
- Enter the note ID to update
- Enter the new text
- Click "Update"
- Shows both old and new text

### 4. Delete - Remove Notes
- Enter the note ID to delete
- Click "Delete"
- The note is removed from the database

## Initial Data

The database starts with three example notes:
1. "First note - Welcome to the notes CRUD example!"
2. "Second note - Add, edit, and delete notes."
3. "Third note - All data is stored in memory."

## Technical Details

### Python Implementation
- Uses FastAPI for the web framework
- Form handling with FastAPI's `Form` dependency
- In-memory dictionary for storage
- Redirect pattern after POST operations (Post/Redirect/Get)

### Go Implementation
- Uses standard library `net/http`
- Form parsing with `r.ParseForm()`
- In-memory map for storage
- HTTP 303 redirects after POST operations

### Shared Features
- Bulma CSS for styling
- lofigui table rendering
- Controller pattern for template management
- No external database required

## Learning Points

This example demonstrates:
1. **CRUD Operations**: The four fundamental database operations
2. **Form Handling**: POST requests with form data
3. **State Management**: In-memory data structure
4. **User Feedback**: Success/error messages for each operation
5. **Data Display**: Formatted table rendering
6. **Web Patterns**: Post/Redirect/Get pattern to prevent duplicate submissions

## Extending the Example

This example can be extended to:
- Add persistence (save to file/database)
- Implement search functionality
- Add timestamps to notes
- Support markdown in note content
- Add categories or tags
- Implement user authentication
- Add pagination for large datasets

## File Structure

```
06_notes_crud/
├── README.md                 # This file
├── templates/
│   └── notes.html           # Shared HTML template
├── python/
│   ├── notes.py             # Python implementation
│   └── pyproject.toml       # Python dependencies
└── go/
    ├── main.go              # Go implementation
    └── go.mod               # Go dependencies
```

## License

This example is part of the lofigui project and shares the same license.
