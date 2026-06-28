# gospeed

A Go terminal UI for internet speed testing via [LibreSpeed](https://librespeed.org/)-compatible servers.

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
| `r` | Redo test / reset config values |
| `c` | Open/close configuration menu |
| `↑` `↓` | Navigate config menu |
| `+` `-` | Change config value |
| `q` | Quit |

## Configuration

| Setting | Default | Min – Max |
|---|---|---|
| ping samples | 200 | 10 – 1000 |
| transfer duration | 15s | 5s – 120s |
| download streams | 8 | 1 – 64 |
| upload streams | 4 | 1 – 64 |
| download buffer | 1024 KB | 64 KB – 16384 KB |
| upload buffer | 256 KB | 64 KB – 16384 KB |
| client info timeout | 5.0s | 1s – 30s |
| server list timeout | 10.0s | 1s – 30s |
| ping timeout | 2.0s | 1s – 15s |
| max concurrent pings | 20 | 5 – 100 |
| ping attempts | 3 | 1 – 20 |

## Structure

```
cmd/gospeed/       main.go
internal/
├── endpoints/     API client, ping, server selection
├── speedtest/     download/upload measurement
└── tui/           terminal UI (Bubbletea)
```
cmd/gospeed/       main.go
internal/
├── endpoints/     API client, ping, server selection
├── speedtest/     download/upload measurement
└── tui/           terminal UI (Bubbletea)
```
