//go:build js && wasm

package lofigui

import (
	"net/http"
	"syscall/js"

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
  <link rel="stylesheet" href="/assets/bulma.min.css">
</head>
<body>
  <nav class="navbar is-primary" role="navigation">
    <div class="navbar-brand">
      <span class="navbar-item has-text-weight-bold">{{.version}}</span>
    </div>
    <div class="navbar-end">
      {{if .build_date}}
      <div class="navbar-item">
        <span class="is-size-7 has-text-black" title="WASM boot time">{{.build_date}}</span>
      </div>
      {{end}}
      <div class="navbar-item">
        {{if eq .polling "Running"}}
        <span class="tag is-warning">Running</span>
        <form action="cancel" method="post" style="display:inline" class="ml-1">
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
      <form action="start" method="post" class="mt-4">
        <button class="button is-success" type="submit">Start</button>
      </form>
      {{end}}
    </div>
  </section>
</body>
</html>`

// WASMScopePath returns the service-worker scope path under which the current
// WASM build is running, as reported by go-wasm-http-server's `wasmhttp.path`
// JS global. Use this for [App.SetDisplayURL] and for [App.HandleCancel]'s
// redirect target so that POST-then-redirect responses land back inside the
// scope instead of at the origin root.
//
// Returns "/" if no SW scope is set (e.g. during testing without a real SW).
//
// The full path is needed because wasmhttp internally strips it before the
// Go mux sees a request, but http.Redirect writes the raw Location header,
// so the browser needs the scope-prefixed URL to stay inside the SW.
func WASMScopePath() string {
	scopePath := "/"
	if w := js.Global().Get("wasmhttp"); !w.IsUndefined() {
		if p := w.Get("path"); p.Type() == js.TypeString && p.String() != "" {
			scopePath = p.String()
		}
	}
	return scopePath
}

// RunWASM is the service worker equivalent of [App.Run]. It registers the
// same routes (display, start, cancel, favicon) on an [http.ServeMux] and
// serves them via a service worker using go-wasm-http-server.
//
// The browser page needs a service worker (sw.js) that loads the WASM binary
// and calls registerWasmHTTPListener. See example 01 for the full setup.
//
// RunWASM blocks forever so the Go runtime stays alive to service handler
// callbacks. The ServeMux pattern syntax requires go.mod to declare go 1.22+.
//
// Example:
//
//	app := lofigui.NewApp()
//	app.RunWASM(model)
func (app *App) RunWASM(modelFunc func(*App)) {
	scopePath := WASMScopePath()
	app.SetDisplayURL(scopePath)

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
		wrapped := func(a *App) {
			modelFunc(a)
			a.EndAction() // stop polling so the Start button reappears
		}
		app.HandleRoot(w, r, wrapped, true)
	})

	mux.HandleFunc("POST /cancel", app.HandleCancel(scopePath))

	mux.HandleFunc("GET /favicon.ico", ServeFavicon)
	mux.HandleFunc("GET /assets/bulma.min.css", ServeBulma)

	if _, err := wasmhttp.Serve(mux); err != nil {
		panic(err)
	}
	select {} // block forever so the Go runtime stays alive to service handler callbacks
}
