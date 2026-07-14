# Installation

## Prerequisites

- **Go 1.26.5** ([download](https://go.dev/dl/))
- **Task 3.x** — `go install github.com/go-task/task/v3/cmd/task@latest`

Optionally (for development): `golangci-lint v2`, `govulncheck`, `goimports`.

## Build from source

```sh
git clone https://github.com/lorenzobandini/shakespeare-interpreter-go.git
cd shakespeare-interpreter-go
task build
./bin/shpl.exe --help
```

`task build` produces a Windows `.exe`. For other platforms:

```sh
go build -o bin/shpl ./cmd/shpl/...
```

## Docker

```sh
docker build -t shpl .
docker run --rm shpl --help
```

## Verify

```sh
./bin/shpl.exe version
```
