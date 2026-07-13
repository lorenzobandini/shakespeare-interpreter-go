# Architecture Overview

The interpreter follows a four-stage pipeline: **Lexer → Parser → Semantic Analyzer → Runtime**.

```mermaid
graph LR
  Source["Source (.shpl)"]
  Lexer
  Parser
  Semantic["Semantic Analyzer"]
  Runtime
  Tokens["Token stream ([]Token)
          L001, L002"]
  AST["Abstract Syntax Tree
       (*ast.Program)
       S001–S018"]
  Validated["Validated AST
             SymbolTable
             Stage state
             M001–M008"]
  Output["Program output
          (stdout)
          R001–R004"]

  Source --> Lexer
  Lexer -->|Tokens| Parser
  Parser -->|AST| Semantic
  Semantic -->|Validated| Runtime
  Lexon -.-> Tokens
  Parser -.-> AST
  Semantic -.-> Validated
  Runtime -.-> Output

  linkStyle 0,1,2,3 stroke:#89b4fa,stroke-width:2px
  linkStyle 4,5,6,7 stroke:#6c7086,stroke-width:1px,stroke-dasharray:3
```

## Package dependency graph

```mermaid
graph TD
  CLI["cmd/shpl/
      Cobra CLI entry points"]
  LEX["internal/lexer/
      Token scanner"]
  LOG["internal/logger/
      Structured slog handler"]
  PAR["internal/parser/
      Recursive descent parser
      AST definitions
      Dictionary"]
  SEM["internal/semantic/
      Symbol table
      Stage manager
      Validation"]
  RUNT["internal/runtime/
      Environment
      Instruction trampoline
      I/O"]

  CLI --> LEX
  CLI --> LOG
  LEX --> PAR
  PAR --> SEM
  SEM --> RUNT

  linkStyle 0,1,2,3,4 stroke:#89b4fa,stroke-width:2px
```

Dependencies flow downward only. No package imports from a higher layer.

## Key design decisions

- **Recursive descent parser** — one method per grammar production, no parser generator.
- **Type-switch dispatch** in semantic analyzer and runtime (no Visitor pattern — YAGNI).
- **Trampoline execution** — program flattened to `[]instr`, integer PC, no nested traversal.
- **Replay-based REPL** — full pipeline rerun on each submission (O(n²) at human scale).
