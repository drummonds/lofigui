//go:build !(js && wasm)

package main

import (
	"fmt"
	"net/http"
)

func main() {
	fmt.Println("Style Sampler running at http://localhost:1340")
	http.ListenAndServe(":1340", buildMux("/"))
}
