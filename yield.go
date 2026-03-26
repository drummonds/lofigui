//go:build !js || !wasm

package lofigui

import "runtime"

// Yield cooperatively yields the processor. On native builds this calls
// runtime.Gosched(). In WASM builds (yield_wasm.go) it sleeps for 1ms
// so the browser event loop can run.
//
// Call Yield inside tight background loops that update the lofigui buffer
// to keep the UI responsive.
func Yield() {
	runtime.Gosched()
}
