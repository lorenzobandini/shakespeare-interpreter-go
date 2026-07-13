# Architecture Overview

The interpreter follows a four-stage pipeline: **Lexer → Parser → Semantic Analyzer → Runtime**.

```text
Source (.shpl)
    │
    ▼
┌─────────────┐
│   Lexer     │  → Token stream ([]Token)
│             │    Errors: L001, L002
└─────────────┘
    │
    ▼
┌─────────────┐
│   Parser    │  → Abstract Syntax Tree (*ast.Program)
│             │    Errors: S001–S018
└─────────────┘
    │
    ▼
┌─────────────┐
│  Semantic   │  → Validated AST + SymbolTable + Stage state
│  Analyzer   │    Errors: M001–M008
└─────────────┘
    │
    ▼
┌─────────────┐
│  Runtime    │  → Program output (stdout) + exit code
│  (Evaluate) │    Errors: R001–R004
└─────────────┘
```

## Package dependency graph

```
cmd/shpl/          ← Cobra CLI entry points
    │
    ▼
internal/lexer/    ← Token scanner
internal/logger/   ← Structured slog handler
    │
    ▼
internal/parser/   ← Recursive descent parser, AST definitions, dictionary
    │
    ▼
internal/semantic/ ← Symbol table, stage manager, validation
    │
    ▼
internal/runtime/  ← Environment, instruction trampoline, I/O
```

Dependencies flow downward only. No package imports from a higher layer.

## Key design decisions

- **Recursive descent parser** — one method per grammar production, no parser generator.
- **Type-switch dispatch** in semantic analyzer and runtime (no Visitor pattern — YAGNI).
- **Trampoline execution** — program flattened to `[]instr`, integer PC, no nested traversal.
- **Replay-based REPL** — full pipeline rerun on each submission (O(n²) at human scale).
