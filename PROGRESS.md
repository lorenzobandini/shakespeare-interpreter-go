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
- [x] `docs/spl/specification.md` — canonical SPL language reference
- [x] AGENTS.md with agent workflow + conventions
- [x] README.md with quick start + dev commands

## Phase 1 — Lexer ✅

Tokenize `.shpl` source into a token stream.  
**Dependency**: `docs/spl/specification.md`.

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

## Phase 4 — Runtime / Evaluator ✅

Execute the program.  
**Dependency**: Phase 3 (semantic analysis) ✅

- [x] 4.1: Package scaffold, `RuntimeError` type, `env` struct, `NewEnv` constructor, `TestMain`
- [x] 4.2: Constant expression evaluation (`ConstExpr`)
- [x] 4.3: `CharRefExpr`, `PronounExpr`, binary/unary operations (R001)
- [x] 4.4: Stage manager reuse — `Enter`/`Exit`/`Exeunt` + listener derivation
- [x] 4.5: Assignment and I/O statements (R002, R003)
- [x] 4.6: Stack operations (`Remember`/`Recall`)
- [x] 4.7: Questions, comparison flag, `If`/`Goto` branch resolution
- [x] 4.8: Program flattening + trampoline `Execute`
- [x] 4.9: Public `Execute` entry
- [x] 4.10: `shpl run` CLI subcommand
- [x] 4.11: Canonical fixtures + golden-file tests + `task check` gate

### Decisions
- **R-D1**: Flatten program to single `[]instr` + integer PC (not nested traversal)
- **R-D2**: I/O formatting: Speak writes 1 byte, OpenHeart writes `%d\n`
- **R-D3**: Reuse `semantic.Stage` for runtime stage state; ignore its returned `[]SemanticError`
- **R-D4**: Per-act scene label map (`map[actRoman]map[string]int`) rebuilt by flatten
- **R-D5**: Comparison flag is a single `bool`, not a stack
- **R-D6**: Integer overflow does not error (Go ints wrap)
- **R-D7**: `OpenMind` reads 1 byte (0–255); EOF → R003
- **R-D8**: `Listen` implements hand-rolled numeric parser skipping leading whitespace

### Fixtures
| Fixture | Expected |
|---------|----------|
| `hello.shpl` | byte(2) (STX) |
| `branch.shpl` | `"1\n"` (if-not → jump to Scene II, OpenHeart) |
| `stack.shpl` | `"1\n0\n0\n"` (push/pop/empty-pop) |
| `io-ascii.shpl` (stdin "X") | `"X"` (read byte, echo) |
| `truth-machine.shpl` (stdin "0\n") | `"0\n"` (greater-than comparison false → halt) |
| `truth-machine-1.shpl` (stdin "1\n") | `"1\n"` then R003 (greater-than comparison true → loop → EOF) |
| `divzero.shpl` | R001 error (divide by zero) |

### Post-hoc fix: truth-machine fixture
- **Bug**: `testdata/{semantic,interpreter}/truth-machine.shpl` used `as good as` (equal) instead of canonical `better than` (greater). Introduced during Phase 2/3 fixture creation; invisible to semantic checks (only validates syntax, not runtime behavior).
- **Fix**: Replaced `"Am I as good as you?"` with `"Am I better than you?"`, restructured to 2 scenes matching the original 2001 Hasselström & Åslund spec. Golden files regenerated.
- **Impact**: truth-machine now correctly loops on input "1" and halts on input "0".

## Phase 5 — CLI Integration ✅

Wire everything into the Cobra CLI.  
**Dependency**: Phase 1-4 complete.

- [x] `shpl run <file>` — lex → parse → analyze → execute
- [x] `shpl tokens <file>` — lex only, output token stream
- [x] `shpl ast <file>` — lex + parse, output AST
- [x] `shpl repl` — interactive REPL (Replay-based Accumulating Buffer Model)
- [x] `shpl version` — version info (ldflags-injected, also via `--version`)
- [x] `shpl about` — about / credits
- [x] Global flags: `--debug` (slog LevelDebug), `--trace` (pipeline stage markers on stderr + debug logging)
- [x] Integration tests (17 tests): tokens, ast, run, version, about, trace, repl (basic, auto-declaration, rollback, trace-integration), help output

### Decisions
- **D1**: Replay-based Accumulating Buffer Model chosen for REPL. No Phase 2/4 modifications — the full pipeline (lex → parse → analyze → execute) reruns from scratch on each submission. This avoids refactoring `runtime.env` (unexported) or `parser.*` (fragment methods unexported).
- **D2**: `--trace` implies debug logging + pipeline stage markers on stderr (`--- TOKENS ---`, `--- AST ---`, `--- SEMANTIC ---`, `--- EXECUTE ---`). `--debug` enables debug logging only. Stage markers use `cmd.ErrOrStderr()` for testability.
- **D3**: Buffer + input-cursor checkpoint/rollback on replay failure. Before each replay, the buffer length and recorded-input cursor are checkpointed. On failure, both are restored so the REPL remains usable after errors.
- **D4**: Skeleton (`The REPL Session.\n\nAct I: The REPL Session.\nScene I: The REPL Session.\n\n`) is always prepended on the first submission, guaranteeing a valid SPL structure for auto-declared characters.
- **D5**: Auto-declaration scans for character names in `[Enter]`, `[Exit]`, `[Exeunt]`, and dialogue speaker prefixes. Declarations are inserted before Act I. Characters explicitly declared by the user in the input text are detected via the `declared` map and not duplicated.
- **D6**: Output slicing via byte-length prefix (`newOutput[lastOutputLen:]`), not line-by-line diff. Because stdin replay is deterministic and the buffer only grows, earlier outputs are always byte-prefixes of later outputs.
- **D7**: `os.Exit(1)` eliminated from all command handlers — errors flow through Cobra's `RunE` return to `main()`'s error handler. Testability: all tests use `rootCmd.Execute() error`.

### Trade-offs
- **O(n²) replay complexity**: Each submission re-runs the full pipeline on the accumulated buffer. At human-typing scale (dozens of lines), this overhead is negligible.
- **Infinite loops**: Programs with infinite loops (e.g., unterminated truth-machine) are incompatible with the replay model and are considered out of scope for the REPL.
- **No stdin prompt customization**: The `input>` prompt is used when the program's `Listen`/`OpenMind` reads from stdin, with no configuration option.

---

## Phase 6 — Documentation Platform ✅

Stand up official MkDocs + Material documentation site deployed to GitHub Pages.

- [x] `mkdocs.yml` with Material slate theme, nav tree, GitHub integration
- [x] Migrate `SPL_SPECIFICATION.md` and `ERROR_TAXONOMY.md` into MkDocs tree
- [x] Content pages: Home, Getting Started, Architecture, CLI, Contributing, About
- [x] CI/CD workflow: build + deploy on push to `main`
- [x] Local verification: `mkdocs build --strict` passes

---

## Future Phases (post-v1)

- Language extensions / dialects
- LSP server for editor integration
- WASM build for browser playground
- Performance profiling and optimization
