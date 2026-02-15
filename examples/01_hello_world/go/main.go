package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/drummonds/lofigui"
)

// Model function - contains business logic
// The model receives the App, which manages the singleton active model state
func model(app *lofigui.App) {
	lofigui.Print("Hello world.")
	for i := 0; i < 5; i++ {
		time.Sleep(1 * time.Second)
		lofigui.Print(fmt.Sprintf("Count %d", i))
	}
	lofigui.Markdown("<a href='/'>Restart</a>")
	lofigui.Print("Done.")
	app.EndAction() // Signal that the action is complete (app-level state)
}

func main() {
	// Create an App which provides safe controller management
	// The App uses composition to integrate the controller
	app := lofigui.NewApp()
	app.Version = "Hello World v1.0"

	// Create controller with custom template directory and settings
	// The template directory can be anywhere, not just the default location
	// The controller is integrated into the app via composition
	ctrl, err := lofigui.NewController(lofigui.ControllerConfig{
		Name:         "Hello World Controller",  // Name displayed in app
		TemplatePath: "../templates/hello.html", // Custom location
	})
	if err != nil {
		log.Fatalf("Failed to create controller: %v", err)
	}

	// Set the controller in the app (composition pattern)
	// App provides safe controller replacement - if we later replace the controller,
	// it will ensure any running action is stopped first
	app.SetController(ctrl)

	// Configure app-level settings (singleton active model)
	app.SetRefreshTime(1) // Refresh every 1 second while action is running
	app.SetDisplayURL("/display")

	// Root endpoint - starts the action
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		app.HandleRoot(w, r, model, true)
	})

	// Display endpoint - shows progress
	http.HandleFunc("/display", func(w http.ResponseWriter, r *http.Request) {
		app.HandleDisplay(w, r)
	})

	// Favicon endpoint
	http.HandleFunc("/favicon.ico", lofigui.ServeFavicon)

	addr := ":1340"
	log.Printf("Starting server on http://localhost%s", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}
