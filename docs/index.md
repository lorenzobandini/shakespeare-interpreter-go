# Shakespeare Interpreter

An interpreter for the [Shakespeare Programming Language](http://shakespearelang.com/) (SPL),
written in Go. SPL is a Turing-complete esolang where programs look like Elizabethan plays
— characters are variables, dialogue is arithmetic, stage directions control flow.

## Quick start

```sh
task build
./bin/spl run testdata/runtime/hello.spl
```

For detailed instructions see [Installation](getting-started/installation.md) and [Usage](getting-started/usage.md).

## Project status

| Phase | Status |
|-------|--------|
| Scaffolding & tooling (Taskfile, lint, CI, Docker) | ✅ |
| Lexer (L001–L002) | ✅ |
| Parser (S001–S018, AST, dictionary) | ✅ |
| Semantic analysis (M001–M008, symbol table, stage manager) | ✅ |
| Runtime / Evaluator (R001–R004, trampoline execution) | ✅ |
| CLI integration (Cobra, 6 subcommands, REPL, debug/trace) | ✅ |
| Documentation | ✅ this site |
