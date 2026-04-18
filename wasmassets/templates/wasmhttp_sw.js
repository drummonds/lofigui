function registerWasmHTTPListener(wasm, { base, cacheName, passthrough, args = [] } = {}) {
  let path = new URL(registration.scope).pathname
  if (base && base !== '') path = `${trimEnd(path, '/')}/${trimStart(base, '/')}`

  const handlerPromise = new Promise(setHandler => {
    self.wasmhttp = { path, setHandler }
  })

  self.addEventListener('fetch', e => {
    const { url, method } = e.request

    if (passthrough && passthrough(e.request)) return

    if (new URL(url).pathname.startsWith(path)) {
      e.respondWith(handlerPromise.then(handler => handler(e.request)))
    }
  })

  const run = async () => {
    const go = new Go()
    go.argv = [wasm, ...args]

    let wasmResponse
    if (cacheName) {
      const cache = await caches.open(cacheName)
      wasmResponse = await cache.match(wasm)
    }
    if (!wasmResponse) wasmResponse = await fetch(wasm)

    const { instance } = await WebAssembly.instantiateStreaming(wasmResponse, go.importObject)
    go.run(instance)
  }

  run()
}

function trimStart(s, c) { return s.startsWith(c) ? s.slice(c.length) : s }
function trimEnd(s, c) { return s.endsWith(c) ? s.slice(0, -c.length) : s }
