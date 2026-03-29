//go:build !(js && wasm)

package main

import (
	"log"
	"net/http"

	"codeberg.org/hum3/lofigui"
)

func main() {
	app := lofigui.NewApp()
	http.HandleFunc("/", app.Handle(model))
	if err := app.ListenAndServe(":1340", nil); err != nil {
		log.Fatal(err)
	}
}
