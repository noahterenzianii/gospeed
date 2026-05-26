# gospeed

A Go terminal UI for measuring internet speed via [LibreSpeed](https://librespeed.org/)-compatible servers. Automatically selects the best endpoint and displays real-time results with sparklines.

## Quick start

```bash
make run
```

## Usage

| Command | Description |
|---|---|
| `make run` | Build and launch the TUI |
| `make build` | Build binary to `./gospeed` |
| `make lint` | Run `go vet ./...` |
| `make clean` | Remove binary |
### TUI controls

| Key | Action |
|---|---|
| `q` / `Ctrl+C` | Quit |
| `r` | Redo the test (shown after completion) |

## Features

- **Adaptive color palette** — automatically adjusts contrast for light and dark terminals
- **Real-time sparklines** — visual trend of download/upload speed during the test
- **Organized sections** — client info, server, latency, download, upload, results
## Project layout

```text
cmd/              # entrypoints
└── gospeed/
internal/
├── app/          # CLI-mode orchestration
├── endpoints/    # LibreSpeed API client, ping, server selection
├── speedtest/    # download/upload measurement with progress callbacks
└── tui/          # Bubbletea TUI — model, views, styling, commands
```
