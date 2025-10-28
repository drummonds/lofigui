//go:build js && wasm
// +build js,wasm

package main

import (
	"syscall/js"

	"github.com/drummonds/lofigui"
)

// model generates the basic output
func model() string {
	lofigui.Reset()
	lofigui.Print("Hello from TinyGo WASM running in your browser!")
	lofigui.Print("This code is compiled to WebAssembly using TinyGo.")

	lofigui.Markdown(`## Why TinyGo?

- **Ultra-small**: ~100KB (vs ~2MB standard Go WASM)
- **Instant startup**: Near-instant loading
- **Optimized for WASM**: Designed specifically for WebAssembly targets
- **Go syntax**: Write real Go code, run in the browser`)

	// Add a comparison table
	data := [][]string{
		{"TinyGo WASM", "~100KB", "Ultra-compact", "Instant"},
		{"Go WASM", "~2MB", "Compact", "<100ms"},
		{"Pyodide", "~10MB", "Full Python", "3-5s"},
		{"JavaScript", "0KB", "Native", "Instant"},
	}
	lofigui.Table(data, lofigui.WithHeader([]string{"Technology", "Size", "Description", "Startup"}))

	lofigui.Markdown(`## Perfect For

- Lightweight web apps
- GitHub Pages deployment
- Mobile-friendly pages
- Fast-loading demos`)

	return lofigui.Buffer()
}

// advancedModel generates more complex output
func advancedModel() string {
	lofigui.Reset()
	lofigui.Markdown("## Advanced Processing with TinyGo")

	// Simple computation
	numbers := []int{1, 1, 2, 3, 5, 8, 13, 21, 34, 55}
	total := 0
	for _, n := range numbers {
		total += n
	}
	average := float64(total) / float64(len(numbers))

	lofigui.Printf("Fibonacci sequence sum: %d", total)
	lofigui.Printf("Average: %.2f", average)

	// Show the data
	lofigui.Markdown("### Data Table")

	lofigui.HTML("<table class='table is-striped'><thead><tr><th>Index</th><th>Value</th><th>Cumulative</th></tr></thead><tbody>")
	cumsum := 0
	for i, num := range numbers {
		cumsum += num
		lofigui.HTML("<tr>")
		lofigui.Printf("<td>%d</td><td>%d</td><td>%d</td>", i, num, cumsum)
		lofigui.HTML("</tr>")
	}
	lofigui.HTML("</tbody></table>")

	lofigui.Markdown(`### Why This Works

TinyGo supports most standard Go features:
- Full Go syntax and stdlib (with some limitations)
- Fast execution in the browser
- Direct DOM manipulation possible
- Perfect for computational demos`)

	return lofigui.Buffer()
}

// runModel is called from JavaScript when user clicks "Run Basic Example"
func runModel(this js.Value, args []js.Value) interface{} {
	result := model()
	return js.ValueOf(result)
}

// runAdvancedModel is called from JavaScript when user clicks "Run Advanced Example"
func runAdvancedModel(this js.Value, args []js.Value) interface{} {
	result := advancedModel()
	return js.ValueOf(result)
}

// getSourceCode returns the Go source for display
func getSourceCode(this js.Value, args []js.Value) interface{} {
	source := `// Model function
func model() string {
    lofigui.Reset()
    lofigui.Print("Hello from TinyGo WASM!")

    lofigui.Markdown("## Features")
    // ... more code ...

    return lofigui.Buffer()
}

// Compile to WASM with TinyGo:
// tinygo build -o main.wasm -target wasm main.go

// File size comparison:
// - TinyGo: ~100KB (this example!)
// - Standard Go: ~2MB
// - Pyodide: ~10MB`

	return js.ValueOf(source)
}

func main() {
	// Expose Go functions to JavaScript
	js.Global().Set("goRunModel", js.FuncOf(runModel))
	js.Global().Set("goRunAdvancedModel", js.FuncOf(runAdvancedModel))
	js.Global().Set("goGetSourceCode", js.FuncOf(getSourceCode))

	// Signal that WASM is ready
	js.Global().Call("wasmReady")

	// Keep the program running
	<-make(chan struct{})
}
