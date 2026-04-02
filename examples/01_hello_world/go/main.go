//go:build !(js && wasm)

package main

import (
	"errors"
	"log"
	"net/http"
	"os"

	"codeberg.org/hum3/lofigui"
)

func main() {
	app := lofigui.NewApp()
	http.HandleFunc("/", app.Handle(model))
	http.HandleFunc("/cancel", app.HandleCancel("/"))
	if err := app.ListenAndServe(":1340", nil); err != nil {
		if errors.Is(err, lofigui.ErrCancelled) {
			log.Println(err)
			os.Exit(1)
		}
		log.Fatal(err)
	}
}
