//go:build js && wasm

package main

import (
	"log"

	wasmhttp "github.com/nlepage/go-wasm-http-server/v2"

	"codeberg.org/hum3/lofigui"
)

func main() {
	app := lofigui.NewApp()
	app.Version = "Hello World Explicit + gzip v1.0"

	// See the matching explanation in 01a_hello_world_explicit/go/main_wasm.go
	// — redirects must target the SW scope, not the origin root.
	scopePath := lofigui.WASMScopePath()
	app.SetDisplayURL(scopePath)

	ctrl, err := lofigui.NewController(lofigui.ControllerConfig{
		TemplateString: helloTemplate,
		Name:           "Hello World Explicit + gzip",
	})
	if err != nil {
		log.Fatal(err)
	}
	app.SetController(ctrl)

	mux := setupRoutes(app, scopePath)
	if _, err := wasmhttp.Serve(mux); err != nil {
		log.Fatal(err)
	}
	select {} // block forever so the Go runtime stays alive to service fetch events
}
