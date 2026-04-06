//go:build js && wasm

package main

import "codeberg.org/hum3/lofigui"

func model(app *lofigui.App) {
	lofigui.Print("Hello from Go WASM running in your browser!")
	lofigui.Print("This code is compiled to WebAssembly.")

	lofigui.Markdown(`## Key Features

- **Native WASM**: Go compiled directly to WebAssembly
- **Small size**: ~2MB (vs 8-12MB for Pyodide)
- **Fast startup**: Instant (vs 3-5 seconds for Pyodide)
- **No runtime needed**: Pure WASM, no interpreter
- **GitHub Pages ready**: Deploy anywhere static files work`)

	data := [][]string{
		{"Go WASM", "~2MB", "Compiled, native speed"},
		{"Pyodide", "~10MB", "Interpreter, full Python"},
		{"JavaScript", "0MB", "Native browser language"},
	}
	lofigui.Table(data, lofigui.WithHeader([]string{"Technology", "Size", "Type"}))
}

func main() { lofigui.RunWASM(model) }
