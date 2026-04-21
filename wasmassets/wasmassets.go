// Package wasmassets emits the service-worker bootstrap files needed to run
// a lofigui App as WASM in the browser. It's the "easy deployment" path: one
// call produces bootstrap.html, sw.js, wasmhttp_sw.js, and any requested
// recovery stubs, all with correct paths and cache scoping.
//
// Use [Deploy] from a build step (or run the cmd/wasm-deploy CLI wrapper).
// For hand-written bootstraps (e.g. to demonstrate mechanics or to add gz
// decompression), copy the templates out of this package and edit them.
package wasmassets

import (
	"embed"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"text/template"
	"time"

	"codeberg.org/hum3/lofigui"
)

//go:embed templates
var templatesFS embed.FS

// Config controls what [Deploy] writes and where.
type Config struct {
	// Dir is the output directory. Created if it doesn't exist.
	Dir string

	// Title appears in the bootstrap's <title> and <h1> (e.g. "01 — Hello World").
	Title string

	// WASMPath is the path to the compiled main.wasm on disk. Copied into Dir.
	WASMPath string

	// WASMExecJS is the path to Go's wasm_exec.js. Copied into Dir.
	// Typical value: filepath.Join(os.Getenv("GOROOT"), "lib", "wasm", "wasm_exec.js").
	WASMExecJS string

	// BuildDate, if empty, defaults to today in YYYY-MM-DD. Shown in the
	// bootstrap UI so stale caches are immediately obvious.
	BuildDate string

	// BuildVersion is an arbitrary short string shown in the bootstrap's
	// console log (e.g. "01 v0.17.34"). Empty is fine.
	BuildVersion string

	// PassthroughPaths are extra URL path suffixes the SW should pass
	// through to the network (in addition to index.html). Use this for
	// recovery stubs so the Go handler doesn't try to serve them.
	PassthroughPaths []string

	// RecoveryStubs are filenames to write inside Dir that redirect to the
	// bootstrap after unregistering any SWs scoped at or above their URL.
	// Each filename's redirect target is "./".
	RecoveryStubs []string
}

// Deploy writes the WASM bootstrap assets into cfg.Dir.
//
// Files written:
//   - index.html          — the SW bootstrap page
//   - sw.js               — service worker script
//   - wasmhttp_sw.js      — go-wasm-http-server runtime (vendored)
//   - wasm_exec.js        — copied from cfg.WASMExecJS
//   - main.wasm           — copied from cfg.WASMPath
//   - <stub>.html         — for each cfg.RecoveryStubs entry
//
// Returns the list of files written on success.
func Deploy(cfg Config) ([]string, error) {
	if cfg.Dir == "" {
		return nil, fmt.Errorf("wasmassets: Dir is required")
	}
	if cfg.Title == "" {
		cfg.Title = "lofigui"
	}
	if cfg.WASMPath == "" {
		return nil, fmt.Errorf("wasmassets: WASMPath is required")
	}
	if cfg.WASMExecJS == "" {
		return nil, fmt.Errorf("wasmassets: WASMExecJS is required")
	}
	if cfg.BuildDate == "" {
		cfg.BuildDate = time.Now().UTC().Format("2006-01-02")
	}
	if cfg.BuildVersion == "" {
		cfg.BuildVersion = cfg.Title
	}

	if err := os.MkdirAll(cfg.Dir, 0o755); err != nil {
		return nil, fmt.Errorf("wasmassets: mkdir %s: %w", cfg.Dir, err)
	}

	var written []string

	// Render bootstrap.html → index.html
	if err := renderTemplate("templates/bootstrap.html.tmpl",
		filepath.Join(cfg.Dir, "index.html"),
		map[string]any{
			"Title":        cfg.Title,
			"BuildDate":    cfg.BuildDate,
			"BuildVersion": cfg.BuildVersion,
		}); err != nil {
		return nil, err
	}
	written = append(written, "index.html")

	// Render sw.js
	if err := renderTemplate("templates/sw.js.tmpl",
		filepath.Join(cfg.Dir, "sw.js"),
		map[string]any{
			"PassthroughPaths": passthroughList(cfg),
		}); err != nil {
		return nil, err
	}
	written = append(written, "sw.js")

	// Copy wasmhttp_sw.js verbatim
	if err := copyEmbedded("templates/wasmhttp_sw.js",
		filepath.Join(cfg.Dir, "wasmhttp_sw.js")); err != nil {
		return nil, err
	}
	written = append(written, "wasmhttp_sw.js")

	// Copy wasm_exec.js and main.wasm from disk
	if err := copyFile(cfg.WASMExecJS, filepath.Join(cfg.Dir, "wasm_exec.js")); err != nil {
		return nil, fmt.Errorf("wasmassets: copy wasm_exec.js: %w", err)
	}
	written = append(written, "wasm_exec.js")

	if err := copyFile(cfg.WASMPath, filepath.Join(cfg.Dir, "main.wasm")); err != nil {
		return nil, fmt.Errorf("wasmassets: copy main.wasm: %w", err)
	}
	written = append(written, "main.wasm")

	// Vendored Bulma so the bootstrap loads without an outbound CDN fetch.
	if err := os.WriteFile(filepath.Join(cfg.Dir, "bulma.min.css"), lofigui.BulmaCSS, 0o644); err != nil {
		return nil, fmt.Errorf("wasmassets: write bulma.min.css: %w", err)
	}
	written = append(written, "bulma.min.css")

	for _, stub := range cfg.RecoveryStubs {
		if err := renderTemplate("templates/demo_stub.html.tmpl",
			filepath.Join(cfg.Dir, stub),
			map[string]any{"RedirectTo": "./"}); err != nil {
			return nil, err
		}
		written = append(written, stub)
	}

	return written, nil
}

func passthroughList(cfg Config) []string {
	seen := map[string]bool{}
	var out []string
	for _, p := range cfg.PassthroughPaths {
		if !seen[p] {
			seen[p] = true
			out = append(out, p)
		}
	}
	// Recovery stubs must be passthrough so the SW doesn't swallow them.
	for _, s := range cfg.RecoveryStubs {
		key := "/" + s
		if !seen[key] {
			seen[key] = true
			out = append(out, key)
		}
	}
	return out
}

func renderTemplate(embeddedPath, outPath string, data any) error {
	src, err := fs.ReadFile(templatesFS, embeddedPath)
	if err != nil {
		return fmt.Errorf("wasmassets: read %s: %w", embeddedPath, err)
	}
	tmpl, err := template.New(filepath.Base(embeddedPath)).Parse(string(src))
	if err != nil {
		return fmt.Errorf("wasmassets: parse %s: %w", embeddedPath, err)
	}
	out, err := os.Create(outPath)
	if err != nil {
		return fmt.Errorf("wasmassets: create %s: %w", outPath, err)
	}
	defer out.Close()
	if err := tmpl.Execute(out, data); err != nil {
		return fmt.Errorf("wasmassets: execute %s: %w", embeddedPath, err)
	}
	return nil
}

func copyEmbedded(embeddedPath, outPath string) error {
	src, err := fs.ReadFile(templatesFS, embeddedPath)
	if err != nil {
		return fmt.Errorf("wasmassets: read %s: %w", embeddedPath, err)
	}
	return os.WriteFile(outPath, src, 0o644)
}

func copyFile(srcPath, dstPath string) error {
	src, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer src.Close()
	dst, err := os.Create(dstPath)
	if err != nil {
		return err
	}
	defer dst.Close()
	_, err = io.Copy(dst, src)
	return err
}
