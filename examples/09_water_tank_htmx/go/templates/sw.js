importScripts('wasm_exec.js');
importScripts('https://cdn.jsdelivr.net/gh/nlepage/go-wasm-http-server@v2.2.1/sw.js');

registerWasmHTTPListener('main.wasm', {
    passthrough: function(request) {
        var url = new URL(request.url);
        // Let external CDN resources through (Bulma, HTMX, etc.)
        if (url.hostname !== self.location.hostname) return true;
        // Let the bootstrap page through to the static host
        if (url.pathname.endsWith('demo-sw.html')) return true;
        return false;
    }
});
