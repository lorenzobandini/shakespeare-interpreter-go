# Usage

The `shpl` binary has six subcommands. `--debug` and `--trace` work on all of them.

## `shpl run <file>`

Execute an SPL file through the full pipeline: lex → parse → analyze → run.

```sh
./bin/shpl.exe run testdata/runtime/hello.shpl
```

## `shpl tokens <file>`

Show what the lexer produces — one token per line.

```sh
./bin/shpl.exe tokens testdata/runtime/hello.shpl
```

Output: `TYPE:LN:COL Lexeme`.

## `shpl ast <file>`

Lex and parse, then dump the AST as indented JSON.

```sh
./bin/shpl.exe ast testdata/runtime/hello.shpl
```

## `shpl repl`

Interactive session. Type SPL line by line, submit with a blank line.

```sh
./bin/shpl.exe repl
```

See [REPL](../cli/repl.md) for the full workflow.

## `shpl version`, `shpl about`

Build info and credits.

## Pipeline

```text
Source (.shpl)  →  Lexer  →  Token stream  →  Parser  →  AST
                                                    ↓
                                              Semantic Analyzer
                                                    ↓
                                              Runtime / Execute
```

Errors use consistent formatting:

```
error[<CODE>]: <message>
  --> <file>:<line>:<col>
```

See [Error Taxonomy](../spl/errors.md) for all codes.
