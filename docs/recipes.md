# Recipes: Screenshot Capture

How to automatically capture screenshots from running examples for documentation.

## How it works

1. Start the example server with `LOFIGUI_HOLD=1` (keeps server alive after model completes)
2. Trigger the model by hitting the server URL
3. Capture at timed intervals using [url2svg](https://github.com/nicholasgasior/url2svg) (headless Chrome → SVG)
4. Kill the server

## Taskfile commands

| Command | Description |
|---------|-------------|
| `task docs:capture:01` | Auto-capture polling + complete screenshots for example 01 |
| `task docs:capture` | Capture all example screenshots |
| `task docs:open:01` | Start example 01 in HOLD mode for manual inspection |

## Per-example capture config

| # | Port | Captures | Timing |
|---|------|----------|--------|
| 01 | 1340 | `01_polling.svg` (mid-run), `01_complete.svg` (done) | 2s delay for polling, 7s for complete |

## Adding captures for a new example

1. Add a `docs:capture:NN` task to `Taskfile.yml`:
   - Set `dir` to the example's `go/` directory
   - Start server with `LOFIGUI_HOLD=1 go run . &`
   - Wait for port with curl loop
   - Trigger model, then capture at appropriate delays
   - Use `trap` to kill server on exit

2. Add the task to `docs:capture`'s dependency list

3. Reference the SVGs from the example's docs page

## Manual inspection

`task docs:open:NN` starts the server in HOLD mode without auto-capture. The server stays running until Ctrl-C, so you can:

- Open the page in a browser to check layout
- Run `url2svg` manually with different options
- Test the cancel button mid-flow

## url2svg options

```bash
url2svg --url http://localhost:1340 -o output.svg           # default 1024x768
url2svg --url http://localhost:1340 -o output.svg -width 800  # custom width
url2svg --url http://localhost:1340 -o output.svg -full-page  # full scroll height
```
