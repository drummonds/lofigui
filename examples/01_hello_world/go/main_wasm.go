//go:build js && wasm

package main

import (
	"fmt"
	"sync"
	"syscall/js"
	"time"

	"github.com/drummonds/lofigui"
)

var (
	mu      sync.Mutex
	running bool
)

func goStart(this js.Value, args []js.Value) any {
	mu.Lock()
	if running {
		mu.Unlock()
		return nil
	}
	running = true
	mu.Unlock()

	lofigui.Reset()
	lofigui.Print("Hello world.")

	go func() {
		for i := 0; i < 5; i++ {
			<-time.After(1 * time.Second)
			mu.Lock()
			if !running {
				mu.Unlock()
				return
			}
			mu.Unlock()
			lofigui.Print(fmt.Sprintf("Count %d", i))
		}
		lofigui.Print("Done.")
		mu.Lock()
		running = false
		mu.Unlock()
	}()

	return nil
}

func goRender(this js.Value, args []js.Value) any {
	return js.ValueOf(lofigui.Buffer())
}

func goIsRunning(this js.Value, args []js.Value) any {
	mu.Lock()
	defer mu.Unlock()
	return js.ValueOf(running)
}

func main() {
	js.Global().Set("goStart", js.FuncOf(goStart))
	js.Global().Set("goRender", js.FuncOf(goRender))
	js.Global().Set("goIsRunning", js.FuncOf(goIsRunning))

	js.Global().Call("wasmReady")

	<-make(chan struct{})
}
