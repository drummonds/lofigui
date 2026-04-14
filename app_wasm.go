//go:build js && wasm

package lofigui

import (
	"net/http"

	wasmhttp "github.com/nlepage/go-wasm-http-server/v2"
)

// defaultWASMTemplate is the built-in template for service worker WASM builds.
// Uses HTML forms for Start/Cancel instead of JavaScript polling.
const defaultWASMTemplate = `<!DOCTYPE html>
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

// RunWASM is the service worker equivalent of [App.Run]. It registers the
// same routes (display, start, cancel, favicon) on an [http.ServeMux] and
// serves them via a service worker using go-wasm-http-server.
//
// The browser page needs a service worker (sw.js) that loads the WASM binary
// and calls registerWasmHTTPListener. See example 01 for the full setup.
//
// Example:
//
//	app := lofigui.NewApp()
//	app.RunWASM(model)
func (app *App) RunWASM(modelFunc func(*App)) {
	app.SetDisplayURL("/")

	app.mu.Lock()
	if app.controller == nil {
		ctrl, err := NewController(ControllerConfig{
			TemplateString: defaultWASMTemplate,
			Name:           "lofigui",
		})
		if err != nil {
			panic(err)
		}
		app.controller = ctrl
	}
	app.mu.Unlock()

	mux := http.NewServeMux()

	mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		app.HandleDisplay(w, r)
	})

	mux.HandleFunc("POST /start", func(w http.ResponseWriter, r *http.Request) {
		app.HandleRoot(w, r, modelFunc, true)
	})

	mux.HandleFunc("POST /cancel", app.HandleCancel("/"))

	mux.HandleFunc("GET /favicon.ico", ServeFavicon)

	wasmhttp.Serve(mux)
}
