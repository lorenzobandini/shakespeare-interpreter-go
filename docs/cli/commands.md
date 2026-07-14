# CLI Commands

All commands are subcommands of `spl`. `spl --help` for a quick overview.

## Global flags

| Flag | Effect |
|------|--------|
| `--debug` | Enable debug-level logging (`slog.LevelDebug`) |
| `--trace` | Debug logging + pipeline stage markers on stderr |

## `spl run <file>`

Run an SPL file through the full pipeline.

```sh
./bin/spl run testdata/runtime/hello.spl
```

Exit code 0 on success, 1 on any error.

## `spl tokens <file>`

Dump the token stream.

```sh
./bin/spl tokens testdata/runtime/hello.spl
```

Output: `TYPE:LN:COL Lexeme` per line.

## `spl ast <file>`

Dump the AST as indented JSON.

```sh
./bin/spl ast testdata/runtime/hello.spl
```

## `spl repl`

Interactive REPL. See [REPL](repl.md) for details.

## `spl version`, `spl about`

Build info, credits, and licensing.

## Error format

```
error[<CODE>]: <message>
  --> <file>:<line>:<col>
```
