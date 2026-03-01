# Example 09 — Water Tank HTMX

Same water tank simulator as example 08, but replaces HTTP `Refresh` header polling with **HTMX partial updates**.

## How it works

- Full page loads on initial visit and form submissions (Start/Stop/Pump/Valve)
- The `#results` div has `hx-get` + `hx-trigger="every 1s"` — HTMX fetches a fragment endpoint and swaps just the inner HTML
- No full page reload on each poll — only the results content updates
- Schematic (`/`) and diagnostics (`/diagnostics`) each poll their own fragment endpoint

## Routes

| Route | Returns |
|---|---|
| `GET /` | Full page — schematic view |
| `GET /diagnostics` | Full page — diagnostics view |
| `GET /fragment` | HTML fragment — schematic only |
| `GET /fragment/diagnostics` | HTML fragment — diagnostics only |
| `POST /start` | Start simulation, redirect `/` |
| `POST /stop` | Stop simulation, redirect `/` |
| `POST /pump` | Toggle pump, redirect `/` |
| `POST /valve` | Toggle valve, redirect `/` |

## Running

```bash
task go-example:09
```

Then visit http://localhost:1349
