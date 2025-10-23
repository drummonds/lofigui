package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/drummonds/lofigui"
)

// Model function - contains business logic
func model(ctrl *lofigui.Controller) {
	lofigui.Print("Hello world.")
	for i := 0; i < 5; i++ {
		time.Sleep(1 * time.Second)
		lofigui.Print(fmt.Sprintf("Count %d", i))
	}
	lofigui.Markdown("<a href='/'>Restart</a>")
	lofigui.Print("Done.")
	ctrl.EndAction()
}

func main() {
	// Create controller with custom template directory and settings
	// The template directory can be anywhere, not just the default location
	ctrl, err := lofigui.NewController(lofigui.ControllerConfig{
		TemplatePath: "../templates/hello.html", // Custom location
		RefreshTime:  1,                         // Refresh every 1 second
		DisplayURL:   "/display",                // Where to show results
	})
	if err != nil {
		log.Fatalf("Failed to create controller: %v", err)
	}

	// Root endpoint - starts the action
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		ctrl.HandleRoot(w, r, model, true)
	})

	// Display endpoint - shows progress
	http.HandleFunc("/display", func(w http.ResponseWriter, r *http.Request) {
		ctrl.HandleDisplay(w, r, nil)
	})

	// Favicon endpoint
	http.HandleFunc("/favicon.ico", lofigui.ServeFavicon)

	addr := ":1340"
	log.Printf("Starting server on http://localhost%s", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}
