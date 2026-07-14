# Usage

The `spl` binary has six subcommands. `--debug` and `--trace` work on all of them.

## `spl run <file>`

Execute an SPL file through the full pipeline: lex → parse → analyze → run.

```sh
./bin/spl run testdata/runtime/hello.spl
```

## `spl tokens <file>`

Show what the lexer produces — one token per line.

```sh
./bin/spl tokens testdata/runtime/hello.spl
```

Output: `TYPE:LN:COL Lexeme`.

## `spl ast <file>`

Lex and parse, then dump the AST as indented JSON.

```sh
./bin/spl ast testdata/runtime/hello.spl
```

## `spl repl`

Interactive session. Type SPL line by line, submit with a blank line.

```sh
./bin/spl repl
```

See [REPL](../cli/repl.md) for the full workflow.

## `spl version`, `spl about`

Build info and credits.

## Pipeline

```text
Source (.spl)  →  Lexer  →  Token stream  →  Parser  →  AST
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
