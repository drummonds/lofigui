//go:build js && wasm

package main

import "codeberg.org/hum3/lofigui"

func main() { lofigui.RunWASM(model) }
