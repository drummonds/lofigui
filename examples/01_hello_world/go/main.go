//go:build !(js && wasm)

package main

import "codeberg.org/hum3/lofigui"

func main() {
	app := lofigui.NewApp()
	app.Run(":1340", model)
}
