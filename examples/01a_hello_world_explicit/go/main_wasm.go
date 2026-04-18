//go:build js && wasm

package main

import (
	"log"

	wasmhttp "github.com/nlepage/go-wasm-http-server/v2"

	"codeberg.org/hum3/lofigui"
)

func main() {
	app := lofigui.NewApp()
	app.Version = "Hello World Explicit v1.0"

	// The SW scope is the directory the browser loaded sw.js from
	// (e.g. /01a_hello_world_explicit/wasm_demo/sw/). Redirect targets
	// must carry that prefix so POST-then-redirect responses stay inside
	// the scope instead of bouncing to the origin root.
	scopePath := lofigui.WASMScopePath()
	app.SetDisplayURL(scopePath)

	ctrl, err := lofigui.NewController(lofigui.ControllerConfig{
		TemplateString: helloTemplate,
		Name:           "Hello World Explicit",
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
