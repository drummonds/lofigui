//go:build js && wasm

package main

import (
	wasmhttp "github.com/nlepage/go-wasm-http-server/v2"
)

func main() {
	sim := &Simulation{pumpOn: true}
	mux := setupRoutes(sim)
	wasmhttp.Serve(mux)
}
