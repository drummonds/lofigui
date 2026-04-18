// Gzipped variant: the bootstrap page has already decompressed main.wasm.gz
// and written it into cache 'wasm-gz-01b' under the URL 'main.wasm'. The
// go-wasm-http-server library reads from that cache (instead of the network)
// because of the cacheName option below.
//
// The cache name is example-scoped so other gzipped examples can't stomp on
// this one's cached binary.

importScripts('wasm_exec.js');
importScripts('wasmhttp_sw.js');

self.addEventListener('install', function() { self.skipWaiting(); });
self.addEventListener('activate', function(event) {
  event.waitUntil(clients.claim());
});

registerWasmHTTPListener('main.wasm', {
    cacheName: 'wasm-gz-01b',
    passthrough: function(request) {
        var url = new URL(request.url);
        if (url.hostname !== self.location.hostname) return true;
        if (url.pathname.endsWith('/index.html')) return true;
        return false;
    }
});
