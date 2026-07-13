# Installation

## Prerequisites

- **Go 1.26.5** ([download](https://go.dev/dl/))
- **Task 3.x** — `go install github.com/go-task/task/v3/cmd/task@latest`
- **golangci-lint v2** — `go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest`
- **govulncheck** — `go install golang.org/x/vuln/cmd/govulncheck@latest`
- **goimports** — `go install golang.org/x/tools/cmd/goimports@latest`

## Build from source

```sh
git clone https://github.com/lorenzobandini/shakespeare-interpreter-go.git
cd shakespeare-interpreter-go
task build
./bin/shpl.exe --help
```

The `task build` command compiles a Windows `.exe`. For other platforms, use `go build` directly:

```sh
go build -o bin/shpl ./cmd/shpl/...
```

## Docker

A portable Linux binary is built via Docker:

```sh
docker build -t shpl .
docker run --rm shpl --help
```

## Verify installation

```sh
./bin/shpl.exe version
```

Should print a version string.
