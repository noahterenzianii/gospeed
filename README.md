# gospeed

A Go terminal UI for internet speed testing via [LibreSpeed](https://librespeed.org/)-compatible servers,
with auto-tuning of stream count and buffer size.

## Install

```bash
go install github.com/noahterenzianii/gospeed/cmd/gospeed@latest
```

## Quick start

```bash
make run
```

## Commands

| Command | Description |
|---|---|
| `make run` | Build and launch |
| `make build` | Build binary |
| `make lint` | Run go vet |
| `make clean` | Remove binary |

## Keys

| Key | Action |
|---|---|
| `s` | Start test |
| `c` | Open/close configuration menu |
| `t` | Auto-tune stream count and buffer size |
| `r` | Reset config to defaults (in config menu only) |
| `↑` `↓` or `k` `j` | Navigate config menu |
| `+` `-` | Change config value |
| `esc` | Close config / dismiss tuning results |
| `q` | Quit |

## Configuration

| Setting | Default | Min – Max |
|---|---|---|
| ping samples | 200 | 10 – 1000 |
| transfer duration | 15s | 5s – 120s |
| download streams | 8 | 1 – 64 |
| upload streams | 4 | 1 – 64 |
| download buffer | 1024 KB | 64 KB – max buffer |
| upload buffer | 256 KB | 64 KB – max buffer |
| max streams | 64 | 4 – 256 |
| max buffer | 4096 KB | 256 KB – 65536 KB |
| client info timeout | 5.0s | 1s – 30s |
| server list timeout | 10.0s | 1s – 30s |
| ping timeout | 2.0s | 1s – 15s |
| max concurrent pings | 20 | 5 – 100 |
| ping attempts | 3 | 1 – 20 |

## Structure

```
cmd/gospeed/        main.go
internal/
├── endpoints/      API client, ping, server selection
├── speedtest/      download/upload measurement
├── tuning/         auto-tuning (stream count, buffer size)
└── tui/            terminal UI (Bubbletea)
```
