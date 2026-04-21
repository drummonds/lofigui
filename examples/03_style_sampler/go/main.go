//go:build !(js && wasm)

package main

import (
	"fmt"
	"net/http"

	"codeberg.org/hum3/lofigui"
)

func main() {
	app := lofigui.NewApp()
	app.SetRefreshTime(1)
	app.RunModel(model) // kick off the teletype in a background goroutine

	fmt.Println("Style Sampler running at http://localhost:1340")
	http.ListenAndServe(":1340", buildMux(app, "/"))
}
