# Lexer & Parser Implementation Plan (Phases 1 & 2)

> **For agentic workers:** REQUIRED SUB-SKILL: Use `superpowers:subagent-driven-development` (recommended) or `superpowers:executing-plans` to implement this plan task-by-task. Steps use `- [ ]` checkbox syntax for tracking.

**Goal:** Tokenize `.shpl` source into a token stream (Phase 1) and parse it into a typed AST covering the full canonical SPL grammar (Phase 2).

**Architecture:** A deliberately dumb, stateless lexer emits `WORD` + structural-punctuation tokens. A recursive-descent parser owns a curated noun/adjective/comparative dictionary plus the declared character-name set, and classifies words into typed AST nodes. Unidirectional pipeline: `lexer → parser → ast`, no back-edges. All internal state transitions are traced via `slog.Debug` through the existing `internal/logger`.

**Tech Stack:** Go 1.26.5, stdlib only (`log/slog`, `strings`, `unicode`, `encoding/json`, `testing`, `os`, `fmt`). Cobra (already in `go.mod`) wired minimally for `tokens`/`ast` subcommands. No new dependencies.

---

## Global Constraints

- **Go 1.26.5**; module `github.com/lorenzobandini/shakespeare-interpreter-go`.
- **Gate after every step:** `task check` (fmt → lint → vuln → test). Lefthook pre-commit and CI both run exactly this. `coverage.out` is gitignored — never commit it.
- **golangci-lint v2** enables ONLY `errcheck, govet, ineffassign, staticcheck, unused, bodyclose` (`default: none`). Code must be errcheck-clean with zero unused symbols. Every error return path must be handled or explicitly discarded with a comment.
- **Tests:** `go test -race` always (the suite uses `-race`). Single-test reruns: `go test -race -run TestName ./internal/<pkg>/...`. Fixtures live in `testdata/lexer/`, `testdata/parser/`.
- **Logger:** `logger.Init(logger.LevelDebug)` must be called in test `TestMain` (or a test helper) before any `slog.Debug` output is expected. Emit `slog.Debug` at each token emission (lexer) and each AST node construction (parser) with `type`, `value`/`lexeme`, `line`, `col`.
- **Conventional Commits** per step. Small, frequent, atomic commits. Format: `feat(lexer): ...`, `feat(parser): ...`, `test(lexer): ...`, `docs: ...`.
- **Update `PROGRESS.md`** after each step (check off items, note decisions). Update `docs/ERROR_TAXONOMY.md` only in Step 1.0 (reconciliation) and Step 2.2 (new S-codes).
- **No comments in code** unless explicitly requested (AGENTS.md convention). `ponytail:` comments are the sole exception for deliberate simplifications.

---

## Canonical Reference Authority

The local `docs/SPL_SPECIFICATION.md` and `docs/ERROR_TAXONOMY.md` were LLM-generated in a prior session and contain **known errors** (documented in Step 1.0). The **canonical** SPL reference is:

- **Grokipedia** (community-vetted): https://grokipedia.com/page/Shakespeare_Programming_Language
- **Esolangs wiki**: https://esolangs.org/wiki/Shakespeare
- **spl2c manual** (original 2001 spec by Åslund & Hasselström)

Where the local docs conflict with the canonical reference, **the canonical reference wins**. Step 1.0 reconciles the local docs; all subsequent steps implement the reconciled (canonical) grammar.

### Key canonical facts confirmed (from Grokipedia cross-check)

These differ from or extend the local `docs/SPL_SPECIFICATION.md`:

1. **`[Exeunt]` may optionally name characters.** `[Exeunt Ophelia and Hamlet]` is canonical (appears in the "Infamous Hello World Program"). Bare `[Exeunt]` exits all on stage. `error_taxonomy.md` S012 ("Exeunt must stand alone") is **wrong** and must be fixed.
2. **Assignment without a copula is valid.** `You lying stupid fatherless big smelly half-witted coward!` is an assignment (no "are"/"art"). The local spec only documents `You are <expr>.` — the canonical Hello World fixture would not parse without the no-copula form.
3. **Similes evaluate to the referenced value, NOT 0.** `You are as stupid as the difference between...` assigns the computed value. The local spec's "evaluates to 0 but maintains grammatical flow" is **wrong** and would break every Hello World assignment.
4. **`a vile coward` = -2, not -4.** Each adjective doubles the magnitude once. `vile` (1 adjective) × `coward` (-1) = -2. The local spec's example "vile = 2 adj" is arithmetically wrong (vile is 1 adjective).
5. **Stack operations exist.** `Remember <expr>.` pushes a value onto the listener's stack; `Recall <ignored text>.` pops the listener's stack into the speaker. Entirely missing from the local spec.
6. **Additional unary operations.** `the square root of`, `the factorial of` — missing from the local spec (which only lists square, cube, twice).
7. **Seven comparative forms → six relations.** Positive/negative adjective polarity + `not` negation + `as...as` vs `...than` produce 7 syntactic forms mapping to 6 comparison relations. The local spec lists only 3.
8. **Possessive and reflexive pronouns.** `my/thy/your/his/her/mine/thine` (possessive, ignored like articles in constants), `me/myself` (speaker), `thyself/yourself` (listener). The local spec only mentions `yourself/thyself`.
9. **L001 `@` example is dubious.** Character descriptions are free text (`Andersen Insulting A/S` contains a slash). A strict denylist lexer would reject valid programs. L001 should fire only for control characters; other printable chars fold into `WORD` tokens and surface as parser (S-code) errors.
10. **"Punctuation appended to output"** (Grokipedia prose) is almost certainly wrong — it would corrupt `Hello World!` into `H!e!l!l!o!...`. Terminators (`. ! ?`) are treated as purely syntactic; runtime output semantics are a Phase 4 question, flagged but not resolved here.

---

## Design Decisions (locked from brainstorming)

| # | Decision | Rationale |
|---|----------|-----------|
| D1 | **Lexer dumb, parser classifies.** Lexer emits generic `WORD` for all words (incl. `Act`/`Scene`/`Enter`/`You`/`Romeo`/nouns). Parser owns the noun/adjective dictionary + declared character names and builds typed AST nodes. | Clean unidirectional dependencies; matches reference SPL implementations; keeps the lexer stateless and trivially testable. |
| D2 | **Case-insensitive classification.** Normalize to lowercase for lookup; preserve original `Lexeme` on the token for output/debug. | Matches SPL's flexible grammar (`Speak YOUR mind!` vs `Speak your mind!`). |
| D3 | **Curated dictionary (~80 words).** Positive/negative/neutral nouns + positive/negative adjectives + comparative adjectives, drawn from spec examples + common SPL vocab. Unknown noun → neutral (+1); unknown adjective → counted as a doubler; unknown comparative → positive. | Balance of accuracy and maintenance. Spec says unknown nouns default to 1. |
| D4 | **Exeunt supports optional names.** Bare `[Exeunt]` = exit all on stage; `[Exeunt A and B]` = exit named. Fix S012. | Canonical spec + canonical Hello World fixture require it. |
| D5 | **L001 fires only for control characters** (0x00–0x08, 0x0B, 0x0C, 0x0E–0x1F, 0x7F). All other non-structural printable chars fold into `WORD` tokens. | Taxonomy's own note: "SPL has very few lexical errors." Avoids rejecting valid free-text descriptions. |
| D6 | **Full canonical grammar implemented in Phase 2** (incl. stack ops `Remember`/`Recall`, unary `square root`/`factorial`, 7 comparative forms, no-copula assignment, possessive pronouns). Evaluation of these is Phase 4; Phase 2 only builds the AST. | "No omissions" requirement. The canonical Hello World fixture exercises no-copula assignment, similes, and `[Exeunt A and B]` — all must parse. |
| D7 | **Terminators** (`. ! ?`) recorded on each statement node; runtime output semantics deferred to Phase 4. | Grokipedia's "append punctuation to output" is almost certainly a documentation error. |
| D8 | **Character names are single `WORD` tokens** (Romeo, Juliet, Hamlet, Ophelia). Multi-word names ("Sir Andrew") not supported in v1. | All canonical fixtures use single-word names. |
| D9 | **`spl_specification.md` gets an appended "Canonical Grammar (vetted)" section** rather than a rewrite. | Preserves original prose as a record; adds vetted grammar as the parser's source of truth. |
| D10 | **No `Title`/`Character`/`Number`/`RomanNumeral` token types.** PROGRESS.md listed these; they are parser-side classifications, not lexical tokens. Dropped. | The lexer cannot know a word is a character name before declarations are parsed. |

---

## Open Assumptions (confirm at review before execution)

- **A1:** L001 reduced to control-chars-only (drops the `@` example from `error_taxonomy.md`). Approve?
- **A2:** Include `Remember`/`Recall` stack ops in Phase 2 (parse-only) for full canonical coverage, even though the two core fixtures don't exercise them. Approve, or defer to a later phase?
- **A3:** Character names are single `WORD` (Romeo, Juliet). Multi-word names not supported in v1. Approve?
- **A4:** `spl_specification.md` gets an appended "Canonical Grammar (vetted)" section rather than a rewrite. Approve?
- **A5:** Include minimal Cobra CLI wiring (`tokens`, `ast` subcommands) at the end of Phases 1 and 2, deferring the full CLI (`run`, `repl`, `version`, `about`, global flags) to Phase 5. Approve?

---

## File Structure

### Phase 1 — `internal/lexer/`

| File | Responsibility |
|------|---------------|
| `token.go` | `TokenType`, `Token` struct, `String()` method. |
| `lexer.go` | `Lexer` struct, `New()`, `ScanTokens()`, `ScanToken()`. Rewrites the current stub. |
| `errors.go` | `LexError{Code,Line,Col,Msg}` + `Error()` in taxonomy format. |
| `lexer_test.go` | Table-driven unit tests + golden snapshot tests. |
| `testdata/lexer/*.shpl` | Fixture input files. |
| `testdata/lexer/*.golden.txt` | Expected token-stream snapshots. |

### Phase 2 — `internal/parser/`

| File | Responsibility |
|------|---------------|
| `ast.go` | All AST node types + `String()` + `json` struct tags. |
| `dictionary.go` | Curated noun/adjective/comparative/pronoun sets + lookup functions. |
| `errors.go` | `ParseError{Code,Line,Col,Msg}` + `Error()` in taxonomy format. |
| `parser.go` | `Parser` struct, `New()`, `Parse()`, cursor helpers, each `parseXxx` method. |
| `parser_test.go` | Table-driven unit tests + golden JSON snapshot tests. |
| `testdata/parser/*.shpl` | Fixture input files. |
| `testdata/parser/*.golden.json` | Expected AST JSON snapshots. |

### CLI — `cmd/shpl/` (minimal wiring, Phases 1 & 2)

| File | Responsibility |
|------|---------------|
| `main.go` | Cobra root command + `tokens` subcommand (Phase 1) + `ast` subcommand (Phase 2). Rewrites the current stub. |
| `main_test.go` | CLI integration tests (run subcommand on a fixture, check output). |

---

## Phase 1: Lexer

### Objective

Turn `.shpl` source text into an ordered slice of typed tokens (`WORD` + structural punctuation + `NEWLINE` + `EOF`) with line/column tracking, emitting `L001` only for control characters and `L002` for unterminated stage directions. The lexer is deliberately dumb: no keyword recognition, no word classification, no dictionary. It recognizes only structural punctuation (`.` `,` `:` `!` `?` `[` `]`), newlines, and folds everything else into `WORD` tokens.

---

### Step 1.0: Reconcile docs against canonical SPL

* **Goal:** Before any parsing logic, cross-check `docs/ERROR_TAXONOMY.md` and `docs/SPL_SPECIFICATION.md` against the canonical reference and fix the 10 documented discrepancies so both phases build on vetted ground truth. This is the verification step: do not assume the taxonomy doc is fully reliable.
* **Context Files to Reference:** `docs/SPL_SPECIFICATION.md`, `docs/ERROR_TAXONOMY.md`, canonical reference (https://grokipedia.com/page/Shakespeare_Programming_Language, https://esolangs.org/wiki/Shakespeare).
* **Implementation Details:**
  1. **Edit `docs/ERROR_TAXONOMY.md`:**
     - Rewrite S012 from "Exeunt must stand alone" to: "Exeunt may stand alone (exits all characters currently on stage) or optionally name one or more characters to exit (`[Exeunt A and B]`). Use `Exit` for a single character."
     - Refine L001: change the example/message to reflect control-char-only firing. New message: `unexpected control character 0x%02X at line X, col Y`. Add a note: "Other unexpected printable characters fold into WORD tokens and surface as S-code parse errors."
     - Add S017 `InvalidComparative`: `expected comparative phrase (e.g., 'as good as', 'better than'), got '<word>'`.
     - Add S018 `InvalidStackOp`: `expected expression after 'Remember'` or `expected '.' after 'Recall'`.
  2. **Append a `## Canonical Grammar (vetted)` section to `docs/SPL_SPECIFICATION.md`** documenting (do not rewrite existing prose, append only):
     - No-copula assignment: `You <constant>!` / `Thou <constant>!` (assigns constant to listener, no "are"/"art").
     - Simile semantics: `as <adj> as <expr>` evaluates to `<expr>`'s value (NOT 0).
     - Corrected example: `a vile coward` = -2 (1 adjective × noun -1).
     - Stack operations: `Remember <expr>.` (push listener's eval of expr onto listener's stack); `Recall <ignored text>.` (pop listener's stack into speaker's value).
     - Unary operations: `the square of`, `the cube of`, `the square root of`, `the factorial of`, `twice`.
     - The 7 comparative forms → 6 relations table (see Step 2.14 for the full table).
     - Pronoun reference: possessive `my/thy/your/his/her/mine/thine` (ignored like articles in constants); reflexive `me/myself` (speaker), `thyself/yourself` (listener).
     - Question forms: `Am I <comparative> you?` and `Is <X> <comparative> <Y>?`.
     - Branch forms: `If so, let us proceed to scene/act X.`, `If not, let us proceed to scene/act X.`, `Let us proceed to scene/act X.`, `Let us return to scene/act X.`.
  3. **Do not rewrite existing prose; append only.** The existing prose stays as a historical record; the vetted addendum is the parser's source of truth.
* **Error Handling:** Establishes the corrected code set (L001 refined, S012 rewritten, S017 + S018 added) used by all later steps.
* **Testing & Verification:** No code; doc review. Verify every construct used by `testdata/lexer/hello.shpl` (the canonical Hello World, incl. `[Exeunt Ophelia and Hamlet]`, `You lying... coward!`, `Thou art as sweet as...`, `summer's day`, `A/S`) is documented in the vetted addendum. Manual checklist.
* **Documentation Update:** Note reconciliation + all 10 fixes in `PROGRESS.md` Phase 1 "Decisions" section. Commit: `docs: reconcile SPL spec and error taxonomy with canonical reference`.

---

### Step 1.1: Token types & Token struct

* **Goal:** Define the lexer's vocabulary: 9 token types and a `Token` value type with position info.
* **Context Files to Reference:** `docs/ERROR_TAXONOMY.md` (L001 scope, reconciled in 1.0).
* **Implementation Details:** Create `internal/lexer/token.go`:
  - `type TokenType string` with constants:
    - `TokenEOF TokenType = "EOF"`
    - `TokenNewline TokenType = "NEWLINE"`
    - `TokenWord TokenType = "WORD"`
    - `TokenPeriod TokenType = "PERIOD"` (lexeme `.`)
    - `TokenComma TokenType = "COMMA"` (lexeme `,`)
    - `TokenColon TokenType = "COLON"` (lexeme `:`)
    - `TokenBang TokenType = "BANG"` (lexeme `!`)
    - `TokenQuestion TokenType = "QUESTION"` (lexeme `?`)
    - `TokenLBracket TokenType = "LBRACKET"` (lexeme `[`)
    - `TokenRBracket TokenType = "RBRACKET"` (lexeme `]`)
  - `type Token struct { Type TokenType; Lexeme string; Line, Col int }`
  - `func (t Token) String() string` returns `"{<Type> <Lexeme> <line>:<col>}"` for debug/golden files (e.g., `"{WORD Romeo 3:5}"`). EOF renders as `"{EOF  3:1}"`.
  - No `Number`/`RomanNumeral`/`Title`/`Character` token types (D10).
* **Error Handling:** None yet (pure types).
* **Testing & Verification:** TDD:
  - [ ] Write `TestTokenString` in `internal/lexer/token_test.go` asserting `Token{TokenWord, "Romeo", 3, 5}.String()` == `"{WORD Romeo 3:5}"` and `Token{TokenEOF, "", 1, 1}.String()` == `"{EOF  1:1}"`.
  - [ ] Run `go test -race -run TestTokenString ./internal/lexer/` → fails (no file).
  - [ ] Create `token.go`; run → pass.
  - [ ] `task fmt && git add internal/lexer/token.go internal/lexer/token_test.go && git commit -m "feat(lexer): add token types and Token struct"`.
* **Documentation Update:** `PROGRESS.md` Phase 1 — check off "Define token types".

---

### Step 1.2: Lexer core — source scan, position, whitespace, NEWLINE, L001

* **Goal:** `Lexer` struct tracking source bytes, line, col; scan skipping spaces/tabs/`\r`; emit `NEWLINE` on `\n`; emit `L001` for control characters.
* **Context Files to Reference:** `docs/ERROR_TAXONOMY.md` L001 (reconciled: control chars only).
* **Implementation Details:** Create `internal/lexer/errors.go` and rewrite `internal/lexer/lexer.go`:
  - **`errors.go`:**
    - `type LexError struct { Code string; Line, Col int; Msg string }`
    - `func (e LexError) Error() string` returns taxonomy format:
      ```
      error[L001]: unexpected control character 0x07 at line 3, col 1
        --> input:3:1
      ```
      (Filename is cosmetic for now — use `"input"` since the lexer doesn't track source filename. Phase 5 CLI will pass the real filename.)
  - **`lexer.go`:**
    - `type Lexer struct { source []byte; pos, line, col int }`
    - `func New(src string) *Lexer` — initializes `source: []byte(src)`, `line: 1`, `col: 1`, `pos: 0`.
    - `func (l *Lexer) ScanTokens() ([]Token, error)` — drives `ScanToken()` in a loop until `TokenEOF` is emitted or an error is returned. Collects tokens into a slice. Returns `(nil, err)` on first error; returns `(tokens, nil)` including the trailing `TokenEOF` on success.
    - `func (l *Lexer) ScanToken() (Token, error)` — main dispatch. The loop in `ScanTokens` skips "nil sentinel" tokens (whitespace) by checking: if `ScanToken` returns a token with `Type == ""` (empty sentinel), continue without appending.
      - `c := l.peek()`; switch on `c`:
        - `0` (EOF sentinel) → handled in Step 1.5; for now return `Token{TokenEOF, "", l.line, l.col}, nil`.
        - space (`0x20`), tab (`0x09`), `\r` (`0x0D`) → `l.advance()`, return empty-sentinel `Token{Type: ""}` (skip, don't append). `col++` via `advance()`.
        - `\n` (`0x0A`) → emit `Token{TokenNewline, "\\n", l.line, l.col}`, then `l.line++; l.col = 1; l.advance()`.
        - control char (0x00–0x08, 0x0B, 0x0C, 0x0E–0x1F, 0x7F) → return `LexError{Code: "L001", Line: l.line, Col: l.col, Msg: fmt.Sprintf("unexpected control character 0x%02X at line %d, col %d", c, l.line, l.col)}`.
        - default → deferred to Steps 1.3/1.4 (for now, to make this step testable in isolation, treat default as `advance()` + skip; will be replaced).
    - `func (l *Lexer) advance() byte` — returns `source[l.pos]`, increments `l.pos`, increments `l.col`. Does NOT cross newlines.
    - `func (l *Lexer) peek() byte` — returns `0` if `l.pos >= len(source)`, else `source[l.pos]`. Zero is the EOF sentinel (safe because NUL is a control char that would error before reaching here).
    - Keep `slog.Debug("scan token", "type", t.Type, "lexeme", t.Lexeme, "line", t.Line, "col", t.Col)` at each token emission.
* **Error Handling:** `L001` via `LexError`. The error format follows `docs/ERROR_TAXONOMY.md` "Error Format" section.
* **Testing & Verification:** TDD:
  - [ ] `TestNewlineAndWhitespace`: input `"  \t\n  "` → expect one `TokenNewline` at `1:4` (after spaces/tab), rest skipped. Assert the returned slice has exactly 2 tokens: the NEWLINE + EOF.
  - [ ] `TestControlCharL001`: input `"abc\x07"` → but letters not handled yet in this step (default skip); use `"\x07"` alone → expect `L001` error with code `L001`, line 1, col 1, mentioning `0x07`.
  - [ ] Run → fail; implement; run → pass.
  - [ ] `task fmt && git commit -m "feat(lexer): core scan loop with newline and L001 control-char error"`.
* **Documentation Update:** `PROGRESS.md` — check off L001, `New()`, position tracking.

---

### Step 1.3: Structural punctuation tokens

* **Goal:** Emit `.` `,` `:` `!` `?` `[` `]` as single-char tokens.
* **Context Files to Reference:** Reconciled spec (bracket usage, sentence terminators, comma in declarations, colon after speaker).
* **Implementation Details:** Extend the `ScanToken` default switch in `lexer.go`:
  - `.` → emit `Token{TokenPeriod, ".", l.line, l.col}`; `advance()`.
  - `,` → emit `Token{TokenComma, ",", l.line, l.col}`; `advance()`.
  - `:` → emit `Token{TokenColon, ":", l.line, l.col}`; `advance()`.
  - `!` → emit `Token{TokenBang, "!", l.line, l.col}`; `advance()`.
  - `?` → emit `Token{TokenQuestion, "?", l.line, l.col}`; `advance()`.
  - `[` → emit `Token{TokenLBracket, "[", l.line, l.col}`; `advance()`.
  - `]` → emit `Token{TokenRBracket, "]", l.line, l.col}`; `advance()`.
  - Each emits with current line/col, then advances.
* **Error Handling:** None new.
* **Testing & Verification:** TDD table `TestPunctuation` mapping each input rune to its `TokenType` + Lexeme + position:
  - [ ] Input `"."` → `[TokenPeriod ".", 1:1, EOF]`.
  - [ ] Input `","` → `[TokenComma, 1:1, EOF]`.
  - [ ] Input `":"` → `[TokenColon, 1:1, EOF]`.
  - [ ] Input `"!"` → `[TokenBang, 1:1, EOF]`.
  - [ ] Input `"?"` → `[TokenQuestion, 1:1, EOF]`.
  - [ ] Input `"[]"` → `[TokenLBracket, 1:1, TokenRBracket, 1:2, EOF]`.
  - [ ] Input `".:,!?"` → all six in order + EOF.
  - [ ] Run/fail/implement/pass.
  - [ ] `task fmt && git commit -m "feat(lexer): structural punctuation tokens"`.
* **Documentation Update:** `PROGRESS.md` — check off punctuation handling.

---

### Step 1.4: WORD token scanning (fold free-text chars)

* **Goal:** Emit `WORD` for maximal runs of printable non-whitespace, non-newline, non-structural chars. Handles `summer's`, `half-witted`, `A/S`, `coward`, `@`, digits, etc.
* **Context Files to Reference:** Step 1.0 addendum (descriptions are free text), `docs/ERROR_TAXONOMY.md` L001 (printable chars fold into WORD).
* **Implementation Details:** In `ScanToken`, the default branch (not whitespace, not newline, not control char, not structural punctuation):
  - Mark `startLine := l.line; startCol := l.col`.
  - Collect a `WORD`: while `l.peek()` is printable, not whitespace (`0x20`/`0x09`/`0x0D`/`0x0A`), not newline (`0x0A`), and not one of `. , : ! ? [ ]`:
    - Append `l.peek()` to a `strings.Builder`; `l.advance()`.
  - Emit `Token{TokenWord, builder.String(), startLine, startCol}`.
  - `slog.Debug("emitted WORD", "lexeme", lexeme, "line", startLine, "col", startCol)`.
  - **Behavior notes:**
    - `Romeo @ Juliet` → WORDs `Romeo`, `@`, `Juliet` (no L001; `@` is a WORD; parser will reject `@` as invalid with an S-code).
    - `A/S` → single WORD `A/S` (slash is not structural).
    - `summer's` → single WORD `summer's` (aprophe is not structural).
    - `half-witted` → single WORD `half-witted` (hyphen is not structural).
    - `Speak` and `YOUR` → two separate WORDs `Speak`, `YOUR` (space separates).
    - Case is preserved in `Lexeme` (classification/normalization happens in the parser).
* **Error Handling:** None at lexer level. Invalid words surface as S-codes in the parser.
* **Testing & Verification:** TDD:
  - [ ] `TestWordSimple`: `"Romeo"` → `[WORD Romeo 1:1, EOF]`.
  - [ ] `TestWordApostropheHyphen`: `"summer's half-witted"` → `[WORD summer's 1:1, WORD half-witted 1:9, EOF]`.
  - [ ] `TestWordFreeText`: `"A/S"` → `[WORD A/S 1:1, EOF]`.
  - [ ] `TestAtSignFoldsToWord`: `"a @ b"` → `[WORD a 1:1, WORD @ 1:3, WORD b 1:5, EOF]` (no error).
  - [ ] `TestWordBeforePunctuation`: `"Romeo,"` → `[WORD Romeo 1:1, COMMA 1:6, EOF]`.
  - [ ] `TestMixedCasePreserved`: `"Speak YOUR mind"` → `[WORD Speak, WORD YOUR, WORD mind]` (case preserved in Lexeme).
  - [ ] Run/fail/implement/pass.
  - [ ] `task fmt && git commit -m "feat(lexer): WORD token scanning folding free-text chars"`.
* **Documentation Update:** `PROGRESS.md` — check off "Handle multi-word nouns/adjectives" (the lexer folds compound words like `summer's`, `half-witted` into single WORD tokens; the parser handles multi-word noun phrases).

---

### Step 1.5: EOF + ScanTokens driver + LexError L002

* **Goal:** Emit `TokenEOF` at end; `ScanTokens()` returns `[]Token` ending with EOF; define `L002 UnterminatedToken` for an open `[` stage direction that hits EOF without a closing `]`.
* **Context Files to Reference:** `docs/ERROR_TAXONOMY.md` L002.
* **Implementation Details:**
  - **EOF emission:** When `l.peek()` returns `0` (past end): emit `Token{TokenEOF, "", l.line, l.col}`. `ScanTokens` appends it and stops the loop.
  - **L002 tracking:** Add `inBracket bool` field to `Lexer`. When `[` is scanned, set `inBracket = true`. When `]` is scanned, set `inBracket = false`. If `ScanToken` is about to emit EOF while `inBracket == true`, return `LexError{Code: "L002", Line: <bracket-open-line>, Col: <bracket-open-col>, Msg: fmt.Sprintf("unterminated stage direction starting at line %d", <bracket-open-line>)}`. Track the bracket-open line/col in fields `bracketLine`, `bracketCol`.
  - **`ScanTokens` driver loop logic:**
    ```
    for {
        tok, err := l.ScanToken()
        if err != nil { return nil, err }
        if tok.Type == "" { continue }   // skip sentinel (whitespace)
        tokens = append(tokens, tok)
        if tok.Type == TokenEOF { break }
    }
    return tokens, nil
    ```
  - **Edge case:** Multiple `[` without `]` — the lexer only tracks a single `inBracket` flag (not depth). SPL stage directions don't nest. If a second `[` appears while `inBracket` is true, it's a WORD inside the bracket (but `[` is structural, so it would emit `LBRACKET`... hmm). Actually `[` is always structural. Two consecutive `[` without `]` → second `[` emits LBRACKET while `inBracket` is already true. This is a malformed program; the parser will catch it (S011/S013). The lexer stays dumb: just set `inBracket = true` again (no depth tracking). `ponytail: no bracket-depth tracking; SPL stage directions don't nest, malformed nesting surfaces as S-codes`.
* **Error Handling:** `L002` message: `unterminated stage direction starting at line X`. Error format follows taxonomy.
* **Testing & Verification:** TDD:
  - [ ] `TestEOF`: `"Romeo"` → `[WORD Romeo, EOF]` (EOF at `1:6`).
  - [ ] `TestEOFEmpty`: `""` → `[EOF]` (single EOF at `1:1`).
  - [ ] `TestEOFAfterNewline`: `"a\n"` → `[WORD a, NEWLINE 1:2, EOF 2:1]`.
  - [ ] `TestL002UnterminatedBracket`: `"[Enter Romeo"` → `L002` error (no `]` before EOF).
  - [ ] `TestL002NotTriggered`: `"[Enter Romeo]"` → `[LBRACKET, WORD Enter, WORD Romeo, RBRACKET, EOF]` (no error).
  - [ ] Run/fail/implement/pass.
  - [ ] `task fmt && git commit -m "feat(lexer): EOF token, ScanTokens driver, L002 unterminated bracket"`.
* **Documentation Update:** `PROGRESS.md` — check off `ScanToken`/`ScanTokens`, L002.

---

### Step 1.6: Fixtures

* **Goal:** Create canonical + edge fixtures under `testdata/lexer/`.
* **Context Files to Reference:** Reconciled spec examples (canonical Hello World from Grokipedia, Truth Machine from local spec).
* **Implementation Details:** Create the following fixture files (verbatim canonical text):
  - `testdata/lexer/hello.shpl` — the canonical "Infamous Hello World Program" verbatim from the reconciled spec / Grokipedia. Must include: `[Exeunt Ophelia and Hamlet]`, `A/S` (in Hamlet's description), `summer's day`, `You lying stupid fatherless big smelly half-witted coward!` (no-copula assignment), `Thou art as sweet as...`, `Speak YOUR mind!` (mixed case). This is the master lexer fixture.
  - `testdata/lexer/truth-machine.shpl` — the Truth Machine from `docs/SPL_SPECIFICATION.md` (lines 242–264). Includes `Listen to your heart!`, `Open your heart!`, `Am I better than you?`, `If so, let us proceed to scene II.`, `[Exeunt]`.
  - `testdata/lexer/minimal.shpl` — a 4-line synthetic program for fast tests:
    ```
    The Minimal Program.

    Romeo, a young man.

    Act I: A single scene.
    Scene I: The only scene.

    [Enter Romeo and Juliet]

    Juliet:
    Speak your mind.

    [Exeunt]
    ```
    (Note: this references Juliet without declaring her — the lexer doesn't care; the parser/semantic phases catch that. For lexer testing, only token structure matters.)
  - `testdata/lexer/bad-char.shpl` — a file containing a BEL byte (0x07) embedded after a valid word, e.g., `Romeo\x07Juliet` → triggers L001.
* **Error Handling:** None (fixtures are data).
* **Testing & Verification:** 
  - [ ] `go test -race ./internal/lexer/` with a smoke test that reads `hello.shpl` and ensures `ScanTokens` returns no error (just verifies the canonical fixture is lexable).
  - [ ] Smoke test that reads `bad-char.shpl` and asserts `L001` error.
  - [ ] `task fmt && git commit -m "test(lexer): add canonical and edge fixtures"`.
* **Documentation Update:** `PROGRESS.md` — check off "Fixtures: testdata/lexer/...".

---

### Step 1.7: Table-driven + golden snapshot tests

* **Goal:** Lock lexer behavior: a unit table covering every token type and edge case, plus golden-file snapshots of the full token stream for each fixture.
* **Context Files to Reference:** `testdata/lexer/*.shpl`.
* **Implementation Details:** In `internal/lexer/lexer_test.go`:
  - **`TestScanTokenTable`** — table-driven, one case per behavior:
    - Each case: `{name, input, expectedTokens, expectedError}`. 
    - Cover: empty string → `[EOF]`; single punctuation of each type; single WORD; WORD + punctuation; apostrophe/hyphen WORD; free-text `A/S`; `@` as WORD; newline; whitespace skip; mixed case preserved; control char → L001; unterminated bracket → L002; multi-line (line/col tracking across `\n`).
  - **`TestGoldenSnapshots`** — for each fixture in `testdata/lexer/`:
    - Read the `.shpl` file, run `ScanTokens`, format each token as `tok.String()` joined by `\n`, compare against `testdata/lexer/<name>.golden.txt`.
    - **Golden file generation:** If the golden file does not exist, write it and fail the test with message `"golden file <path> created; inspect and re-run"`. If it exists, compare. This lets the developer generate goldens on first run, review them, then lock them.
    - `bad-char.shpl` has no golden (it errors); assert the error instead.
  - **`TestMain`** in `lexer_test.go`: call `logger.Init(logger.LevelDebug)` so `slog.Debug` output appears during test runs (and can be inspected if a test fails).
* **Error Handling:** Tests assert error codes L001/L002 match expected.
* **Testing & Verification:**
  - [ ] Write `TestScanTokenTable` with all cases above.
  - [ ] Write `TestGoldenSnapshots` with the golden-file-generation pattern.
  - [ ] Write `TestMain` with logger init.
  - [ ] Run `go test -race -v ./internal/lexer/` → golden files created on first run; inspect them; re-run → all pass.
  - [ ] `task fmt && git add testdata/lexer/*.golden.txt && git commit -m "test(lexer): table-driven and golden snapshot tests"`.
* **Documentation Update:** `PROGRESS.md` — check off "Tests: table-driven, snapshot".

---

### Step 1.8: Minimal `tokens` CLI subcommand

* **Goal:** Wire Cobra minimally so `shpl tokens <file.shpl>` reads a file, runs the lexer, and prints the token stream to stdout. This provides a runnable deliverable for Phase 1.
* **Context Files to Reference:** `cmd/shpl/main.go` (current stub), `go.mod` (cobra already listed).
* **Implementation Details:** Rewrite `cmd/shpl/main.go`:
  - Import `github.com/spf13/cobra`.
  - Root command `shpl` with `Short: "Shakespeare Programming Language interpreter"`.
  - Subcommand `tokens`:
    - `Use: "tokens <file>"`, `Args: cobra.ExactArgs(1)`.
    - `RunE`: read the file via `os.ReadFile`, create `lexer.New(string(content))`, call `ScanTokens()`, print each token via `fmt.Println(tok.String())`. On error, print to stderr and return non-zero.
  - Global persistent flag `--debug` on root: if set, call `logger.Init(logger.LevelDebug)` before running; else `logger.Init(logger.LevelInfo)`.
  - `main()` calls `rootCmd.Execute()` and exits with code 1 on error.
  - Keep the existing `TestMainEngineSetup` test passing (it's a trivial boolean check; don't break it).
* **Error Handling:** File-not-found → print `error: cannot read file '<path>': <err>` to stderr, exit 1. Lexer error → print the `LexError.Error()` string, exit 1.
* **Testing & Verification:**
  - [ ] Write `TestTokensCommand` in `cmd/shpl/main_test.go`: run `rootCmd` with args `["tokens", "testdata/lexer/minimal.shpl"]`, capture stdout, assert it contains `WORD` and `EOF` token lines.
  - [ ] Run `go test -race -run TestTokensCommand ./cmd/shpl/` → pass.
  - [ ] Run `go test -race ./cmd/shpl/` → all pass (including existing `TestMainEngineSetup`).
  - [ ] Manual smoke: `task build && ./bin/shpl.exe tokens testdata/lexer/hello.shpl | head -5` → see token lines.
  - [ ] `task fmt && git commit -m "feat(cli): add tokens subcommand for lexer output"`.
* **Documentation Update:** `PROGRESS.md` Phase 1 — check off "CLI subcommand: shpl tokens".

---

## Phase 2: Parser

### Objective

Parse the lexer's token stream into a typed AST covering the full canonical SPL grammar. The parser is a hand-written recursive-descent parser that owns a curated noun/adjective/comparative dictionary plus the set of declared character names. It classifies `WORD` tokens into typed AST nodes (constants, character references, pronouns, operations) and validates structural syntax (acts, scenes, stage directions, dialogue, statements). It emits `S-code` errors from `error_taxonomy.md` for invalid arrangements.

The parser does **not** evaluate expressions, check semantic validity (undeclared characters, stage occupancy — that's Phase 3), or execute anything. It only builds the AST.

---

### Step 2.1: AST node types

* **Goal:** Define all AST node types as Go structs with `json` tags for snapshot serialization and `String()` for debug.
* **Context Files to Reference:** Reconciled spec "Canonical Grammar (vetted)" section (from Step 1.0), `PROGRESS.md` Phase 2 AST node list.
* **Implementation Details:** Create `internal/parser/ast.go`. All nodes implement a common `Node` interface with `String() string` and a `Pos() (int, int)` method (line, col). Define:

  **Top-level:**
  - `type Program struct { Title Title; Characters []CharacterDecl; Acts []Act; Line, Col int }`
  - `type Title struct { Text string; Line, Col int }` — the title text (everything before the terminating `.`).

  **Declarations:**
  - `type CharacterDecl struct { Name string; Description string; Line, Col int }` — Name is the character name (WORD), Description is the free-text after the comma.

  **Structure:**
  - `type Act struct { Number int; RomanNumeral string; Description string; Scenes []Scene; Line, Col int }` — Number is the parsed integer (1, 2, 3...), RomanNumeral is the original text ("I", "II").
  - `type Scene struct { Number int; RomanNumeral string; Description string; Statements []Statement; Line, Col int }`

  **Stage directions:**
  - `type EnterStmt struct { Characters []string; Line, Col int }` — 1 or 2 character names.
  - `type ExitStmt struct { Character string; Line, Col int }` — exactly 1 character name.
  - `type ExeuntStmt struct { Characters []string; Line, Col int }` — 0 names (exit all) or 1+ names (exit named). Empty slice = exit all.

  **Dialogue:**
  - `type Dialogue struct { Speaker string; Statements []Statement; Line, Col int }` — Speaker is the character name before `:`.

  **Statement (sum type via interface):**
  - `type Statement interface { Node; stmtNode() }` — marker method.
  - `type AssignStmt struct { Target string; SimileAdj string; Expr Expr; Terminator string; Line, Col int }` — Target is always the listener ("you"/"thou", normalized to "you"); SimileAdj is the adjective in a simile (empty if no simile); Expr is the assigned expression; Terminator is `.`, `!`, or `?`.
  - `type SpeakStmt struct { Terminator string; Line, Col int }` — "Speak your/thy mind." → output char.
  - `type OpenHeartStmt struct { Terminator string; Line, Col int }` — "Open your/thy heart." → output number.
  - `type OpenMindStmt struct { Terminator string; Line, Col int }` — "Open your/thy mind." → input char.
  - `type ListenStmt struct { Terminator string; Line, Col int }` — "Listen to your/thy heart." → input number.
  - `type QuestionStmt struct { Left Expr; Comparative Comparative; Right Expr; Line, Col int }` — "Am I/Is X <comparative> Y?" Left and Right are expressions (for "Am I", Left = speaker pronoun; for "Is X", Left = X).
  - `type IfStmt struct { BranchIfTrue bool; Target string; TargetKind string; Line, Col int }` — `If so` (BranchIfTrue=true) or `If not` (BranchIfTrue=false); Target is the Roman numeral; TargetKind is "scene" or "act".
  - `type GotoStmt struct { Target string; TargetKind string; Line, Col int }` — "Let us proceed/return to scene/act X." Target is Roman numeral; TargetKind is "scene" or "act".
  - `type RememberStmt struct { Expr Expr; Line, Col int }` — "Remember <expr>." → push expr onto listener's stack.
  - `type RecallStmt struct { IgnoredText string; Line, Col int }` — "Recall <ignored text>." → pop listener's stack into speaker.

  **Expression (sum type via interface):**
  - `type Expr interface { Node; exprNode() }` — marker method.
  - `type ConstExpr struct { AdjectiveCount int; Noun string; Polarity int; Line, Col int }` — Polarity is +1 or -1; value (computed in Phase 4) = Polarity × 2^AdjectiveCount. Noun is the noun word.
  - `type CharRefExpr struct { Name string; Line, Col int }` — reference to a declared character variable.
  - `type PronounExpr struct { Ref string; Line, Col int }` — Ref is "speaker" or "listener" (normalized from me/myself → speaker, thyself/yourself → listener).
  - `type BinaryOpExpr struct { Op string; Left, Right Expr; Line, Col int }` — Op is one of: "sum", "difference", "product", "quotient", "remainder".
  - `type UnaryOpExpr struct { Op string; Operand Expr; Line, Col int }` — Op is one of: "square", "cube", "square_root", "factorial", "twice".

  **Comparative:**
  - `type Comparative struct { Form string; Adjective string; Negated bool; Relation string; Line, Col int }` — Form is "as-as" or "than"; Adjective is the comparative word; Negated is whether "not" preceded; Relation is the resolved relation: "equal", "not_equal", "greater", "less", "greater_or_equal", "less_or_equal".

  - Each struct gets `json:"..."` tags on all fields. Each gets a `String()` method returning a compact one-line representation (e.g., `AssignStmt{target=you, expr=ConstExpr{adj=2, noun=flower, pol=+1}, term=!}`). Each gets `stmtNode()` or `exprNode()` empty methods. Each gets `Pos()` returning `(Line, Col)`.
* **Error Handling:** None yet (pure types).
* **Testing & Verification:** TDD:
  - [ ] Write `TestNodeString` asserting a few representative `String()` outputs (e.g., `ConstExpr{AdjectiveCount: 2, Noun: "flower", Polarity: 1}.String()` == `ConstExpr{adj=2, noun=flower, pol=+1}`).
  - [ ] Write `TestASTJSON` asserting `json.Marshal` of a small hand-built AST produces expected JSON (verifies tags).
  - [ ] Run/fail/create/pass.
  - [ ] `task fmt && git commit -m "feat(parser): add AST node types"`.
* **Documentation Update:** `PROGRESS.md` Phase 2 — check off "Define AST node types".

---

### Step 2.2: Parser errors

* **Goal:** Define `ParseError` with S-codes from `error_taxonomy.md` (including S017/S018 added in Step 1.0).
* **Context Files to Reference:** `docs/ERROR_TAXONOMY.md` (reconciled), all S-codes.
* **Implementation Details:** Create `internal/parser/errors.go`:
  - `type ParseError struct { Code string; Line, Col int; Msg string }`
  - `func (e ParseError) Error() string` returns:
    ```
    error[S001]: expected program title ending with '.'
      --> input:1:1
    ```
    (Same format as `LexError`; filename `"input"` for now.)
  - Helper constructor functions for each S-code (one per code) to keep messages consistent:
    - `func errMissingTitle(line, col int) ParseError` → S001
    - `func errMissingCharacterDecl(line, col int) ParseError` → S002
    - `func errInvalidCharacterName(name string, line, col int) ParseError` → S003 (note: this is a **warning**, not a hard error; see below)
    - `func errMissingAct(line, col int) ParseError` → S004
    - `func errInvalidActNumber(got string, line, col int) ParseError` → S005
    - `func errActOrder(expected, got string, line, col int) ParseError` → S006
    - `func errMissingScene(actRoman string, line, col int) ParseError` → S007
    - `func errInvalidSceneNumber(got string, line, col int) ParseError` → S008
    - `func errSceneOrder(expected, got string, line, col int) ParseError` → S009
    - `func errMissingEnter(name string, line, col int) ParseError` → S010
    - `func errInvalidEnter(line, col int) ParseError` → S011
    - `func errInvalidExeunt(line, col int) ParseError` → S012 (updated: Exeunt with invalid args, e.g., `[Exeunt ,]`)
    - `func errMissingStage(line, col int) ParseError` → S013
    - `func errMissingSpeaker(line, col int) ParseError` → S014
    - `func errInvalidExpression(line, col int) ParseError` → S015
    - `func errInvalidIf(line, col int) ParseError` → S016
    - `func errInvalidComparative(got string, line, col int) ParseError` → S017
    - `func errInvalidStackOp(msg string, line, col int) ParseError` → S018
  - **S003 is a warning:** `errInvalidCharacterName` returns a `Warning` type (not a `ParseError`) — `type Warning struct { Code string; Line, Col int; Msg string }`. Warnings are collected but don't stop parsing. The parser accumulates `[]Warning` alongside the AST. `ponytail: S003 is advisory only; spec doesn't enforce Shakespeare-character names`.
* **Error Handling:** This step defines the error vocabulary for the entire parser.
* **Testing & Verification:** TDD:
  - [ ] `TestParseErrorFormat`: assert `errMissingTitle(1, 1).Error()` starts with `error[S001]:` and contains `input:1:1`.
  - [ ] `TestWarningIsNotError`: assert `Warning` does not implement `error` (or if it does, it's wrapped separately — decide: Warning is NOT an error, just a struct collected in a slice).
  - [ ] Run/fail/create/pass.
  - [ ] `task fmt && git commit -m "feat(parser): add ParseError and Warning types with S-codes"`.
* **Documentation Update:** `PROGRESS.md` Phase 2 — note S017/S018 added. `docs/ERROR_TAXONOMY.md` already updated in Step 1.0.

---

### Step 2.3: Dictionary — nouns, adjectives, comparatives, pronouns

* **Goal:** Curated word classification data for the parser: noun polarity, adjective recognition, comparative adjective polarity, pronoun classification, structural keyword recognition.
* **Context Files to Reference:** Reconciled spec, canonical SPL vocab from Grokipedia examples.
* **Implementation Details:** Create `internal/parser/dictionary.go`:
  - **Noun polarity:** `var positiveNouns = map[string]bool{...}` and `var negativeNouns = map[string]bool{...}`. A noun is positive if in `positiveNouns`, negative if in `negativeNouns`, neutral (+1) otherwise. Curated entries (~30 each):
    - Positive: `flower, hero, king, angel, rose, cat, horse, tree, sky, squirrel, hamster, pony, town, purse, kingdom, day, brother, father, heart, mind, lord, liege, cousin, friend, star, sun, moon, rose, joy, treasure, gold`
    - Negative: `coward, liar, fool, pig, blister, leech, codpiece, beggar, thief, villain, traitor, knave, rat, toad, plague, famine, pestilence, misery, bastard, idiot, moron, trash, garbage, dirt, shadow, venom, poison, scum`
  - `func nounPolarity(word string) int` — lowercases `word`; returns -1 if in `negativeNouns`, +1 otherwise (positive + neutral both +1). `ponytail: unknown noun defaults to +1 per spec`.
  - **Adjective set:** `var adjectives = map[string]bool{...}` — used to count adjectives in constants. Curated (~40):
    - `red, hot, big, handsome, rich, brave, beautiful, fair, warm, peaceful, sunny, sweet, good, great, fine, lovely, amazing, bold, cute, fat, little, stuffed, misused, dusty, old, rotten, green, huge, large, rural, bottomless, embroidered, bluest, clearest, sweetest, reddest, smooth, small, furry, white, black, half-witted, stupid, lying, fatherless, smelly, vile, cowardly, worried, bad, sick, dead, young, remarkable, infected, oozing, summer's, handsome, proud, mighty, healthy`
    - Note: `summer's` is an adjective (modifies `day` in `summer's day`).
  - `func isAdjective(word string) bool` — lowercases; returns true if in `adjectives` map. Unknown word treated as NOT an adjective (it becomes the noun). `ponytail: unknown adjective stops the adjective scan; the word becomes the noun`.
  - **Comparative adjectives:** `var positiveComparatives = map[string]bool{...}` and `var negativeComparatives = map[string]bool{...}`. Curated (~15 each):
    - Positive: `better, bigger, greater, larger, fairer, braver, sweeter, nicer, richer, warmer, brighter, stronger, faster, higher, lower` (wait, "lower" is negative... let me reconsider). Positive comparatives convey "more/better": `better, bigger, greater, larger, fairer, braver, sweeter, nicer, richer, warmer, brighter, stronger, more, beautiful`
    - Negative comparatives convey "less/worse": `worse, smaller, poorer, uglier, meaner, baser, lower, weaker, slower, colder, darker, fouler, viler, sicker, deader`
    - Also non-comparative-form adjectives used in "as ADJ as": `good, bad, big, small, large, little, fair, brave, sweet, lovely, stupid, cowardly, beautiful, disgusting, healthy, worried, loving` — these appear in the "as <adj> as" comparative form AND in similes. The parser context disambiguates (simile in assignment vs. comparative in question).
  - `func comparativePolarity(word string) string` — returns "positive" or "negative". Unknown comparative → "positive" (default).
  - **Pronouns:**
    - `var possessivePronouns = map[string]bool{...}`: `my, thy, your, his, her, mine, thine` — ignored like articles in constant parsing.
    - `var speakerPronouns = map[string]bool{...}`: `me, myself` → PronounExpr("speaker").
    - `var listenerPronouns = map[string]bool{...}`: `thyself, yourself, thou, you` → PronounExpr("listener"). (Note: `you`/`thou` as subjects in assignments are handled by the statement parser, not the expression parser. In expression context, `yourself`/`thyself` are the listener refs.)
    - `func isPossessive(word string) bool`, `func isSpeakerPronoun(word string) bool`, `func isListenerPronoun(word string) bool`.
  - **Articles:** `var articles = map[string]bool{...}`: `a, an, the` — skipped in constant parsing. (Note: `the` is also an operator keyword prefix — `the sum of`. The expression parser checks for operator keywords before treating `the` as an article.)
  - **Structural keywords** (for the parser's statement dispatch): these are recognized by the parser's `peek`/`match` logic, not by the dictionary. But list them here for reference:
    - Act/Scene headers: `act, scene`
    - Stage directions: `enter, exit, exeunt`
    - Statement keywords: `remember, recall, listen, open, speak, if, let, am, is`
    - Assignment: `you, thou, are, art, as`
    - Operator keywords: `the, sum, difference, product, quotient, remainder, square, cube, root, factorial, twice, of, between, and, than`
    - Branch keywords: `so, not, proceed, return, to`
* **Error Handling:** None (pure data + lookups).
* **Testing & Verification:** TDD:
  - [ ] `TestNounPolarity`: `nounPolarity("flower")` == 1, `nounPolarity("coward")` == -1, `nounPolarity("chrysanthemum")` == 1 (unknown → neutral +1).
  - [ ] `TestIsAdjective`: `isAdjective("big")` == true, `isAdjective("flower")` == false (it's a noun, not an adj), `isAdjective("summer's")` == true.
  - [ ] `TestComparativePolarity`: `comparativePolarity("better")` == "positive", `comparativePolarity("worse")` == "negative", `comparativePolarity("chrysanthemum")` == "positive" (unknown → default).
  - [ ] `TestPronounClassification`: `isSpeakerPronoun("me")` == true, `isListenerPronoun("thyself")` == true, `isPossessive("my")` == true.
  - [ ] Run/fail/create/pass.
  - [ ] `task fmt && git commit -m "feat(parser): add curated word dictionary"`.
* **Documentation Update:** `PROGRESS.md` Phase 2 — note dictionary approach (curated ~80 words).

---

### Step 2.4: Parser core — cursor, helpers, New(), Parse()

* **Goal:** The `Parser` struct with token cursor, lookahead helpers, and the top-level `Parse()` entry point.
* **Context Files to Reference:** `internal/lexer/token.go` (Token type).
* **Implementation Details:** Create `internal/parser/parser.go`:
  - `type Parser struct { tokens []lexer.Token; pos int; characters map[string]bool; warnings []Warning }`
    - `tokens` is the token slice from the lexer.
    - `pos` is the current cursor index.
    - `characters` is the set of declared character names (populated during `parseCharacterDecls`), lowercased for case-insensitive lookup.
    - `warnings` accumulates S003 warnings.
  - `func New(tokens []lexer.Token) *Parser` — initializes with empty `characters` map and `pos: 0`.
  - **Cursor helpers:**
    - `func (p *Parser) peek() lexer.Token` — returns `tokens[p.pos]` (or a synthetic EOF token if past end).
    - `func (p *Parser) peekAt(offset int) lexer.Token` — returns `tokens[p.pos+offset]` (or synthetic EOF).
    - `func (p *Parser) advance() lexer.Token` — returns `tokens[p.pos]`, increments `p.pos`.
    - `func (p *Parser) at(t lexer.TokenType) bool` — `peek().Type == t`.
    - `func (p *Parser) match(t lexer.TokenType) bool` — if `at(t)`, `advance()` and return true; else false.
    - `func (p *Parser) expect(t lexer.TokenType) (lexer.Token, error)` — if `at(t)`, return `advance()`; else return `ParseError{Code: "S015", ...}` (generic syntax error; specific steps use specific S-codes).
    - `func (p *Parser) checkWord(value string) bool` — `peek().Type == TokenWord && strings.ToLower(peek().Lexeme) == value`. Case-insensitive keyword matching.
    - `func (p *Parser) matchWord(value string) bool` — if `checkWord(value)`, `advance()` and return true; else false.
    - `func (p *Parser) skipNewlines()` — while `at(TokenNewline)`, `advance()`. SPL ignores blank lines between structural elements.
    - `func (p *Parser) isEOF() bool` — `peek().Type == TokenEOF`.
  - **`func (p *Parser) Parse() (*Program, error)`** — top-level entry:
    1. `p.skipNewlines()`.
    2. Parse title → `title, err := p.parseTitle()`.
    3. `p.skipNewlines()`.
    4. Parse character declarations → `chars, err := p.parseCharacterDecls()`. Populate `p.characters` map.
    5. `p.skipNewlines()`.
    6. Parse acts → `acts, err := p.parseActs()`.
    7. Return `&Program{Title: title, Characters: chars, Acts: acts, Line: ..., Col: ...}`.
    8. Each sub-parse returns on first error (fail-fast).
  - `slog.Debug` at each major parse step: `"parsing title"`, `"parsing character decls"`, `"parsing acts"`, etc.
  - **Note on `TokenEOF` and `TokenNewline`:** The parser skips `NEWLINE` tokens between structural elements via `skipNewlines()`. Within dialogue, newlines are also skipped (statements are terminated by `. ! ?`, not newlines). The parser treats newlines as insignificant whitespace throughout.
* **Error Handling:** `Parse()` returns `(*Program, error)` where error is a `ParseError`. Warnings are stored in `p.warnings` and returned alongside the program (via a method `p.Warnings() []Warning` or as a field — decide: return `(*Program, []Warning, error)` from `Parse()`. Actually, to keep the signature clean, add a `Warnings` field to `Program`.)
* **Testing & Verification:** TDD (minimal — the full `Parse()` is tested via golden snapshots in Step 2.17; here test just the helpers):
  - [ ] `TestParserHelpers`: feed a token slice, test `peek`, `advance`, `at`, `match`, `checkWord`, `skipNewlines`.
  - [ ] `TestParseEmpty`: empty token slice (just EOF) → `Parse()` returns S001 error (no title).
  - [ ] Run/fail/create/pass.
  - [ ] `task fmt && git commit -m "feat(parser): add Parser struct, cursor helpers, and Parse entry point"`.
* **Documentation Update:** `PROGRESS.md` Phase 2 — check off "Recursive descent parser" start.

---

### Step 2.5: Parse Title + Character Declarations

* **Goal:** Parse the program title (first line ending with `.`) and character declarations (`Name, description.`). Apply S001, S002, S004.
* **Context Files to Reference:** Reconciled spec "Program Title and Dramatis Personae", `docs/ERROR_TAXONOMY.md` S001/S002/S003/S004.
* **Implementation Details:** In `parser.go`, add:
  - **`parseTitle() (Title, error)`:**
    - Collect all WORD tokens (and any non-period, non-newline, non-EOF tokens) into the title text, separated by spaces, until hitting `TokenPeriod`.
    - If the first token is NOT a `TokenWord` (e.g., immediately a period or EOF) → S001 `errMissingTitle`.
    - If we hit `TokenEOF` or `TokenNewline` before `TokenPeriod` → S001 (title must end with `.`).
    - Consume the `TokenPeriod`.
    - Skip any trailing `TokenNewline`.
    - Return `Title{Text: collectedText, Line: firstLine, Col: firstCol}`.
    - **Behavior:** The title text includes everything before the `.` — words, free-text chars, etc. — joined by spaces. The `.` terminates. E.g., `"The Infamous Hello World Program."` → `Title{Text: "The Infamous Hello World Program"}`.
  - **`parseCharacterDecls() ([]CharacterDecl, error)`:**
    - Loop: while the next token is `TokenWord` AND it is NOT `act` (case-insensitive) — because character declarations are followed by acts:
      - Parse one declaration:
        - Read WORD tokens for the character name until `TokenComma`. The name is the first WORD (single word, D8). Actually, character names in the canonical examples are single words (Romeo, Juliet, Hamlet, Ophelia). But the name could be multi-word? The spec says "Names must be characters from Shakespeare plays." Some Shakespeare characters have multi-word names (e.g., "Sir Andrew"). For v1, single WORD only (D8). If there are multiple WORDs before the comma, take only the first as the name and treat the rest as part of the description? No — that would misparse. Decision (D8): the name is exactly one WORD. If multiple WORDs appear before the comma, it's a parse error (or take the first and include rest in description). `ponytail: single-word character names only; multi-word names are a v2 concern`.
        - `name := p.advance().Lexeme` (the first WORD).
        - Expect `TokenComma` — if not, S014 `errMissingSpeaker`? No, that's for dialogue. For declarations, if no comma after name → S002 `errMissingCharacterDecl` (malformed declaration).
        - Read the description: collect all tokens (WORD, and any non-period tokens) until `TokenPeriod`. Join with spaces. The description must have at least one non-whitespace token — if we hit `TokenPeriod` immediately (empty description), S002.
        - Expect `TokenPeriod`.
        - `p.skipNewlines()`.
        - Add `CharacterDecl{Name: name, Description: desc, Line, Col}`.
        - Add `strings.ToLower(name)` to `p.characters` map.
        - **S003 warning:** if `name` (lowercased) is not in a curated list of known Shakespeare characters (a small set: romeo, juliet, hamlet, ophelia, mercutio, tybalt, etc.) → append S003 Warning. `ponytail: S003 is advisory; the spec doesn't enforce this strictly`. Include a small `shakespeareCharacters` set in `dictionary.go` (~20 names). Unknown → warning, not error.
    - If no declarations were parsed (the loop never executed) → S002 `errMissingCharacterDecl` (at least one declaration required before acts).
    - If after declarations, the next token is NOT `TokenWord` with value `act` → S004 `errMissingAct`.
    - Return the slice of declarations.
  - **Edge cases:**
    - Title with no period (just EOF) → S001.
    - Character decl with no comma → S002.
    - Character decl with empty description → S002.
    - Zero character declarations → S002.
    - Characters declared but no `Act` follows → S004 (detected when `parseActs` finds no act keyword).
* **Error Handling:** S001 (missing title), S002 (missing/malformed character decl), S003 (invalid character name — warning), S004 (missing act).
* **Testing & Verification:** TDD table:
  - [ ] `TestParseTitle`: input tokens for `"Hello World."` → `Title{Text: "Hello World"}`.
  - [ ] `TestParseTitleNoPeriod`: tokens `"Hello World"` (no period, just EOF) → S001.
  - [ ] `TestParseCharacterDecls`: tokens for `"Romeo, a young man.\nJuliet, a young woman.\n"` → 2 declarations, `p.characters` has "romeo" and "juliet".
  - [ ] `TestParseCharacterDeclNoComma`: `"Romeo a young man."` (no comma) → S002.
  - [ ] `TestParseCharacterDeclEmptyDesc`: `"Romeo,."` → S002 (empty description).
  - [ ] `TestParseCharacterDeclS003Warning`: `"Mario, a plumber."` → parses successfully, but `p.warnings` contains S003 for "Mario".
  - [ ] `TestParseNoCharacters`: `"Title."` followed by `Act I:` → S002 (no character decls).
  - [ ] Run/fail/implement/pass.
  - [ ] `task fmt && git commit -m "feat(parser): parse title and character declarations"`.
* **Documentation Update:** `PROGRESS.md` Phase 2 — check off title + character decl parsing.

---

### Step 2.6: Parse Acts + Scenes + Roman numerals

* **Goal:** Parse `Act N: Description.` and `Scene N: Description.` headers with Roman numeral validation and sequential ordering. Apply S004–S009.
* **Context Files to Reference:** Reconciled spec "Acts, Scenes", `docs/ERROR_TAXONOMY.md` S004–S009.
* **Implementation Details:** In `parser.go`, add:
  - **Roman numeral parser:** `func parseRoman(s string) (int, bool)` — standard Roman numeral parser. Handles I, V, X (acts/scenes rarely go beyond X). Returns `(value, true)` if valid, `(0, false)` if invalid. Algorithm: map each char to its value, iterate, if a smaller value precedes a larger, subtract; else add. Validate against a known set or pattern.
  - **`parseActs() ([]Act, error)`:**
    - `expectedAct := 1`.
    - Loop while `p.checkWord("act")`:
      - Consume `act` keyword.
      - Expect a WORD token (the Roman numeral). If not a WORD → S005 `errInvalidActNumber(got, line, col)`.
      - `num, ok := parseRoman(roman)`. If `!ok` → S005.
      - If `num != expectedAct` → S006 `errActOrder(expectedRoman, gotRoman, line, col)`.
      - `expectedAct++`.
      - Expect `TokenColon`.
      - Parse act description: collect tokens until `TokenPeriod` (like title). Join with spaces.
      - Expect `TokenPeriod`.
      - `p.skipNewlines()`.
      - Parse scenes → `scenes, err := p.parseScenes(actNum)`.
      - If `len(scenes) == 0` → S007 `errMissingScene(roman, line, col)`.
      - Append `Act{Number: num, RomanNumeral: roman, Description: desc, Scenes: scenes}`.
      - `p.skipNewlines()`.
    - If no acts were parsed → S004 `errMissingAct`.
    - Return the slice.
  - **`parseScenes(actNum int) ([]Scene, error)`:**
    - `expectedScene := 1`.
    - Loop while `p.checkWord("scene")` (and NOT `p.checkWord("act")` — acts end the scene list):
      - Consume `scene` keyword.
      - Expect a WORD (Roman numeral). If not → S008 `errInvalidSceneNumber`.
      - `num, ok := parseRoman(roman)`. If `!ok` → S008.
      - If `num != expectedScene` → S009 `errSceneOrder(expectedRoman, gotRoman, line, col)`.
      - `expectedScene++`.
      - Expect `TokenColon`.
      - Parse scene description (tokens until `TokenPeriod`). 
      - Expect `TokenPeriod`.
      - `p.skipNewlines()`.
      - Parse statements → `stmts, err := p.parseStatements()`. (Statements = stage directions + dialogue. Implemented in Step 2.7+.)
      - Append `Scene{Number: num, RomanNumeral: roman, Description: desc, Statements: stmts}`.
      - `p.skipNewlines()`.
    - Return the slice (may be empty — caller checks S007).
  - **Edge cases:**
    - `Act 1:` (Arabic numeral) → S005 (not Roman).
    - `Act I:` then `Act III:` (skips II) → S006.
    - Act with no scenes → S007.
    - `Scene 1:` → S008.
    - `Scene I:` then `Scene III:` → S009.
    - Scene description is optional? The spec shows `Scene I: Description.` — the description is free text. If no description (just `Scene I:.`), allow it (empty description). `ponytail: empty scene/act descriptions allowed; not worth enforcing`.
* **Error Handling:** S004 (no acts), S005 (invalid act number), S006 (act order), S007 (no scenes), S008 (invalid scene number), S009 (scene order).
* **Testing & Verification:** TDD:
  - [ ] `TestParseRoman`: `parseRoman("I")` == (1, true), `parseRoman("III")` == (3, true), `parseRoman("IV")` == (4, true), `parseRoman("1")` == (0, false), `parseRoman("X")` == (10, true).
  - [ ] `TestParseActs`: tokens for `"Act I: Desc.\nScene I: Desc.\n"` → 1 act with 1 scene.
  - [ ] `TestParseActsS005`: `"Act 1: Desc."` → S005.
  - [ ] `TestParseActsS006`: `"Act I: ...\nAct III: ..."` → S006 (expected II).
  - [ ] `TestParseScenesS007`: `"Act I: Desc."` with no scene → S007.
  - [ ] `TestParseScenesS008`: `"Scene 1:"` → S008.
  - [ ] `TestParseScenesS009`: `"Scene I: ...\nScene III: ..."` → S009.
  - [ ] Run/fail/implement/pass.
  - [ ] `task fmt && git commit -m "feat(parser): parse acts, scenes, and Roman numerals"`.
* **Documentation Update:** `PROGRESS.md` Phase 2 — check off act/scene parsing.

---

### Step 2.7: Parse Stage Directions (Enter / Exit / Exeunt)

* **Goal:** Parse `[Enter ...]`, `[Exit ...]`, `[Exeunt ...]` stage directions. Apply S010, S011, S012 (reconciled), S013.
* **Context Files to Reference:** Reconciled spec "Stage Management", `docs/ERROR_TAXONOMY.md` S010–S013.
* **Implementation Details:** In `parser.go`, add `parseStageDirection() (Statement, error)`:
  - Expect `TokenLBracket`. If not → not a stage direction (caller handles).
  - Read the keyword: `p.advance().Lexeme` (must be a WORD). Lowercase it.
  - Switch on keyword:
    - **`enter`:**
      - Read character names: one or two WORDs separated by `and`.
      - `var chars []string`.
      - Read first WORD (character name). If not a WORD → S011 `errInvalidEnter`.
      - `chars = append(chars, firstWord.Lexeme)`.
      - If `p.matchWord("and")`: read second WORD. If not a WORD → S011. `chars = append(chars, secondWord.Lexeme)`.
      - Expect `TokenRBracket`.
      - Return `EnterStmt{Characters: chars}`.
      - **Note:** Semantic check for max 2 / duplicate names is Phase 3 (M002, M007). Parser just records the names.
    - **`exit`:**
      - Read one WORD (character name). If not a WORD → S011 `errInvalidEnter` (or a new code? S011 covers "expected character name after Enter"; for Exit use a similar message). `ponytail: reuse S011 for both Enter and Exit missing-name errors`.
      - Expect `TokenRBracket`.
      - Return `ExitStmt{Character: name}`.
      - If there's an `and` after the first name → error (Exit takes exactly one). Use S012 `errInvalidExeunt`? No, S012 is for Exeunt. For Exit with too many args, use S011 `errInvalidEnter` with message "Exit takes exactly one character". `ponytail: S011 reused for stage-direction arity errors`.
    - **`exeunt`:**
      - Check if next is `TokenRBracket` (bare `[Exeunt]`) → return `ExeuntStmt{Characters: nil}` (exit all).
      - Else: read one or more character names separated by `and`:
        - Read first WORD. If not → S012 `errInvalidExeunt` (updated: "expected character name after Exeunt").
        - `chars = append(chars, firstWord)`.
        - While `p.matchWord("and")`: read next WORD, append. If not a WORD → S012.
        - Expect `TokenRBracket`.
        - Return `ExeuntStmt{Characters: chars}`.
      - **S012 (reconciled):** fires for malformed Exeunt args (e.g., `[Exeunt ,]` or `[Exeunt and]`), NOT for `[Exeunt A and B]` which is valid.
  - If the keyword is not enter/exit/exeunt → S013 `errMissingStage` or a generic syntax error. `ponytail: unknown bracket keyword → S013`.
  - After the stage direction, `p.skipNewlines()`.
  - **Integration with `parseStatements`:** In `parseStatements()`, if `p.at(TokenLBracket)` → call `parseStageDirection()`.
* **Error Handling:** S010 (character speaks without Enter — detected in dialogue parsing, Step 2.8), S011 (invalid Enter/Exit — missing character name), S012 (invalid Exeunt — malformed args, reconciled), S013 (missing stage — dialogue without prior Enter, detected in Step 2.8).
* **Testing & Verification:** TDD:
  - [ ] `TestParseEnterSingle`: `"[Enter Romeo]"` → `EnterStmt{Characters: ["Romeo"]}`.
  - [ ] `TestParseEnterDouble`: `"[Enter Romeo and Juliet]"` → `EnterStmt{Characters: ["Romeo", "Juliet"]}`.
  - [ ] `TestParseExit`: `"[Exit Romeo]"` → `ExitStmt{Character: "Romeo"}`.
  - [ ] `TestParseExeuntBare`: `"[Exeunt]"` → `ExeuntStmt{Characters: nil}`.
  - [ ] `TestParseExeuntNamed`: `"[Exeunt Ophelia and Hamlet]"` → `ExeuntStmt{Characters: ["Ophelia", "Hamlet"]}` (valid per reconciled spec).
  - [ ] `TestParseExeuntSingle`: `"[Exeunt Romeo]"` → `ExeuntStmt{Characters: ["Romeo"]}` (single name, also valid).
  - [ ] `TestParseEnterS011`: `"[Enter]"` (no name) → S011.
  - [ ] `TestParseExitTooMany`: `"[Exit Romeo and Juliet]"` → S011 (Exit takes one).
  - [ ] `TestParseExeuntMalformed`: `"[Exeunt and]"` → S012.
  - [ ] Run/fail/implement/pass.
  - [ ] `task fmt && git commit -m "feat(parser): parse stage directions Enter/Exit/Exeunt"`.
* **Documentation Update:** `PROGRESS.md` Phase 2 — check off stage direction parsing.

---

### Step 2.8: Parse Dialogue (speaker + statement list)

* **Goal:** Parse `Name:` dialogue blocks: identify the speaker, then parse a sequence of statements until the next speaker/stage-direction/act/scene/EOF. Apply S010, S013, S014.
* **Context Files to Reference:** Reconciled spec "Dialogue Format", `docs/ERROR_TAXONOMY.md` S010/S013/S014.
* **Implementation Details:** In `parser.go`, add:
  - **`parseStatements() ([]Statement, error)`:** (called by `parseScenes` — collects all stage directions + dialogue in a scene)
    - Loop:
      - `p.skipNewlines()`.
      - If `p.isEOF()` → break.
      - If `p.at(TokenLBracket)` → `parseStageDirection()`, append to statements.
      - If `p.at(TokenWord)` and `peekAt(1).Type == TokenColon` and `p.checkWord` is NOT a keyword that starts a statement (like `you`, `speak`, etc.) → it's a speaker line → `parseDialogue()`, append the `Dialogue` node to statements.
        - **Disambiguation:** A WORD followed by `:` is a speaker label (e.g., `Romeo:`). But some statement keywords could theoretically be followed by `:`? In SPL, `:` only appears after character names (speaker labels) and after act/scene numbers. Within a scene, a WORD + `:` = speaker. So: if `peek().Type == TokenWord && peekAt(1).Type == TokenColon` → it's a dialogue speaker. This is unambiguous because no SPL statement starts with `WORD:`. `ponytail: WORD+COLON in a scene is always a speaker label`.
      - If `p.checkWord("act")` or `p.checkWord("scene")` → break (next act/scene header, not part of this scene's statements).
      - Else → break (unexpected token; could be EOF or end of scene).
    - Return the statement slice.
  - **`parseDialogue() (Dialogue, error)`:**
    - `speaker := p.advance().Lexeme` (the WORD).
    - Expect `TokenColon` — consume it.
    - `p.skipNewlines()`.
    - Parse a sequence of statements until the next speaker/stage-direction/act/scene/EOF:
      - Loop: call `parseStatement()` (Step 2.9+). Append to `stmts`.
      - Break when: `p.at(TokenLBracket)`, or `p.at(TokenWord) && peekAt(1).Type == TokenColon` (next speaker), or `p.checkWord("act")`, or `p.checkWord("scene")`, or `p.isEOF()`.
    - **S010/S013 check:** If a dialogue is parsed but no preceding `EnterStmt` was seen in this scene → S013 `errMissingStage`. Track `enterSeen bool` in `parseStatements` (set when an Enter is parsed). If dialogue appears before any Enter → S013. (Full S010 — "character must enter before speaking" — requires knowing WHICH character entered; that's Phase 3 semantic. Parser-level: just check that at least one Enter preceded dialogue → S013.)
    - Return `Dialogue{Speaker: speaker, Statements: stmts, Line, Col}`.
  - **S014:** If a line starts with a non-WORD, non-LBRACKET token (e.g., a stray `.`) within a scene → S014 `errMissingSpeaker`.
* **Error Handling:** S010 (character not on stage — Phase 3), S013 (no Enter before dialogue — parser-level), S014 (missing speaker — malformed line).
* **Testing & Verification:** TDD:
  - [ ] `TestParseDialogue`: tokens for `"Romeo:\nSpeak your mind.\n"` → `Dialogue{Speaker: "Romeo", Statements: [SpeakStmt]}`.
  - [ ] `TestParseDialogueMultipleStatements`: `"Hamlet:\nSpeak your mind!\nYou are as good as a flower.\n"` → 2 statements.
  - [ ] `TestParseDialogueS013`: dialogue with no prior Enter → S013.
  - [ ] `TestParseMultipleSpeakers`: `"Romeo:\nSpeak your mind.\nJuliet:\nSpeak your mind.\n"` → 2 Dialogue nodes.
  - [ ] `TestParseDialogueEndsAtStageDirection`: `"Romeo:\nSpeak your mind.\n[Exit Romeo]"` → Dialogue + EnterStmt/ExitStmt.
  - [ ] Run/fail/implement/pass.
  - [ ] `task fmt && git commit -m "feat(parser): parse dialogue blocks with speaker labels"`.
* **Documentation Update:** `PROGRESS.md` Phase 2 — check off dialogue parsing.

---

### Step 2.9: Parse Expressions — constants, CharRef, Pronoun

* **Goal:** Parse the leaf expressions: constants (article* adjective* noun), character references, and pronouns. These are the atoms of the expression grammar.
* **Context Files to Reference:** Reconciled spec "Constant Expressions" + vetted pronoun section, `docs/ERROR_TAXONOMY.md` S015.
* **Implementation Details:** In `parser.go`, add `parseExpr() (Expr, error)` (the expression dispatcher — operator forms are added in Steps 2.10–2.11; this step implements the leaf fallback):
  - **`parseExpr()` dispatch order:**
    1. If `p.checkWord("the")` → binary/unary op (Step 2.10–2.11). For now, not implemented; will be added.
    2. If `p.checkWord("twice")` → unary op (Step 2.11). For now, not implemented.
    3. If `p.checkWord("as")` → simile (Step 2.12). For now, not implemented.
    4. If `p.peek().Type == TokenWord` and `isSpeakerPronoun(lower(peek().Lexeme))` → `PronounExpr{Ref: "speaker"}`, advance.
    5. If `p.peek().Type == TokenWord` and `isListenerPronoun(lower(peek().Lexeme))` → `PronounExpr{Ref: "listener"}`, advance.
    6. If `p.peek().Type == TokenWord` and `p.characters[lower(peek().Lexeme)]` → `CharRefExpr{Name: peek().Lexeme}`, advance. (Declared character name.)
    7. Else → `parseConstant()`.
  - **`parseConstant() (Expr, error)`:**
    - `adjCount := 0`.
    - `noun := ""`, `polarity := 1`, `hasNoun := false`.
    - Loop:
      - `word := strings.ToLower(p.peek().Lexeme)`.
      - If `p.peek().Type != TokenWord` → break (end of constant).
      - If `word` is an article (`a`, `an`, `the`) → `advance()`, continue. (Note: `the` as an article in constants is rare but valid; in practice `the` starts operations and is caught by step 1. But if `parseConstant` is reached with `the`, it's an article here.) Actually, `the` should NOT be treated as an article in `parseConstant` because `the` is the operator prefix. If `parseExpr` dispatches to `parseConstant` with `the` as the first word, something is wrong. `ponytail: treat 'the' as article in parseConstant; the expression dispatcher catches 'the' before reaching here, so this is defensive`.
      - If `word` is a possessive pronoun (`my`, `thy`, `your`, `his`, `her`, `mine`, `thine`) → `advance()`, continue (skip like articles).
      - If `isAdjective(word)` → `adjCount++`, `advance()`, continue.
      - Else (not article, not possessive, not adjective) → this is the noun. `noun = p.advance().Lexeme`. `polarity = nounPolarity(word)`. `hasNoun = true`. Break.
    - If `!hasNoun` → S015 `errInvalidExpression` (expected a noun).
    - Return `ConstExpr{AdjectiveCount: adjCount, Noun: noun, Polarity: polarity, Line, Col}`.
    - **Behavior notes:**
      - `"flower"` → `ConstExpr{0, "flower", +1}`.
      - `"red flower"` → `ConstExpr{1, "flower", +1}` (value = 2).
      - `"red hot flower"` → `ConstExpr{2, "flower", +1}` (value = 4).
      - `"coward"` → `ConstExpr{0, "coward", -1}` (value = -1).
      - `"big coward"` → `ConstExpr{1, "coward", -1}` (value = -2).
      - `"a vile coward"` → `ConstExpr{1, "coward", -1}` (value = -2; `a` skipped, `vile` is 1 adjective, `coward` is -1 noun).
      - `"summer's day"` → `ConstExpr{1, "day", +1}` (value = 2; `summer's` is an adjective, `day` is +1 noun).
      - `"my father"` → `ConstExpr{0, "father", +1}` (`my` skipped as possessive, `father` is noun).
      - `"chrysanthemum"` (unknown) → `ConstExpr{0, "chrysanthemum", +1}` (unknown noun → neutral +1).
* **Error Handling:** S015 (invalid expression — no noun found).
* **Testing & Verification:** TDD:
  - [ ] `TestParseConstantSimple`: `"flower"` → `ConstExpr{0, "flower", 1}`.
  - [ ] `TestParseConstantWithAdjectives`: `"red hot flower"` → `ConstExpr{2, "flower", 1}`.
  - [ ] `TestParseConstantNegativeNoun`: `"coward"` → `ConstExpr{0, "coward", -1}`.
  - [ ] `TestParseConstantWithArticle`: `"a flower"` → `ConstExpr{0, "flower", 1}`.
  - [ ] `TestParseConstantVileCoward`: `"a vile coward"` → `ConstExpr{1, "coward", -1}` (value -2, confirming the corrected arithmetic).
  - [ ] `TestParseConstantSummersDay`: `"summer's day"` → `ConstExpr{1, "day", 1}`.
  - [ ] `TestParseConstantUnknownNoun`: `"chrysanthemum"` → `ConstExpr{0, "chrysanthemum", 1}`.
  - [ ] `TestParseCharRef`: with `"Romeo"` declared → `CharRefExpr{Name: "Romeo"}`.
  - [ ] `TestParsePronounSpeaker`: `"me"` → `PronounExpr{Ref: "speaker"}`; `"myself"` → `PronounExpr{Ref: "speaker"}`.
  - [ ] `TestParsePronounListener`: `"thyself"` → `PronounExpr{Ref: "listener"}`; `"yourself"` → `PronounExpr{Ref: "listener"}`.
  - [ ] `TestParseConstantNoNoun`: `"the"` (just an article, no noun) → S015.
  - [ ] Run/fail/implement/pass.
  - [ ] `task fmt && git commit -m "feat(parser): parse constant, CharRef, and Pronoun expressions"`.
* **Documentation Update:** `PROGRESS.md` Phase 2 — check off expression parsing (leaf nodes).

---

### Step 2.10: Parse Expressions — binary operations

* **Goal:** Parse the five binary operations: `sum`, `difference`, `product`, `quotient`, `remainder`. Each has a specific syntactic form with `of`/`between`/`and` keywords. Support nesting.
* **Context Files to Reference:** Reconciled spec "Operations", vetted addendum.
* **Implementation Details:** Extend `parseExpr()` in `parser.go`:
  - At the top of `parseExpr()`, if `p.matchWord("the")`:
    - Check the next word to determine the operation:
    - **`sum`:** `p.matchWord("sum")` → expect `of` → `left := parseExpr()` → expect `and` → `right := parseExpr()` → return `BinaryOpExpr{Op: "sum", Left: left, Right: right}`.
    - **`product`:** `p.matchWord("product")` → expect `of` → `left := parseExpr()` → expect `and` → `right := parseExpr()` → return `BinaryOpExpr{Op: "product", Left, Right}`.
    - **`difference`:** `p.matchWord("difference")` → expect `between` → `left := parseExpr()` → expect `and` → `right := parseExpr()` → return `BinaryOpExpr{Op: "difference", Left, Right}`.
    - **`quotient`:** `p.matchWord("quotient")` → expect `between` → `left := parseExpr()` → expect `and` → `right := parseExpr()` → return `BinaryOpExpr{Op: "quotient", Left, Right}`.
    - **`remainder`:** `p.matchWord("remainder")` → expect `of` → expect `the` → expect `quotient` → expect `between` → `left := parseExpr()` → expect `and` → `right := parseExpr()` → return `BinaryOpExpr{Op: "remainder", Left, Right}`.
      - The full phrase is `the remainder of the quotient between A and B` → A % B.
    - If none of these matched after `the` → fall through to other `the`-prefixed ops (unary, Step 2.11) or S015.
    - **`expect` for keywords:** `expectWord(value string) error` — if `p.checkWord(value)`, `advance()` and return nil; else return S015 `errInvalidExpression`. E.g., after `sum`, expecting `of`: if next word is not `of` → S015.
  - **Nesting:** Since each operand calls `parseExpr()` recursively, nesting is automatic:
    - `"the sum of Romeo and the difference between Juliet and a flower"` → `BinaryOpExpr{sum, CharRefExpr{Romeo}, BinaryOpExpr{difference, CharRefExpr{Juliet}, ConstExpr{0, flower, 1}}}`.
  - **Precedence:** All binary operations have equal precedence in SPL — the syntax is fully explicit (`the sum of A and B`). There is no ambiguity. The recursive descent handles this naturally: each operand is a full `parseExpr()`, so `the sum of A and the product of B and C` → `sum(A, product(B, C))` because the `and` after A binds to the `sum`, and the `product`'s `and` binds to the `product`. Wait — is that right? `the sum of A and the product of B and C` — the parser reads `sum of`, then `parseExpr()` for left → reads `A` (a constant/charref). Then expects `and`. Then `parseExpr()` for right → reads `the product of B and C` (a full product expression). So: `sum(A, product(B, C))`. Correct.
  - **But what about** `the sum of the product of A and B and C`? This is ambiguous in SPL. The parser reads `sum of`, then `parseExpr()` for left → reads `the product of A and B` → `product(A, B)`. Then expects `and` → yes. Then `parseExpr()` for right → `C`. So: `sum(product(A, B), C)`. The first `and` after `A` binds to `product`, the second `and` binds to `sum`. This is the natural left-to-right greedy parse. `ponytail: greedy left-to-right; first 'and' binds to the innermost open operation`.
* **Error Handling:** S015 (invalid expression — missing `of`/`between`/`and`, or missing operand).
* **Testing & Verification:** TDD:
  - [ ] `TestParseSum`: `"the sum of Romeo and a flower"` → `BinaryOpExpr{sum, CharRef{Romeo}, Const{0, flower, 1}}`.
  - [ ] `TestParseDifference`: `"the difference between Romeo and Juliet"` → `BinaryOpExpr{difference, CharRef{Romeo}, CharRef{Juliet}}`.
  - [ ] `TestParseProduct`: `"the product of Romeo and Juliet"` → `BinaryOpExpr{product, ...}`.
  - [ ] `TestParseQuotient`: `"the quotient between Romeo and Juliet"` → `BinaryOpExpr{quotient, ...}`.
  - [ ] `TestParseRemainder`: `"the remainder of the quotient between Romeo and Juliet"` → `BinaryOpExpr{remainder, ...}`.
  - [ ] `TestParseNested`: `"the sum of Romeo and the difference between Juliet and a flower"` → nested BinaryOpExpr.
  - [ ] `TestParseDeeplyNested`: `"the sum of the sum of A and B and C"` → `sum(sum(A, B), C)`.
  - [ ] `TestParseSumMissingOf`: `"the sum Romeo and a flower"` → S015.
  - [ ] `TestParseSumMissingAnd`: `"the sum of Romeo a flower"` → S015.
  - [ ] Run/fail/implement/pass.
  - [ ] `task fmt && git commit -m "feat(parser): parse binary operation expressions"`.
* **Documentation Update:** `PROGRESS.md` Phase 2 — check off binary ops.

---

### Step 2.11: Parse Expressions — unary operations + similes

* **Goal:** Parse the five unary operations (`square`, `cube`, `square root`, `factorial`, `twice`) and the simile prefix (`as <adj> as`).
* **Context Files to Reference:** Reconciled spec vetted addendum (unary ops + similes).
* **Implementation Details:** Extend `parseExpr()` in `parser.go`:
  - After the `the`-prefixed binary ops (Step 2.10), also within the `the` branch:
    - **`square`:** `p.matchWord("square")`:
      - If `p.matchWord("root")` → expect `of` → `operand := parseExpr()` → return `UnaryOpExpr{Op: "square_root", Operand}`.
      - Else → expect `of` → `operand := parseExpr()` → return `UnaryOpExpr{Op: "square", Operand}`.
    - **`cube`:** `p.matchWord("cube")` → expect `of` → `operand := parseExpr()` → return `UnaryOpExpr{Op: "cube", Operand}`.
    - **`factorial`:** `p.matchWord("factorial")` → expect `of` → `operand := parseExpr()` → return `UnaryOpExpr{Op: "factorial", Operand}`.
  - **`twice`** (NOT preceded by `the`): before the `the` check, if `p.matchWord("twice")` → `operand := parseExpr()` → return `UnaryOpExpr{Op: "twice", Operand}`.
  - **Simile** (`as <adj> as <expr>`): before the pronoun/charref/constant checks, if `p.matchWord("as")`:
    - Read the adjective: `adj := p.advance().Lexeme` (must be a WORD; if not → S015). (The adjective is grammatical filler; its value is not used for evaluation. It's stored in the AssignStmt's SimileAdj field when in assignment context, but in pure expression context it's discarded.)
    - Expect `as`: `p.matchWord("as")` — if not → S017 `errInvalidComparative` (or S015). Use S015 since this is expression context.
    - `inner := p.parseExpr()`.
    - Return `inner` (the simile is transparent — it just evaluates to the inner expression). The adjective is discarded in expression context. `ponytail: simile in expression context is transparent; returns the inner expression directly`.
    - **Note:** In assignment context (Step 2.12), the simile adjective IS captured (for the AssignStmt.SimileAdj field). That's handled in the assignment parser, not here. Here, `parseExpr`'s simile just returns the inner expr.
  - **Behavior notes:**
    - `"the square of a flower"` → `UnaryOpExpr{square, ConstExpr{0, flower, 1}}`.
    - `"the square root of a flower"` → `UnaryOpExpr{square_root, ConstExpr{0, flower, 1}}`.
    - `"the cube of a flower"` → `UnaryOpExpr{cube, ...}`.
    - `"the factorial of a flower"` → `UnaryOpExpr{factorial, ...}`.
    - `"twice a flower"` → `UnaryOpExpr{twice, ConstExpr{0, flower, 1}}`.
    - `"twice the sum of Romeo and a flower"` → `UnaryOpExpr{twice, BinaryOpExpr{sum, CharRef{Romeo}, ConstExpr{0, flower, 1}}}`.
    - `"as brave as Hamlet"` (in expression context) → `CharRefExpr{Hamlet}` (simile transparent).
* **Error Handling:** S015 (missing `of`/`as`, missing operand).
* **Testing & Verification:** TDD:
  - [ ] `TestParseSquare`: `"the square of a flower"` → `UnaryOpExpr{square, ...}`.
  - [ ] `TestParseSquareRoot`: `"the square root of a flower"` → `UnaryOpExpr{square_root, ...}`.
  - [ ] `TestParseCube`: `"the cube of a flower"` → `UnaryOpExpr{cube, ...}`.
  - [ ] `TestParseFactorial`: `"the factorial of a flower"` → `UnaryOpExpr{factorial, ...}`.
  - [ ] `TestParseTwice`: `"twice a flower"` → `UnaryOpExpr{twice, ...}`.
  - [ ] `TestParseTwiceNested`: `"twice the sum of Romeo and a flower"` → nested.
  - [ ] `TestParseSimileExpr`: `"as brave as Hamlet"` → `CharRefExpr{Hamlet}` (transparent).
  - [ ] `TestParseSquareMissingOf`: `"the square a flower"` → S015.
  - [ ] Run/fail/implement/pass.
  - [ ] `task fmt && git commit -m "feat(parser): parse unary operations and similes"`.
* **Documentation Update:** `PROGRESS.md` Phase 2 — check off unary ops + similes.

---

### Step 2.12: Parse Assignment statements

* **Goal:** Parse the three assignment forms: `You/Thou are/art [as <adj> as] <expr>.` (copula) and `You/Thou <constant>!` (no copula). Apply S015.
* **Context Files to Reference:** Reconciled spec vetted addendum (no-copula assignment, simile semantics).
* **Implementation Details:** In `parser.go`, add `parseAssignStmt() (Statement, error)`:
  - This is called from `parseStatement()` (Step 2.8's statement dispatcher) when the current word is `you` or `thou` (case-insensitive).
  - Consume `you`/`thou` → `advance()`. (Target is always the listener; normalized to `"you"`.)
  - Check the next word:
    - **Copula path:** if `p.matchWord("are")` or `p.matchWord("art")`:
      - Check for simile: if `p.checkWord("as")`:
        - `p.advance()` (consume `as`).
        - `simileAdj := p.advance().Lexeme` (the adjective WORD).
        - Expect `as`: `if !p.matchWord("as")` → S015.
      - Else `simileAdj = ""`.
      - `expr, err := p.parseExpr()`.
      - Expect a terminator: `.` or `!` (assignments end with `.` or `!`). `terminator := p.advance()` — must be `TokenPeriod` or `TokenBang`. If not → S015.
      - Return `AssignStmt{Target: "you", SimileAdj: simileAdj, Expr: expr, Terminator: terminator.Lexeme}`.
    - **No-copula path:** else (next word is not `are`/`art`):
      - Parse a constant directly: `expr, err := p.parseConstant()`. (Only constants, not full expressions — the no-copula form is for insult/praise constants like `You lying stupid... coward!`.)
      - If `parseConstant` fails → S015.
      - Expect a terminator: `!` or `.`. The no-copula form typically ends with `!` (insult) but `.` is also valid.
      - Return `AssignStmt{Target: "you", SimileAdj: "", Expr: expr, Terminator: terminator.Lexeme}`.
  - **Behavior notes:**
    - `"You are as good as a flower."` → `AssignStmt{Target: "you", SimileAdj: "good", Expr: ConstExpr{0, flower, 1}, Terminator: "."}`.
    - `"You are the sum of Romeo and a flower."` → `AssignStmt{Target: "you", SimileAdj: "", Expr: BinaryOpExpr{sum, ...}, Terminator: "."}`.
    - `"Thou art as sweet as the sum of Romeo and his horse!"` → `AssignStmt{Target: "you", SimileAdj: "sweet", Expr: BinaryOpExpr{sum, ...}, Terminator: "!"}`.
    - `"You lying stupid fatherless big smelly half-witted coward!"` → `AssignStmt{Target: "you", SimileAdj: "", Expr: ConstExpr{6, "coward", -1}, Terminator: "!"}`. (6 adjectives: lying, stupid, fatherless, big, smelly, half-witted; value = -2^6 = -64. Wait — the canonical Hello World's first line assigns -64 to Romeo. Let me verify: 6 adjectives × coward(-1) = -1 × 2^6 = -64. Yes.)
  - **`parseStatement()` dispatch** (the statement dispatcher called by `parseDialogue`):
    - If `p.checkWord("you")` or `p.checkWord("thou")` → `parseAssignStmt()`.
    - (Other statement types are added in subsequent steps.)
* **Error Handling:** S015 (invalid expression — missing copula context, missing expr, missing terminator).
* **Testing & Verification:** TDD:
  - [ ] `TestParseAssignCopula`: `"You are as good as a flower."` → AssignStmt with SimileAdj="good", ConstExpr.
  - [ ] `TestParseAssignCopulaExpr`: `"You are the sum of Romeo and a flower."` → AssignStmt with BinaryOpExpr.
  - [ ] `TestParseAssignThouArt`: `"Thou art as sweet as the sum of Romeo and his horse!"` → AssignStmt with SimileAdj="sweet", BinaryOpExpr, Terminator="!".
  - [ ] `TestParseAssignNoCopula`: `"You lying stupid fatherless big smelly half-witted coward!"` → AssignStmt with ConstExpr{6, "coward", -1}, Terminator="!".
  - [ ] `TestParseAssignNoCopulaSimple`: `"You are as good as a flower."` → confirmed.
  - [ ] `TestParseAssignMissingExpr`: `"You are ."` → S015.
  - [ ] `TestParseAssignMissingTerminator`: `"You are as good as a flower"` (no `.` or `!`) → S015.
  - [ ] Run/fail/implement/pass.
  - [ ] `task fmt && git commit -m "feat(parser): parse assignment statements (copula and no-copula)"`.
* **Documentation Update:** `PROGRESS.md` Phase 2 — check off AssignStmt.

---

### Step 2.13: Parse I/O statements (Speak, Open, Listen)

* **Goal:** Parse the four I/O command phrases: `Speak your/thy mind.` (output char), `Open your/thy heart.` (output number), `Open your/thy mind.` (input char), `Listen to your/thy heart.` (input number).
* **Context Files to Reference:** Reconciled spec "I/O Commands".
* **Implementation Details:** Extend `parseStatement()` dispatch in `parser.go`:
  - **`Speak`:** if `p.matchWord("speak")`:
    - Expect `your` or `thy`: `if !p.matchWord("your") && !p.matchWord("thy")` → S015.
    - Expect `mind`: `if !p.matchWord("mind")` → S015.
    - Expect terminator (`.` or `!`): `terminator := p.advance()` — must be `TokenPeriod` or `TokenBang`.
    - Return `SpeakStmt{Terminator: terminator.Lexeme}`.
  - **`Open`:** if `p.matchWord("open")`:
    - Expect `your` or `thy`: `if !p.matchWord("your") && !p.matchWord("thy")` → S015.
    - Check next word:
      - If `p.matchWord("heart")` → expect terminator (`.` or `!`) → return `OpenHeartStmt{Terminator}`. (Output number.)
      - If `p.matchWord("mind")` → expect terminator (`.` or `!`) → return `OpenMindStmt{Terminator}`. (Input char.)
      - Else → S015.
  - **`Listen`:** if `p.matchWord("listen")`:
    - Expect `to`: `if !p.matchWord("to")` → S015.
    - Expect `your` or `thy`: `if !p.matchWord("your") && !p.matchWord("thy")` → S015.
    - Expect `heart`: `if !p.matchWord("heart")` → S015.
    - Expect terminator (`.` or `!`).
    - Return `ListenStmt{Terminator}`.
  - **Behavior notes:**
    - Case-insensitive: `"Speak YOUR mind!"` matches (YOUR matches `your` via `matchWord`).
    - `"Speak thy mind."` → SpeakStmt.
    - `"Open your heart!"` → OpenHeartStmt.
    - `"Open your mind."` → OpenMindStmt.
    - `"Listen to your heart!"` → ListenStmt.
* **Error Handling:** S015 (malformed I/O command — wrong pronoun, wrong body part, missing terminator).
* **Testing & Verification:** TDD:
  - [ ] `TestParseSpeak`: `"Speak your mind."` → SpeakStmt{Terminator: "."}.
  - [ ] `TestParseSpeakThy`: `"Speak thy mind!"` → SpeakStmt{Terminator: "!"}.
  - [ ] `TestParseSpeakMixedCase`: `"Speak YOUR mind!"` → SpeakStmt (case-insensitive).
  - [ ] `TestParseOpenHeart`: `"Open your heart."` → OpenHeartStmt.
  - [ ] `TestParseOpenMind`: `"Open your mind."` → OpenMindStmt.
  - [ ] `TestParseListen`: `"Listen to your heart!"` → ListenStmt.
  - [ ] `TestParseSpeakWrongPronoun`: `"Speak his mind."` → S015.
  - [ ] `TestParseSpeakWrongWord`: `"Speak your soul."` → S015.
  - [ ] `TestParseListenMissingTo`: `"Listen your heart."` → S015.
  - [ ] Run/fail/implement/pass.
  - [ ] `task fmt && git commit -m "feat(parser): parse I/O statements"`.
* **Documentation Update:** `PROGRESS.md` Phase 2 — check off I/O statements.

---

### Step 2.14: Parse Questions, If-statements, and Goto-statements

* **Goal:** Parse comparative questions (`Am I/Is X <comparative> Y?`), conditional branches (`If so/not, let us proceed/return to scene/act X.`), and unconditional gotos (`Let us proceed/return to scene/act X.`). Apply S016, S017.
* **Context Files to Reference:** Reconciled spec "Control Flow", vetted addendum (7 comparative forms → 6 relations).
* **Implementation Details:** In `parser.go`:
  - **Comparative parsing** — `parseComparative() (Comparative, error)`:
    - `negated := false`.
    - If `p.matchWord("not")` → `negated = true`.
    - Check for `as...as` form: if `p.matchWord("as")`:
      - `adj := p.advance().Lexeme` (the comparative adjective WORD).
      - Expect `as`: `if !p.matchWord("as")` → S017.
      - `pol := comparativePolarity(lower(adj))`.
      - Determine relation:
        - If `!negated && pol == "positive"` → relation = "equal".
        - If `!negated && pol == "negative"` → relation = "not_equal".
        - If `negated && pol == "positive"` → relation = "not_equal".
        - If `negated && pol == "negative"` → relation = "equal". (not as <neg> as → equal. Double negative.)
      - Return `Comparative{Form: "as-as", Adjective: adj, Negated: negated, Relation: relation}`.
    - Else (than form): read the comparative adjective: `adj := p.advance().Lexeme`.
      - Expect `than`: `if !p.matchWord("than")` → S017.
      - `pol := comparativePolarity(lower(adj))`.
      - Determine relation:
        - If `!negated && pol == "positive"` → relation = "greater".
        - If `!negated && pol == "negative"` → relation = "less".
        - If `negated && pol == "positive"` → relation = "less_or_equal".
        - If `negated && pol == "negative"` → relation = "greater_or_equal".
      - Return `Comparative{Form: "than", Adjective: adj, Negated: negated, Relation: relation}`.
    - **The 7 forms → 6 relations table:**
      | # | Form | Polarity | Negated | Relation |
      |---|------|----------|---------|----------|
      | 1 | `as <pos> as` | positive | false | equal |
      | 2 | `as <neg> as` | negative | false | not_equal |
      | 3 | `<pos> than` | positive | false | greater |
      | 4 | `<neg> than` | negative | false | less |
      | 5 | `not as <pos> as` | positive | true | not_equal |
      | 6 | `not <pos> than` | positive | true | less_or_equal |
      | 7 | `not <neg> than` | negative | true | greater_or_equal |
  - **Question** — `parseQuestion() (Statement, error)`:
    - If `p.matchWord("am")`:
      - Expect `I`: `if !p.matchWord("i")` → S017.
      - `comp, err := p.parseComparative()`.
      - Expect `you`/`thou`: `if !p.matchWord("you") && !p.matchWord("thou")` → S017. (The right operand is the listener.)
      - Expect `TokenQuestion` (questions end with `?`).
      - Return `QuestionStmt{Left: PronounExpr{"speaker"}, Comparative: comp, Right: PronounExpr{"listener"}}`.
    - If `p.matchWord("is")`:
      - `left, err := p.parseExpr()` (or parse a character name / pronoun). Actually, `Is X better than Y?` — X and Y are expressions (character names or pronouns). Use `parseExpr()` for both.
      - `comp, err := p.parseComparative()`.
      - `right, err := p.parseExpr()`.
      - Expect `TokenQuestion`.
      - Return `QuestionStmt{Left: left, Comparative: comp, Right: right}`.
  - **If-statement** — `parseIfStmt() (Statement, error)`:
    - Expect `if`: `p.matchWord("if")`.
    - Check `so` or `not`:
      - If `p.matchWord("so")` → `branchIfTrue = true`.
      - Else if `p.matchWord("not")` → `branchIfTrue = false`.
      - Else → S016 `errInvalidIf`.
    - Expect `,` (comma): `if !p.match(TokenComma)` → S016.
    - Expect `let us`: `if !p.matchWord("let")` → S016. `if !p.matchWord("us")` → S016.
    - Expect `proceed` or `return`: `proceed := p.matchWord("proceed")`. If not, `if !p.matchWord("return")` → S016.
    - Expect `to`: `if !p.matchWord("to")` → S016.
    - Expect `scene` or `act`: 
      - If `p.matchWord("scene")` → `targetKind = "scene"`.
      - Else if `p.matchWord("act")` → `targetKind = "act"`.
      - Else → S016.
    - Expect a WORD (Roman numeral target): `if p.peek().Type != TokenWord` → S016 `errInvalidIf`. `target := p.advance().Lexeme`.
    - Expect `TokenPeriod`.
    - Return `IfStmt{BranchIfTrue: branchIfTrue, Target: target, TargetKind: targetKind}`.
  - **Goto-statement** — `parseGotoStmt() (Statement, error)`:
    - Expect `let`: `p.matchWord("let")`.
    - Expect `us`: `if !p.matchWord("us")` → S015.
    - Expect `proceed` or `return` (either is valid for goto).
    - Expect `to`.
    - Expect `scene` or `act`.
    - Expect WORD (Roman numeral).
    - Expect `TokenPeriod`.
    - Return `GotoStmt{Target: target, TargetKind: targetKind}`.
  - **`parseStatement()` dispatch additions:**
    - If `p.checkWord("am")` or `p.checkWord("is")` → `parseQuestion()`.
    - If `p.checkWord("if")` → `parseIfStmt()`.
    - If `p.checkWord("let")` → `parseGotoStmt()`.
  - **Note:** The question and if-statement are separate statements (often spoken by different characters). The runtime (Phase 4) links them: after a `QuestionStmt`, the next statement may be an `IfStmt` that branches based on the question's result. The parser does NOT link them — it just builds both nodes independently.
* **Error Handling:** S016 (invalid If — malformed branch), S017 (invalid Comparative — unrecognized comparative phrase).
* **Testing & Verification:** TDD:
  - [ ] `TestParseComparativeTable`: all 7 forms → correct relations:
    - `"as good as"` → equal (pos, not negated, as-as).
    - `"as bad as"` → not_equal (neg, not negated, as-as).
    - `"better than"` → greater (pos, not negated, than).
    - `"worse than"` → less (neg, not negated, than).
    - `"not as good as"` → not_equal (pos, negated, as-as).
    - `"not better than"` → less_or_equal (pos, negated, than).
    - `"not worse than"` → greater_or_equal (neg, negated, than).
  - [ ] `TestParseQuestionAmI`: `"Am I better than you?"` → QuestionStmt with PronounExpr{speaker}, Comparative{greater}, PronounExpr{listener}.
  - [ ] `TestParseQuestionIs`: `"Is Romeo as good as Juliet?"` → QuestionStmt with CharRef{Romeo}, Comparative{equal}, CharRef{Juliet}.
  - [ ] `TestParseIfSo`: `"If so, let us proceed to scene II."` → IfStmt{BranchIfTrue: true, Target: "II", TargetKind: "scene"}.
  - [ ] `TestParseIfNot`: `"If not, let us proceed to scene I."` → IfStmt{BranchIfTrue: false, Target: "I", TargetKind: "scene"}.
  - [ ] `TestParseGotoProceed`: `"Let us proceed to scene II."` → GotoStmt{Target: "II", TargetKind: "scene"}.
  - [ ] `TestParseGotoReturn`: `"Let us return to scene I."` → GotoStmt{Target: "I", TargetKind: "scene"}.
  - [ ] `TestParseGotoAct`: `"Let us proceed to act II."` → GotoStmt{Target: "II", TargetKind: "act"}.
  - [ ] `TestParseIfS016`: `"If so, let us proceed to scene"` (missing Roman numeral) → S016.
  - [ ] `TestParseComparativeS017`: `"Am I as good you?"` (missing second `as`) → S017.
  - [ ] Run/fail/implement/pass.
  - [ ] `task fmt && git commit -m "feat(parser): parse questions, if-statements, and goto-statements"`.
* **Documentation Update:** `PROGRESS.md` Phase 2 — check off control flow parsing.

---

### Step 2.15: Parse Stack operations (Remember / Recall)

* **Goal:** Parse `Remember <expr>.` (push) and `Recall <ignored text>.` (pop). Apply S018.
* **Context Files to Reference:** Reconciled spec vetted addendum (Stack Operations from Grokipedia).
* **Implementation Details:** Extend `parseStatement()` dispatch in `parser.go`:
  - **`Remember`:** if `p.matchWord("remember")`:
    - `expr, err := p.parseExpr()`. If err → S018 `errInvalidStackOp("expected expression after 'Remember'")`.
    - Expect terminator (`.` or `!`).
    - Return `RememberStmt{Expr: expr}`.
    - **Behavior:** `"Remember me."` → `RememberStmt{Expr: PronounExpr{speaker}}`. `"Remember yourself."` → `RememberStmt{Expr: PronounExpr{listener}}`. `"Remember the sum of Romeo and a flower."` → `RememberStmt{Expr: BinaryOpExpr{...}}`.
  - **`Recall`:** if `p.matchWord("recall")`:
    - Collect all tokens (WORD and any non-period tokens) until `TokenPeriod` into `ignoredText` (joined by spaces). The text is semantically ignored — it's just dramatic filler.
    - Expect `TokenPeriod`.
    - Return `RecallStmt{IgnoredText: ignoredText}`.
    - **Behavior:** `"Recall your tragic fate."` → `RecallStmt{IgnoredText: "your tragic fate"}`. `"Recall."` → `RecallStmt{IgnoredText: ""}` (empty recall text).
* **Error Handling:** S018 (invalid stack op — missing expr after Remember, missing period after Recall).
* **Testing & Verification:** TDD:
  - [ ] `TestParseRememberMe`: `"Remember me."` → `RememberStmt{Expr: PronounExpr{speaker}}`.
  - [ ] `TestParseRememberYourself`: `"Remember yourself."` → `RememberStmt{Expr: PronounExpr{listener}}`.
  - [ ] `TestParseRememberExpr`: `"Remember the sum of Romeo and a flower."` → `RememberStmt{Expr: BinaryOpExpr{...}}`.
  - [ ] `TestParseRecall`: `"Recall your tragic fate."` → `RecallStmt{IgnoredText: "your tragic fate"}`.
  - [ ] `TestParseRecallEmpty`: `"Recall."` → `RecallStmt{IgnoredText: ""}`.
  - [ ] `TestParseRememberS018`: `"Remember ."` (no expr) → S018.
  - [ ] `TestParseRecallNoPeriod`: `"Recall your fate"` (no period) → S018.
  - [ ] Run/fail/implement/pass.
  - [ ] `task fmt && git commit -m "feat(parser): parse Remember and Recall stack operations"`.
* **Documentation Update:** `PROGRESS.md` Phase 2 — check off stack operations.

---

### Step 2.16: Parser fixtures

* **Goal:** Create parser test fixtures under `testdata/parser/` covering the full canonical grammar.
* **Context Files to Reference:** `testdata/lexer/hello.shpl` (canonical Hello World), `testdata/lexer/truth-machine.shpl`.
* **Implementation Details:** Create:
  - `testdata/parser/hello.shpl` — same as `testdata/lexer/hello.shpl` (canonical Hello World). The master parser fixture; exercises no-copula assignment, similes, binary ops, unary (square, cube), `[Exeunt A and B]`, multiple acts/scenes, I/O commands.
  - `testdata/parser/truth-machine.shpl` — same as `testdata/lexer/truth-machine.shpl`. Exercises `Listen to your heart!`, `Open your heart!`, questions (`Am I better than you?`), if-statements (`If so, let us proceed to scene II.`), bare `[Exeunt]`.
  - `testdata/parser/arithmetic.shpl` — a synthetic program exercising all binary ops (sum, difference, product, quotient, remainder), all unary ops (square, cube, square root, factorial, twice), and nested expressions. Example:
    ```
    The Arithmetic Test.

    Romeo, a young man.
    Juliet, a young woman.

    Act I: Arithmetic.
    Scene I: Operations.

    [Enter Romeo and Juliet]

    Juliet:
    You are the sum of a flower and a flower.
    You are the difference between a flower and a cowards.
    You are the product of a flower and a flower.
    You are the quotient between a flower and a flower.
    You are the remainder of the quotient between a flower and a flower.
    You are the square of a flower.
    You are the cube of a flower.
    You are the square root of a flower.
    You are the factorial of a flower.
    You are twice a flower.
    You are the sum of the sum of a flower and a flower and a flower.

    [Exeunt]
    ```
  - `testdata/parser/stack.shpl` — a synthetic program exercising Remember/Recall:
    ```
    The Stack Test.

    Romeo, a young man.
    Juliet, a young woman.

    Act I: Stack.
    Scene I: Push and Pop.

    [Enter Romeo and Juliet]

    Romeo:
    You are as good as a flower.
    Remember me.
    You are as good as a cowards.
    Recall your fate.

    [Exeunt]
    ```
  - `testdata/parser/conditionals.shpl` — a synthetic program exercising all 7 comparative forms:
    ```
    The Conditional Test.

    Romeo, a young man.
    Juliet, a young woman.

    Act I: Conditionals.
    Scene I: All forms.

    [Enter Romeo and Juliet]

    Juliet:
    Am I as good as you?
    Am I as bad as you?
    Am I better than you?
    Am I worse than you?
    Am I not as good as you?
    Am I not better than you?
    Am I not worse than you?

    Romeo:
    If so, let us proceed to scene I.

    [Exeunt]
    ```
  - `testdata/parser/error-*.shpl` — small malformed programs for error tests (one per S-code that's feasible to trigger in isolation): `error-no-title.shpl` (S001), `error-no-chars.shpl` (S002), `error-bad-act-num.shpl` (S005), `error-bad-scene-num.shpl` (S008), `error-no-enter.shpl` (S013), `error-bad-expr.shpl` (S015), `error-bad-if.shpl` (S016), `error-bad-comparative.shpl` (S017), `error-bad-stack.shpl` (S018).
* **Error Handling:** Fixtures are data; error fixtures trigger specific S-codes.
* **Testing & Verification:**
  - [ ] Smoke test: each valid fixture parses without error (`Parse()` returns non-nil Program, nil error).
  - [ ] Smoke test: each error fixture produces the expected S-code.
  - [ ] `task fmt && git commit -m "test(parser): add parser fixtures"`.
* **Documentation Update:** `PROGRESS.md` Phase 2 — check off "Fixtures: testdata/parser/...".

---

### Step 2.17: Table-driven + golden JSON snapshot tests

* **Goal:** Lock parser behavior: unit tables for each statement/expression type, plus golden JSON snapshots of the full AST for each valid fixture.
* **Context Files to Reference:** `testdata/parser/*.shpl`.
* **Implementation Details:** In `internal/parser/parser_test.go`:
  - **`TestMain`**: call `logger.Init(logger.LevelDebug)`.
  - **Unit tests:** table-driven for each `parseXxx` — most already written in Steps 2.5–2.15 (move/consolidate into `parser_test.go`). Each test feeds a token slice to the parser and asserts the returned AST node.
  - **`TestGoldenSnapshots`**: for each valid fixture in `testdata/parser/`:
    - Read the `.shpl` file, lex it (using `lexer.New` + `ScanTokens`), then parse (using `parser.New` + `Parse`).
    - Marshal the resulting `*Program` to JSON via `json.MarshalIndent(program, "", "  ")`.
    - Compare against `testdata/parser/<name>.golden.json`.
    - **Golden file generation:** If the golden file does not exist, write it and fail with `"golden file <path> created; inspect and re-run"`. If it exists, compare. (Same pattern as lexer golden tests.)
    - Normalize before comparison: sort character declarations? No — order matters. Just compare exact JSON.
  - **`TestParseErrors`**: table-driven, one case per error fixture:
    - Read the `.shpl` file, lex + parse, assert the error is a `ParseError` with the expected `Code`.
    - Cases: `{name, fixturePath, expectedCode}`.
  - **Full pipeline test:** `TestParseHelloWorld` — parse `testdata/parser/hello.shpl` and assert the AST has: 1 title, 4 character decls, 2 acts, correct scene counts, at least one no-copula AssignStmt, at least one simile AssignStmt, at least one `[Exeunt Ophelia and Hamlet]` ExeuntStmt with 2 characters.
* **Error Handling:** Tests assert S-code matches expected.
* **Testing & Verification:**
  - [ ] Write `TestMain`, `TestGoldenSnapshots`, `TestParseErrors`, `TestParseHelloWorld`.
  - [ ] Run `go test -race -v ./internal/parser/` → golden JSON files created on first run; inspect them; re-run → all pass.
  - [ ] `task fmt && git add testdata/parser/*.golden.json && git commit -m "test(parser): table-driven and golden JSON snapshot tests"`.
* **Documentation Update:** `PROGRESS.md` Phase 2 — check off "Tests: table-driven, snapshot".

---

### Step 2.18: Minimal `ast` CLI subcommand

* **Goal:** Wire the `ast` subcommand so `shpl ast <file.shpl>` reads a file, lexes + parses, and prints the AST as JSON. Completes the Phase 2 runnable deliverable.
* **Context Files to Reference:** `cmd/shpl/main.go` (from Step 1.8).
* **Implementation Details:** Extend `cmd/shpl/main.go`:
  - Add subcommand `ast`:
    - `Use: "ast <file>"`, `Args: cobra.ExactArgs(1)`.
    - `RunE`: read the file, lex (`lexer.New` + `ScanTokens`), parse (`parser.New` + `Parse`), marshal the `*Program` to `json.MarshalIndent(program, "", "  ")`, print to stdout. On lexer or parser error, print to stderr, exit 1.
  - Add subcommand `tokens` (already from Step 1.8 — just ensure both coexist on the root command).
  - Import `internal/lexer` and `internal/parser` packages.
* **Error Handling:** File-not-found → stderr + exit 1. Lexer/parser error → print the error string, exit 1.
* **Testing & Verification:**
  - [ ] Write `TestASTCommand` in `cmd/shpl/main_test.go`: run `rootCmd` with `["ast", "testdata/parser/minimal.shpl"]` (create a minimal parser fixture or reuse `testdata/parser/truth-machine.shpl`), capture stdout, assert it contains valid JSON with `"Title"` and `"Acts"` fields.
  - [ ] Run `go test -race -run TestASTCommand ./cmd/shpl/` → pass.
  - [ ] Run `go test -race ./...` → all tests pass (full suite).
  - [ ] Manual smoke: `task build && ./bin/shpl.exe ast testdata/parser/hello.shpl | head -20` → see AST JSON.
  - [ ] `task check` → all gates pass (fmt → lint → vuln → test).
  - [ ] `task fmt && git commit -m "feat(cli): add ast subcommand for parser output"`.
* **Documentation Update:** `PROGRESS.md` Phase 2 — check off "CLI subcommand: shpl ast". Mark Phase 2 as complete (pending self-review).

---

## Self-Review Checklist (run after Step 2.18, before declaring complete)

Before declaring Phases 1 & 2 complete, run this checklist:

1. **Spec coverage:** Does every construct in the reconciled "Canonical Grammar (vetted)" section have a corresponding AST node and parser step?
   - Title → Step 2.5 ✓
   - Character decls → Step 2.5 ✓
   - Acts/Scenes/Roman numerals → Step 2.6 ✓
   - Enter/Exit/Exeunt (with optional names) → Step 2.7 ✓
   - Dialogue (speaker + statements) → Step 2.8 ✓
   - Constants (article* adj* noun) → Step 2.9 ✓
   - CharRef → Step 2.9 ✓
   - Pronouns (speaker/listener/possessive) → Step 2.9 ✓
   - Binary ops (sum/difference/product/quotient/remainder) → Step 2.10 ✓
   - Unary ops (square/cube/square root/factorial/twice) → Step 2.11 ✓
   - Similes (as adj as) → Step 2.11 ✓
   - Assignment (copula + no-copula) → Step 2.12 ✓
   - I/O (Speak/Open heart/Open mind/Listen) → Step 2.13 ✓
   - Questions (Am I / Is X) + comparatives (7 forms) → Step 2.14 ✓
   - If-statements (If so/not) → Step 2.14 ✓
   - Goto (proceed/return) → Step 2.14 ✓
   - Stack ops (Remember/Recall) → Step 2.15 ✓
2. **Error code coverage:** Does every S-code in `error_taxonomy.md` have a corresponding `errXxx` function and at least one test?
   - S001–S018: check each.
3. **Placeholder scan:** Search the plan for "TBD", "TODO", "implement later", "fill in". Fix any.
4. **Type consistency:** Do the AST struct field names match across steps? (e.g., `AssignStmt.Target` not `AssignStmt.Listener`).
5. **Fixture coverage:** Does `testdata/parser/hello.shpl` parse successfully end-to-end? (The canonical Hello World is the integration test.)
6. **Gate:** `task check` passes clean (fmt → lint → vuln → test, all green).

---

## Execution Handoff

After saving the plan and the user approves, offer execution choice:

**"Plan complete and saved to `docs/superpowers/plans/2026-07-09-lexer-parser.md`. Two execution options:**

**1. Subagent-Driven (recommended)** — I dispatch a fresh subagent per step, review between steps, fast iteration.

**2. Inline Execution** — Execute steps in this session using `executing-plans`, batch execution with checkpoints.

**Which approach?"**

---

## Notes for the Implementer

- **Read `docs/SPL_SPECIFICATION.md` (including the vetted addendum from Step 1.0) before starting any parser step.** The original prose has known errors; the vetted addendum is authoritative.
- **Read `docs/ERROR_TAXONOMY.md` (reconciled in Step 1.0) before implementing any error.** S012 and L001 have been corrected; S017 and S018 are new.
- **The canonical Hello World fixture (`testdata/lexer/hello.shpl` / `testdata/parser/hello.shpl`) is the integration test.** If it doesn't lex + parse successfully end-to-end, Phases 1 & 2 are not complete.
- **Every step ends with a commit.** Small, atomic, Conventional Commits. Don't batch multiple steps into one commit.
- **Every step ends with `task fmt` at minimum.** Run `task check` before committing (the pre-commit hook runs it anyway).
- **Update `PROGRESS.md` after each step.** Check off the item, note any decisions or deviations. This guards against context loss.
- **The lexer and parser are in `internal/` — they cannot be imported by external packages.** Only `cmd/shpl/` imports them. This enforces unidirectional dependencies.
- **Ponytail mode is active.** Prefer the standard library over custom code. Prefer one line over fifty. Prefer deletion over addition. Mark deliberate simplifications with `ponytail:` comments. But never simplify away correctness — the canonical Hello World must parse.
