# Research: WASM Accessing External APIs

How a Go WASM program running inside a service worker can talk to external HTTP APIs, and why the obvious approach doesn't work.

## The problem

Once you've adopted the [WASM service worker pattern](research-wasm-service-workers.html), every request inside your scope is intercepted and routed into Go's `net/http` mux. The WASM is no longer a leaf in the browser --- it sits *behind* the service worker and *in front of* the page. This is great for serving HTML fragments to HTMX, but it creates a subtle problem the moment your handler needs to call an upstream service:

```go
func handleWeather(w http.ResponseWriter, r *http.Request) {
    // Inside a WASM service worker handler, this is the problem case:
    resp, err := http.Get("https://api.openweathermap.org/data/2.5/weather?q=London")
    ...
}
```

The intuition that "WASM can't make network calls" is wrong --- Go's `net/http` on `GOOS=js GOARCH=wasm` is a perfectly real HTTP transport that wraps the browser's `fetch()` API. The actual issue is more interesting and lives in three layers.

## Why the naive call breaks

### Layer 1: The routing loop

When Go code inside the service worker calls `http.Get(...)`, the `js/wasm` HTTP transport calls `fetch()` via `syscall/js`. **Fetches initiated from inside a service worker still pass through that service worker's own fetch event listeners.** This is documented browser behaviour and surprises almost everyone the first time.

So the call sequence is:

```
WASM handler -> http.Get("https://api.example.com/foo")
             -> Go js/wasm transport -> fetch()
             -> Service worker fetch event fires
             -> registerWasmHTTPListener routes it back to the WASM mux
             -> mux returns 404 (no handler for api.example.com/foo)
             -> Go transport sees a 404 response
```

The WASM is talking to itself. No request ever leaves the browser.

### Layer 2: Passthrough is not enough

`go-wasm-http-server` provides a `passthrough` callback that lets requests bypass the WASM router and go to the real network. Lofigui already uses it for CDN assets:

```js
passthrough: function(request) {
    var url = new URL(request.url);
    if (url.hostname !== self.location.hostname) return true;  // CDN, etc.
    return false;
}
```

This solves the routing loop --- a passthrough request goes to the network --- but the response goes back to **whoever issued the fetch**. If the original caller was the page (e.g. an `<img src="https://cdn.../bulma.css">`), the page receives the response. If the original caller was Go code inside the WASM, Go *also* receives the response, because `passthrough` does not care who initiated the fetch. So technically this can work for outbound calls from Go.

The catch is that passthrough is a single global function evaluated for every request. Mixing "WASM-internal external calls" with "page-direct CDN requests" in the same predicate gets messy fast, and there is no opportunity to inject auth headers, rewrite URLs, cache responses, or apply per-call policy.

### Layer 3: CORS is still in force

Service worker fetches are subject to the same CORS rules as page fetches. If `api.example.com` does not return `Access-Control-Allow-Origin: <your-static-host>`, the browser blocks the response body. You can request `mode: 'no-cors'` to get an opaque response back, but you cannot read it --- which defeats the point.

CORS is the constraint that no amount of WASM cleverness fixes. If the upstream API does not permit your origin, you need a real server somewhere (a CORS-relaxing proxy) and you've left the static-hosting model behind. This research page is about everything *up to* that wall.

## Approach A: Passthrough (the lightweight option)

For an external API that supports CORS for your origin and that Go code can call without any header massaging, the existing passthrough mechanism is enough. Add the upstream host to the predicate and Go's `http.Get` will work transparently:

```js
registerWasmHTTPListener('main.wasm', {
    base: scope,
    passthrough: function(request) {
        var url = new URL(request.url);
        if (url.hostname === 'api.openweathermap.org') return true;
        if (url.hostname !== self.location.hostname)   return true;
        return false;
    }
});
```

```go
resp, err := http.Get("https://api.openweathermap.org/data/2.5/weather?q=London")
```

| Pros | Cons |
|------|------|
| Zero new code on the Go side | No place to inject auth headers (Go has to do it) |
| Zero new code on the JS side beyond the predicate | No URL rewriting, caching, allowlisting |
| Standard `net/http` semantics | Predicate gets crowded with mixed concerns |
| | API key ends up baked into the WASM binary |

Use this when: the upstream API is public, supports CORS, and you don't need secrets. Weather feeds, public datasets, status endpoints, GitHub's public API.

## Approach B: HTTP-path proxy in the service worker (recommended)

Install a fetch listener *before* `registerWasmHTTPListener` that claims any request matching a reserved URL prefix and forwards it to the real upstream. Go code calls `/_proxy/<encoded-target>` and gets back exactly what the upstream returned.

```js
// sw.js --- this listener must be added BEFORE the WASM router listener
self.addEventListener('fetch', function(event) {
    var url = new URL(event.request.url);
    if (url.pathname.startsWith('/_proxy/')) {
        event.respondWith(handleProxy(event.request));
    }
    // else: don't call respondWith --- let the WASM listener handle it
});

async function handleProxy(req) {
    var url = new URL(req.url);
    // /_proxy/https://api.example.com/v1/foo
    var target = url.pathname.slice('/_proxy/'.length) + url.search;

    var allowed = ['api.openweathermap.org', 'api.example.com'];
    var targetUrl = new URL(target);
    if (!allowed.includes(targetUrl.hostname)) {
        return new Response('upstream not allowed', { status: 403 });
    }

    var init = {
        method: req.method,
        headers: req.headers,
        body: ['GET','HEAD'].includes(req.method) ? undefined : await req.blob(),
    };

    try {
        var upstream = await fetch(targetUrl.toString(), init);
        return new Response(upstream.body, {
            status:  upstream.status,
            headers: upstream.headers,
        });
    } catch (err) {
        return new Response(JSON.stringify({ error: err.message }), {
            status:  502,
            headers: { 'Content-Type': 'application/json' },
        });
    }
}

importScripts('wasm_exec.js');
importScripts('https://cdn.jsdelivr.net/gh/nlepage/go-wasm-http-server@v2.2.1/sw.js');
registerWasmHTTPListener('main.wasm', { base: '/' });
```

```go
// Inside the WASM handler --- looks like a normal HTTP call
resp, err := http.Get("/_proxy/https://api.openweathermap.org/data/2.5/weather?q=London")
```

### Why this is the cleanest fit for lofigui

- **The Go side stays standard `net/http`.** No `syscall/js`, no Promise/channel plumbing, no separate code path for "external" vs "internal" calls. The same handler can run on the server build by stripping the `/_proxy/` prefix in a tiny middleware (or by setting up an equivalent route on the real server).
- **One place to enforce policy.** Allowlists, auth header injection, response caching via the Cache API, request budgeting, log-and-trace --- all live in `handleProxy`. The Go code stays oblivious.
- **No double-fetch.** The proxy listener calls `respondWith` before the WASM router listener gets a chance, so the request never enters the WASM mux. No loop.
- **Auth without baking secrets into the WASM binary.** API keys can live in `sw.js` (visible to anyone who fetches it from your static host --- so still not *secret*, but at least one layer removed from the WASM blob). For real secrets you still need a server.

### The listener ordering subtlety

This pattern relies on a specific browser quirk: a service worker can have multiple `fetch` event listeners, they all fire, but only the first one to call `event.respondWith()` wins. Subsequent calls to `respondWith` on the same event throw. So the proxy listener does not need to "intercept" anything special --- it just needs to be registered *before* `registerWasmHTTPListener`, and to call `respondWith` only for the proxy paths it cares about.

If you accidentally register the WASM listener first, the WASM router will see proxy paths, hit a 404, and respond 404 --- and your proxy listener will never get a chance. The fix is just import order in `sw.js`.

## How the two servers coexist

Once you've added the proxy listener you have **two HTTP servers living inside the same service worker**: the existing `go-wasm-http-server` (which runs your Go `net/http` mux as WASM) and the new proxy listener (which runs as plain JavaScript). They share the global scope, both register `fetch` event listeners, and both see every request that touches the SW. This section spells out exactly how they relate, because the mental model is what makes or breaks debugging.

### Two listeners, one event, one winner

Every fetch in the SW's scope fires a `FetchEvent`. The browser dispatches that event to **all** registered listeners --- there is no priority, no short-circuit, and no concept of "the request was already handled". What `respondWith()` actually does is mark the event as having a response promise; the *first* call wins, and any subsequent call on the same event throws `InvalidStateError`. Both listeners therefore have to make a synchronous decision at the top of their handler: "is this mine, or do I let it fall through?"

```js
// Listener 1 (proxy) --- runs first because it was added first
self.addEventListener('fetch', function(event) {
    if (new URL(event.request.url).pathname.startsWith('/_proxy/')) {
        event.respondWith(handleProxy(event.request));   // <-- claim
    }
    // else: return without calling respondWith --- listener 2 gets to decide
});

// Listener 2 (registerWasmHTTPListener, added by importScripts) --- runs second
// Internally does the equivalent of:
self.addEventListener('fetch', function(event) {
    if (passthrough(event.request)) return;              // <-- network
    event.respondWith(routeIntoWasmMux(event.request));  // <-- claim
});
```

The key invariant: **the proxy listener must decide synchronously**. You cannot `await` anything before calling `respondWith` --- if you do, the WASM listener has already run, and either it claimed the event (in which case your `respondWith` throws) or it didn't (in which case the browser has already moved on and is giving the caller a network error). Async work happens *inside* the promise passed to `respondWith`, not before.

### Routing matrix

For a request hitting the SW, here is which subsystem actually handles it:

| Request                                          | Proxy listener | WASM listener (`registerWasmHTTPListener`) | Where it ends up                |
|--------------------------------------------------|----------------|--------------------------------------------|---------------------------------|
| `/index.html`, `/main.wasm`, `/sw.js`            | ignores        | claims                                     | WASM mux (or static fallback)   |
| `/fragment/tank` (HTMX)                          | ignores        | claims                                     | WASM `net/http` handler         |
| `/_proxy/https://api.example.com/v1/foo`         | **claims**     | never sees it                              | Real upstream via `fetch()`     |
| `https://cdn.jsdelivr.net/...`                   | ignores        | passthrough returns true                   | Real network                    |
| `/favicon.ico`                                   | ignores        | claims                                     | WASM mux                        |
| WASM-internal `http.Get("/fragment/tank")`       | ignores        | claims                                     | Same WASM mux (loops back in)   |
| WASM-internal `http.Get("/_proxy/https://...")`  | **claims**     | never sees it                              | Real upstream                   |

The last two rows are the interesting ones. A `http.Get` from inside the Go handler is itself a fetch through the SW. If the path is `/_proxy/...` the proxy claims it and the upstream call happens. If the path is a normal mux route the WASM listener claims it and the call recurses into the same Go process. Both cases are well-defined --- there is no ambiguity --- as long as the proxy listener was registered first.

### The two servers are independent

This is the design property that matters most:

- **No code sharing.** The proxy listener is plain JavaScript that uses only standard service worker APIs (`fetch`, `Response`, `URL`, `Cache`). It does not call any function from `go-wasm-http-server`, does not read or set any of its globals, and does not need to know the WASM is there.
- **No build coupling.** Editing `sw.js` to change proxy behaviour does not require rebuilding the WASM binary. Conversely, rebuilding the WASM does not invalidate the proxy listener --- the SW just has a fresh `main.wasm` to instantiate the next time the WASM listener is invoked.
- **No version pinning.** You can upgrade `go-wasm-http-server` to a new tag and the proxy listener is unaffected, because it does not depend on the library's internal API. The only contract is "the WASM listener does not call `respondWith` for `/_proxy/*` paths", which is guaranteed by the proxy listener running first.
- **No shared state.** Each listener has its own closures. They do not share counters, caches, or auth tokens unless you explicitly thread them through `self.*` globals or the Cache API.

In effect, the proxy is a separate microservice that happens to share an address space with the WASM HTTP server. You could lift it into its own `proxy.js` file and `importScripts` it from `sw.js` --- which is what the recommendation at the bottom of this page suggests for lofigui.

### Lifecycle interaction

Service workers have a non-trivial lifecycle (install, activate, idle, terminate, restart) and the two servers behave differently at each stage:

| Stage            | Proxy listener                                  | WASM HTTP server                                                |
|------------------|-------------------------------------------------|-----------------------------------------------------------------|
| Install          | Available as soon as `sw.js` finishes parsing    | Library code loaded; WASM binary not yet instantiated           |
| Activate         | Ready                                            | Ready (registers fetch listener)                                |
| First fetch      | Handles `/_proxy/*` immediately                  | Lazily instantiates `main.wasm` on the first matching request   |
| Idle/terminate   | Listener persists across termination             | WASM heap is lost; reinstantiated on next fetch                 |
| Update           | New `sw.js` replaces both listeners atomically   | Same                                                            |

There is a useful side-effect here: **the proxy listener is available before the WASM is loaded**. If your app needs to make an external API call during a "loading" page that runs before the WASM is ready (e.g. fetching a license check, prefetching config), the proxy works without waiting for `main.wasm` to instantiate. This is hard to achieve any other way on a static host.

The same property cuts the other way during failure: if the WASM panics on startup or the binary is corrupt, **the proxy still works**. You can build a tiny error page that fetches diagnostics from an external service through `/_proxy/...` even when the main app is broken.

### Failure modes and how they propagate

| Failure                                                  | What the caller sees                                                                              |
|----------------------------------------------------------|---------------------------------------------------------------------------------------------------|
| `handleProxy` throws synchronously                       | The promise passed to `respondWith` rejects → caller gets a network error. WASM does **not** get a fallback chance --- `respondWith` was already called. |
| `handleProxy` returns a rejected promise                 | Same as above --- network error to caller.                                                       |
| `handleProxy` returns a 502 `Response`                   | Caller sees a normal HTTP 502 and can handle it. **This is the preferred failure shape** --- prefer returning an error `Response` over throwing. |
| Upstream is unreachable                                  | `fetch()` rejects inside `handleProxy`; the `try/catch` converts it to a 502.                    |
| Upstream is reachable but blocks CORS                    | `fetch()` returns an opaque response with `status: 0`; if you forward it, the caller gets an empty body. The fix is server-side (Approach D), not in the proxy. |
| Proxy listener registered *after* WASM listener          | WASM listener 404s on `/_proxy/*` paths before the proxy ever sees them. Symptom: "my proxy never runs". Fix: import order in `sw.js`. |
| WASM is still loading when proxy is hit                  | Proxy works fine --- it doesn't depend on WASM.                                                  |
| WASM crashes during the request                          | Affects only the WASM listener's response. Proxy is unaffected.                                  |

The takeaway: **return error `Response` objects from `handleProxy`, never throw**. Throwing makes the failure invisible to Go code (it just sees `connection refused`) and bypasses any logging or telemetry you might add.

### Interaction with a real backend (Approach D)

If you also adopt Approach D --- a real BFF server somewhere on the internet --- the picture becomes a three-tier system:

```
Browser page
  -> Service worker
       -> Proxy listener (JS)         -> Real backend (Approach D)
       -> WASM HTTP server (Go)        -> Lofigui handlers
                                            -> http.Get("/_proxy/https://backend.example/api/...")
                                                 -> back through the SW, claimed by proxy listener
                                                      -> outbound fetch to real backend
```

The Go handler does not know whether `/_proxy/https://backend.example/...` ends up in a real upstream or in a Cloudflare Worker --- it just calls `http.Get` and reads the response. The proxy listener is the seam where deployment-specific routing lives. This means you can develop locally with the proxy pointing at `localhost:8080`, ship to production with the proxy pointing at the real BFF, and the WASM binary is identical in both deployments.

### Trade-offs

| Pros | Cons |
|------|------|
| Standard Go `net/http` calls | One file (`sw.js`) becomes load-bearing |
| Single place for auth, allowlist, caching, logging | API keys in `sw.js` are visible to anyone on the origin |
| Same Go handler runs on server (with a trivial mux entry) | Adds ~50 lines of JS to maintain |
| Composes naturally with existing `passthrough` (CDN passthrough still works) | Listener ordering is a sharp edge to remember |
| Easy to test in DevTools (Network tab shows real upstream calls) | |

Use this when: you need any combination of auth injection, allowlisting, caching, or "the handler runs on both server and WASM and I don't want a fork in the code".

## Approach C: `syscall/js` bridge to a JS function

Skip HTTP entirely. Define a JS function on `self`, call it from Go via `syscall/js`, await the returned Promise via a channel.

```js
// sw.js
self.proxyFetch = async function(url, opts) {
    var resp = await fetch(url, opts || {});
    return await resp.text();
};
```

```go
// Go side --- helper that turns a JS Promise into a blocking Go call
func proxyFetch(url string) (string, error) {
    promise := js.Global().Call("proxyFetch", url)
    done := make(chan struct{})
    var result string
    var fetchErr error

    promise.Call("then",
        js.FuncOf(func(this js.Value, args []js.Value) any {
            result = args[0].String()
            close(done)
            return nil
        }),
        js.FuncOf(func(this js.Value, args []js.Value) any {
            fetchErr = errors.New(args[0].Get("message").String())
            close(done)
            return nil
        }),
    )

    <-done
    return result, fetchErr
}
```

| Pros | Cons |
|------|------|
| No HTTP serialization overhead | Forks the Go code path --- server build needs different code |
| Can pass binary data, streams, structured types directly | Needs a `js.FuncOf`/channel helper at every call site |
| Free from the listener-ordering subtlety | `syscall/js` types are clunky and easy to leak |
| | Doesn't compose with the rest of the `net/http`-shaped lofigui surface |

Use this when: you're calling from low-level WASM code that isn't already inside an `http.Handler`, or you need binary streaming that HTTP serialization would mangle. For ordinary lofigui handlers, Approach B is strictly nicer.

## Approach D: A real backend somewhere

If the upstream API does not support CORS for your origin and you need to read the response body, no amount of service worker cleverness will save you. The browser will block the body. You need a server that:

1. Receives the request from your static-hosted WASM
2. Calls the upstream on your behalf (server-to-server, no CORS in play)
3. Returns the response with `Access-Control-Allow-Origin` set for your origin

This is just a CORS proxy / BFF and is a well-trodden path --- a 30-line Go program on Fly, a Cloudflare Worker, or `cors-anywhere` self-hosted. The relevant point for this research page is that **adopting Approach D moves you off pure static hosting**, which is the whole reason you went down the WASM service worker route in the first place. So if you reach for Approach D, ask whether the simpler answer is to drop the static-hosting constraint and run a real lofigui server.

## Comparison

| | A. Passthrough | B. Path proxy | C. JS bridge | D. Real backend |
|---|---|---|---|---|
| Static hosting only | Yes | Yes | Yes | **No** |
| Bypasses CORS | No | No | No | Yes |
| Standard Go `net/http` | Yes | Yes | No | Yes |
| Auth injection in JS | No | Yes | Yes | Yes (server) |
| Code shared with server build | Trivial | Trivial (small middleware) | Forked | Same as today |
| Lines of new code | ~3 | ~50 (JS) | ~30 (Go + JS) | ~30 (server) |
| Sharp edges | Predicate sprawl | Listener ordering | `syscall/js` plumbing | Hosting and ops |

## Security considerations

Anything in `sw.js` is fetchable from the static host. Treat it as **public configuration**, not as a secret store:

- API keys placed in `sw.js` are visible to anyone who knows where to look. They are *one step removed from being baked into the WASM*, which is a small improvement (you can rotate keys without rebuilding the binary) but not a real security boundary.
- For genuinely secret keys, use Approach D --- the secret lives on the server, never on any static asset.
- An open path proxy (`/_proxy/<anything>`) turns your origin into a CORS-bypass relay for whatever upstream you forward to. Always allowlist hostnames in `handleProxy`. An attacker can't directly use this from another origin (browser CORS would still apply at the original-page level), but an XSS on your own pages becomes much more useful.
- Set `Cache-Control` headers on proxy responses deliberately. The default behaviour (whatever the upstream sent) may be wildly inappropriate for a service-worker-cached environment.
- Consider rate-limiting at the proxy layer if the upstream charges per request.

## Operational notes

- **Allowlist with intent.** A growing allowlist in `handleProxy` is a code smell. If you have more than half a dozen upstreams, generate the list from a config file at build time.
- **Cache responses with the Cache API.** Service workers have direct access to `caches.open(...)` and `cache.match(...)`. The proxy is the natural place to add a "30-second cache for GET requests with no auth" policy.
- **Log to the page via `postMessage`.** `console.log` inside the service worker shows up in DevTools but not on the page. For demos, post a message back to clients on every proxied call so the user sees what just happened.
- **Test with `chrome://serviceworker-internals/`.** Forcing a service worker update is annoying; bookmark this page.
- **Force-reload bypasses everything.** `Ctrl+Shift+R` skips the service worker entirely, so neither the WASM nor the proxy will run. Plain reload is what you want during development.

## Recommendation for lofigui

Adopt **Approach B** as the standard pattern, with Approach A reserved for the trivial public-API case. Specifically:

1. Document the `/_proxy/<absolute-url>` convention as the lofigui way to call external APIs from a WASM service worker handler.
2. Provide a small `proxy.js` snippet in the `examples/` tree that other examples can `importScripts` from `sw.js` --- so the listener-ordering and allowlist boilerplate lives in one place.
3. On the server build, add a one-line middleware that strips the `/_proxy/` prefix and proxies upstream. This makes handlers identical between the two builds.
4. Build a pilot example (provisional name `13_external_api`) that fetches from one or two real public APIs (e.g. weather, GitHub status) using the path proxy, with both server and service worker entry points.
5. Update the [WASM Service Workers research page](research-wasm-service-workers.html) to cross-reference this page from the "No real network access" limitation.

If a future example needs CORS-bypassed access to an API that does not allow the origin, that's the trigger to step back and consider whether the example should run on a real server instead of in a service worker. Adding a hosted CORS proxy to a static site is the worst of both worlds.

## References

- [Fetch API in service workers](https://developer.mozilla.org/en-US/docs/Web/API/Service_Worker_API) --- MDN reference for service worker fetch interception
- [FetchEvent.respondWith()](https://developer.mozilla.org/en-US/docs/Web/API/FetchEvent/respondWith) --- the "first listener wins" semantics
- [CORS](https://developer.mozilla.org/en-US/docs/Web/HTTP/CORS) --- the constraint that no client-side trick removes
- [go-wasm-http-server](https://github.com/nlepage/go-wasm-http-server) --- the library lofigui uses to run `net/http` inside a service worker
- [Cache API](https://developer.mozilla.org/en-US/docs/Web/API/Cache) --- for caching proxied responses inside the service worker
- [WASM Service Workers research](research-wasm-service-workers.html) --- the broader context this page sits inside
