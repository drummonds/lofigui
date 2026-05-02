# 06 — Notes CRUD

In-memory notes app with a **master / detail UI** (per-row Read / Edit / Delete buttons), the **Post / Redirect / Get** pattern, and a 4 KiB cap per note. The Go build runs in two targets (server and WASM) from the same `*http.ServeMux`; the Python build uses FastAPI.

See [`index.md`](./index.md) for the rendered documentation page (three screenshots — initial / populated / detail — plus the curl-driven integration test).

## Quick start

```bash
task go-example:06       # Go server                http://localhost:1340
task example-06          # Python (FastAPI)         http://localhost:1340

# WASM demo (browser-only, served from docs/):
task docs:build-wasm     # builds main.wasm into docs/06_notes_crud/wasm_demo/
tp pages                 # serves docs/ on http://localhost:8080
# Visit http://localhost:8080/06_notes_crud/wasm_demo/
```

The Go build also has a curl-driven CRUD test that doubles as the screenshot capture:

```bash
task docs:capture:06     # asserts every POST returns 303 + greps the resulting SVG
```

## Project structure

```
06_notes_crud/
├── README.md
├── index.md                    # Rendered docs page (screenshots, walkthrough)
├── go/
│   ├── main.go                 # //go:build !(js && wasm) — server entry
│   ├── main_wasm.go            # //go:build js && wasm    — WASM entry (SW)
│   ├── model.go                # Shared: notesDB, flash, CRUD ops, buildMux
│   ├── go.mod / go.sum
│   └── templates/
│       └── notes.html          # <base href="{{.base}}"> + {{.content}}
└── python/
    ├── notes.py                # FastAPI implementation (no flash bridge)
    ├── pyproject.toml          # uv-managed; lofigui pinned to ../../../
    ├── uv.lock
    └── templates/
        └── notes.html
```

## What the Go and Python builds share

- Same in-memory map (`{1: "First note…", 2: "Second…", 3: "Third…"}`, `nextID = 4`)
- Same four CRUD form actions (`create`, `read`, `update`, `delete`)
- Same redirect-after-POST flow (`303 See Other` → `GET /`)
- Same Bulma-styled table and notification colours

## What only the Go build has

- A single shared `*http.ServeMux` between server and WASM via build tags
- `lofigui.NewControllerFromFS` reading `//go:embed templates`
- A package-level **flash variable** so notifications survive the redirect (the simpler Python version puts notifications inside the POST handler before the redirect, so the message is lost on next render — see `index.md` for why the Go version splits these)

## License

Part of the lofigui project; shares the same license.
