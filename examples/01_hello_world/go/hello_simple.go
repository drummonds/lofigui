package main

import (
	"log"
	"net/http"
	"time"

	"github.com/drummonds/lofigui"
)

// This example shows the simplest possible usage of lofigui Controller
// using the convenience function NewControllerFromDir

func modelSimple(ctrl *lofigui.Controller) {
	lofigui.Print("Hello from simple example!")
	for i := 0; i < 3; i++ {
		time.Sleep(1 * time.Second)
		lofigui.Printf("Step %d of 3", i+1)
	}
	lofigui.Markdown("**Done!** <a href='/'>Restart</a>")
	ctrl.EndAction()
}

func main() {
	// Create controller using directory + filename convenience function
	ctrl, err := lofigui.NewControllerFromDir("../templates", "hello.html", 1)
	if err != nil {
		log.Fatalf("Failed to create controller: %v", err)
	}

	// Setup routes using the helper methods
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		ctrl.HandleRoot(w, r, modelSimple, true)
	})

	http.HandleFunc("/display", ctrl.ServeHTTP) // Can use ServeHTTP directly!

	http.HandleFunc("/favicon.ico", lofigui.ServeFavicon)

	addr := ":1341"
	log.Printf("Starting simple server on http://localhost%s", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}
