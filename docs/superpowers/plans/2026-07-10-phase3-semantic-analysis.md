# Phase 3: Semantic Analysis — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use `superpowers:subagent-driven-development` (recommended) or `superpowers:executing-plans` to implement this plan task-by-task. Steps use `- [ ]` checkbox syntax for tracking.

**Goal:** Add a static validation pass (`internal/semantic`) that walks the parsed AST and enforces SPL's *meaning* rules (the M001–M008 error codes from `docs/ERROR_TAXONOMY.md`), producing a symbol table + act/scene registries and a collected list of semantic errors. Gate before Phase 4 (runtime).

**Architecture:** A single `Analyzer` struct holds mutable stage state and immutable registries. It walks the AST via a **two-level type-switch dispatch** (idiomatic Go; no formal Visitor interface — single consumer, so an interface would be YAGNI). Level 1 dispatches `Scene.Statements` (stage directions + dialogues); level 2 dispatches `Dialogue.Statements` (assignments/IO/questions/gotos) carrying a `{speaker, listener}` context. Errors are **collected, not fail-fast** (linter-style), so one run reports all M-codes.

**Tech Stack:** Go 1.26.5, `log/slog` (shared `internal/logger`), table-driven tests + `testdata/semantic/*.shpl` fixtures. No new dependencies.

## Global Constraints

- Go 1.26.5; gate is `task check` (fmt → lint → vuln → test). Lint set: `errcheck, govet, ineffassign, staticcheck, unused, bodyclose` (`.golangci.yaml` v2).
- Tests always pass `-race` (suite convention). Mirror `internal/parser/parser_test.go`: a `TestMain` that calls `logger.Init(logger.LevelDebug)`.
- Character-name lookups are **case-insensitive**: key by `lower(name)`, preserve original lexeme for display (parser convention D2).
- Error format (from `docs/ERROR_TAXONOMY.md`): `error[<CODE>]: <message>\n  --> <file>:<line>:<col>`. Default `<file>` = `input` to match `parser.ParseError`; allow override via an `Analyzer` filename field.
- **Boundary:** Semantic analysis owns **M-codes only**. Do NOT re-check S-codes the parser already owns (act/scene ordering S006/S009, missing scene S007, missing stage S010/S013, etc.). Two exceptions are defensive assertions explicitly flagged below (M008 overlaps S007; M002 is parser-precluded).
- All code lives under `internal/semantic/`. Testdata under `testdata/semantic/`.

## File Structure

```
internal/semantic/
├── errors.go           — SemanticError struct + M001–M008 constructors
├── symbol_table.go     — SymbolTable, ActRegistry, per-act SceneRegistry
├── stage.go            — Stage state manager (Enter/Exit/Exeunt, listener resolution)
├── analyzer.go         — Analyzer struct, New(), Analyze(), type-switch dispatch, Result
├── analyzer_test.go    — table-driven + fixture tests (TestMain with logger.Init)

testdata/semantic/
├── hello.shpl                     — canonical Hello World (zero errors)
├── truth-machine.shpl             — canonical Truth Machine (zero errors)
├── minimal-valid.shpl             — minimal valid program (zero errors)
├── m001-undeclared-enter.shpl     — [Enter Banquo] undeclared → M001
├── m001-undeclared-speaker.shpl   — speaker not in dramatis personae → M001
├── m003-stage-overflow.shpl       — [Enter A and B] then [Enter C] → M003
├── m004-speaker-not-on-stage.shpl — speaker speaks after exiting → M004
├── m004-empty-stage-cross-act.shpl— Act I ends [Exeunt], Act II dialogue no Enter → M004
├── m005-exit-not-on-stage.shpl    — [Exit B] where B not on stage → M005
├── m006-undefined-scene.shpl      — goto scene V that doesn't exist → M006
├── m006-undefined-act.shpl        — goto act IX that doesn't exist → M006
├── m007-self-enter.shpl           — [Enter Romeo and Romeo] → M007
├── self-talk.shpl                 — single character on stage speaks to self (zero errors)
├── off-stage-value-read.shpl      — value read of Romeo while Romeo off stage (zero errors, D2)
├── goto-cross-act-scene.shpl      — per-act scene scoping (zero errors + error variant)
├── primes-persistence.shpl        — cross-act carryover (zero errors, D1)
├── multiple-errors.shpl           — M001 + M003 + M006 collected in one pass
└── valid-operations.shpl          — exercises all ops + comparative relations (zero errors)
```

## Decision Points (agreed after Q&A)

| # | Decision | Value | Evidence / Rationale |
|---|----------|-------|---------------------|
| **D1** | Stage does **NOT** clear at Act boundaries | `stage` persists across acts, mutated **only** by `Enter`/`Exit`/`Exeunt` | `primes.spl` (official zmbc/shakespearelang): Act I ends with {Juliet, Romeo} and no `[Exeunt]`; Act II Scene I opens with no `[Enter]` — Juliet speaks immediately. **REVERSED from the earlier recommendation.** |
| **D2** | Off-stage character value reads are **ALLOWED** | `CharRefExpr` in an expression does **not** require on-stage presence; M004 applies only to `Dialogue.Speaker` and listener resolution | Canonical Hello World: Act I Scene II/III reads `Romeo`'s value after `[Exit Romeo]`. Also primes.spl references "the Ghost" after `[Exit the Ghost]`. |
| **D3** | M008 (NoSceneInAct) is unreachable on parser-produced AST | Include as cheap defensive assert; note S007 overlap | Parser already enforces S007 (`errMissingScene`) at `parser.go:219,238`. |
| **D4** | Collect-all errors, continue best-effort | Return `[]SemanticError` — guard against panics after errors | Modern static analyzers (Go vet, Rust/TS type-checkers) all collect-all. Matches multiple-errors test design. |
| **D5** | Type-switch dispatch (not formal Visitor interface) | Idiomatic Go; single consumer → YAGNI | Parser uses raw type switches internally. |
| **D6** | Re-entering an already-on-stage character → M007 error | Treat entry-of-already-present as `cannot enter 'X': already on stage` | Also catches same-name duplicate within one `[Enter]`. Minor; could soften to no-op later. |
| **D7** | Annotated AST deferred | `Result` carries symbols + registries + errors only | Runtime (Phase 4) re-derives listener from stage at execution time. Pronoun→name annotation is YAGNI. |

---

## Phase 2 Pre-requisite: fix `enterSeen` act-boundary reset (Step 3.0)

### Discovery: Case B — real Phase 2 bug

`internal/parser/parser.go` resets `enterSeen` to `false` at the start of **every Act** AND requires `enterSeen == true` to accept a `Dialogue` (else S013). This rejects valid canonical programs where stage state carries across an act boundary with no new `[Enter]`.

**The code** (`internal/parser/parser.go`):

Act-loop calls (two branches, both reset per act):
```go
scenes, enterSeen, err := p.parseScenes(roman, false)   // line ~214 — reset per act
...
_ = enterSeen                                             // line ~221 — discarded
...
scenes, _, err := p.parseScenes(roman, false)            // line ~233 — reset per act, discarded
```

Dialogue acceptance gate (`parseStatements`):
```go
if _, ok := stmt.(EnterStmt); ok {
    enterSeen = true
}
...
if p.peek().Type == lexer.TokenWord && p.peekAt(1).Type == lexer.TokenColon {
    if !enterSeen {
        return nil, false, errMissingStage(p.peek().Line, p.peek().Col)  // S013, line ~327
    }
    dlg, err := p.parseDialogue()
```

`errMissingStage` = S013 `"expected [Enter ...] before dialogue"`.

### Independent verification: primes.spl (official zmbc/shakespearelang)

Source: `shakespearelang/tests/sample_plays/primes.spl` at `zmbc/shakespearelang` (the official Python interpreter linked from `shakespearelang.com`, the canonical modern SPL reference).

**Fact (a)** — Act I ends with Juliet and Romeo on stage, **no closing Exeunt**:

Consecutive lines at the Act I/II boundary:
```
[Enter Romeo]                          <- stage: {Juliet, Romeo}

Juliet:
 Thou art as sweet as a sunny summer's day!


                    Act II: Determining divisibility.
```

Stage trace: `[Enter the Ghost and Juliet]` → `[Exit the Ghost]` leaves {Juliet} → `[Enter Romeo]` gives {Juliet, Romeo} → Act I ends with Juliet's line, then `Act II:` — **no `[Exeunt]`** between.

**Fact (b)** — Act II Scene I opens with dialogue and **no `[Enter]`** before Juliet speaks:

```
                    Act II: Determining divisibility.

                    Scene I: A private conversation.

Juliet:
 Art thou more cunning than the Ghost?
```

No `[Enter ...]` between the scene header and Juliet's speech.

**Both facts confirmed independently from the authoritative reference source.**

### Complication: "The Ghost" multi-word name

`primes.spl` declares **`The Ghost`** — a two-word character name. Our v1 parser takes the first `WORD` as the character name (`parser.go:157`) then expects a comma. `"The"` would become the name, and the subsequent `"Ghost"` would miss the comma → **S002**. Therefore `primes.spl` cannot be used verbatim as a regression fixture. The fixture below is an **adapted excerpt** that reproduces *exactly* the cross-act stage-carryover structure being fixed, but with single-word names our parser can handle.

### The fix: thread enterSeen across the whole program

**File:** `internal/parser/parser.go`  
**Function:** `parseActs`

**Before** (two branches both reset and discard):
```go
	expected := 1
	for p.checkWord("act") {
		// ... period-branch:
			scenes, enterSeen, err := p.parseScenes(roman, false)
			// ...
			_ = enterSeen
			acts = append(acts, Act{...})
			continue
		// ... colon-branch:
		scenes, _, err := p.parseScenes(roman, false)
		// ...
	}
```

**After** (hoist once, thread + capture in both branches):
```go
	expected := 1
	enterSeen := false // threaded across acts: stage entry persists across act boundaries
	for p.checkWord("act") {
		// ... period-branch:
			scenes, newSeen, err := p.parseScenes(roman, enterSeen)
			if err != nil { return nil, err }
			enterSeen = newSeen
			// ... (delete the `_ = enterSeen` line)
			acts = append(acts, Act{...})
			continue
		// ... colon-branch:
		scenes, newSeen, err := p.parseScenes(roman, enterSeen)
		if err != nil { return nil, err }
		enterSeen = newSeen
		// ...
	}
```

The change only ever flips `enterSeen` false→true across acts — it **cannot** make a currently-passing program fail, only admit previously-rejected ones.

### Regression fixture: `testdata/parser/cross-act-persistence.shpl`

Adapted excerpt from primes.spl with single-word names:

```
Cross Act Stage Persistence.

Romeo, a young man.
Juliet, a young woman.

Act I: The first act.
Scene I: The opening.

[Enter Romeo and Juliet]

Juliet:
 You are a flower.

Act II: The second act.
Scene I: The carrying over.

Juliet:
 You are a flower.

[Exeunt]
```

**Expected behavior:**
- Pre-fix → `errMissingStage` (S013) at Act II Scene I (`enterSeen` reset to `false` per act)
- Post-fix → parses successfully, `Acts == 2`, Act II has 1 scene

### Regression test (in `internal/parser/parser_test.go`)

Parse the fixture above — must use a test helper that asserts **no error** (`parseTokens` pattern). This will fail pre-fix because parse returns `err`, and pass post-fix. The test acts as an isolation guard against future regressions.

### Safety check (no existing test breaks)

Grepped `testdata/**` for `Act II`: only `testdata/lexer/hello.shpl` is multi-act, and its Act II opens with its own `[Enter Romeo and Juliet]` — so the fix leaves its parse unchanged. `error-no-enter.shpl` is single-act with no prior Enter, so S013 still fires there (enterSeen stays `false` the whole time). Run the full `go test -race ./internal/parser/...` and `task check` after applying the fix to confirm.

### Documentation update

Add a "Phase 2 bugfix" subsection to `PROGRESS.md` under Phase 2 "Decisions" recording:
- The bug: `enterSeen` reset per act → S013 on valid cross-act carryover programs (primes.spl)
- The fix: thread `enterSeen` across the whole program (not per act)
- The primes.spl verification source

---

## Phase 3 Steps

### Step 3.1: Package scaffold + `SemanticError` type + M-code constructors

* **Goal:** Create `internal/semantic/errors.go` with the error type and all eight M-code constructors, matching the taxonomy format exactly.
* **Context Files to Reference:** `docs/ERROR_TAXONOMY.md` (Level 3 table — M001–M008 names, example messages); `internal/parser/errors.go` (style to mirror — `ParseError` struct, `Error()` format, `errXxx` constructor helpers).
* **Implementation Details:**

  Define struct:
  ```go
  type SemanticError struct {
      Code, Msg string
      Line, Col  int
      Filename   string
  }
  func (e SemanticError) Error() string
  ```
  `Error()` mirrors `parser.ParseError.Error()` but uses `e.Filename` (default `"input"` if empty).

  Constructor helpers (signatures only — bodies are one-line struct returns):
  ```go
  func errUndefinedCharacter(name string, l, c int) SemanticError           // M001
  func errTooManyOnStage(got int, l, c int) SemanticError                   // M002
  func errStageOverflow(name string, onStage []string, l, c int) SemanticError // M003
  func errCharacterNotOnStage(name, role string, l, c int) SemanticError    // M004  role="speaker"|"listener"
  func errExitNotOnStage(name string, l, c int) SemanticError               // M005
  func errUndefinedScene(target, kind, ctx string, l, c int) SemanticError  // M006  ctx=current act roman
  func errSelfReferenceEnter(name string, l, c int) SemanticError           // M007
  func errNoSceneInAct(actRoman string, l, c int) SemanticError             // M008
  ```
  Message text must match the taxonomy column exactly (e.g., M001 → `character 'Banquo' is not declared`; M003 → `cannot enter: stage is full (Romeo, Juliet already on stage)`).

* **Error Handling:** N/A (this step *defines* the error vocabulary).
* **Testing:**
  - [ ] Table-driven `TestSemanticErrorFormat` mapping each M-code constructor to its expected formatted string.
  - [ ] Run `go test -race -run TestSemanticErrorFormat ./internal/semantic/...` → PASS.
  - [ ] Run `task lint` → no findings.
* **Documentation Update:** Add entry under "Phase 3 — Semantic Analysis" in `PROGRESS.md`.

---

### Step 3.2: Symbol Table & registries (Dramatis Personae + act/scene indexes)

* **Goal:** Build the immutable lookup tables the rest of the pass queries: declared characters (for M001), a global act registry, and a per-act scene registry (for M006).
* **Context Files to Reference:** `docs/SPL_SPECIFICATION.md` (Character Declarations, Acts, Scenes); `internal/parser/ast.go` (Program, CharacterDecl, Act, Scene field shapes).
* **Implementation Details:**

  ```go
  type CharacterSymbol struct {
      Name, Description string
      Line, Col int
  }

  type SymbolTable struct {
      chars map[string]CharacterSymbol // keyed by lower(Name)
  }

  func newSymbolTable(decls []parser.CharacterDecl) SymbolTable
  func (s SymbolTable) Has(lowerName string) bool
  func (s SymbolTable) Get(lowerName string) (CharacterSymbol, bool)
  ```

  `newSymbolTable` algorithm: iterate `decls`, insert `lower(d.Name) → CharacterSymbol{...}`; log each insertion via `slog.Debug("symbol inserted", "name", d.Name, "line", d.Line)`. Duplicate declaration (same lowercased name) → flag as M001 "character 'X' declared twice" — records error and continues (skip the duplicate).

  ```go
  type ActRegistry map[string]*parser.Act           // keyed by lower(RomanNumeral)
  type SceneRegistry map[string]*parser.Scene        // keyed by lower(RomanNumeral), built per act

  func buildActRegistry(acts []parser.Act) ActRegistry
  func buildSceneRegistry(act *parser.Act) SceneRegistry
  func (r ActRegistry) Resolve(lowerRoman string) (*parser.Act, bool)
  func (r SceneRegistry) Resolve(lowerRoman string) (*parser.Scene, bool)
  ```

* **Error Handling:** No M-code emitted here (pure construction). Duplicate-declaration handling emits M001 inline.
* **Testing:**
  - [ ] `TestNewSymbolTable`: feed `[]CharacterDecl` with mixed case (Romeo, JULIET, Hamlet) → assert Has/Get works case-insensitively, returns original-case Name.
  - [ ] `TestBuildRegistries`: 2-act, 3-scene program (Act I: I,II; Act II: I,III) → ActRegistry resolves "ii" to Act II; Act I's SceneRegistry resolves "ii" to its Scene II but Act II's does **not** have "ii" (proves per-act scoping).
  - [ ] Run `go test -race ./internal/semantic/...` → PASS.
* **Documentation Update:** `PROGRESS.md` bullet.

---

### Step 3.3: M001 — UndefinedCharacter (Enter/Exit/Exeunt/Speaker must be declared)

* **Goal:** Verify every character *named* in stage directions and every dialogue *speaker* was declared in the Dramatis Personae. The parser does **not** check this for these fields (raw lexemes stored).
* **Context Files to Reference:** `internal/parser/ast.go` (EnterStmt.Characters, ExitStmt.Character, ExeuntStmt.Characters, Dialogue.Speaker are all raw string/[]string); `docs/ERROR_TAXONOMY.md` (M001).
* **Implementation Details:**

  ```go
  func (a *Analyzer) checkDeclared(name string, line, col int) *SemanticError
  ```
  Algorithm: `if !a.symbols.Has(lower(name)) { return &errUndefinedCharacter(name, line, col) }`. Log lookup: `slog.Debug("symbol lookup", "name", name, "found", ok)`.

  Call sites in the dispatch:
  - `EnterStmt`: for each `c in Characters` → `checkDeclared(c, ...)`.
  - `ExitStmt`: `checkDeclared(Character, ...)`.
  - `ExeuntStmt`: for each `c in Characters` (skip when `nil` = exit all).
  - `Dialogue`: `checkDeclared(Speaker, ...)`. An undeclared speaker is M001.

* **Error Handling:** M001 with original-case name. Collect, continue.
* **Testing:**
  - [ ] Unit `TestCheckDeclared`: SymbolTable with Romeo/Juliet; checkDeclared("Romeo") → nil; checkDeclared("Banquo") → M001.
  - [ ] Fixture `m001-undeclared-enter.shpl`: declares Romeo only; `[Enter Banquo]` → M001.
  - [ ] Fixture `m001-undeclared-speaker.shpl`: declares Romeo/Juliet; `[Enter Romeo and Juliet]`; speaker `Mercutio` → M001.
  - [ ] Run `go test -race -run TestM001 ./internal/semantic/...` → PASS.
* **Documentation Update:** `PROGRESS.md` bullet.

---

### Step 3.4: Stage manager + Enter semantics (M007, M002, M003)

* **Goal:** Implement the mutable stage state and the `Enter` mutation with all overflow/duplication rules. This is the heart of "max 2 on stage".
* **Context Files to Reference:** `docs/SPL_SPECIFICATION.md` (Stage Management — max two; Enter single/two); `docs/ERROR_TAXONOMY.md` (M002, M003, M007); `internal/parser/ast.go` (EnterStmt).
* **Implementation Details:**

  ```go
  type Stage struct {
      names []string // len 0..2, insertion order
  }

  func (s *Stage) Clear()
  func (s *Stage) Has(name string) bool               // case-insensitive
  func (s *Stage) Size() int
  func (s *Stage) OnStage() []string                  // snapshot (original case)
  func (s *Stage) Enter(chars []string, syms SymbolTable, line, col int) []SemanticError
  func (s *Stage) Exit(name string, syms SymbolTable, line, col int) []SemanticError
  func (s *Stage) Exeunt(chars []string, syms SymbolTable, line, col int) []SemanticError
  func (s *Stage) Listener(speaker string) (string, bool) // "other" on stage or self-if-alone
  ```

  `Enter` algorithm (order — cheapest structural first, then mutations):
  1. If `len(chars) > 2` → emit **M002** (`too many characters on stage (max 2), got N`). Defensive: parser's `parseEnter` caps at 2.
  2. For each `c` in `chars`: `checkDeclared(c)` → M001 if not.
  3. Same-enter duplicate: for each `c`, if it appears earlier in this slice → **M007** (`cannot enter the same character twice`).
  4. Already-on-stage: for each `c`, if `s.Has(c)` → **M007** (`cannot enter 'X': already on stage`) per D6.
  5. Overflow: if `s.Size() + len(chars) > 2` → **M003** (`cannot enter: stage is full (<current names> already on stage)`).
  6. Only if no error for a given `c`: append `c` to `s.names`. Best-effort: still mutate non-offending names.
  7. Log `slog.Debug("stage enter", "chars", chars, "size", s.Size())`.

* **Error Handling:** M001, M002, M003, M007 as above. Return all errors from this one Enter.
* **Testing:**
  - [ ] `TestStageEnter`: table — `{[Romeo]}`→size1; `{[Romeo,Juliet]}`→size2; `{[Romeo,Romeo]}`→M007; `{[Romeo,Juliet,Hamlet]}`→M002 (hand-built slice); pre-fill stage `[A,B]` then `Enter[C]`→M003; pre-fill `[Romeo]` then `Enter[Romeo]`→M007-already-on-stage.
  - [ ] Fixture `m007-self-enter.shpl`: `[Enter Romeo and Romeo]` → M007.
  - [ ] Fixture `m003-stage-overflow.shpl`: `[Enter Romeo and Juliet]` then `[Enter Hamlet]` → M003.
  - [ ] Run `go test -race -run TestStageEnter ./internal/semantic/...` → PASS.
* **Documentation Update:** `PROGRESS.md` bullet.

---

### Step 3.5: Exit/Exeunt semantics (M005) — stage persists across acts (D1)

* **Goal:** Implement character *removal* and confirm the stage does **not** auto-clear at act boundaries (D1 — only `Exit`/`Exeunt` mutate stage).
* **Context Files to Reference:** `docs/SPL_SPECIFICATION.md` (Exit Romeo, [Exeunt] exits all, [Exeunt A and B] exits named); `docs/ERROR_TAXONOMY.md` (M005); `internal/parser/ast.go` (ExitStmt, ExeuntStmt where `nil` = exit all).
* **Implementation Details:**

  `Exit` algorithm:
  1. `checkDeclared(Character)` → M001 if not.
  2. If `!s.Has(Character)` → **M005** (`character 'X' is not on stage`).
  3. Else remove Character (case-insensitive match) from `s.names`. Log.

  `Exeunt` algorithm:
  1. If `chars == nil` → `s.Clear()` (exit all). Log "stage exeunt all".
  2. Else for each `c`: `checkDeclared(c)` → M001; if `!s.Has(c)` → **M005**; else remove.

  **D1 enforcement:** Do **NOT** call `s.Clear()` at act boundaries. The stage only changes via `Enter`/`Exit`/`Exeunt`. Stage starts empty at program start (the `Stage` zero value). This is the key revision from the original plan — confirmed by `primes.spl` evidence.

* **Error Handling:** M001 + M005. Collect, continue.
* **Testing:**
  - [ ] `TestStageExitExeunt`: table — `[Romeo]`+`Exit Romeo`→size0; `Exit Hamlet`(absent)→M005; `[Romeo,Juliet]`+`Exeunt nil`→size0; `[Romeo,Juliet]`+`Exeunt[Juliet]`→names=[Romeo]; `Exeunt[Hamlet]`(absent)→M005.
  - [ ] Fixture `m005-exit-not-on-stage.shpl`: `[Enter Romeo]` then `[Exit Juliet]` (Juliet not on stage) → M005.
  - [ ] Fixture `primes-persistence.shpl`: Act I ends with characters on stage (no Exeunt), Act II Scene I opens with a carried-over speaker (no `[Enter]`) → **zero** M004 (proves D1 persistence).
  - [ ] Fixture `m004-empty-stage-cross-act.shpl`: Act I ends with `[Exeunt]` (stage empty), Act II opens with dialogue and no `[Enter]` → parser now allows it (`enterSeen` persists true after Step 3.0 fix); semantic emits M004 (no one on stage). This is the correct parser/semantic division of labor.
  - [ ] Run `go test -race -run TestStageExit ./internal/semantic/...` → PASS.
* **Documentation Update:** `PROGRESS.md` bullet noting D1 (stage persists across acts, per primes.spl evidence).

---

### Step 3.6: Act/Scene traversal driver + M008 defensive assert (S007 overlap)

* **Goal:** Implement the outer walk that iterates acts and scenes in order, builds the per-act `SceneRegistry`, dispatches each scene's top-level statements, and asserts M008 defensively.
* **Context Files to Reference:** `internal/parser/ast.go` (Program.Acts, Act.Scenes, Scene.Statements); `docs/ERROR_TAXONOMY.md` (M008); note parser already enforces S007 (`errMissingScene`).
* **Implementation Details:**

  ```go
  type Analyzer struct {
      filename     string
      symbols      SymbolTable
      acts         ActRegistry
      currentAct   *parser.Act       // set during per-act walk
      sceneReg     SceneRegistry     // built per-act, used for goto resolution
      stage        Stage
      errs         []SemanticError
  }

  type Result struct {
      Symbols SymbolTable
      Acts    ActRegistry
      Errors  []SemanticError
  }

  func (r Result) OK() bool  // len(Errors) == 0

  func New(filename string, prog *parser.Program) *Analyzer
  func (a *Analyzer) Analyze(prog *parser.Program) Result
  ```

  `Analyze` algorithm:
  1. Build `symbols` from `prog.Characters`; build `acts` registry from `prog.Acts`.
  2. For each `act := range prog.Acts`:
     a. Set `a.currentAct = &act`. Log "act begin", act.RomanNumeral.
     b. **Do NOT clear the stage** (D1).
     c. **M008 defensive:** if `len(act.Scenes) == 0` → emit M008 (`Act <roman> has no scenes`). *Parser already enforces S007; this catches hand-built/evolving ASTs.*
     d. Build `a.sceneReg = buildSceneRegistry(&act)`.
     e. For each `scene := range act.Scenes`: call `a.analyzeSceneStatements(scene.Statements)`.
  3. Return `Result{Symbols: a.symbols, Acts: a.acts, Errors: a.errs}`.

  `analyzeSceneStatements` is the **level-1 dispatch** (type-switch over `parser.Statement`):
  - `parser.EnterStmt` → `a.stage.Enter(...)` → append errors.
  - `parser.ExitStmt` → `a.stage.Exit(...)`.
  - `parser.ExeuntStmt` → `a.stage.Exeunt(...)`.
  - `parser.Dialogue` → defer to Step 3.7 (`a.analyzeDialogue`).
  - default → ignore (no other node types appear at scene top level).

* **Error Handling:** M008 defensive. All stage errors appended to `a.errs`.
* **Testing:**
  - [ ] `TestAnalyzeEmptyScenesDefensive`: hand-build a `Program` with an `Act` containing zero scenes → expect M008.
  - [ ] `TestAnalyzeWalksMultipleActsWithPersistence`: 2-act program, Act I ends with characters on stage (no Exeunt), Act II Scene I opens with carried-over speaker, no `[Enter]` → zero errors (proves D1, no stage.Clear at act boundary).
  - [ ] Run `go test -race -run TestAnalyze ./internal/semantic/...` → PASS.
* **Documentation Update:** `PROGRESS.md` bullet.

---

### Step 3.7: Dialogue — speaker on stage (M004) + listener resolution + pronoun verification

* **Goal:** For each `Dialogue`, assert the speaker is on stage, compute the listener, and verify pronoun references resolve to a character genuinely present.
* **Context Files to Reference:** `docs/SPL_SPECIFICATION.md` (Characters must be on stage to speak; pronoun reference — me/myself/I=speaker, you/thou/yourself/thyself=listener; single-on-stage self-reference); `docs/ERROR_TAXONOMY.md` (M004); `internal/parser/ast.go` (Dialogue.Speaker, PronounExpr.Ref ∈ {"speaker","listener"}).
* **Implementation Details:**

  `analyzeDialogue(d parser.Dialogue, act *parser.Act, sceneReg SceneRegistry)`:
  1. `checkDeclared(d.Speaker, d.Line, d.Col)` → M001 if undeclared.
  2. If `!a.stage.Has(d.Speaker)` → **M004** with `role="speaker"` (`character '<Speaker>' is not on stage`). Best-effort: continue with `listener = Speaker`.
  3. Determine listener: `listener, ok := a.stage.Listener(d.Speaker)`:
     - `Listener` algorithm: `others = names where lower(n) != lower(speaker)`. If `len(others)==1` → `others[0]`. If `len(others)==0` → `speaker` (self-talk, canonical).
     - `ok` is false only if speaker not on stage (already M004'd) — then return speaker, false.
  4. Log `slog.Debug("dialogue", "speaker", d.Speaker, "listener", listener)`.
  5. Build `ctx := dialogCtx{speaker: d.Speaker, listener: listener}` and call level-2 dispatch (Steps 3.8/3.9).

  **Pronoun verification:** Inside expression analysis (3.8), when a `PronounExpr` is visited:
  - `Ref=="speaker"` → resolved = `ctx.speaker`; assert on stage (defensive — already guaranteed by step 2).
  - `Ref=="listener"` → resolved = `ctx.listener`; assert listener exists. If not, M004 already emitted.
  - *Framing:* Pronoun verification collapses to speaker-on-stage check. The per-pronoun pass records resolved names for optional future annotation and asserts `Ref ∈ {"speaker","listener"}` (catches corrupt AST).

* **Error Handling:** M001 (undeclared speaker) + M004 (speaker not on stage). Never emit M004 for a value-read `CharRefExpr` (D2).
* **Testing:**
  - [ ] Fixture `m004-speaker-not-on-stage.shpl`: `[Enter Romeo]`, `[Exit Romeo]`, then `Romeo:` speaks → M004 (parser allows because `enterSeen` stays true; semantic catches speaker off stage).
  - [ ] `TestListenerResolution`: stage `[Romeo,Juliet]`, speaker Romeo → listener Juliet; stage `[Romeo]` → listener Romeo (self); stage `[]` → ok=false.
  - [ ] Fixture `self-talk.shpl`: single character on stage speaks `You are yourself.` → **zero** errors (proves self-listener accepted).
  - [ ] Run `go test -race -run TestDialogue ./internal/semantic/...` → PASS.
* **Documentation Update:** `PROGRESS.md` bullet.

---

### Step 3.8: Expression verification + assignment/operation invariant assertions (D2)

* **Goal:** Walk every expression with the dialogue context, asserting AST invariants that align with SPL rules. Per D2, value-read `CharRefExpr`s do **not** require on-stage presence.
* **Context Files to Reference:** `docs/SPL_SPECIFICATION.md` (Operations, Assignment, Pronoun Reference, corrected constant/simile/unary sections); `internal/parser/ast.go` (all Expr nodes, Comparative, AssignStmt, RememberStmt).
* **Implementation Details:**

  `analyzeExpr(e parser.Expr, ctx dialogCtx) []SemanticError` — recursive:
  - `parser.ConstExpr` → no checks. Return nil.
  - `parser.CharRefExpr` → `checkDeclared(e.Name)` → M001 if not declared. **Do NOT check on-stage (D2).** Log `slog.Debug("charref", "name", e.Name, "on_stage", a.stage.Has(e.Name))` for observability only.
  - `parser.PronounExpr` → assert `e.Ref ∈ {"speaker","listener"}` (else M000 defensive). Resolve name via ctx; assert resolved name declared (defensive).
  - `parser.BinaryOpExpr` → assert `Op ∈ {sum,difference,product,quotient,remainder}` (defensive); recurse left, right.
  - `parser.UnaryOpExpr` → assert `Op ∈ {square,cube,square_root,factorial,twice}` (defensive); recurse operand.

  **Level-2 dispatch** (type-switch over inner dialogue statements):
  - `parser.AssignStmt` → assert `Target == "you"` (parser hardcodes; defensive). Call `analyzeExpr(Expr, ctx)`.
  - `parser.SpeakStmt` / `OpenHeartStmt` / `OpenMindStmt` / `ListenStmt` → operate on listener; no expr; covered by listener existence (3.7).
  - `parser.RememberStmt` → `analyzeExpr(Expr, ctx)` (Remember pushes onto **listener's** stack; listener covered).
  - `parser.RecallStmt` → no expr; pops **listener's** stack into speaker; covered by speaker+listener existence.
  - `parser.QuestionStmt` → assert `Comparative.Relation ∈ {equal,not_equal,greater,less,less_or_equal,greater_or_equal}` (defensive); `analyzeExpr(Left)`, `analyzeExpr(Right)`.
  - `parser.IfStmt` / `parser.GotoStmt` → handled in Step 3.9.

* **Error Handling:** M001 for undeclared `CharRefExpr` (defensive — parser normally can't produce these). Defensive (non-M) asserts for invalid Op/Ref/Relation/Target: emit `SemanticError{Code:"M000", Msg:"internal: invalid AST node"}` so it surfaces visibly.
* **Testing:**
  - [ ] `TestAnalyzeExpr`: CharRefExpr{Name:"Banquo"} undeclared → M001; CharRefExpr{Name:"Romeo"} declared but off-stage → **no error** (proves D2); BinaryOpExpr{Op:"bogus"} → M000 defensive; PronounExpr{Ref:"other"} → M000 defensive.
  - [ ] Fixture `off-stage-value-read.shpl`: Romeo off-stage, referenced in `the sum of Romeo and a flower` → zero errors (proves D2 end-to-end).
  - [ ] Fixture `valid-operations.shpl`: exercises sum/difference/product/quotient/remainder/square/cube/square_root/factorial/twice + a question with each comparative relation → zero errors.
  - [ ] Run `go test -race -run TestAnalyzeExpr ./internal/semantic/...` → PASS.
* **Documentation Update:** `PROGRESS.md` bullet noting D2 (off-stage value reads allowed, per canonical Hello World + primes.spl).

---

### Step 3.9: Goto/If target resolution (M006)

* **Goal:** Verify every `GotoStmt` and `IfStmt` target refers to a real scene (within the current act) or a real act (program-wide). The parser does **not** check existence.
* **Context Files to Reference:** `docs/SPL_SPECIFICATION.md` (Control Flow, Scenes are goto labels); `docs/ERROR_TAXONOMY.md` (M006); `internal/parser/ast.go` (GotoStmt.Target/TargetKind, IfStmt.Target/TargetKind — TargetKind ∈ {"scene","act"}).
* **Implementation Details:**

  ```go
  func (a *Analyzer) resolveGoto(target, kind string, line, col int) *SemanticError
  ```
  Algorithm:
  1. If `kind == "scene"`: look up `lower(target)` in `a.sceneReg` (the **current act's** scene registry). If missing → **M006** (`scene <target> is not defined in Act <currentAct.RomanNumeral>`).
  2. If `kind == "act"`: look up `lower(target)` in `a.acts` (global act registry). If missing → **M006** (`act <target> is not defined`).
  3. If `kind` is neither (corrupt AST) → M000 defensive.
  4. Log `slog.Debug("goto resolve", "kind", kind, "target", target, "found", ok)`.

  Wire into level-2 dispatch: `IfStmt` and `GotoStmt` both call `resolveGoto(s.Target, s.TargetKind, s.Line, s.Col)`.

  *Scope correctness:* scene numerals restart per act; the traversal driver (3.6) sets `a.sceneReg` per act, so `goto scene II` resolves against the **current act's** scenes only.

* **Error Handling:** M006 per missing target. Collect.
* **Testing:**
  - [ ] Fixture `m006-undefined-scene.shpl`: Act I has Scene I,II; `let us proceed to scene V` → M006.
  - [ ] Fixture `m006-undefined-act.shpl`: `let us return to act IX` (only I,II exist) → M006.
  - [ ] Fixture `goto-cross-act-scene.shpl`: Act II `goto scene I` resolves to Act II's Scene I → zero errors. Variant: Act II `goto scene III` where III not in Act II → M006 (proves per-act scoping).
  - [ ] Truth Machine fixture: contains `If so, let us proceed to scene II` — Scene II exists → zero errors.
  - [ ] Run `go test -race -run TestM006 ./internal/semantic/...` → PASS.
* **Documentation Update:** `PROGRESS.md` bullet.

---

### Step 3.10: Top-level `Analyze` wiring + `Result` + integration tests + docs

* **Goal:** Assemble the full pass, expose a clean `Result`, run the two canonical programs end-to-end (zero errors), and update all tracking docs. Gate with `task check`.
* **Context Files to Reference:** `PROGRESS.md` (Phase 3 section); `docs/ERROR_TAXONOMY.md` (severity: semantic errors stop pipeline before execution); `internal/parser/testutil_test.go` + `parser_test.go` (fixture-loading + TestMain pattern).
* **Implementation Details:**

  `Analyze(prog *parser.Program) Result` — the full flow from Step 3.6 (build symbols/acts → loop acts → M008 → loop scenes → level-1 dispatch → Dialogues → level-2 dispatch + expression walk + goto resolve). Append every emitted `SemanticError` to `r.Errors`.

  Test harness: `analyzer_test.go` with `TestMain` calling `logger.Init(logger.LevelDebug)`; helpers `analyzeFile(t, path) Result` that lex+parse+analyze a `.shpl` (mirror `parser_test.go`'s `lex`/`parseTokens` helpers).

* **Error Handling:** Full collection; `Result.OK()` is the pipeline gate that the future CLI (Phase 5) will check before invoking runtime.
* **Testing:**
  - [ ] `TestAnalyzeHelloWorld`: `testdata/semantic/hello.shpl` → `Result.OK() == true` (zero M-errors). Strongest evidence the D1/D2 interpretations are correct.
  - [ ] `TestAnalyzeTruthMachine`: → `OK() == true`.
  - [ ] `TestAnalyzeMultipleErrors`: `testdata/semantic/multiple-errors.shpl` (combines undeclared enter + stage overflow + bad goto) → code set equals `{M001, M003, M006}` (collect-all, D4).
  - [ ] Error-format snapshot: each returned `SemanticError.Error()` matches `error[Mxxx]: ...\n  --> input:L:C`.
* **Verification:**
  - [ ] Run `go test -race -coverprofile=coverage.out ./internal/semantic/...` → all PASS.
  - [ ] Run `task check` (fmt → lint → vuln → test) → **must pass clean**.
  - [ ] Self-review: re-read all four source files; confirm no S-code duplication; confirm every M001–M008 has at least one test that triggers it.
* **Documentation Update:**
  - Fill "Phase 3 — Semantic Analysis" section of `PROGRESS.md`: check off all M-codes, list fixtures, record decisions D1–D7 under "Decisions".
  - `docs/ERROR_TAXONOMY.md`: no new codes needed; optionally add note that M008 overlaps S007, M002 is parser-precluded.

---

## Self-Review Checklist (run before declaring Phase 3 complete)

1. **Spec coverage:** Stage rules (max 2, enter/exit/exeunt) → 3.4/3.5. On-stage to speak/be listener → 3.7. Declared characters → 3.2/3.3. Goto labels exist → 3.9. Pronoun resolution → 3.7. Assignment/operation alignment → 3.8. All eight M-codes: M001(3.3), M002(3.4), M003(3.4), M004(3.7), M005(3.5), M006(3.9), M007(3.4), M008(3.6). ✓
2. **Boundary check:** No S-code re-validation except the two flagged defensive asserts (M008, M002) with explicit overlap notes. No stage.Clear at act boundaries (D1). ✓
3. **D2 integrity:** No M004 issued for CharRefExpr in any expression — verified by `off-stage-value-read.shpl` and `TestAnalyzeExpr`. ✓
4. **D1 integrity:** No `stage.Clear()` at act boundaries — verified by `primes-persistence.shpl` (zero errors) and `TestAnalyzeWalksMultipleActsWithPersistence`. ✓
5. **Collect-all (D4):** `multiple-errors.shpl` triggers three distinct M-codes in one pass → all returned. ✓
6. **Task check green:** `task check` passes with no linter/vuln/test failures. ✓
7. **Coverage:** Every branch of `Stage.Enter`/`Exit`/`Exeunt`, every analyzer dispatch case, and every expression type is hit by at least one test. ✓

## Execution Handoff

This plan is complete and saved to `docs/superpowers/plans/2026-07-10-phase3-semantic-analysis.md`.

**Required first action (Phase 2 pre-req, Step 3.0):** Apply the `enterSeen` threading fix to `internal/parser/parser.go`, add `testdata/parser/cross-act-persistence.shpl` + regression test, run `task check`, update `PROGRESS.md`.

Then proceed through **Steps 3.1 → 3.10** in order, each with its own test gate.
