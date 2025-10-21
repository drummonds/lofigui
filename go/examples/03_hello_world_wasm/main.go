// +build js,wasm

package main

import (
	"syscall/js"

	"github.com/drummonds/lofigui/go/lofigui"
)

// model generates the basic output
func model() string {
	lofigui.Reset()
	lofigui.Print("Hello from Go WASM running in your browser!")
	lofigui.Print("This code is compiled to WebAssembly.")

	lofigui.Markdown(`## Key Features

- **Native WASM**: Go compiled directly to WebAssembly
- **Small size**: ~2MB (vs 8-12MB for Pyodide)
- **Fast startup**: Instant (vs 3-5 seconds for Pyodide)
- **No runtime needed**: Pure WASM, no interpreter
- **GitHub Pages ready**: Deploy anywhere static files work`)

	// Add a simple table
	data := [][]string{
		{"Go WASM", "~2MB", "Compiled, native speed"},
		{"Pyodide", "~10MB", "Interpreter, full Python"},
		{"JavaScript", "0MB", "Native browser language"},
	}
	lofigui.Table(data, lofigui.WithHeader([]string{"Technology", "Size", "Type"}))

	return lofigui.Buffer()
}

// advancedModel generates more complex output
func advancedModel() string {
	lofigui.Reset()
	lofigui.Markdown("## Advanced Processing in Go WASM")

	// Simple computation
	numbers := []int{1, 1, 2, 3, 5, 8, 13, 21, 34, 55}
	total := 0
	for _, n := range numbers {
		total += n
	}
	average := float64(total) / float64(len(numbers))

	lofigui.Printf("Fibonacci sequence sum: %d", total)
	lofigui.Printf("Average: %.2f", average)

	// Show the computation
	lofigui.Markdown("### Data Analysis")

	tableData := make([][]string, len(numbers))
	cumsum := 0
	for i, num := range numbers {
		cumsum += num
		tableData[i] = []string{
			string(rune('0' + i)),      // Index
			string(rune('0' + num)),     // Value (simple conversion for demo)
			string(rune('0' + cumsum)),  // Cumulative sum
		}
	}

	// For proper number formatting, we need Printf
	lofigui.HTML("<table class='table'><thead><tr><th>Index</th><th>Value</th><th>Cumulative Sum</th></tr></thead><tbody>")
	for i, num := range numbers {
		cumsum := 0
		for j := 0; j <= i; j++ {
			cumsum += numbers[j]
		}
		lofigui.HTML("<tr>")
		lofigui.Printf("<td>%d</td><td>%d</td><td>%d</td>", i, num, cumsum)
		lofigui.HTML("</tr>")
	}
	lofigui.HTML("</tbody></table>")

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
    lofigui.Print("Hello from Go WASM!")

    lofigui.Markdown("## Features")
    // ... more code ...

    return lofigui.Buffer()
}

// Compile to WASM:
// GOOS=js GOARCH=wasm go build -o main.wasm`

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
