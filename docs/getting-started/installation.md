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
./bin/spl --help
```

The build target for your current platform (`.exe` on Windows, no extension on Linux/macOS).
For cross-compilation:

```sh
task build:linux   # Linux amd64 → bin/spl
task build:mac     # macOS amd64  → bin/spl
task build:win     # Windows amd64 → bin/spl.exe
```

## Docker

```sh
docker build -t spl .
docker run --rm spl --help
```

## Verify

```sh
./bin/spl version
```
