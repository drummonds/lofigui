package main

import (
	"html/template"
	"log"
	"net/http"

	"github.com/drummonds/lofigui"
)

// Model function - contains business logic
func model() {
	lofigui.Print("Hello world from Go!")
	lofigui.Print("This is the lofigui Go version.")

	lofigui.Markdown("## Key Features\n\n- **Simple API**: Similar to Python version\n- **Type-safe**: Go's strong typing\n- **Fast**: Compiled performance\n- **Concurrent**: Safe for concurrent use")

	// Add a table
	data := [][]string{
		{"Alice", "30", "Engineer"},
		{"Bob", "25", "Designer"},
		{"Carol", "35", "Manager"},
	}
	lofigui.Table(data, lofigui.WithHeader([]string{"Name", "Age", "Role"}))
}

// Controller manages state and routing
type Controller struct {
	templates *template.Template
}

func NewController() *Controller {
	tmpl := template.Must(template.ParseFiles("templates/hello.html"))
	return &Controller{
		templates: tmpl,
	}
}

func (ctrl *Controller) handleRoot(w http.ResponseWriter, r *http.Request) {
	// Reset buffer and run model
	lofigui.Reset()
	model()

	// Prepare template data
	data := struct {
		Results string
	}{
		Results: lofigui.Buffer(),
	}

	// Render template
	if err := ctrl.templates.ExecuteTemplate(w, "hello.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func main() {
	ctrl := NewController()

	http.HandleFunc("/", ctrl.handleRoot)
	http.HandleFunc("/favicon.ico", lofigui.ServeFavicon)

	addr := ":1340"
	log.Printf("Starting server on http://localhost%s", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}
