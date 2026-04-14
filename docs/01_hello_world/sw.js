importScripts('wasm_exec.js');
importScripts('https://cdn.jsdelivr.net/gh/nlepage/go-wasm-http-server@v2.2.1/sw.js');

const scope = new URL('./', self.location.href).pathname;

registerWasmHTTPListener('main.wasm', {
    base: scope,
    passthrough: function(request) {
        var url = new URL(request.url);
        if (url.hostname !== self.location.hostname) return true;
        if (url.pathname.endsWith('demo.html')) return true;
        if (url.pathname.endsWith('demo-gz.html')) return true;
        return false;
    }
});
