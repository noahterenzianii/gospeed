# gospeed

A small Go CLI that measures download speed by automatically selecting the endpoint with the lowest latency.

## Requirements

- Go (version defined in `go.mod`)

## Quick start

```bash
make run
```

Useful commands:

```bash
make build   # build ./gospeed from ./cmd/gospeed
make run     # build and run
make lint    # go vet ./...
make clean   # remove binary
```

## Project layout

```text
cmd/gospeed/main.go        # CLI entrypoint
internal/app/run.go        # app orchestration
internal/endpoints/        # endpoint discovery, ping, and selection
internal/speedtest/        # download speed measurement
```
