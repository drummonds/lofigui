package main

import (
	"log"
	"net/http"

	"github.com/drummonds/lofigui"
	"github.com/flosch/pongo2/v6"
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
	template *pongo2.Template
}

func NewController() *Controller {
	// Load Jinja2-style template using pongo2
	tmpl, err := pongo2.FromFile("../templates/hello.html")
	if err != nil {
		log.Fatalf("Failed to load template: %v", err)
	}
	return &Controller{
		template: tmpl,
	}
}

func (ctrl *Controller) handleRoot(w http.ResponseWriter, r *http.Request) {
	// Reset buffer and run model
	lofigui.Reset()
	model()

	// Prepare template data using lowercase 'results' to match Python version
	// pongo2 automatically marks content as safe when using the 'safe' filter in templates
	data := pongo2.Context{
		"results": lofigui.Buffer(),
		"refresh": "", // Empty refresh for now
	}

	// Render template
	if err := ctrl.template.ExecuteWriter(data, w); err != nil {
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
