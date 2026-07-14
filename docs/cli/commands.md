# CLI Commands

All commands are subcommands of `shpl`. `shpl --help` for a quick overview.

## Global flags

| Flag | Effect |
|------|--------|
| `--debug` | Enable debug-level logging (`slog.LevelDebug`) |
| `--trace` | Debug logging + pipeline stage markers on stderr |

## `shpl run <file>`

Run an SPL file through the full pipeline.

```sh
./bin/shpl.exe run testdata/runtime/hello.shpl
```

Exit code 0 on success, 1 on any error.

## `shpl tokens <file>`

Dump the token stream.

```sh
./bin/shpl.exe tokens testdata/runtime/hello.shpl
```

Output: `TYPE:LN:COL Lexeme` per line.

## `shpl ast <file>`

Dump the AST as indented JSON.

```sh
./bin/shpl.exe ast testdata/runtime/hello.shpl
```

## `shpl repl`

Interactive REPL. See [REPL](repl.md) for details.

## `shpl version`, `shpl about`

Build info, credits, and licensing.

## Error format

```
error[<CODE>]: <message>
  --> <file>:<line>:<col>
```
