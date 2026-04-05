//go:build js && wasm

package main

import "codeberg.org/hum3/lofigui"

func model(app *lofigui.App) {
	lofigui.Print("Hello from TinyGo WASM running in your browser!")
	lofigui.Print("This code is compiled to WebAssembly using TinyGo.")

	lofigui.Markdown(`## Why TinyGo?

- **Ultra-small**: ~2.5MB (vs ~8MB standard Go WASM)
- **Instant startup**: Near-instant loading
- **Optimized for WASM**: Designed specifically for WebAssembly targets
- **Go syntax**: Write real Go code, run in the browser`)

	data := [][]string{
		{"TinyGo WASM", "~2.5MB", "Ultra-compact", "Instant"},
		{"Go WASM", "~8MB", "Compact", "<100ms"},
		{"Pyodide", "~10MB", "Full Python", "3-5s"},
		{"JavaScript", "0KB", "Native", "Instant"},
	}
	lofigui.Table(data, lofigui.WithHeader([]string{"Technology", "Size", "Description", "Startup"}))
}

func main() { lofigui.RunWASM(model) }
