package main

import (
	"time"

	"codeberg.org/hum3/lofigui"
)

// model contains the business logic — shared by server and WASM builds.
// Just like a terminal program: print output and sleep between steps.
func model(app *lofigui.App) {
	lofigui.Print("Hello world.")
	for i := 0; i < 5; i++ {
		app.Sleep(1 * time.Second)
		lofigui.Printf("Count %d", i)
	}
	lofigui.Print("Done.")
}
