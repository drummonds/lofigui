package main

import (
	"net/http"

	"codeberg.org/hum3/lofigui"
)

// helloTemplate is the page template, embedded as a string so it works
// in both server and WASM builds (WASM has no filesystem access).
const helloTemplate = `<!DOCTYPE html>
<html>
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>{{.version}}</title>
  <link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/bulma@1.0.4/css/bulma.min.css">
</head>
<body>
  <nav class="navbar is-primary" role="navigation">
    <div class="navbar-brand">
      <span class="navbar-item has-text-weight-bold">{{.version}}</span>
    </div>
    <div class="navbar-end">
      <div class="navbar-item">
        {{if eq .polling "Running"}}
        <span class="tag is-warning">Running</span>
        <form action="/cancel" method="post" style="display:inline" class="ml-1">
          <button class="tag is-danger is-light" type="submit" style="cursor:pointer;border:none">Cancel</button>
        </form>
        {{else}}
        <span class="tag is-success">Ready</span>
        {{end}}
      </div>
    </div>
  </nav>
  <section class="section">
    <div class="container content">
      {{.results}}
      {{if ne .polling "Running"}}
      <form action="/start" method="post" class="mt-4">
        <button class="button is-success" type="submit">Start</button>
      </form>
      {{end}}
    </div>
  </section>
</body>
</html>`

// setupRoutes registers all HTTP handlers on a new ServeMux.
// Used by both the server (main.go) and WASM (main_wasm.go) builds.
func setupRoutes(app *lofigui.App) *http.ServeMux {
	mux := http.NewServeMux()

	// GET / — display current state (results + start button or running indicator)
	mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		app.HandleDisplay(w, r)
	})

	// POST /start — reset buffer, start model in background, redirect to /
	mux.HandleFunc("POST /start", func(w http.ResponseWriter, r *http.Request) {
		app.HandleRoot(w, r, model, true)
	})

	// POST /cancel — cancel running action, redirect to /
	mux.HandleFunc("POST /cancel", app.HandleCancel("/"))

	mux.HandleFunc("GET /favicon.ico", lofigui.ServeFavicon)

	return mux
}
