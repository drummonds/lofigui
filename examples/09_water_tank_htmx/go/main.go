//go:build !(js && wasm)

package main

import (
	"log"
	"net/http"
)

func main() {
	sim := &Simulation{pumpOn: true}
	mux := setupRoutes(sim)

	addr := ":1349"
	log.Printf("Starting Water Tank HTMX on http://localhost%s", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}
