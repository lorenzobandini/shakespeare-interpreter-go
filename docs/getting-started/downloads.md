# Downloads

Pre-built binaries are published on the [GitHub Releases](https://github.com/lorenzobandini/shakespeare-interpreter-go/releases) page. Each release includes Windows, Linux, and macOS builds.

## Docker

```sh
docker pull ghcr.io/lorenzobandini/shakespeare-interpreter-go:latest
docker run --rm shpl --help
```

To run a program:

```sh
docker run --rm -v $(pwd)/program.shpl:/program.shpl shpl run /program.shpl
```

## Build from source

See [Installation](installation.md) for building from source with Go 1.26.5.
