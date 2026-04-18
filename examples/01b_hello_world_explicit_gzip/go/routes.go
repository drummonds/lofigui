package main

import (
	_ "embed"
	"net/http"

	"codeberg.org/hum3/lofigui"
)

// helloTemplate is loaded from templates/hello.html via go:embed so the
// same Go binary works for the server and WASM builds (WASM has no
// filesystem at runtime — the embed captures the bytes at build time).
//
//go:embed templates/hello.html
var helloTemplate string

// setupRoutes registers all HTTP handlers on a new ServeMux.
// Used by both the server (main.go) and WASM (main_wasm.go) builds.
//
// redirectPath is passed to HandleCancel so the post-cancel redirect stays
// inside the SW scope in the WASM build (and is just "/" for the server).
func setupRoutes(app *lofigui.App, redirectPath string) *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		app.HandleDisplay(w, r)
	})

	// POST /start — reset buffer, run model in background, redirect to redirectPath.
	mux.HandleFunc("POST /start", func(w http.ResponseWriter, r *http.Request) {
		wrapped := func(a *lofigui.App) {
			model(a)
			// EndAction flips .polling back to "Stopped" so the template
			// re-shows the Start button. app.RunWASM does this wrap for
			// you; the explicit build does it here so it's visible.
			a.EndAction()
		}
		app.HandleRoot(w, r, wrapped, true)
	})

	mux.HandleFunc("POST /cancel", app.HandleCancel(redirectPath))

	mux.HandleFunc("GET /favicon.ico", lofigui.ServeFavicon)

	return mux
}
