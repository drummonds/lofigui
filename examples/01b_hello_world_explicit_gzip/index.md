<style>
.annotation { border-left: 3px solid #3273dc; background: #f0f4ff; padding: 0.75em 1em; margin: 0.75em 0; border-radius: 0 4px 4px 0; font-size: 0.9em; }
.annotation strong { color: #3273dc; }
</style>

# 01b — Hello World Explicit + gzip

Same as [01a](../01a_hello_world_explicit/index.html) — explicit routes, hand-written service worker — plus one optimisation: the WASM binary is shipped gzipped and decompressed in the browser. The Go code is byte-identical to 01a; all the interesting work happens in the SW bootstrap.

<div class="buttons">
<a href="wasm_demo/sw/" class="button is-primary">Launch Demo</a>
<a target="_blank" href="https://codeberg.org/hum3/lofigui/src/branch/main/examples/01b_hello_world_explicit_gzip" class="button is-light">Source on Codeberg</a>
</div>

---

## Why do this?

A Go WASM "hello world" is ~11 MB uncompressed and ~2.8 MB gzipped. Production static hosts (GitHub Pages, Cloudflare Pages, Netlify) negotiate `Content-Encoding: gzip` for you and you don't have to care. Hosts that don't — or when you want explicit control over which variant the client gets — this pattern serves a raw `.wasm.gz`, decompresses it client-side, and feeds the bytes to the SW.

<div class="annotation">
<strong>Why layer this on top of 01a rather than hide it behind an option?</strong> The decompression step has real moving parts: <code>DecompressionStream</code>, a named Cache entry whose URL must match the SW's resolved <code>main.wasm</code> path, and a hand-set <code>Content-Type</code>. Hiding those behind a flag makes debugging awful; showing them turns them into exhibits. If any of the pieces goes wrong, the symptom (usually a stuck "Compiling WASM..." spinner) is much easier to diagnose when you can read the five-step promise chain in front of you.
</div>

---

## The five-step bootstrap

Read `go/templates/sw/index.html` alongside this. Each numbered comment in that file corresponds to one step below:

1. **Fetch `main.wasm.gz` as-is.** The file is served with `Content-Type: application/gzip` (not `application/wasm` + `Content-Encoding: gzip`), so the browser hands us the raw compressed bytes instead of auto-decoding.
2. **Pipe through `DecompressionStream('gzip')`.** The browser's native streaming decompressor. Output is a `ReadableStream` of raw WASM bytes, consumed via `new Response(...).blob()`.
3. **Cache the decompressed blob.** Open `caches.open('wasm-gz-01b')`, `cache.put(wasmUrl, new Response(blob, {headers: {'Content-Type': 'application/wasm'}}))`. Two details matter here:
   - The cache name is example-scoped (`wasm-gz-01b`) so other gzipped examples can't collide.
   - The URL key is resolved from `location.href`, so it matches whatever `registerWasmHTTPListener('main.wasm', …)` asks for at runtime.
4. **Register the service worker** with `cacheName: 'wasm-gz-01b'`. The go-wasm-http-server runtime reads from the named cache first and only falls back to the network if that's a miss. Because step 3 just wrote a match, the SW finds the bytes instantly.
5. **Poll `./favicon.ico`.** The SW needs one fetch to instantiate the WASM and register its handler. Favicon is the cheapest probe; when it comes back 200, redirect to `./` which the SW now serves from the Go app.

```js
// go/templates/sw/index.html — the core promise chain
fetch('main.wasm.gz')
  .then(resp => {
    const ds = new DecompressionStream('gzip');
    return new Response(resp.body.pipeThrough(ds)).blob();
  })
  .then(blob => caches.open('wasm-gz-01b').then(cache => {
    const wasmUrl = new URL('main.wasm', location.href).href;
    return cache.put(wasmUrl, new Response(blob, {
      headers: {'Content-Type': 'application/wasm'}
    }));
  }))
  .then(() => navigator.serviceWorker.register('sw.js'))
  .then(() => navigator.serviceWorker.ready)
  .then(() => waitForWasm());
```

---

## What the SW itself looks like

Almost the same as 01a — just two lines different (the `cacheName` and the variant-specific cache scope):

```js
// go/templates/sw/sw.js
importScripts('wasm_exec.js');
importScripts('wasmhttp_sw.js');

self.addEventListener('install',  () => self.skipWaiting());
self.addEventListener('activate', e => e.waitUntil(clients.claim()));

registerWasmHTTPListener('main.wasm', {
    cacheName: 'wasm-gz-01b',   // ← the one meaningful line vs 01a
    passthrough: function(request) {
        var url = new URL(request.url);
        if (url.hostname !== self.location.hostname) return true;
        if (url.pathname.endsWith('/index.html')) return true;
        return false;
    }
});
```

---

## Recovering from a stuck service worker

Every SW-based demo in this repo ships a tiny recovery stub next to its main entry point. For 01b that's `wasm_demo/demo-sw.html` — a static HTML page served by the host (not the SW), which runs a short script to unregister any SW whose scope is a prefix of the current URL and then redirects to the canonical `./sw/` entry.

If the demo gets stuck ("Compiling WASM..." spinning indefinitely, or a blank page), navigate to:

```
/01b_hello_world_explicit_gzip/wasm_demo/demo-sw.html
```

That path is outside the SW's `sw/` scope, so the static host serves it unconditionally. It cleans up any ancestor-scoped registration and redirects — the bootstrap page it lands on does its own unregister-before-register, so a child-scoped stale SW gets replaced too. If even that fails, you're in DevTools territory (Application → Service Workers → Unregister + Cache Storage → Delete `wasm-gz-01b`).

---

## Failure modes to know about

| Symptom | Usual cause |
|---------|-------------|
| Stuck "Compiling WASM..." forever | Cache URL mismatch. The SW is asking for `…/main.wasm` but the bootstrap wrote to a different URL. Make sure `new URL('main.wasm', location.href).href` resolves the same way in both places. |
| "Incorrect response MIME type" error | Forgot to set `Content-Type: application/wasm` on the cached `Response`. `WebAssembly.instantiateStreaming` refuses to parse non-wasm content types. |
| Works once, fails after reload | Stale cached blob from a previous build. Bump the cache name (append a version), or clear site data in DevTools. |
| Works in Chrome, stuck in Firefox | Usually a loading-order issue — register the SW only AFTER `cache.put` resolves, and use `updateViaCache: 'none'` to prevent HTTP-cache shadowing. |

---

## Running

```bash
# Server mode (no gzip involved — server serves the normal binary)
task go-example:01b
# Opens http://localhost:1342

# WASM demo (via docs)
task docs:build-wasm
tp pages
# Navigate to http://localhost:8080/01b_hello_world_explicit_gzip/wasm_demo/sw/
```
