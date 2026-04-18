//go:build !(js && wasm)

package main

import (
	"log"
	"net/http"

	"codeberg.org/hum3/lofigui"
)

func main() {
	app := lofigui.NewApp()
	app.Version = "Hello World Explicit + gzip v1.0"
	app.SetDisplayURL("/")

	ctrl, err := lofigui.NewController(lofigui.ControllerConfig{
		TemplateString: helloTemplate,
		Name:           "Hello World Explicit + gzip",
	})
	if err != nil {
		log.Fatal(err)
	}
	app.SetController(ctrl)

	mux := setupRoutes(app, "/")

	addr := ":1342"
	log.Printf("Starting on http://localhost%s", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}
