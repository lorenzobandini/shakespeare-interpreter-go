# Shakespeare Interpreter

![Shakespeare Interpreter](assets/images/spl_go_logo.png)

A **Go-based interpreter** for the [Shakespeare Programming Language](http://shakespearelang.com/) (SPL),
a Turing-complete esoteric language where programs read like Elizabethan plays.

## Features

- **Full pipeline** — lexer, parser, semantic analyzer, runtime interpreter
- **Complete SPL support** — all canonical operations, control flow, I/O, stack
- **Cobra CLI** — `run`, `tokens`, `ast`, `repl`, `version`, `about` subcommands
- **Debugging** — `--debug` and `--trace` flags for development
- **Portable** — cross-platform Go binary, Docker image available

## Quick start

```sh
task build
./bin/shpl.exe run examples/hello.shpl
```

For detailed instructions see [Installation](getting-started/installation.md) and [Usage](getting-started/usage.md).

## Project status

| Phase | Status |
|-------|--------|
| Scaffolding | ✅ |
| Lexer | ✅ |
| Parser | ✅ |
| Semantic Analysis | ✅ |
| Runtime / Evaluator | ✅ |
| CLI Integration | ✅ |
| **Documentation** | ✅ **this site** |
