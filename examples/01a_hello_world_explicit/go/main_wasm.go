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
	app.SetDisplayURL("/")

	ctrl, err := lofigui.NewController(lofigui.ControllerConfig{
		TemplateString: helloTemplate,
		Name:           "Hello World Explicit",
	})
	if err != nil {
		log.Fatal(err)
	}
	app.SetController(ctrl)

	mux := setupRoutes(app)
	wasmhttp.Serve(mux)
}
