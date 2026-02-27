# Example 08: Multi-Page Water Tank

Extends [Example 07](../07_water_tank/) with a second page, demonstrating that the HTTP `Refresh` header reloads the **current** page rather than always redirecting to `/`.

## Pages

- `/` — SVG schematic with controls (same as 07)
- `/diagnostics` — pump/valve cycle counts, on-time, float switch trips, level history

## Run

```bash
task go-example:08
# http://localhost:1348
```

## Key concept

Both pages call `app.HandleDisplay(w, r)`, which sets the HTTP `Refresh` header when polling is active. Each page reloads itself without being redirected elsewhere.
