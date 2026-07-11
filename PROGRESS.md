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
- [x] Expanded table-driven unit tests for each parse method
- [x] Golden JSON snapshot tests for all 6 parser fixtures
- [x] Parser fixtures: arithmetic.shpl, stack.shpl, conditionals.shpl + error fixtures
- [x] Error fixture tests for S002, S005, S008, S013, S017

### Decisions
- *(none yet)*

### Phase 2 bugfix: `enterSeen` act-boundary reset
- **Bug:** `parseActs()` reset `enterSeen` to `false` at every act boundary, causing S013 (`errMissingStage`) on valid programs where stage state carries across acts without a new `[Enter]` (e.g., `primes.spl` from the official zmbc/shakespearelang reference).
- **Fix:** Hoisted `enterSeen` to the `parseActs()` scope (was per-act), threaded it through both the period-branch and colon-branch of act parsing.
- **Verification:** `primes.spl` adapted excerpt (`testdata/parser/cross-act-persistence.shpl`) now parses without S013. Existing tests unaffected.
- **Reference:** `docs/superpowers/plans/2026-07-10-phase3-semantic-analysis.md` Step 3.0.

### Remaining
- Full parser implementation + tests

## Phase 3 — Semantic Analysis ✅

Validate AST for semantic correctness (M001–M008).  
**Dependency**: Phase 2 (parser) ✅

- [x] `internal/semantic/` package: `errors.go`, `symbol_table.go`, `stage.go`, `analyzer.go`
- [x] SemanticError struct with M001–M008 constructors, matching taxonomy format
- [x] SymbolTable (case-insensitive, preserves original-case Name), ActRegistry, SceneRegistry (per-act scoping)
- [x] Stage manager: Enter/Exit/Exeunt with max-2 enforcement, overflow, duplicate detection
- [x] M001: Undeclared character in stage directions or dialogue speaker
- [x] M002: Too many characters in single Enter (defensive — parser caps at 2)
- [x] M003: Stage overflow (2 already on stage, trying to add more)
- [x] M004: Speaker not on stage (role="speaker"), listener not reachable
- [x] M005: Exit/Exeunt targeting character not on stage
- [x] M006: Goto/If target scene/act does not exist (per-act scene scoping)
- [x] M007: Self-reference Enter (same name twice in one Enter, or re-enter already-on-stage)
- [x] M008: Act with zero scenes (defensive — parser enforces S007)
- [x] D1: Stage does NOT clear at act boundaries (primes.spl evidence)
- [x] D2: Off-stage CharRefExpr in expressions is ALLOWED (no M004)
- [x] Two-level type-switch dispatch (level 1: Scene statements; level 2: Dialogue statements with dialogCtx)
- [x] Collect-all errors (linter-style, D4), returns `Result` with `OK()` method
- [x] 16 fixtures under `testdata/semantic/` (valid + each M-code + multiple errors)
- [x] 97.9% statement coverage
- [x] `task check` green

### Decisions
- **D1**: Stage persists across act boundaries, mutated only by Enter/Exit/Exeunt (confirmed by `primes.spl`).
- **D2**: CharRefExpr value reads allowed even if character is off-stage (confirmed by canonical Hello World, Act I Scenes II/III).
- **D3**: M008 is defensive (S007 already enforced by parser).
- **D4**: All errors collected in one pass; no fail-fast.
- **D5**: Type-switch dispatch (no formal Visitor interface — YAGNI).
- **D6**: Re-entering an already-on-stage character → M007 (`cannot enter 'X': already on stage`).
- **D7**: No annotated AST produced — `Result` carries symbols + registries + errors only (runtime re-derives listener from stage).

### Fixtures
| Fixture | Expected |
|---------|----------|
| `hello.shpl` | zero errors |
| `truth-machine.shpl` | zero errors |
| `minimal-valid.shpl` | zero errors |
| `self-talk.shpl` | zero errors |
| `off-stage-value-read.shpl` | zero errors (D2) |
| `valid-operations.shpl` | zero errors |
| `primes-persistence.shpl` | zero errors (D1) |
| `m001-undeclared-enter.shpl` | M001 |
| `m001-undeclared-speaker.shpl` | M001 |
| `m003-stage-overflow.shpl` | M003 |
| `m004-speaker-not-on-stage.shpl` | M004 |
| `m004-empty-stage-cross-act.shpl` | M004 |
| `m005-exit-not-on-stage.shpl` | M005 |
| `m006-undefined-scene.shpl` | M006 |
| `m006-undefined-act.shpl` | M006 |
| `m007-self-enter.shpl` | M007 |
| `multiple-errors.shpl` | M001 + M003 + M006 |
| `goto-cross-act-scene.shpl` | M006 (scene III in Act II) |

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
