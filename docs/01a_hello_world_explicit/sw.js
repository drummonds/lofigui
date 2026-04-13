importScripts('wasm_exec.js');
importScripts('https://cdn.jsdelivr.net/gh/nlepage/go-wasm-http-server@v2.2.1/sw.js');

// Determine base path from service worker scope (matches deployment directory)
const scope = new URL('./', self.location.href).pathname;

registerWasmHTTPListener('main.wasm', {
    base: scope,
    passthrough: function(request) {
        var url = new URL(request.url);
        // Let external CDN resources through (Bulma CSS, etc.)
        if (url.hostname !== self.location.hostname) return true;
        // Let the bootstrap page through to the static host
        if (url.pathname.endsWith('demo-sw.html')) return true;
        return false;
    }
});
