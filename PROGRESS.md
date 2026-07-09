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

## Phase 1 — Lexer ✅

Tokenize `.shpl` source into a token stream.  
**Dependency**: `docs/SPL_SPECIFICATION.md`.

- [x] Define token types (9 types: EOF, NEWLINE, WORD, PERIOD, COMMA, COLON, BANG, QUESTION, LBRACKET, RBRACKET)
- [x] Implement lexer in `internal/lexer/`:
  - `ScanTokens()`, `ScanToken()`, `Token` struct with type/lexeme/line/col
  - Skip whitespace, emit NEWLINE tokens
  - L001 (control chars), L002 (unterminated bracket)
- [x] Fixtures: `testdata/lexer/{hello,truth-machine,minimal,bad-char}.shpl`
- [x] Tests: table-driven + golden snapshot (3 golden files generated)
- [x] CLI subcommand: `shpl tokens <file.shpl>`

### Decisions
- [x] **Step 1.0** — Doc reconciliation complete. Cross-checked `spl_specification.md` and `error_taxonomy.md` against canonical reference (Grokipedia/Esolang). Found and corrected 10 material errors: (1) Exeunt with names is valid; (2) no-copula assignment is valid; (3) similes evaluate to the value, not 0; (4) `a vile coward` = -2; (5) stack ops `Remember`/`Recall` exist; (6) unary `square root`/`factorial` exist; (7) 7 comparative forms → 6 relations; (8) possessive pronouns are ignored like articles; (9) L001 fires only for control chars; (10) L001 example `@` is dubious (descriptions are free text). Added S017/S018 to `error_taxonomy.md`. Appended "Canonical Grammar (vetted)" section to `spl_specification.md`. Reference: `docs/superpowers/plans/2026-07-09-lexer-parser.md`.
- Token types: **9 types** (EOF, NEWLINE, WORD, PERIOD, COMMA, COLON, BANG, QUESTION, LBRACKET, RBRACKET). Dropped `Number`/`RomanNumeral`/`Title`/`Character` — these are parser-side classifications, not lexical tokens. (D10)
- Lexer is **deliberately dumb**: emits generic `WORD` for all words (including `Act`, `Scene`, `Enter`, `You`, `Romeo`, nouns, adjectives). Parser classifies by value + declared character names. (D1)
- **Case-insensitive** classification; preserve original `Lexeme` on the token. (D2)
- **`[Exeunt A and B]`** is valid (bare = exit all; named = exit those). Fix S012. (D4)
- **L001 fires only for control characters** (0x00–0x08, 0x0B, 0x0C, 0x0E–0x1F, 0x7F). Other printable chars fold into `WORD`. (D5)

## Phase 2 — Parser (partial ✅)

Parse token stream into AST.  
**Dependency**: Phase 1 (lexer) ✅

- [x] Define AST node types: Program, Title, CharacterDecl, Act, Scene, EnterStmt, ExitStmt, ExeuntStmt, Dialogue, AssignStmt, SpeakStmt, OpenHeartStmt, OpenMindStmt, ListenStmt, QuestionStmt, IfStmt, GotoStmt, RememberStmt, RecallStmt, ConstExpr, CharRefExpr, PronounExpr, BinaryOpExpr, UnaryOpExpr, Comparative
- [x] Parser errors: `ParseError` with S001–S018, `Warning` for S003
- [x] Dictionary: curated ~80 words (nouns positive/negative, adjectives, comparatives, pronouns, articles, Shakespeare characters)
- [x] Parser core: cursor helpers, New(), Parse()
- [x] Parse title + character declarations (S001–S003)
- [x] Parse acts + scenes + Roman numerals (S004–S009)
- [x] Parse stage directions Enter/Exit/Exeunt (S010–S012)
- [x] Parse dialogue (speaker + statement list) (S013/S014)
- [x] Parse expressions: constants, CharRef, Pronoun (S015)
- [x] Parse binary operations: sum, difference, product, quotient, remainder
- [x] Parse unary operations: square, cube, square_root, factorial, twice
- [x] Parse assignment statements (copula + no-copula)
- [x] Parse I/O statements (Speak, Open heart, Open mind, Listen)
- [x] Parse questions, if-statements, goto (S016, S017)
- [x] Parse stack operations Remember/Recall (S018)
- [x] CLI subcommand: `shpl ast <file.shpl>`
- [x] Canonical Hello World fixture parses successfully
- [x] Canonical Truth Machine fixture parses successfully

### Key design notes (Phase 2)
- Recursive descent parser with case-insensitive keyword matching.
- Stage state (`enterSeen`) persists across scenes within an act (SPL semantics).
- "the" is disambiguated as operator prefix vs. article by peeking past newlines at the next word.
- Newlines are skipped within expressions (speech spans multiple lines) but not between sentences.
- The simile `as <adj> as <value>` evaluates to the inner value (not 0, correcting the local spec).
- Exeunt accepts optional names: bare = exit all; named = exit those (per canonical spec).
- No-copula assignment `You <constant>!` is supported (per canonical Hello World).

### Remaining (Phase 2)
- [ ] Expanded table-driven unit tests for each parse method (Steps 2.5–2.15 individual tests)
- [ ] Golden JSON snapshot tests for hello.shpl and truth-machine.shpl ASTs
- [ ] Parser fixtures beyond lexer-shared ones (arithmetic.shpl, stack.shpl, conditionals.shpl)

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
