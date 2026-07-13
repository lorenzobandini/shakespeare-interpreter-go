# shakespeare-interpreter-go

A Go-based interpreter for the [Shakespeare Programming Language](http://shakespearelang.com/) (SPL).

## Features

- **Full pipeline** — lexer, parser, semantic analyzer, runtime interpreter
- **Complete SPL support** — all canonical operations, control flow, I/O, stack
- **Cobra CLI** — `run`, `tokens`, `ast`, `repl`, `version`, `about` subcommands
- **Debugging** — `--debug` and `--trace` flags
- **Cross-platform** — Go binary + Docker image

## Quick start

```sh
task build
./bin/shpl.exe run examples/hello.shpl
```

## Documentation

Full docs at **[lorenzobandini.github.io/shakespeare-interpreter-go](https://lorenzobandini.github.io/shakespeare-interpreter-go)** — includes installation, usage, architecture, SPL spec, error reference, and an interactive [WASM playground](https://lorenzobandini.github.io/shakespeare-interpreter-go/playground/editor.html).

## Development

```sh
task check          # fmt → lint → vuln → test   (pre-commit gate)
```

## License

[GPL-3.0](LICENSE)
