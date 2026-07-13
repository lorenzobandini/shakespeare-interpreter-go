# CLI Commands

All commands are subcommands of `shpl`. Use `shpl --help` for an overview.

## Global flags

| Flag | Effect |
|------|--------|
| `--debug` | Enable debug-level logging (`slog.LevelDebug`) |
| `--trace` | Enable debug logging + pipeline stage markers on stderr |

## `shpl run <file>`

Execute a `.shpl` file through the full pipeline.

```sh
./bin/shpl.exe run examples/hello.shpl
```

Exit codes: 0 on success, 1 on any error (lexical, syntax, semantic, or runtime).

## `shpl tokens <file>`

Print the token stream from lexing.

```sh
./bin/shpl.exe tokens examples/hello.shpl
```

Output format: `TYPE:LN:COL Lexeme` on each line.

## `shpl ast <file>`

Print the AST as indented JSON.

```sh
./bin/shpl.exe ast examples/hello.shpl
```

## `shpl repl`

Launch an interactive REPL. See [REPL](repl.md) for details.

## `shpl version`

Print the build version.

## `shpl about`

Print credits and license information.

## Error format

All errors use this format:

```
error[<CODE>]: <message>
  --> <file>:<line>:<col>
```
