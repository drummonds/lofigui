package lofigui

import (
	_ "embed"
	"net/http"
)

// BulmaCSS is Bulma v1.0.4 (MIT) vendored into the library. Served by
// [ServeBulma] at "/assets/bulma.min.css" so lofigui apps work without
// an outbound CDN fetch.
//
//go:embed bulma.min.css
var BulmaCSS []byte

// ServeBulma is an http.HandlerFunc that serves the vendored Bulma CSS
// with a long cache lifetime. Register at "/assets/bulma.min.css":
//
//	http.HandleFunc("/assets/bulma.min.css", lofigui.ServeBulma)
//
// Automatically registered by [App.Run] and [App.RunWASM].
func ServeBulma(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/css; charset=utf-8")
	w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	w.Write(BulmaCSS)
}
