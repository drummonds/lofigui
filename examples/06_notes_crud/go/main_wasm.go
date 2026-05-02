//go:build js && wasm

package main

import (
	"strings"

	"codeberg.org/hum3/lofigui"
	wasmhttp "github.com/nlepage/go-wasm-http-server/v2"
)

func main() {
	// notes.html uses <base href="{{.base}}"> so the form action="" hrefs and
	// the POST handlers' "./" redirects all stay inside the SW scope.
	base := strings.TrimSuffix(lofigui.WASMScopePath(), "/") + "/"

	if _, err := wasmhttp.Serve(buildMux(base)); err != nil {
		panic(err)
	}
	select {} // keep the Go runtime alive to service SW fetches
}
