# Example 11 — SeaweedFS Integration

Demonstrates lofigui + HTMX with a live [SeaweedFS](https://github.com/seaweedfs/seaweedfs) cluster. Upload test files and watch them appear in real-time via 1-second HTMX polling.

## Prerequisites

Install SeaweedFS: download a release binary from https://github.com/seaweedfs/seaweedfs/releases

Start a local server:

```bash
weed server -dir=/tmp/seaweed-data
```

This starts master (port 9333) + volume (port 8080) + filer (port 8888).

## Running

```bash
task go-example:11
```

Open http://localhost:1350 — the UI shows connection status and lets you create/verify/delete files.

## Routes

| Method | Route | Action |
|--------|-------|--------|
| GET | `/` | Full page |
| GET | `/fragment` | HTMX fragment |
| POST | `/create` | Create a test file |
| POST | `/verify?i=N` | Read file back and verify |
| POST | `/delete?i=N` | Delete file from SeaweedFS |
| POST | `/auto/start` | Auto-create a file every 3s |
| POST | `/auto/stop` | Stop auto-create |
