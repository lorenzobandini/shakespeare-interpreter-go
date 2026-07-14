# Architecture Overview

The interpreter is a four-stage pipeline: **Lexer → Parser → Semantic Analyzer → Runtime**.

```mermaid
graph LR
  Source["Source (.spl)"]
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
  Lexer -.-> Tokens
  Parser -.-> AST
  Semantic -.-> Validated
  Runtime -.-> Output
```

## Package dependency graph

```mermaid
graph TD
  CLI["cmd/spl/
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
```

Dependencies flow downward. No package imports from a higher layer.

## Key decisions

- **Recursive descent parser** — one method per grammar production, no parser generator.
- **Type-switch dispatch** in the semantic analyzer and runtime — no Visitor pattern.
- **Trampoline execution** — the program is flattened to `[]instr` with an integer PC.
- **Replay-based REPL** — reruns the full pipeline on each submission.
