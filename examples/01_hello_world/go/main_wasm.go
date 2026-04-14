//go:build js && wasm

package main

import "codeberg.org/hum3/lofigui"

func main() {
	app := lofigui.NewApp()
	app.RunWASM(model)
}
