# shakespeare-interpreter-go

A Go-based interpreter for the [Shakespeare Programming Language](http://shakespearelang.com/) (SPL).  
**Phase 0** — scaffolding and tooling setup.

## Prerequisites

- **Go 1.26.5**
- **Task 3.x** — task runner
- **golangci-lint v2** — static analysis
- **govulncheck** — vulnerability scanning
- **goimports** — import formatting

## Quick start

```sh
task build          # compile → bin/shpl.exe  (Windows)
./bin/shpl.exe      # or: go run cmd/shpl/main.go

# Execute an SPL source file:
./bin/shpl.exe run examples/hello.shpl
```

For a portable Linux binary:

```sh
docker build -t shpl . && docker run --rm shpl
```

## Development

```sh
task check          # fmt → lint → vuln → test   (the pre-commit gate)
task test           # go test -v -race -coverprofile=coverage.out ./...
task fmt            # gofmt + goimports
task lint           # golangci-lint run ./...
```

## License

[GPL-3.0](LICENSE)
