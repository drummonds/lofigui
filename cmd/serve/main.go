// Command serve is a static file server with correct WASM MIME type.
// Usage: go run ./cmd/serve -dir docs -port 8000
package main

import (
	"flag"
	"fmt"
	"log"
	"mime"
	"net/http"
)

func main() {
	dir := flag.String("dir", ".", "directory to serve")
	port := flag.Int("port", 8000, "port to listen on")
	flag.Parse()

	mime.AddExtensionType(".wasm", "application/wasm")

	addr := fmt.Sprintf(":%d", *port)
	log.Printf("Serving %s on http://localhost%s", *dir, addr)
	log.Fatal(http.ListenAndServe(addr, http.FileServer(http.Dir(*dir))))
}
