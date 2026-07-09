# PROGRESS.md

## Phase 0 — Scaffolding & Tooling ✅

- [x] Go module initialized (`github.com/lorenzobandini/shakespeare-interpreter-go`), Go 1.26.5
- [x] Directory layout: `cmd/shpl/`, `internal/{lexer,logger}`, `testdata/{lexer,parser,interpreter}`, `docs/`
- [x] Taskfile: `fmt`, `lint`, `vuln`, `test`, `check`, `build`
- [x] golangci-lint v2 config (errcheck, govet, ineffassign, staticcheck, unused, bodyclose)
- [x] Lefthook pre-commit → `task check`
- [x] CI (GitHub Actions): checkout → setup-go → task → golangci-lint v7 → govulncheck → goimports → `task check`
- [x] Dockerfile: static linux binary, alpine base
- [x] Structured logger (`internal/logger`, slog TextHandler, LevelInfo/LevelDebug)
- [x] `docs/SPL_SPECIFICATION.md` — canonical SPL language reference
- [x] AGENTS.md with agent workflow + conventions
- [x] README.md with quick start + dev commands

## Phase 1 — Lexer

Tokenize `.shpl` source into a token stream.  
**Dependency**: `docs/SPL_SPECIFICATION.md`.

- [ ] Define token types: Title, Character, Act, Scene, Enter, Exit, Exeunt, Word, Number, RomanNumeral, Punctuation, Newline, EOF, etc.
- [ ] Implement lexer in `internal/lexer/`:
  - `ScanToken()`, `ScanTokens()`, `Token` struct with type/lexeme/line/col
  - Skip whitespace and comments (lines within brackets? none in SPL)
  - Handle multi-word nouns/adjectives
- [ ] Fixtures: `testdata/lexer/hello.shpl`, `testdata/lexer/truth-machine.shpl`
- [ ] Tests: table-driven, snapshot (input `.shpl` → expected token stream)
- [ ] Wire `--debug` flag to enable lexer trace logging
- [ ] CLI subcommand: `shpl tokens <file.shpl>`

### Decisions
- [x] **Step 1.0** — Doc reconciliation complete. Cross-checked `spl_specification.md` and `error_taxonomy.md` against canonical reference (Grokipedia/Esolang). Found and corrected 10 material errors: (1) Exeunt with names is valid; (2) no-copula assignment is valid; (3) similes evaluate to the value, not 0; (4) `a vile coward` = -2; (5) stack ops `Remember`/`Recall` exist; (6) unary `square root`/`factorial` exist; (7) 7 comparative forms → 6 relations; (8) possessive pronouns are ignored like articles; (9) L001 fires only for control chars; (10) L001 example `@` is dubious (descriptions are free text). Added S017/S018 to `error_taxonomy.md`. Appended "Canonical Grammar (vetted)" section to `spl_specification.md`. Reference: `docs/superpowers/plans/2026-07-09-lexer-parser.md`.
- Token types: **9 types** (EOF, NEWLINE, WORD, PERIOD, COMMA, COLON, BANG, QUESTION, LBRACKET, RBRACKET). Dropped `Number`/`RomanNumeral`/`Title`/`Character` — these are parser-side classifications, not lexical tokens. (D10)
- Lexer is **deliberately dumb**: emits generic `WORD` for all words (including `Act`, `Scene`, `Enter`, `You`, `Romeo`, nouns, adjectives). Parser classifies by value + declared character names. (D1)
- **Case-insensitive** classification; preserve original `Lexeme` on the token. (D2)
- **`[Exeunt A and B]`** is valid (bare = exit all; named = exit those). Fix S012. (D4)
- **L001 fires only for control characters** (0x00–0x08, 0x0B, 0x0C, 0x0E–0x1F, 0x7F). Other printable chars fold into `WORD`. (D5)

### Remaining
- Steps 1.1–1.8 (token types, scan loop, punctuation, WORD, EOF, fixtures, tests, CLI)

## Phase 2 — Parser

Parse token stream into AST.  
**Dependency**: Phase 1 (lexer).

- [ ] Define AST node types: Program, Title, CharacterDecl, Act, Scene, EnterStmt, ExitStmt, AssignStmt, SpeakStmt, OpenStmt, IfStmt, GotoStmt, Expr (constant, binary op, similes)
- [ ] Implement parser in `internal/parser/`:
  - Recursive descent or Pratt parser
  - Handle operation precedence (sum, difference, product, quotient)
  - Nested expressions
- [ ] Fixtures: `testdata/parser/hello.shpl`, `testdata/parser/truth-machine.shpl`
- [ ] Tests: table-driven, snapshot (input `.shpl` → expected AST JSON)
- [ ] CLI subcommand: `shpl ast <file.shpl>`

### Decisions
- *(none yet)*

### Remaining
- Full parser implementation + tests

## Phase 3 — Semantic Analysis

Validate AST for semantic correctness.  
**Dependency**: Phase 2 (parser).

- [ ] Validate: all referenced characters are declared
- [ ] Validate: max 2 characters on stage at any time
- [ ] Validate: characters are on stage before being spoken to
- [ ] Validate: scenes referenced in `goto` exist
- [ ] Validate: at least one act and one scene per act
- [ ] Validate: Roman numeral ordering (Act I, then II, etc.)
- [ ] Produce annotated AST or symbol table

### Decisions
- *(none yet)*

### Remaining
- Full semantic analysis + tests

## Phase 4 — Runtime / Evaluator

Execute the program.  
**Dependency**: Phase 3 (semantic analysis).

- [ ] Implement runtime in `internal/runtime/`:
  - Character state (variable store: name → int)
  - Stage manager (who's on stage, max 2)
  - Expression evaluator (nouns, adjectives, operations, similes)
  - I/O: stdin/stdout for Speak/Open/Listen commands
  - Control flow: `if so` jumps, `goto` scenes
  - Program counter: act/scene traversal
- [ ] Fixtures: `testdata/interpreter/hello.shpl` → stdout "Hello World!"
- [ ] Fixtures: `testdata/interpreter/truth-machine.shpl`
- [ ] Tests: golden file (input `.shpl` + stdin → expected stdout)
- [ ] CLI subcommand: `shpl run <file.shpl>`

### Decisions
- *(none yet)*

### Remaining
- Full runtime + tests

## Phase 5 — CLI Integration

Wire everything into the Cobra CLI.  
**Dependency**: Phase 1-4 complete.

- [ ] `shpl run <file>` — lex → parse → analyze → execute
- [ ] `shpl tokens <file>` — lex only, output token stream
- [ ] `shpl ast <file>` — lex + parse, output AST
- [ ] `shpl repl` — interactive REPL
- [ ] `shpl version` — version info
- [ ] `shpl about` — about / credits
- [ ] Global flags: `--debug` (enable debug logging), `--trace` (enable trace)

### Decisions
- *(none yet)*

### Remaining
- Cobra wiring + integration tests

---

## Future Phases (post-v1)

- Official docs/wiki (GitHub Pages)
- Language extensions / dialects
- LSP server for editor integration
- WASM build for browser playground
- Performance profiling and optimization
