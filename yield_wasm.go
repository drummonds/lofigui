//go:build js && wasm

package lofigui

import "time"

// Yield cooperatively yields to the browser event loop. In WASM,
// runtime.Gosched() only yields between Go goroutines — it never
// returns control to JavaScript. time.Sleep suspends the Go runtime
// and lets the browser repaint and handle events.
func Yield() {
	defaultContext.checkCancelled()
	time.Sleep(time.Millisecond)
}
