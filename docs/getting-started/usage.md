# Usage

The CLI provides six subcommands. Global flags `--debug` and `--trace` are available on all commands.

## `shpl run <file>`

Execute a `.shpl` source file through the full pipeline: lex → parse → analyze → execute.

```sh
./bin/shpl.exe run examples/hello.shpl
```

With tracing:

```sh
./bin/shpl.exe --trace run examples/hello.shpl
```

## `shpl tokens <file>`

Lex the source file and print the token stream.

```sh
./bin/shpl.exe tokens examples/hello.shpl
```

## `shpl ast <file>`

Lex and parse the source file, then print the AST as nested JSON.

```sh
./bin/shpl.exe ast examples/hello.shpl
```

## `shpl repl`

Launch an interactive REPL session. Each submission is accumulated and replayed
through the full pipeline. Characters are auto-declared on first use.

```sh
./bin/shpl.exe repl
```

Type SPL dialogue line by line. See [REPL](../cli/repl.md) for details.

## `shpl version`

Print the build version (injected via `-ldflags` at build time).

## `shpl about`

Print credits and licensing information.

## Pipeline overview

```text
Source (.shpl)  →  Lexer  →  Token stream  →  Parser  →  AST
                                                    ↓
                                              Semantic Analyzer
                                                    ↓
                                              Runtime / Execute
```

Each stage validates its input and reports errors in a consistent format:

```
error[<CODE>]: <message>
  --> <file>:<line>:<col>
```

See [Error Taxonomy](../spl/errors.md) for all error codes.
