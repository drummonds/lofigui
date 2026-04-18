importScripts('wasm_exec.js');
importScripts('wasmhttp_sw.js');

self.addEventListener('install', function() { self.skipWaiting(); });
self.addEventListener('activate', function(event) {
  event.waitUntil(clients.claim());
});

registerWasmHTTPListener('main.wasm', {
    passthrough: function(request) {
        var url = new URL(request.url);
        if (url.hostname !== self.location.hostname) return true;
        if (url.pathname.endsWith('/index.html')) return true;
        return false;
    }
});
