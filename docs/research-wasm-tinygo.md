# Research: WASM TinyGo

TinyGo as a WASM compilation target for lofigui --- current status, blockers, and a research proposal to fix the core limitation.

## Background

TinyGo is an alternative Go compiler targeting microcontrollers, WASM, and other constrained environments. Its primary appeal for browser WASM is dramatically smaller binary output. Lofigui previously shipped TinyGo-compiled WASM demos alongside standard Go WASM builds for every example, but removed them when adopting the [service worker approach](research-wasm-service-workers.html) because TinyGo cannot compile `net/http` for browser targets.

This page consolidates TinyGo-related findings that were previously scattered across the service workers research.

## Binary size: TinyGo vs standard Go vs compression

TinyGo's headline advantage is binary size:

| Scenario | Standard Go | TinyGo | Reduction |
|----------|-------------|--------|-----------|
| Minimal hello | ~2 MB | ~86 KB | 95% |
| HTTP server app | ~8--10 MB | ~2--3 MB | 70% |

For a service worker, binary size directly affects cold-start time. A 10 MB binary is noticeably slow on first load. Caching the binary in the service worker cache mitigates repeat visits, but the first visit matters.

### Compression closes the gap

Standard Go WASM binaries compress extremely well. Browsers transparently decompress `Content-Encoding: gzip` (or brotli) before passing bytes to `WebAssembly.instantiateStreaming()`, so the user downloads the compressed size, not the raw size.

| Method | Raw size (typical lofigui) | Compressed | Transfer size | Ratio |
|--------|---------------------------|------------|---------------|-------|
| Standard Go, uncompressed | 7.8 MB | --- | 7.8 MB | 100% |
| Standard Go, gzip -9 | 7.8 MB | 2.1 MB | 2.1 MB | 27% |
| Standard Go, brotli -11 | 7.8 MB | ~1.7 MB (est.) | ~1.7 MB | ~22% |
| TinyGo, uncompressed | ~2.5 MB | --- | 2.5 MB | 32% |
| TinyGo, gzip -9 | ~2.5 MB | ~800 KB (est.) | ~800 KB | ~10% |

With gzip, standard Go is ~2.1 MB on the wire --- comparable to TinyGo's *uncompressed* size. With brotli, the gap narrows further. TinyGo + gzip would still be smallest (~800 KB), but the difference between 800 KB and 2.1 MB is far less dramatic than between 2.5 MB and 7.8 MB.

**Compression is the pragmatic answer.** It works today, requires no compiler changes, and is a single build step (`gzip -9 main.wasm`). Most static hosts (GitHub Pages, Netlify, Cloudflare Pages, statichost.eu) apply gzip automatically, so pre-compression may not even be needed for deployment. Lofigui's build pipeline already includes gzip compression for WASM artifacts.

### Build flags for standard Go

These provide modest savings on top of compression:

| Flag | Effect | Saving |
|------|--------|--------|
| `-ldflags="-s"` | Strips WASM `name` section | ~186 KB (2.2%) |
| `-trimpath` | Shortens embedded path strings | ~10 KB |
| `-ldflags="-w"` | No effect on WASM (no DWARF) | 0 |

Total from flags: ~196 KB (~2.3% reduction). Worth including in the build command but not transformative.

## What TinyGo cannot do (browser WASM)

These are the blockers that prevent TinyGo from being used for lofigui's service worker approach:

| Feature | Standard Go | TinyGo | Why it matters |
|---------|-------------|--------|----------------|
| `net/http` in browser | Works (maps to Fetch API) | **Panics at runtime** ([#4420](https://github.com/tinygo-org/tinygo/issues/4420)) | Service workers need `net/http` for `go-wasm-http-server` |
| `html/template` | Works | **Cannot import** (reflection limits) | Server-side template rendering |
| `text/template` | Works | **Cannot import** (reflection limits) | Same |
| `pongo2` | Works | Likely broken (reflection) | Lofigui's Go template engine |
| `encoding/json` | Works | Partial (missing `sync.WaitGroup.Go()`) | API response parsing |
| Full goroutine scheduler | Yes (cooperative on single thread) | Simpler scheduler, edge cases | Complex concurrent handlers |

### The net/http problem is decisive

TinyGo's `net/http` does not work in browser WASM because the Fetch API bridge is not implemented. Standard Go maps `net/http` calls to the browser's `fetch()` API via `syscall/js` --- this is what makes `go-wasm-http-server` possible. TinyGo's `net/http` was built for `wasip1` and bare-metal targets; the browser `js/wasm` transport layer simply does not exist.

Since `go-wasm-http-server` depends on `net/http` types (`http.Request`, `http.ResponseWriter`, `http.Handler`), TinyGo cannot be used for the service worker pattern at all.

### Template engines

| Engine | Standard Go WASM | TinyGo WASM | Approach |
|--------|-----------------|-------------|----------|
| `html/template` | Works | Broken | Reflection-based |
| `text/template` | Works | Broken | Reflection-based |
| pongo2 | Works | Likely broken | Reflection-based |
| [templ](https://github.com/a-h/templ) | Works | **Works** | Code generation, no reflection |

`templ` is the only viable template engine for TinyGo WASM because it uses code generation instead of reflection. For standard Go WASM, pongo2 and `html/template` both work.

## Where TinyGo still works

For lofigui's *original* WASM approach --- direct `syscall/js` function exports without service workers --- TinyGo works. This approach uses `js.Global().Set("goRender", ...)` and custom JavaScript glue, never touching `net/http`. The smaller binary size is genuinely valuable here.

However, lofigui has decided to standardise on the service worker approach (shared `net/http` handlers for server and WASM builds), which makes TinyGo incompatible with the direction of the project. The direct JS export pattern remains available for trivial Level 1--2 examples where code sharing is not needed, but new examples should use the service worker pattern.

## Concurrency

Both Go and TinyGo WASM run on a single thread. Goroutines are cooperatively scheduled:

- `time.Sleep()` suspends the Go runtime and returns control to the browser event loop
- `runtime.Gosched()` yields between goroutines but does NOT return control to the browser
- Channel operations and select statements work normally
- There is no parallelism --- goroutines interleave, they do not run simultaneously

TinyGo's goroutine scheduler is simpler than standard Go's and has known edge cases with deeply nested channel operations. For service worker request handling (where each request can be a goroutine), this would be a concern --- but since TinyGo can't do service workers anyway, it's academic.

## Research proposal: TinyGo net/http for browser WASM

### The gap

Standard Go's `net/http` works in `GOOS=js GOARCH=wasm` because of a transport implementation in `net/http/roundtrip_js.go` that maps HTTP calls to the browser's `fetch()` API via `syscall/js`. TinyGo has no equivalent. The issue is tracked as [tinygo-org/tinygo#4420](https://github.com/tinygo-org/tinygo/issues/4420).

### What would need to happen

1. **Implement the Fetch API transport.** Port or rewrite Go's `roundtrip_js.go` for TinyGo's `syscall/js` implementation. This maps `http.RoundTrip` to `fetch()`, converting Go request/response types to and from JavaScript `Request`/`Response` objects. The standard Go implementation is ~300 lines.

2. **Fix reflection for `net/http` types.** TinyGo's reflection support is incomplete. `net/http` uses reflection in several places (header canonicalization, cookie parsing). Each site would need testing and potentially a TinyGo-compatible alternative.

3. **Test with `go-wasm-http-server`.** Once `net/http` works, verify that `go-wasm-http-server` (which bridges `net/http.Handler` to service worker `FetchEvent`) functions correctly. This library uses `http.Request`, `http.ResponseWriter`, and `http.Handler` --- all of which would need to work under TinyGo.

4. **Template engine.** Even with `net/http` working, `html/template` and `pongo2` would remain broken (reflection). The lofigui WASM build would need to switch to `templ` or raw string concatenation for template rendering. This is a separate, larger migration.

### Effort estimate

The Fetch API transport (#1) is a focused, bounded piece of work --- likely 1--2 weeks for someone familiar with TinyGo internals. The reflection fixes (#2) are unbounded and may require upstream TinyGo changes. The template migration (#4) is a lofigui-specific concern, independent of TinyGo.

### Is it worth it?

With compression, the size argument for TinyGo is less compelling than it was:

| Scenario | Standard Go (gzip) | TinyGo (gzip, hypothetical) | Saving |
|----------|--------------------|-----------------------------|--------|
| Minimal app | ~500 KB | ~50 KB | 450 KB |
| Full lofigui app | ~2.1 MB | ~800 KB | 1.3 MB |

1.3 MB matters on slow mobile connections. It does not matter on broadband or after the service worker has cached the binary. For lofigui's target audience (1--10 users, internal tools, dashboards), it is unlikely to be the deciding factor.

**The stronger argument for TinyGo is startup time**, not transfer size. A 2.5 MB WASM binary instantiates significantly faster than an 8 MB one, even when both were transferred compressed. WASM instantiation is proportional to uncompressed binary size because the runtime must parse and compile all the code.

**Recommendation:** Monitor [tinygo-org/tinygo#4420](https://github.com/tinygo-org/tinygo/issues/4420). If the TinyGo team implements the Fetch API transport, re-evaluate. Contributing the implementation upstream would be a valuable open-source contribution but is not on lofigui's critical path. In the meantime, standard Go with gzip compression is the pragmatic choice.

## Decision: standard Go with compression

Lofigui uses standard Go for all WASM builds. TinyGo examples have been removed from the repository. The rationale:

1. **Service worker compatibility.** TinyGo cannot compile `net/http` for browser WASM. Lofigui's direction requires `net/http` for shared server/WASM handlers.
2. **Compression closes the size gap.** Gzip brings standard Go WASM to ~2.1 MB on the wire, comparable to TinyGo uncompressed. Brotli reduces this further.
3. **Template compatibility.** Pongo2 (lofigui's template engine) uses reflection and will not work under TinyGo.
4. **Maintenance burden.** Dual TinyGo/standard Go builds doubled the build matrix and produced divergent WASM behaviour. Consolidating on one target eliminates this.

If TinyGo gains browser `net/http` support in the future, the decision can be revisited. The service worker architecture does not preclude TinyGo --- it just requires `net/http` to work.

## References

- [TinyGo](https://tinygo.org/) --- the alternative Go compiler
- [TinyGo net/http browser issue](https://github.com/tinygo-org/tinygo/issues/4420) --- the blocking issue
- [templ](https://github.com/a-h/templ) --- code-gen template engine that works with TinyGo
- [Minimizing Go WASM binary size](https://dev.bitolog.com/minimizing-go-webassembly-binary-size/)
- [WASM Service Workers research](research-wasm-service-workers.html) --- the service worker approach that requires `net/http`
