//go:build js && wasm

package main

import (
	"strings"

	"codeberg.org/hum3/lofigui"
	wasmhttp "github.com/nlepage/go-wasm-http-server/v2"
)

func main() {
	// Links in base.html use <base href="{{.base}}"> so relative hrefs
	// resolve inside the service-worker scope. Normalise to a single
	// trailing slash.
	base := strings.TrimSuffix(lofigui.WASMScopePath(), "/") + "/"

	if _, err := wasmhttp.Serve(buildMux(base)); err != nil {
		panic(err)
	}
	select {} // keep the Go runtime alive to service SW fetches
}
