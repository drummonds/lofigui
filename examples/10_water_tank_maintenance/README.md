# Example 10 — Water Tank Maintenance (Background Operations)

Builds on example 09 (HTMX partial updates) by adding long-running background maintenance operations.

## Key concepts

- **Background goroutines** running alongside the simulation tick loop
- **`context.Context` cancellation** for clean shutdown of maintenance operations
- **Equipment lockout** — pump/valve controls disabled during their maintenance
- **Progress tracking** visible through existing HTMX 1-second polling (no new endpoints needed)
- **Single maintenance slot** — only one operation at a time

## Running

```bash
# Server with HTMX
task go-example:10

# WASM demo (built with build-pages)
task build:docs
```

## Routes

| Method | Route | Action |
|--------|-------|--------|
| GET | `/` | Full page with schematic |
| GET | `/diagnostics` | Full page with diagnostics |
| GET | `/fragment` | HTML fragment (schematic) |
| GET | `/fragment/diagnostics` | HTML fragment (diagnostics) |
| POST | `/start` | Start simulation |
| POST | `/stop` | Stop simulation |
| POST | `/pump` | Toggle pump |
| POST | `/valve` | Toggle valve |
| POST | `/maintenance/pump` | Start pump maintenance (~8s) |
| POST | `/maintenance/valve` | Start valve inspection (~5s) |
| POST | `/maintenance/cancel` | Cancel running maintenance |
| POST | `/maintenance/clear` | Dismiss result |
