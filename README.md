<picture>
  <source media="(prefers-color-scheme: dark)" srcset="docs/assets/images/spl_go_logo.png">
  <img alt="SPL Go Logo" src="docs/assets/images/spl_go_logo.png" width="200">
</picture>

# shakespeare-interpreter-go

[![CI](https://github.com/lorenzobandini/shakespeare-interpreter-go/actions/workflows/ci.yml/badge.svg)](https://github.com/lorenzobandini/shakespeare-interpreter-go/actions/workflows/ci.yml)
[![Go Version](https://img.shields.io/badge/Go-1.26.5-blue)](https://go.dev/)
[![License](https://img.shields.io/badge/License-GPL--3.0-green)](LICENSE)

A Go-based interpreter for the [Shakespeare Programming Language](http://shakespearelang.com/) (SPL) — a Turing-complete esolang where programs look like Elizabethan plays.

## Features

- **Full pipeline** — lexer, parser, semantic analyzer, runtime interpreter
- **Complete SPL support** — all canonical operations, control flow, I/O, stack
- **Cobra CLI** — `run`, `tokens`, `ast`, `repl`, `version`, `about` subcommands
- **REPL** — interactive mode with auto-declaration and stdin replay
- **Debugging** — `--debug` and `--trace` flags for pipeline visibility
- **Cross-platform** — Go binary for Windows, Linux, macOS + Docker image
- **WASM playground** — run SPL in your browser

## Quick start

```sh
task build
./bin/spl run testdata/runtime/hello.spl
```

On Windows the binary is `bin/spl.exe`; on Linux/macOS it is `bin/spl`.

## Documentation

Full docs at **[lorenzobandini.github.io/shakespeare-interpreter-go](https://lorenzobandini.github.io/shakespeare-interpreter-go)** — installation, usage, architecture, SPL spec, error reference, and the [WASM playground](https://lorenzobandini.github.io/shakespeare-interpreter-go/playground/editor.html).

## Development

```sh
task check          # fmt → lint → vuln → test (pre-commit gate)
task build:linux    # cross-compile for Linux amd64
task build:mac      # cross-compile for macOS amd64
task docker:build   # build Docker image
```

## License

[GPL-3.0](LICENSE)
