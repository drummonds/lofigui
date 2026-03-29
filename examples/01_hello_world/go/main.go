//go:build !(js && wasm)

package main

import (
	"log"
	"net/http"
	"time"

	"codeberg.org/hum3/lofigui"
)

// Model function - contains business logic.
// Just like a terminal program: print output and sleep between steps.
func model(app *lofigui.App) {
	lofigui.Print("Hello world.")
	for i := 0; i < 5; i++ {
		app.Sleep(1 * time.Second)
		lofigui.Printf("Count %d", i)
	}
	lofigui.Print("Done.")
}

func main() {
	app := lofigui.NewApp()
	http.HandleFunc("/", app.Handle(model))
	if err := app.ListenAndServe(":1340", nil); err != nil {
		log.Fatal(err)
	}
}
