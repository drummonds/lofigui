// wasm-deploy emits the SW bootstrap assets for a lofigui WASM example.
//
// Typical use inside a Taskfile build step:
//
//	GOOS=js GOARCH=wasm go build -o main.wasm .
//	go run codeberg.org/hum3/lofigui/cmd/wasm-deploy \
//	    --dir=docs/NN_example/wasm_demo \
//	    --wasm=main.wasm \
//	    --title="NN — Example"
//
// See [wasmassets.Config] for all flags.
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"codeberg.org/hum3/lofigui/wasmassets"
)

func main() {
	var (
		dir         = flag.String("dir", "", "output directory (required)")
		title       = flag.String("title", "lofigui", "<title>/<h1> shown in the bootstrap")
		wasmPath    = flag.String("wasm", "main.wasm", "path to compiled main.wasm")
		wasmExecJS  = flag.String("wasm-exec-js", "", "path to wasm_exec.js (default: auto-detected in GOROOT)")
		buildDate   = flag.String("build-date", "", "date shown in bootstrap UI (default: today)")
		buildVer    = flag.String("build-version", "", "version shown in bootstrap console (default: title)")
		passthrough = flag.String("passthrough", "", "comma-separated extra URL suffixes for SW passthrough")
		stubs       = flag.String("recovery-stubs", "", "comma-separated filenames to emit as SW-clearing redirect stubs")
	)
	flag.Parse()

	if *dir == "" {
		log.Fatal("wasm-deploy: --dir is required")
	}

	execJS := *wasmExecJS
	if execJS == "" {
		execJS = findWasmExecJS()
		if execJS == "" {
			log.Fatal("wasm-deploy: couldn't auto-detect wasm_exec.js; pass --wasm-exec-js")
		}
	}

	cfg := wasmassets.Config{
		Dir:              *dir,
		Title:            *title,
		WASMPath:         *wasmPath,
		WASMExecJS:       execJS,
		BuildDate:        *buildDate,
		BuildVersion:     *buildVer,
		PassthroughPaths: splitCSV(*passthrough),
		RecoveryStubs:    splitCSV(*stubs),
	}

	written, err := wasmassets.Deploy(cfg)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("wasm-deploy: wrote %d files to %s\n", len(written), *dir)
	for _, f := range written {
		fmt.Printf("  %s\n", f)
	}
}

func splitCSV(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if t := strings.TrimSpace(p); t != "" {
			out = append(out, t)
		}
	}
	return out
}

// findWasmExecJS looks in the usual places in GOROOT for wasm_exec.js.
// Go 1.24+ moved it from misc/wasm/ to lib/wasm/.
func findWasmExecJS() string {
	goroot := os.Getenv("GOROOT")
	if goroot == "" {
		if out, err := exec.Command("go", "env", "GOROOT").Output(); err == nil {
			goroot = strings.TrimSpace(string(out))
		}
	}
	if goroot == "" {
		return ""
	}
	candidates := []string{
		filepath.Join(goroot, "lib", "wasm", "wasm_exec.js"),
		filepath.Join(goroot, "misc", "wasm", "wasm_exec.js"),
	}
	for _, c := range candidates {
		if _, err := os.Stat(c); err == nil {
			return c
		}
	}
	return ""
}
