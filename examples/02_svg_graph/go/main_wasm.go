//go:build js && wasm

package main

import "codeberg.org/hum3/lofigui"

func main() {
	app := lofigui.NewApp()
	app.Version = "Output Showcase v1.0"
	app.RunWASM(model)
}
