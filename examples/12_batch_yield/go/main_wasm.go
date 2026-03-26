//go:build js && wasm

package main

import (
	"fmt"
	"math/rand"
	"strings"
	"sync"
	"syscall/js"
	"time"

	"github.com/drummonds/lofigui"
)

const batchSize = 40

var (
	mu             sync.Mutex
	yieldRunning   bool
	noYieldRunning bool
	yieldBuf       strings.Builder
	noYieldBuf     strings.Builder
	cancelYield    chan struct{}
)

func appendLine(buf *strings.Builder, format string, args ...any) {
	buf.WriteString("<p>")
	buf.WriteString(fmt.Sprintf(format, args...))
	buf.WriteString("</p>\n")
}

// busySpin burns CPU for the given duration without calling time.Sleep,
// so the browser event loop never gets control back.
func busySpin(d time.Duration) {
	end := time.Now().Add(d)
	for time.Now().Before(end) {
		// spin
	}
}

func batchWithYield(cancel chan struct{}) {
	mu.Lock()
	yieldBuf.Reset()
	appendLine(&yieldBuf, "Generating %d customers...", batchSize)
	mu.Unlock()

	for i := 1; i <= batchSize; i++ {
		select {
		case <-cancel:
			mu.Lock()
			appendLine(&yieldBuf, "Reset after %d customers.", i-1)
			yieldRunning = false
			mu.Unlock()
			return
		default:
		}

		name := fmt.Sprintf("Customer-%04d", i)
		balance := rand.Float64() * 10000

		mu.Lock()
		appendLine(&yieldBuf, "%s — balance: £%.2f", name, balance)
		mu.Unlock()

		// Simulate 0.5s of blocking work. In WASM time.Sleep suspends
		// the Go runtime and returns control to the browser, so the
		// page stays responsive and output streams in progressively.
		time.Sleep(500 * time.Millisecond)
		lofigui.Yield()
	}

	mu.Lock()
	appendLine(&yieldBuf, "Done — %d customers created.", batchSize)
	yieldRunning = false
	mu.Unlock()
}

func batchWithoutYield() {
	mu.Lock()
	noYieldBuf.Reset()
	appendLine(&noYieldBuf, "Generating %d customers...", batchSize)
	mu.Unlock()

	for i := 1; i <= batchSize; i++ {
		name := fmt.Sprintf("Customer-%04d", i)
		balance := rand.Float64() * 10000

		mu.Lock()
		appendLine(&noYieldBuf, "%s — balance: £%.2f", name, balance)
		mu.Unlock()

		// CPU-bound busy-spin for 0.5s — never yields to the browser.
		// The page freezes and all output appears at once when done.
		busySpin(500 * time.Millisecond)
	}

	mu.Lock()
	appendLine(&noYieldBuf, "Done — %d customers created.", batchSize)
	noYieldRunning = false
	mu.Unlock()
}

func goStartWithYield(this js.Value, args []js.Value) any {
	mu.Lock()
	if yieldRunning {
		mu.Unlock()
		return nil
	}
	yieldRunning = true
	cancelYield = make(chan struct{})
	mu.Unlock()

	go batchWithYield(cancelYield)
	return nil
}

func goStartWithoutYield(this js.Value, args []js.Value) any {
	mu.Lock()
	if noYieldRunning {
		mu.Unlock()
		return nil
	}
	noYieldRunning = true
	mu.Unlock()

	go batchWithoutYield()
	return nil
}

func goStartBoth(this js.Value, args []js.Value) any {
	goStartWithYield(this, args)
	goStartWithoutYield(this, args)
	return nil
}

func goRenderYield(this js.Value, args []js.Value) any {
	mu.Lock()
	defer mu.Unlock()
	return js.ValueOf(yieldBuf.String())
}

func goRenderNoYield(this js.Value, args []js.Value) any {
	mu.Lock()
	defer mu.Unlock()
	return js.ValueOf(noYieldBuf.String())
}

func goReset(this js.Value, args []js.Value) any {
	mu.Lock()
	defer mu.Unlock()
	// Cancel running yield batch if any.
	if yieldRunning && cancelYield != nil {
		close(cancelYield)
		cancelYield = nil
	}
	yieldBuf.Reset()
	noYieldBuf.Reset()
	yieldRunning = false
	noYieldRunning = false
	return nil
}

func goIsRunning(this js.Value, args []js.Value) any {
	mu.Lock()
	defer mu.Unlock()
	return js.ValueOf(yieldRunning || noYieldRunning)
}

func main() {
	js.Global().Set("goStartWithYield", js.FuncOf(goStartWithYield))
	js.Global().Set("goStartWithoutYield", js.FuncOf(goStartWithoutYield))
	js.Global().Set("goStartBoth", js.FuncOf(goStartBoth))
	js.Global().Set("goRenderYield", js.FuncOf(goRenderYield))
	js.Global().Set("goRenderNoYield", js.FuncOf(goRenderNoYield))
	js.Global().Set("goReset", js.FuncOf(goReset))
	js.Global().Set("goIsRunning", js.FuncOf(goIsRunning))

	js.Global().Call("wasmReady")

	<-make(chan struct{})
}
