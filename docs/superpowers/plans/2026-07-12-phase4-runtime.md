# Plan: Phase 4 — Runtime / Evaluator (2026-07-12)

> **For agentic workers:** REQUIRED SUB-SKILL: Use `superpowers:subagent-driven-development` (recommended) or `superpowers:executing-plans` to implement this plan task-by-task. Steps use `- [ ]` checkbox syntax for tracking.

## Objective

Execute a semantically-validated SPL `*parser.Program`. Walk the program with a **flat-instruction trampoline** (single integer program counter over a pre-flattened instruction slice), evaluate expressions against a mutable character value store + per-character LIFO stacks, perform I/O through injected `io.Reader`/`io.Writer`, and resolve `Goto`/`If` jumps to scene/act program-counter indices obtained from the semantic registries. The phase ends with a `shpl run <file>` CLI subcommand and golden-file tests for the canonical Hello World and Truth Machine programs.

## Architecture

- New package `internal/runtime` imports `internal/parser` (AST), `internal/semantic` (to reuse the exported `Stage` and `SymbolTable` / `ActRegistry` from `semantic.Result`), and `internal/logger`. No new third-party deps; only `io`, `fmt`, `strings`, `strconv`, `log/slog`.
- **Flatten once, run many.** At `Execute` entry, walk `Program.Acts[].Scenes[].Statements` and produce a single `[]instr` slice. Each `instr` carries the original `parser.Statement` plus a `speaker string` (empty for stage directions). Build two label maps: `sceneLabel[lower(sceneRoman)] -> pc` and `actLabel[lower(actRoman)] -> pc` (act label = its first scene's first instruction). This makes Goto/If cost O(1) and lets a backward jump escape a Dialogue by simply overwriting `pc`.
- **State lives on one struct** `env`: character value store (`map[string]int` keyed by lowercased name), stack store (`map[string][]int`), a single `comparison bool` flag (the shared channel between `QuestionStmt` and the next `IfStmt`), the borrowed `*semantic.Stage`, the current dialogue's `speaker` + `listener` (re-derived per instruction from the stage), I/O handles, and the label maps.
- **Listener is always the implicit I/O/assignment target.** Re-derived at runtime via `stage.Listener(speaker)` (rules already implemented in `internal/semantic/stage.go:153-168`): exactly one other char on stage → that char; speaker alone → speaker (self-talk); speaker off stage → `ok=false`.
- **No visitor.** Type-switch dispatch on `instr.stmt`, mirroring the Phase 3 decision (D5 — single consumer, YAGNI).

## Global Constraints

- Go 1.26.5; the single gate is `task check` (fmt → lint → vuln → test). Lint set: `errcheck, govet, ineffassign, staticcheck, unused, bodyclose` (`.golangci.yaml` v2).
- Tests always use `-race` (suite convention). Test files include a `TestMain` calling `logger.Init(logger.LevelDebug)` to mirror `internal/parser`/`internal/semantic` test harnesses.
- Character-name keys are **case-insensitive** (lowercased); display preserves the original lexeme. This is the parser/semantic convention (D2).
- **The AST is pre-validated.** Do NOT re-run lex/parse/semantic checks at runtime. Specifically, the following are pre-computed on AST nodes and consumed directly:
  - `ConstExpr.Polarity` (∈ {-1, +1}), `ConstExpr.AdjectiveCount` (≥ 0) — value = `Polarity * (1 << AdjectiveCount)`.
  - `Comparative.Relation` ∈ {`equal`, `not_equal`, `greater`, `less`, `greater_or_equal`, `less_or_equal`} (parser collapsed the 7 forms to 6 relations).
  - `PronounExpr.Ref` ∈ {`"speaker"`, `"listener"`} (synthesized by the parser).
  - `AssignStmt.Target` is hardcoded to `"you"` (= listener); no speaker-self-assignment form exists.
  - `IfStmt.TargetKind` / `GotoStmt.TargetKind` ∈ {`"scene"`, `"act"`}; `Target` is the raw Roman numeral as-typed (lowercase it before registry lookup).
- **Dead branches of the dictionary are NOT needed**: `nounPolarity`, `comparativePolarity`, `isAdjective` are unexported in `internal/parser` and are intentionally not re-implemented; the runtime consumes only the pre-computed AST fields above.
- I/O is **injected** through `io.Reader`/`io.Writer` so tests can drive it with `bytes.Buffer`. Speak writes a single byte (no newline); OpenHeart writes the decimal integer followed by `\n` (canonical SPL behaviour — golden files pin the exact bytes).
- All new code lives under `internal/runtime/` and `testdata/interpreter/`. `cmd/shpl/main.go` is extended with one new `run` subcommand.

## Key Facts Inherited From Prior Phases (do not re-derive)

| Fact | Source | Runtime consequence |
|------|--------|---------------------|
| `Dialogue` has only `Speaker`; listener is not stored on the AST. | `parser/ast.go:130-135` | Re-derive listener per dialogue instruction via `semantic.Stage.Listener(speaker)`. |
| Value reads of off-stage characters are **allowed** (D2). | `PROGRESS.md` Phase 3 | `CharRefExpr` reads `env.values[lower(name)]` regardless of stage status. Only the *speaker* must be on stage (already enforced at semantic time; no re-check). |
| Stage **does not clear at act boundaries** (D1). | `PROGRESS.md` Phase 3 | `env.stage` is mutated only by `Enter`/`Exit`/`Exeunt` instructions; never reset on act label transitions. |
| `semantic.Result` exposes `Symbols`, `Acts`, `Errors` — **no `Scenes`, no `Stage`** (D7). | `internal/semantic/analyzer.go:22-30` | Runtime builds its own per-act `SceneRegistry` (3-line loop) and instantiates its own `semantic.Stage`. |
| Stage methods return `[]SemanticError`. | `internal/semantic/stage.go` | Runtime calls them with `res.Symbols` but **discards** the returned errors (AST is pre-validated). On the re-enter-after-backward-goto edge case, the no-op semantics of "already on stage → quietly keep stage unchanged" is the correct fallback. → decision R-D3. |
| `Remember` pushes onto the **listener's** stack; `Recall` pops from the **listener's** stack and assigns to the **speaker**. | `SPL_SPECIFICATION.md:324-330`, `parser/ast.go:274,285` | Implement exactly this asymmetry. |
| `QuestionStmt` sets a flag; `IfStmt` reads it. No AST link between them. | `parser/ast.go:205-266` | `env.comparison bool` is the only channel. |

## File Structure

```
internal/runtime/
├── errors.go        — RuntimeError type + R001–R004 constructors (format per taxonomy)
├── env.go           — env struct: value store, stack store, stage, comparison flag,
│                      speaker/listener cache, I/O handles, label maps; NewEnv + Execute entry
├── flatten.go       — flatten Program → []instr; build sceneLabel/actLabel maps
├── eval.go          — recursive eval(Expr) (int, error) over ConstExpr / CharRefExpr /
│                      PronounExpr / BinaryOpExpr / UnaryOpExpr
├── exec.go          — execInstr(instr) (jumpPC int, jumped bool, err error): type-switch dispatch
└── *_test.go        — table-driven unit tests + fixture golden tests (TestMain w/ logger.Init)

testdata/interpreter/
├── hello.shpl                    — canonical Hello World (copy of testdata/semantic/hello.shpl)
├── hello.golden                  — expected stdout bytes
├── truth-machine.shpl            — canonical Truth Machine (copy of testdata/semantic/truth-machine.shpl)
├── truth0.stdin / truth0.golden   — input "0\n" → output "0\n", halt
├── truth1.stdin / truth1.golden   — input "1\n" → output "1\n" repeated (cap in test, e.g. 8 reps)
├── stack.shpl                     — exercises Remember/Recall
├── stack.golden
├── branch.shpl                    — exercises If so / If not / Goto (forward and backward)
├── branch.golden
├── divzero.shpl                   — triggers R001
└── io-ascii.shpl                  — Speak + OpenMind round-trip

cmd/shpl/main.go                   — add `runCmd` mirroring `astCmd`; chain lex → parse → semantic → runtime.Execute
```

## Decision Points

| # | Decision | Value | Rationale |
|---|----------|-------|-----------|
| **R-D1** | Flatten program to a single `[]instr` + integer PC (not nested traversal) | One trampoline loop; backward goto = `pc = label`; dialogue escape = implicit | Truth Machine loops via backward goto; nested traversal would need sub-PC + escape signals. Ponytail: shortest path that holds. |
| **R-D2** | I/O formatting: Speak writes 1 byte, OpenHeart writes `%d\n` | Canonical SPL behaviour | Golden files pin exact bytes; regeneration documented in step 4.11. |
| **R-D3** | Reuse `semantic.Stage` for runtime stage state; **ignore** its returned `[]SemanticError` | Re-enter after backward goto silently no-ops (Stage keeps names unchanged, Listener unaffected). | Re-validation is redundant post-semantic. Avoids a runtime-mirror Stage with duplicate logic (DRY). |
| **R-D4** | Rebuild per-act `SceneRegistry` inside the runtime (3-line loop mirroring `semantic.buildSceneRegistry`) | `semantic.SceneRegistry` and its constructor are unexported. | The runtime needs scene-by-roman lookup against the *current* act when flattening/building labels. The loop is identical to semantic's. |
| **R-D5** | Comparison flag is a single `env.comparison bool` (not a stack) | One Question → one If consumes it (SPL semantics) | Matches SPL truth-machine behaviour and the AST (no link field). |
| **R-D6** | Integer overflow (R004) **does not error** | Go `int` wraps on overflow; spec taxonomy explicitly says "may not error" | Ponytail: do not invent a check the spec declines to mandate. |
| **R-D7** | `OpenMind` (read ASCII) reads exactly one byte; stores it as `int` (0–255). EOF → R003. | Spec: "Read ASCII char from input, store in character" | Simplest faithful mapping. |
| **R-D8** | `Listen` (read number) skips leading ASCII whitespace, accepts an optional `-`/`+` sign, then consecutive digits; no digits before next non-digit/EOF → R002; EOF before any digit → R003. | Matches canonical SPL numeric input parser | `strconv`-freehand; small enough to inline. |

---

## Step 4.1: Package scaffold, `RuntimeError` type, `env` struct, I/O abstraction

- **Goal:** Create `internal/runtime/errors.go` with the R-code error type and constructor helpers, and `internal/runtime/env.go` with the empty-but-typed *environment* struct and its constructor. No execution logic yet — just the typed surface every later step depends on.
- **Files:**
  - Create: `internal/runtime/errors.go`
  - Create: `internal/runtime/env.go`
  - Test: `internal/runtime/errors_test.go`
- **Context Files to Reference:** `docs/ERROR_TAXONOMY.md` (Level 4 table — R001–R004 names + exact message templates); `internal/parser/errors.go` and `internal/semantic/errors.go` (style to mirror: a struct with Code/Msg/Line/Col/Filename and an `Error()` formatter producing `error[<CODE>]: <message>\n  --> <file>:<line>:<col>`).
- **Implementation Details:**
  - Define `type RuntimeError struct { Code, Msg string; Line, Col int; Filename string }` with `func (e RuntimeError) Error() string` matching the taxonomy format exactly. Default `Filename = "input"` when empty (mirrors parser/semantic).
  - One-line constructor helpers (signatures only — bodies are plain struct returns):
    - `func errDivisionByZero(actRoman, sceneRoman string, line, col int) RuntimeError` — R001, message `division by zero at Act <act>, Scene <scene>` (taxonomy text).
    - `func errInputNotANumber(got string, line, col int) RuntimeError` — R002, `expected a number, got '<got>'`.
    - `func errInputEOF(line, col int) RuntimeError` — R003, `unexpected end of input`.
    - `func errIntegerOverflow(op string, line, col int) RuntimeError` — R004, `integer overflow in '<op>'` (currently unused per R-D6 but defined for completeness).
  - Define the environment struct (field sketch only — *no method bodies*):
    - `values map[string]int` — keyed by `lower(name)`.
    - `stacks map[string][]int` — per-character LIFO; missing key = empty (top = 0).
    - `stage *semantic.Stage` — borrowed; constructed by `NewEnv`.
    - `comparison bool` — shared flag (R-D5).
    - `in io.Reader`, `out io.Writer` — injected I/O.
    - `acts semantic.ActRegistry` — for `act` goto resolution.
    - `sceneLabel map[string]int` — `lower(roman) -> flat pc` (per current act scope; see 4.8 — actually this becomes per-act, so the field holds the *currently executing act's* scene label map, rebuilt on act transitions; documented in 4.8).
    - `actLabel map[string]int` — `lower(roman) -> flat pc` (global).
    - `currentActRoman string` — for error messages and scene-scope switches.
    - `filename string` — for runtime errors (`--> <file>:L:C`).
  - Constructor `func NewEnv(prog *parser.Program, res semantic.Result, in io.Reader, out io.Writer, filename string) *env` — initializes `values` from `res.Symbols` (every declared character starts at 0), `stacks` keyed similarly, a fresh `&semantic.Stage{}`, copies `res.Acts` reference, leaves label maps to be filled by `flatten` (4.8). Body is straightforward wiring; do not implement business logic.
  - Add a `TestMain` in `internal/runtime/main_test.go` calling `logger.Init(logger.LevelDebug)` — mirrors `internal/semantic`.
- **Error Handling:** This step *defines* the error vocabulary; no error paths are produced yet.
- **Testing & Verification:**
  - `TestRuntimeErrorFormat`: table mapping each R-code constructor to its exact formatted string `error[Rxxx]: <msg>\n  --> input:L:C`. Verify message text matches the taxonomy column byte-for-byte (e.g., R001 → `division by zero at Act I, Scene II`).
  - `TestNewEnv`: build a tiny `*parser.Program`+`semantic.Result` by hand (or by lex+parse+analyze a 1-character / 1-act / 1-scene string), call `NewEnv`; assert `values["romeo"] == 0`, `stacks["romeo"]` is empty/nil, `stage.Size() == 0`, `in`/`out` are the injected buffers, `acts` non-nil.
  - Run `go test -race -run 'TestRuntimeErrorFormat|TestNewEnv' ./internal/runtime/...` → PASS.
  - Run `task lint` → no findings.
- **Documentation Update:** Append a "Phase 4 — Runtime" subsection to `PROGRESS.md` listing Step 4.1 done and recording decisions R-D1..R-D8.

---

## Step 4.2: Constant expression evaluation (`ConstExpr`)

- **Goal:** Implement the first `eval` branch: turn `parser.ConstExpr` into an integer using the pre-computed `Polarity` and `AdjectiveCount`. This is the only expression that has no external dependency, so it ships first and unblocks every other eval branch.
- **Files:**
  - Create: `internal/runtime/eval.go`
  - Test: `internal/runtime/eval_test.go`
- **Context Files to Reference:** `docs/SPL_SPECIFICATION.md` (Constant evaluation corrected example, lines 303-313 — confirms `Value = Polarity × 2^AdjectiveCount`); `internal/parser/ast.go:296-303` (`ConstExpr` field shapes).
- **Implementation Details:**
  - Declare `func (e *env) eval(expr parser.Expr) (int, error)` — for this step, the body type-switches on `parser.ConstExpr` only and panics-on-default with an `R000`-style defensive error for any other type (later steps fill those in). Defensive `RuntimeError{Code:"R000", Msg:"internal: unsupported expr <T>"}` is allowed even though R000 is not in the taxonomy — it matches the "M000 defensive" pattern from Phase 3 and surfaces clearly if a future AST type slips in.
  - ConstExpr algorithm (pseudo):
    1. Read `expr.Polarity` (∈ {-1, +1}) and `expr.AdjectiveCount` (≥ 0).
    2. Compute `value := expr.Polarity * (1 << uint(expr.AdjectiveCount))` — Go's `<<` on `int` is fine for small counts; cap is academic (counts are tiny, ≤ ~6 in canonical SPL).
    3. Log `slog.Debug("const", "noun", expr.Noun, "value", value)`.
    4. Return `(value, nil)`.
  - No listener/speaker context needed for ConstExpr — later branches (Pronoun/CharRef) will require `(speaker, listener)` to be passed. Keep `eval` signature stable from the start: `eval(expr, speaker, listener string)` — pass empty strings in this step's tests.
- **Error Handling:** No R-codes here. Only R000 defensive for non-ConstExpr (temporary; replaced in 4.3).
- **Testing & Verification:**
  - `TestEvalConst`: table of `(AdjectiveCount, Polarity)` → expected int:
    - `flower` (0, +1) → 1
    - `red flower` (1, +1) → 2
    - `red hot flower` (2, +1) → 4
    - `coward` (0, -1) → -1
    - `big coward` (1, -1) → -2
    - `a vile coward` (1, -1) → -2 (the corrected example from the spec).
  - `TestEvalDefensive`: pass a `CharRefExpr{}` (not yet supported) → expect a `RuntimeError` with `Code == "R000"`.
  - Run `go test -race -run 'TestEvalConst|TestEvalDefensive' ./internal/runtime/...` → PASS.
- **Documentation Update:** `PROGRESS.md` bullet under Phase 4.

---

## Step 4.3: `CharRefExpr`, `PronounExpr`, binary and unary operations (incl. R001)

- **Goal:** Complete the expression evaluator. Add value reads (`CharRefExpr`), pronoun resolution (`PronounExpr`), and the five binary + five unary operations. Implement R001 (division/modulo by zero).
- **Files:**
  - Modify: `internal/runtime/eval.go`
  - Test: `internal/runtime/eval_test.go`
- **Context Files to Reference:** `docs/SPL_SPECIFICATION.md` (Operations list lines 82-91, corrected unary/simile sections lines 333-338, pronoun reference lines 358-372); `internal/parser/ast.go:312-355` (expression node shapes and their pre-set fields); `docs/ERROR_TAXONOMY.md` (R001).
- **Implementation Details:**
  - Change the `eval` signature to `eval(expr parser.Expr, speaker, listener string) (int, error)`. Update the 4.2 tests accordingly (pass `""`,`""` for ConstExpr cases).
  - `CharRefExpr`:
    1. `key := lower(expr.Name)`; `v, ok := e.values[key]`; if missing → R000 defensive (every declared character has an entry in `values` from `NewEnv`, so this is unreachable on validated ASTs).
    2. Return `(v, nil)`. Off-stage reads are allowed (D2) — no stage check.
  - `PronounExpr`:
    1. Switch on `expr.Ref`:
       - `"speaker"` → return `(e.values[lower(speaker)], nil)`.
       - `"listener"` → return `(e.values[lower(listener)], nil)`.
       - else → R000 defensive.
  - `BinaryOpExpr`:
    1. Recursively eval `Left` and `Right` (pass same `speaker, listener`).
    2. Switch on `expr.Op`:
       - `sum` → L + R
       - `difference` → L - R
       - `product` → L * R
       - `quotient` → if `R == 0` → `R001` (use `errDivisionByZero(e.currentActRoman, <scene roman>, expr.Line, expr.Col)`); scene roman is carried on the env as `currentSceneRoman` — add to env struct here); `else` L / R (Go integer division truncates toward zero — matches SPL "integer division").
       - `remainder` → if `R == 0` → `R001`; `else` L % R (Go's `%` truncates toward zero like the quotient; matches canonical SPL).
       - else → R000 defensive.
    3. Return `(result, nil)` or the R001 error.
  - `UnaryOpExpr`:
    1. Recursively eval `Operand`.
    2. Switch on `expr.Op`:
       - `square` → v * v
       - `cube` → v * v * v
       - `square_root` → integer square root via `int(math.Sqrt(float64(v)))` for v ≥ 0; for v < 0 → R000 defensive (canonical SPL doesn't define sqrt of negatives).
       - `factorial` → loop product 1..v; for v < 0 → R000 defensive; for v > 20 → R004 `errIntegerOverflow("factorial", ...)` (since 21! overflows int64; per R-D6 we don't error on overflow, but factorial is the one operation whose overflow is detectable and catastrophic — flag it R004 here as the single exception, document the deviation).
       - `twice` → v * 2
       - else → R000 defensive.
    3. Return `(result, nil)`.
  - Add a `currentSceneRoman string` field to `env` (set by the exec loop in 4.8; unused here but reserved for R001 messages).
  - **Simile note:** the parser already stripped `as <adj> as <expr>` to its inner `Expr` (parser.go:846-859). `eval` never sees a simile node — just the inner expression. Do not implement simile handling.
- **Error Handling:** R001 (division/modulo by zero); R004 (factorial overflow only); R000 defensive for unreachable AST shapes.
- **Testing & Verification:**
  - `TestEvalCharRef`: env with `values["romeo"]=3, values["juliet"]=5`; eval `CharRefExpr{Name:"Romeo"}` → 3, eval `CharRefExpr{Name:"JULIET"}` → 5 (case-insensitive). Eval `CharRefExpr{Name:"Ghost"}` → R000 (defensive, never on validated AST).
  - `TestEvalPronoun`: same env; eval `PronounExpr{Ref:"speaker"}` with speaker="Romeo" → 3; `Ref:"listener"` with listener="Juliet" → 5; `Ref:"other"` → R000.
  - `TestEvalBinary`: table over each op with concrete L/R; include `quotient` with R=0 → R001; verify the error's `Code` and that the message mentions the current act+scene (set `env.currentActRoman="I"`, `currentSceneRoman="II"`).
  - `TestEvalUnary`: table over each op; include `factorial` of 5 → 120; `factorial` of 21 → R004; `square_root` of 16 → 4; `square_root` of -1 → R000.
  - Nesting: `the sum of Romeo and the difference between Juliet and a flower` → build `BinaryOpExpr{sum, CharRef(Romeo), BinaryOpExpr{difference, CharRef(Juliet), ConstExpr(0,+1)}}`, assert equals `3 + (5 - 1) = 7`.
  - Run `go test -race -run 'TestEval' ./internal/runtime/...` → PASS.
- **Documentation Update:** `PROGRESS.md` bullet noting R001 implemented and the factorial-overflow exception (R-D6 partial exception for factorial).

---

## Step 4.4: Stage manager reuse — `Enter`/`Exit`/`Exeunt` instruction handling and listener derivation

- **Goal:** Wire the runtime to drive the borrowed `semantic.Stage` from `Enter`/`Exit`/`Exeunt` statements, and provide the single helper `env.listener(speaker) (string, bool)` that every dialogue statement will use.
- **Files:**
  - Modify: `internal/runtime/env.go` (add `listener` helper)
  - Create: `internal/runtime/exec.go` (start the dispatch; this step implements only the three stage-direction cases)
  - Test: `internal/runtime/exec_test.go`
- **Context Files to Reference:** `internal/semantic/stage.go` (full read — `Stage.Enter/Exit/Exeunt` signatures and the `Listener` rule at lines 153-168); `internal/parser/ast.go:93-121` (`EnterStmt`, `ExitStmt`, `ExeuntStmt` — note `ExeuntStmt.Characters == nil` means "exit all"); `docs/SPL_SPECIFICATION.md` (Stage Management lines 41-51).
- **Implementation Details:**
  - `func (e *env) listener(speaker string) (string, bool)` — one-liner delegating to `e.stage.Listener(speaker)`. (Indirection lets us add logging/observability in one place; not strictly required by ponytail but one line of glue.)
  - Define the `instr` struct (in `exec.go`):
    - `stmt parser.Statement` — the original AST node.
    - `speaker string` — empty for `Enter`/`Exit`/`Exeunt` (stage op); non-empty for Dialogue-derived instructions.
    - `sceneRoman string` — the Roman numeral of the hosting scene (for R001 messages and act-scope tracking).
    - `actRoman string` — the Roman numeral of the hosting act.
  - Begin `func (e *env) execInstr(i instr) (jumpPC int, jumped bool, err error)`. For this step, only the three stage-op cases are implemented; default returns a not-yet-implemented R000 (later steps fill it in).
    - `EnterStmt`: `errs := e.stage.Enter(s.Characters, e.syms, i.Line()` — wait, `stage.Enter` takes a `SymbolTable`; expose `syms` on the env (set in `NewEnv` from `res.Symbols`). **Discard `errs`** (per R-D3). Set `jumped=false`. Log `slog.Debug("stage enter", "chars", s.Characters, "size", e.stage.Size())`.
    - `ExitStmt`: call `e.stage.Exit(s.Character, e.syms, ...)`; discard errors. Log.
    - `ExeuntStmt`: call `e.stage.Exeunt(s.Characters, e.syms, ...)`; discard. Note `s.Characters == nil` ⇒ "exit all" (the Stage method handles this). Log.
  - Add `syms semantic.SymbolTable` field to `env` (set in `NewEnv`).
  - Update `NewEnv` to store `res.Symbols` into `e.syms`.
  - **No act-boundary stage reset (D1):** the loop in 4.8 must NOT clear `e.stage` when `actRoman` changes between consecutive instructions. Document this invariant in a comment on `instr.actRoman`.
- **Error Handling:** No R-codes here. Discarded semantic errors are intentional (R-D3).
- **Testing & Verification:**
  - `TestExecEnterExitExeunt`: drive `execInstr` directly (no flatten needed) on hand-built `instr` values:
    - `instr{stmt: EnterStmt{Characters:["Romeo"]}}` → after call, `stage.Size()==1`, `stage.Has("Romeo")`.
    - Two consecutive `Enter` instrs (Romeo, then [Romeo and Juliet]) → size 2.
    - `instr{stmt: EnterStmt{Characters:["Romeo","Romeo"]}}` on an empty stage → `stage.Size()==1` (second Romeo skipped by semantic, error discarded). *Confirms R-D3 re-enter is a silent no-op, not a runtime crash.*
    - `ExitStmt{Character:"Romeo"}` after Romeo on stage → size 0; `ExitStmt{Character:"Juliet"}` when Juliet not on stage → stage unchanged (error discarded).
    - `ExeuntStmt{Characters:nil}` after {Romeo,Juliet} → size 0.
    - `ExeuntStmt{Characters:["Juliet"]}` after {Romeo,Juliet} → size 1, only Romeo remains.
  - `TestListenerDerivation`: pre-seed `stage` with {Romeo, Juliet}; `listener("Romeo")` → ("Juliet", true); `listener("Mercutio")` → ("Mercutio", false). Pre-seed {Romeo} alone; `listener("Romeo")` → ("Romeo", true) (self-talk).
  - Run `go test -race -run 'TestExecEnter|TestListener' ./internal/runtime/...` → PASS.
- **Documentation Update:** `PROGRESS.md` bullet noting R-D3 (semantic Stage errors discarded at runtime).

---

## Step 4.5: Assignment and I/O statements (Speak / OpenHeart / OpenMind / Listen) — R002, R003

- **Goal:** Implement the four I/O statements and the assignment statement. All target the **listener** (re-derived per instruction via `env.listener(speaker)`). Implement R002 (non-numeric input) and R003 (EOF).
- **Files:**
  - Modify: `internal/runtime/exec.go`
  - Test: `internal/runtime/exec_test.go`
- **Context Files to Reference:** `docs/SPL_SPECIFICATION.md` (I/O Commands table lines 113-120, Assignment lines 98-105); `internal/parser/ast.go:144-202` (`AssignStmt` — `Target=="you"`; the four I/O node types — no operand field); `docs/ERROR_TAXONOMY.md` (R002, R003).
- **Implementation Details:**
  - `AssignStmt`:
    1. `listener, ok := e.listener(i.speaker)`; if `!ok` → R000 defensive ("speaker not on stage" — unreachable post-semantic).
    2. `v, err := e.eval(s.Expr, i.speaker, listener)`; propagate error.
    3. `e.values[lower(listener)] = v`. Log `slog.Debug("assign", "char", listener, "value", v)`. Skip `s.SimileAdj` (the parser already stripped the simile to `Expr`).
    4. Return `(0, false, nil)` (no jump).
  - `SpeakStmt`: (writes one ASCII byte of the **listener's** value)
    1. Resolve `listener`.
    2. `v := e.values[lower(listener)]`.
    3. Write a single byte: `e.out.Write([]byte{byte(v & 0xFF)})` — cast to a single byte; the canonical Hello World only emits values 32–126. For values outside that range the behaviour is undefined by the spec and remains "write the low 8 bits". Log.
    4. No newline (R-D2).
  - `OpenHeartStmt`: (writes the decimal integer of the listener's value, followed by `\n`)
    1. Resolve `listener`, read value.
    2. `fmt.Fprintf(e.out, "%d\n", v)`. Log.
  - `OpenMindStmt`: (reads one ASCII char into the **listener's** value)
    1. Resolve `listener`.
    2. Read exactly one byte from `e.in`: `var buf [1]byte; n, err := e.in.Read(buf[:])`. If `err == io.EOF || n == 0` → return R003 (`errInputEOF(line, col)`). For any other read error → wrap as R003 with the same message.
    3. `e.values[lower(listener)] = int(buf[0])`. Log.
  - `ListenStmt`: (reads an integer from text input into the **listener's** value)
    1. Resolve `listener`.
    2. Implement the R-D8 numeric reader: skip leading ` `/`\t`/`\r`/`\n`; read an optional `+` or `-` sign; read consecutive ASCII digits into a `[]byte`. If EOF is hit before any digit → R003. If at least one non-digit non-whitespace char is hit before any digit (and no EOF) → R002 (`errInputNotANumber(string(gotBytes), line, col)`). If EOF is hit after a sign but before digits → R002.
    3. Parse the digits + sign with `strconv.Atoi`; an `Atoi` error here is defensive (R000) — the hand-rolled scan already validated the format.
    4. `e.values[lower(listener)] = parsed`. Log.
  - Update the `instr` lookup loop: when `i.speaker == ""` (stage op) → only Enter/Exit/Exeunt are valid; otherwise compute `listener` once at the top of `execInstr` (`listener, _ := e.listener(i.speaker)`) and reuse inside the statement's branch.
- **Error Handling:** R002 (non-numeric Listen input), R003 (EOF on OpenMind or Listen). R000 defensive for unreachable mis-shapes.
- **Testing & Verification:**
  - `TestExecAssign`: env with listener=Juliet (values["juliet"]=5); `AssignStmt{Expr: ConstExpr{1,+1}}` → `values["juliet"]==1`. Use `AssignStmt{Expr: BinaryOpExpr{sum, CharRef{Romeo=3}, ConstExpr{0,+1}}}` → `values["juliet"]==4`.
  - `TestExecSpeak`: out=`&bytes.Buffer{}`; listener=Juliet with value 72; `SpeakStmt` → out contains `[]byte{72}` (ASCII 'H'). Repeat with 105 → 'i'.
  - `TestExecOpenHeart`: listener value 42; `OpenHeartStmt` → out == `"42\n"`.
  - `TestExecOpenMind`: in=`strings.NewReader("A")`; listener=Juliet (any value); `OpenMindStmt` → `values["juliet"]==65` (ASCII 'A'). Then with `in=&bytes.Buffer{}` (empty) → R003 (`Code=="R003"`).
  - `TestExecListen`: with `in=strings.NewReader("123\n")` → `values==123`. With `in=strings.NewReader("  -7 xyz")` → `values==-7`. With `in=strings.NewReader("abc")` → R002, message contains `'abc'`. With `in=strings.NewReader("")` → R003. With `in=strings.NewReader("  -")` (sign then EOF) → R002.
  - `TestExecSelfTalkIO`: stage {Romeo} alone; `SpeakStmt` with speaker `"Romeo"` → listener resolves to `"Romeo"`; writes its own value as a byte. Confirms self-talk path.
  - Run `go test -race -run 'TestExec' ./internal/runtime/...` → PASS.
- **Documentation Update:** `PROGRESS.md` bullet noting R002/R003 paths and R-D2 (Speak no-newline, OpenHeart newline).

---

## Step 4.6: Stack operations (`Remember` / `Recall`) — listener/speaker asymmetry

- **Goal:** Implement the two stack operations exactly per the spec's listener/speaker asymmetry: `Remember` pushes onto the **listener's** stack; `Recall` pops from the **listener's** stack and assigns to the **speaker** (0 if empty).
- **Files:**
  - Modify: `internal/runtime/exec.go`
  - Test: `internal/runtime/exec_test.go`
- **Context Files to Reference:** `docs/SPL_SPECIFICATION.md` (Stack operations lines 324-330 — exact push/pop orientation); `internal/parser/ast.go:274-294` (`RememberStmt.Expr`; `RecallStmt.IgnoredText` is free-text filler, semantically discarded).
- **Implementation Details:**
  - `RememberStmt`:
    1. Resolve `listener`.
    2. `v, err := e.eval(s.Expr, i.speaker, listener)`; propagate error.
    3. `e.stacks[lower(listener)] = append(e.stacks[lower(listener)], v)`. Log.
  - `RecallStmt`:
    1. Resolve `listener` (yes — the source is the listener's stack, even though the destination is the speaker).
    2. `stack := e.stacks[lower(listener)]`; if `len(stack)==0` → `v := 0`; else `v := stack[len(stack)-1]; e.stacks[lower(listener)] = stack[:len(stack)-1]`.
    3. `e.values[lower(i.speaker)] = v`. Log `slog.Debug("recall", "from", listener, "to", i.speaker, "value", v)`.
    4. Ignore `s.IgnoredText` entirely.
  - Edge: `Remember yourself.` — `yourself` is a `PronounExpr{Ref:"listener"}` (per parser). `eval` returns the listener's current value, which then gets pushed onto the listener's own stack. Correct.
  - Edge: `Remember me.` — `me` is `PronounExpr{Ref:"speaker"}`; evaluates to the speaker's current value, pushed onto the listener's stack. Correct.
- **Error Handling:** No R-codes. No stack-overflow check (stacks are bounded only by memory — YAGNI).
- **Testing & Verification:**
  - `TestRemember`: env stage {Romeo, Juliet}, values romeo=3 juliet=0; `RememberStmt{Expr: CharRef{Romeo}}` with speaker=Romeo → listener=Juliet, stacks["juliet"]==[3]. Then `RememberStmt{Expr: ConstExpr{0,-1}}` → stacks["juliet"]==[3,-1].
  - `TestRecall`: precondition stacks["juliet"]==[7,8] (8 on top), stage {Romeo,Juliet}, speaker=Romeo (listener=Juliet). `RecallStmt` → values["romeo"]==8, stacks["juliet"]==[7]. Repeat → values["romeo"]==7, stacks["juliet"]==[]. Repeat with empty stack → values["romeo"]==0.
  - `TestRememberRecallRoundTrip`: the canonical SPL idiom `Remember me.` (speaker=Romeo) → push Romeo's value onto Juliet's stack; then later `Recall` with speaker=Juliet (listener=Romeo) → pop from Romeo's stack into Juliet. Build a 4-instr sequence verifying both halves.
  - `TestRememberYourself`: `Remember yourself.` with speaker=Romeo, listener=Juliet (juliet=4) → stacks["juliet"]==[4].
  - Run `go test -race -run 'TestRemember|TestRecall' ./internal/runtime/...` → PASS.
- **Documentation Update:** `PROGRESS.md` bullet confirming the listener/speaker asymmetry is implemented exactly per spec (resolves the discrepancy between the spec and the earlier explore-subagent note).

---

## Step 4.7: Questions, comparison flag, and `If`/`Goto` branch resolution

- **Goal:** Evaluate `QuestionStmt` (set `env.comparison` from the pre-computed `Comparative.Relation`), and resolve `IfStmt` / `GotoStmt` jumps to a flat program-counter index, returning a "jump" signal from `execInstr` rather than advancing the PC.
- **Files:**
  - Modify: `internal/runtime/exec.go`
  - Test: `internal/runtime/exec_test.go`
- **Context Files to Reference:** `docs/SPL_SPECIFICATION.md` (Control Flow lines 122-142, comparative forms table at 339-356); `internal/parser/ast.go:205-266` (`QuestionStmt.Left/Right` are `Expr`, `Comparative.Relation` pre-computed; `IfStmt.BranchIfTrue`, `IfStmt.Target`/`TargetKind`; `GotoStmt.Target`/`TargetKind`); `docs/ERROR_TAXONOMY.md` (M006 — absent at runtime since semantic validated targets; document defensive R000 for unknown `Relation`).
- **Implementation Details:**
  - `QuestionStmt`:
    1. `leftV, err := e.eval(s.Left, i.speaker, listener)`; same for `Right`.
    2. `e.comparison = applyRelation(s.Comparative.Relation, leftV, rightV)`; the helper switches on the 6 relation strings and returns a `bool`:
       - `equal` → `leftV == rightV`
       - `not_equal` → `leftV != rightV`
       - `greater` → `leftV > rightV`
       - `less` → `leftV < rightV`
       - `greater_or_equal` → `leftV >= rightV`
       - `less_or_equal` → `leftV <= rightV`
       - default → defensive R000 (`internal: unknown relation <r>`).
    3. Log `slog.Debug("question", "relation", s.Comparative.Relation, "left", leftV, "right", rightV, "result", e.comparison)`.
    4. Return `(0, false, nil)` — questions do not jump.
  - `IfStmt`:
    1. `matched := e.comparison == s.BranchIfTrue` (BranchIfTrue=true for `If so`, false for `If not`).
    2. If `!matched` → return `(0, false, nil)` (fall through; PC advances).
    3. If matched → resolve target: `pc, err := e.resolveJump(s.Target, s.TargetKind, i)`; on error return it; else return `(pc, true, nil)`.
  - `GotoStmt` (unconditional):
    1. `pc, err := e.resolveJump(s.Target, s.TargetKind, i)`; on error return it; else return `(pc, true, nil)`.
  - `func (e *env) resolveJump(target, kind string, i instr) (int, error)`:
    1. `lt := lower(target)`.
    2. If `kind == "scene"` → look up `e.sceneLabel[lt]`. `i.actRoman` must equal `e.currentActRoman` (guaranteed by the flatten+run loop — the env's `sceneLabel` is rebuilt on act transitions per R-D4). If missing → R000 defensive (semantic M006 already eliminated this case).
    3. If `kind == "act"` → look up `e.actLabel[lt]`; missing → R000 defensive.
    4. Else (corrupt `TargetKind`) → R000.
    5. Return the flat PC.
- **Error Handling:** No R-codes from user error (M006 already prevented undefined targets at runtime). R000 defensive only, signalling an internal bug if hit.
- **Testing & Verification:**
  - `TestApplyRelation`: table over each of the 6 relations with concrete ints; assert the bool. Include `greater_or_equal` with equal operands → true.
  - `TestExecQuestion`: env with romeo=5, juliet=3, stage {Romeo,Juliet}, speaker=Romeo (listener=Juliet). `QuestionStmt{Left: Pronoun{speaker}, Comparative.Relation: "greater", Right: Pronoun{listener}}` → `e.comparison == true`. Same with `equal` → false. Reset comparison before each subtest (the tests must be hermetic — re-seed `e.comparison=false` per case).
  - `TestExecIf`: precondition `e.comparison=true`; `IfStmt{BranchIfTrue:true, Target:"II", TargetKind:"scene"}` with `e.sceneLabel={"ii": 7}` → returns `(7, true, nil)`. Same If with `BranchIfTrue:false` (i.e. `If not`) and `comparison=false` (set explicitly) → returns `(7, true, nil)`. Mismatched `If not` with `comparison=true` → `(0, false, nil)`.
  - `TestExecGoto`: `GotoStmt{Target:"II", TargetKind:"scene"}` → `(7, true, nil)`. `GotoStmt{Target:"III", TargetKind:"act"}` with `actLabel={"iii": 12}` → `(12, true, nil)`.
  - `TestExecQuestionDefensive`: `QuestionStmt` with `Comparative.Relation:"unknown"` → R000.
  - Run `go test -race -run 'TestExecIf|TestExecGoto|TestApplyRelation|TestExecQuestion' ./internal/runtime/...` → PASS.
- **Documentation Update:** `PROGRESS.md` bullet noting R-D5 (single comparison bool) and the goto/escape mechanism returning a (pc, jumped) signal.

---

## Step 4.8: Program flattening + per-act scene label maps + the trampoline `Execute`

- **Goal:** Flatten `Program.Acts[].Scenes[].Statements` into a single `[]instr` slice, attach the correct `speaker` to each Dialogue-derived instruction (empty for stage ops), build `sceneLabel`/`actLabel` maps, and implement the main `Execute` loop that advances/jumps the program counter. This is the integration step that ties 4.1–4.7 together end-to-end.
- **Files:**
  - Create: `internal/runtime/flatten.go`
  - Modify: `internal/runtime/env.go` (add `Execute`, `flatten`, label maps)
  - Test: `internal/runtime/execute_test.go`
- **Context Files to Reference:** `internal/parser/ast.go` (`Program.Acts`, `Act.Scenes`, `Scene.Statements`, `Dialogue.Speaker` / `Dialogue.Statements`); `internal/semantic/symbol_table.go` (`ActRegistry.Resolve`); the run loop pattern in `internal/semantic/analyzer.go` (for layout resemblance).
- **Implementation Details:**
  - Define the flatten algorithm in `func (e *env) flatten(prog *parser.Program)`:
    1. `e.instrs = nil`, `e.actLabel = make(map[string]int)`, `e.sceneLabel = make(map[string]int)`.
    2. For each `actIdx, act := range prog.Acts`:
       a. `actRoman := lower(act.RomanNumeral)`.
       b. If `e.actLabel[actRoman]` is not yet set → `e.actLabel[actRoman] = len(e.instrs)` (the first instruction of the act = first instruction of this act's first scene).
       c. For each `sceneIdx, scene := range act.Scenes`:
          - `sceneRoman := lower(scene.RomanNumeral)`.
          - `e.sceneLabel[sceneRoman] = len(e.instrs)` (regardless of whether the act was seen before — this overwrites the per-act scene map; since the flatten walks acts **in order**, the final `sceneLabel` reflects the *last* act's scenes. **This is wrong** for goto resolution at runtime if gotos can target previous acts' scenes.)
       d. **Correction (R-D4):** `sceneLabel` is **per-act**, not global. Store it as `map[actRoman]map[string]int` keyed by `lower(actRoman)` → `map[lower(sceneRoman)]pc`. At flatten time, build `e.sceneLabels[actRoman][sceneRoman] = len(e.instrs)`. At runtime, `resolveJump` uses `e.sceneLabels[e.currentActRoman][lt]` (see step 4.10 setting `currentActRoman` on act transitions).
       e. For each `stmt in scene.Statements`:
          - Type-switch:
            - `EnterStmt`/`ExitStmt`/`ExeuntStmt` → append `instr{stmt: stmt, speaker:"", sceneRoman: act.RomanNumeral, actRoman: act.RomanNumeral}`.
            - `Dialogue` → for each `inner in d.Statements` append `instr{stmt: inner, speaker: d.Speaker, sceneRoman: act.RomanNumeral, actRoman: act.RomanNumeral}`. (Listener is NOT stored; re-derived per instruction at runtime via `env.listener(speaker)`.)
            - default → defensive panic/R000 (only the four statement kinds live at scene top level).
  - Build `actLabel` as a global map: `lower(actRoman) → pc` of the act's first instruction (set once, first time the act is encountered above).
  - Define `func (e *env) Execute() error` — the trampoline:
    1. `pc := 0`.
    2. `for pc < len(e.instrs) { i := e.instrs[pc] ... }`.
    3. **Act-transition bookkeeping (R-D4):** if `i.actRoman != e.currentActRoman` → set `e.currentActRoman = i.actRoman`; set `e.currentSceneRoman = i.sceneRoman` when it changes too (for R001 messages). Do NOT clear the stage (D1).
    4. `next, jumped, err := e.execInstr(i)`. If `err != nil` → return wrapped `RuntimeError` (use the error's own `Error()` formatting; do not double-wrap).
    5. If `jumped` → `pc = next`. Else `pc++`.
    6. Loop.
    7. On natural exit (`pc >= len(instrs)`) → return `nil`.
  - Move `currentSceneRoman` from the structural addition in 4.3 to a runtime-updated field here (set at each instruction's act/scene transition).
  - Update `resolveJump` (4.7) to use `e.sceneLabels[e.currentActRoman][lt]` for `kind=="scene"` instead of a single global `sceneLabel`.
- **Error Handling:** R-codes propagated from `execInstr` (R001/R002/R003). R000 defensive if a malformed instruction appears or if a goto target is somehow missing (impossible post-semantic).
- **Testing & Verification:**
  - `TestFlatten`: build a small 2-act Program (Act I: Scenes I, II; Act II: Scene I) with mix of Enter/Speak/Dialogue. After flatten:
    - `len(instrs)` equals the count of (stage ops + sum of Dialogue inner statements).
    - `actLabel["i"] == 0`, `actLabel["ii"]` == index of Act II's first instruction.
    - `sceneLabels["i"]["ii"]` == index of Act I Scene II's first instruction; `sceneLabels["ii"]["i"]` == index of Act II Scene I's first instruction. Confirms per-act scoping (D4).
    - Inspect `instr.speaker` is empty for `EnterStmt` and `"Romeo"` for a Romeo dialogue statement.
  - `TestExecuteLinear`: a 3-instruction program (`Enter Romeo; Assign Juliet = ConstExpr(2,+1); Speak`) — wait, listener needs to be Juliet; build with `[Enter Romeo and Juliet]` first → stage {Romeo, Juliet}; `Romeo: You are a flower.` (assigns romeo? no — listener). Use the `self-talk` shorthand: stage {Romeo} alone, `Romeo: You are as good as a flower.` (AssignStmt → listener=Romeo → values["romeo"]=1) then `Romeo: Speak your mind.` → out contains `[]byte{1}`. Assert via a single `Execute()` call.
  - `TestExecuteLoop` (mini truth machine): 1-act, 2-scene program; Scene I is `Romeo: Listen to your heart.` (with two characters on stage so the input goes into the listener), Scene II is `Juliet: Am I as good as you?` + `Romeo: If so, let us proceed to scene II.` Drive `in` with `"0\n"`; assert `Execute` returns and `out` contains exactly `"0\n"` (one OpenHeart writes 0, comparison false, no loop). Drive with `"1\n"`; cap `in` to allow 8 iterations by using a 1-byte Reader that returns EOF after the first read (so the second `Listen` raises R003); assert the error is R003 and `out` contains `"1\n"` repeated 8 times (or as many as the loop runs before the R003). *This proves backward goto + dialogue escape works.*
  - `TestExecuteActTransition`: program with Enter in Act I Scene I that puts {Romeo, Juliet} on stage, no Exeunt, then Act II Scene I has Juliet speak with no new Enter (cross-act persistence, D1). Assert `Execute()` runs and the listener resolves to Romeo.
  - `TestExecuteDivZero`: program that reaches `the quotient between X and a coward` (Y=0) → `Execute()` returns a `RuntimeError{Code:"R001"}` and the message mentions Act I, Scene I.
  - Run `go test -race ./internal/runtime/...` → all PASS.
- **Documentation Update:** `PROGRESS.md` bullet noting the trampoline is live, R-D4 per-act scene label map, and the loop/escape verification.

---

## Step 4.9: Public `Execute` entry + `Run` glue

- **Goal:** Expose a single clean public function `Execute(prog *parser.Program, res semantic.Result, in io.Reader, out io.Writer, filename string) error` that hides the env struct, flattens the program, and runs the trampoline. This is the function the CLI and integration tests call.
- **Files:**
  - Modify: `internal/runtime/env.go` (export `Execute`; or add `runtime.go`)
  - Test: `internal/runtime/execute_test.go`
- **Context Files to Reference:** `internal/lexer/lexer.go` (`New`/`ScanTokens`), `internal/parser/parser.go` (`New`/`Parse`), `internal/semantic/analyzer.go` (`New`/`Analyze`/`Result.OK`); `internal/semantic/stage.go`.
- **Implementation Details:**
  - `func Execute(prog *parser.Program, res semantic.Result, in io.Reader, out io.Writer, filename string) error`:
    1. If `!res.OK()` → return immediately with a sentinel error? No — the CLI gates on `res.OK()` *before* calling `Execute` (per the severity rules in `ERROR_TAXONOMY.md`: semantic errors stop the pipeline before execution). For belt-and-suspenders, assert `res.OK()` here and return `fmt.Errorf("semantic analysis failed: %d error(s)")` if not (do not run).
    2. `e := NewEnv(prog, res, in, out, filename)`.
    3. `e.flatten(prog)`.
    4. `return e.Execute()`.
  - The unexported `env.Execute` (4.8 trampoline) stays unexported; the public surface is `Execute` (capitalized). Keep the trampoline's name internal and pick either `run` or keep both — prefer one exported `Execute` and rename the internal one to `runLoop` to avoid shadowing.
- **Error Handling:** Returned `error` is always a `RuntimeError` (R001/R002/R003) or the belt-and-suspenders "semantic failed" guard. Caller formats to stderr.
- **Testing & Verification:**
  - `TestExecuteEntry`: feed a tiny parsed `*parser.Program` (or one parsed from a 4-line literal string via `parser.New(lexer.New(src).ScanTokens()).Parse()`) plus its `semantic.Result`. Assert the public `Execute` writes expected bytes to the injected buffer and returns `nil`.
  - `TestExecuteRejectsUnvalidatedAST`: pass a `semantic.Result` whose `Errors` is non-empty → `Execute` returns the sentinel error and does not touch `out`.
  - Run `go test -race -run 'TestExecuteEntry|TestExecuteRejects' ./internal/runtime/...` → PASS.
- **Documentation Update:** `PROGRESS.md` bullet noting the public `Execute` entry is available for the CLI.

---

## Step 4.10: `shpl run` CLI subcommand

- **Goal:** Add a `run` subcommand to `cmd/shpl/main.go` mirroring the existing `astCmd` shape; chain `lexer → parser → semantic.Analyze → runtime.Execute`; wire `os.Stdin`/`os.Stdout` for I/O. On lex/parse/semantic error, print to stderr and `os.Exit(1)` (matching existing `tokensCmd`/`astCmd` style).
- **Files:**
  - Modify: `cmd/shpl/main.go` (add `runCmd`; register in `init()`; add `semantic` and `runtime` imports)
  - Test: `cmd/shpl/main_test.go` (add `TestRunCommand` mirroring existing `TestTokensCommand`/`TestAstCommand`)
- **Context Files to Reference:** `cmd/shpl/main.go` (current `astCmd` style, `PersistentPreRun` → `logger.Init`, `os.Exit(1)` on errors); the existing tests in `cmd/shpl/main_test.go` (pattern: `rootCmd.SetOut(&buf)`, `rootCmd.SetArgs(...)`, `rootCmd.Execute()` with reset).
- **Implementation Details:**
  - Add imports: `"github.com/lorenzobandini/shakespeare-interpreter-go/internal/semantic"`, `"github.com/lorenzobandini/shakespeare-interpreter-go/internal/runtime"`.
  - Declare `var runCmd = &cobra.Command{Use:"run <file>", Short:"Execute an SPL source file", Args: cobra.ExactArgs(1), RunE: ...}`.
  - RunE body (pseudo steps):
    1. `src, err := os.ReadFile(args[0])` → on error wrap + return (the rootCmd error handler prints + exits 1).
    2. `tokens, err := lexer.New(string(src)).ScanTokens()` → on err `fmt.Fprintln(os.Stderr, err); os.Exit(1)`.
    3. `prog, err := parser.New(tokens).Parse()` → same.
    4. `res := semantic.New(args[0], prog).Analyze(prog)`; if `!res.OK()` → print every `res.Errors[i].Error()` to stderr (each formatted per taxonomy), then `os.Exit(1)`.
    5. `err = runtime.Execute(prog, res, os.Stdin, cmd.OutOrStdout(), args[0])` → if `err != nil` → `fmt.Fprintln(os.Stderr, err); os.Exit(1)`.
  - Register: `rootCmd.AddCommand(tokensCmd, astCmd, runCmd)` in `init()`.
  - Logger init is already handled by the existing `PersistentPreRun` (Calls `logger.Init(LevelDebug)` if `--debug`); the runtime's `slog.Debug` calls will be silent by default and verbose under `--debug`.
- **Error Handling:** All pipeline errors route through `os.Stderr` + `os.Exit(1)`, matching the existing convention. Runtime `RuntimeError` is printed by `fmt.Fprintln` which calls its `Error()` method (yields the `error[Rxxx]: ...` format).
- **Testing & Verification:**
  - `TestRunCommandHello`: write `testdata/interpreter/hello.shpl` to a temp file (or use the real path), `rootCmd.SetArgs([]string{"run", "<path>"})`, `rootCmd.SetOut(&buf)`, `rootCmd.Execute()`. Assert `buf.String() == string(mustRead(t, "testdata/interpreter/hello.golden"))` and that `Execute()` returned `nil`. Reset `rootCmd.out` after.
  - `TestRunCommandSemanticError`: run a fixture that triggers an M-code (e.g. `testdata/semantic/m004-speaker-not-on-stage.shpl`); assert the command exit code is non-zero and stderr starts with `error[M004]`. Since `os.Exit(1)` aborts the test process, use the standard cobra test pattern: snapshot stderr via a redefinition or use `cobra.Command.SetErr(&buf)` and refactor the RunE to `return fmt.Errorf(...)` instead of `os.Exit(1)`. **Decision:** prefer `return fmt.Errorf("%v", err)` over `os.Exit(1)` inside RunE (cobra prints the returned error via rootCmd's Execute error handler, which is already the pattern at main.go:88-92). This makes the command testable. Update `tokensCmd`/`astCmd` only if needed — but minimal change: keep their existing `os.Exit(1)` and route `runCmd` through `return fmt.Errorf`. Document the divergence in a comment.
  - `TestRunCommandRuntimeError`: run `testdata/interpreter/divzero.shpl`; assert returned `error` wraps an R001 and that stderr contains `error[R001]`.
  - Run `go test -race ./cmd/shpl/...` → PASS.
- **Documentation Update:** `PROGRESS.md` bullet noting the `run` subcommand and the RunE-returns-error (testable) divergence from the older `os.Exit(1)`-style subcommands.

---

## Step 4.11: Canonical fixtures + golden-file tests + `task check` gate + spec handoff

- **Goal:** Lock in the canonical SPL output via golden files, drive end-to-end coverage through the two reference programs and a handful of feature-specific fixtures, then run the full `task check` gate and update all tracking docs.
- **Files:**
  - Create: `testdata/interpreter/{hello,truth-machine,stack,branch,divzero,io-ascii}.shpl` + the corresponding `.golden`/`.stdin` files
  - Test: `internal/runtime/golden_test.go` (and/or extend `cmd/shpl/main_test.go`)
- **Context Files to Reference:** `docs/SPL_SPECIFICATION.md` (Hello World source lines 156-238; Truth Machine source lines 240-265); `testdata/semantic/hello.shpl` and `testdata/semantic/truth-machine.shpl` (canonical content to copy); the canonical output of the reference SPL interpreter at `shakespearelang.com` (the implementer should visit it once to confirm exact byte output before pinning golden files, to avoid baking in a wrong assumption).
- **Implementation Details:**
  - For each fixture pair (`.shpl` source + `.golden` expected stdout, optionally `.stdin`), write a table-driven `TestGoldens` subtest:
    1. `src := mustRead(t, "testdata/interpreter/<case>.shpl")`.
    2. `tokens, _ := lexer.New(src).ScanTokens(); prog, _ := parser.New(tokens).Parse(); res := semantic.New("<case>.shpl", prog).Analyze(prog)`.
    3. `in := io.Reader(&bytes.Buffer{})`; if a `.stdin` exists, `in = bytes.NewReader(mustRead(stdinPath))`.
    4. `out := &bytes.Buffer{}; err := runtime.Execute(prog, res, in, out, "<case>.shpl")`.
    5. For the **non-error** fixtures (hello, truth-machine, stack, branch, io-ascii): assert `err == nil` and `out.String() == string(mustRead(t, "<case>.golden"))`.
    6. For the **error** fixture (divzero): assert `err != nil`, `errors.As(err, &re)` of type `*RuntimeError` or `RuntimeError` with `re.Code == "R001"`, and `out.String()` matches whatever partial output was emitted before the error (golden optional — pin only the error if needed).
  - Fixture content plan:
    - `hello.shpl` — copy verbatim from `testdata/semantic/hello.shpl` (the canonical Hello World). `hello.golden` — exactly the bytes `Hello World!` (no trailing newline; verify against `shakespearelang.com` reference once). **If the reference emits a trailing newline, regenerate the golden to match — the golden file is authoritative, not this prose.**
    - `truth-machine.shpl` — copy from `testdata/semantic/truth-machine.shpl`. `truth0.stdin` = `"0\n"`, `truth0.golden` = `"0\n"`. `truth1.stdin` = `"1\n"`; `truth1.golden` = `"1\n"` repeated 8 times — but the canonical truth machine loops forever on 1. For a deterministic test, replace the truth-machine stdin with a Reader that yields `"1\n"` once then EOF; assert `Execute` returns R003 after the second `Listen` attempt, and that `out` contains `"1\n"` repeated *N* times where *N* is however many iterations ran before the EOF. Pin `N` exactly by reading the actual run once and recording it in a comment on the test. (Pin the count to the observed value, not an assumed one.)
    - `stack.shpl` — short program exercising `Remember` (push a value) and `Recall` (pop into the other character) and `Open your heart` (output the popped value). `stack.golden` pins the decimal output.
    - `branch.shpl` — small program with one `Am I as good as you?` + both `If so` and `If not` branches (using two distinct scenes to assert each branch lands on the right scene). `branch.golden` pins the output of the taken branch.
    - `io-ascii.shpl` — `Open your mind.` (read an ASCII char from stdin) followed by `Speak your mind.` (echo it). Drive with stdin `"X"` → expected stdout `"X"`.
    - `divzero.shpl` — small program that reaches `the quotient between <X> and a coward` with the divisor evaluating to 0 → R001.
  - **Golden regeneration policy:** golden files are committed; if an implementation tweak changes the byte output for a non-error fixture, that is a regression and the test must fail. Regenerate a golden only when (a) the prior content was demonstrably wrong against the canonical reference, and (b) the change is recorded in `PROGRESS.md`.
- **Error Handling:** R-codes are expected and asserted for `divzero` and (transiently) `truth1`.
- **Testing & Verification:**
  - Run `go test -race -run TestGoldens ./internal/runtime/...` → all PASS.
  - Run the CLI test for `run` on `hello.shpl` and `truth0.stdin` → PASS.
  - Run `go test -race -coverprofile=coverage.out ./internal/runtime/...` → all PASS; aim for ≥ 90% statement coverage (Phase 3 hit 97.9%).
  - **Gate:** `task check` (fmt → lint → vuln → test) → must pass clean. Address every lint finding (`errcheck` on the `stage.Enter/Exit/Exeunt` discarded errors needs an explicit `_ =` or a linter-approved ignore comment — do NOT add `// nolint` unless the project already permits it; prefer `_, _ = e.stage.Enter(...)` capturing the return as discarded).
- **Documentation Update:**
  - Fill the "Phase 4 — Runtime / Evaluator" section of `PROGRESS.md`: check off every sub-bullet currently listed (character state, stage manager, expression evaluator, I/O, control flow, program counter); add the 11 new fixtures to a Fixtures table; record decisions R-D1..R-D8 under "Decisions"; mark the `CLI subcommand: shpl run` bullet done.
  - `docs/ERROR_TAXONOMY.md`: add a note under Level 4 that R004 is currently unreachable (R-D6 except the factorial-detected case) and that R001 fires on both `quotient` and `remainder` operations on zero.
  - `README.md`: add the `shpl run <file>` example to the quick-start section.

---

## Self-Review Checklist (run before declaring Phase 4 complete)

1. **Spec coverage — Runtime Environment State:**
   - Character value store (`name → int`) → 4.1 (struct), 4.2/4.3 (reads), 4.5 (writes), 4.6 (Recall writes). ✓
   - Stack storage (`name → []int`) → 4.1 (struct), 4.6 (push/pop). ✓
   - Stage tracking → 4.1 (`stage` field), 4.4 (Enter/Exit/Exeunt mutation, listener derivation). ✓
   - Execution pointers for Acts/Scenes → 4.8 (flatten + actLabel + per-act sceneLabels + integer PC trampoline). ✓
2. **Spec coverage — Expression Evaluation Engine:**
   - Constant word-value calculation → 4.2 (Polarity × 2^AdjectiveCount). ✓
   - Unary/Binary math operations → 4.3 (sum/difference/product/quotient/remainder, square/cube/square_root/factorial/twice). ✓
   - Similes → already stripped by the parser to `Expr`; `eval` consumes the inner expression. Documented in 4.3. ✓
   - Pronoun resolution based on current Speaker/Listener → 4.3 (`PronounExpr.Ref` switch on `speaker`/`listener`). ✓
3. **Spec coverage — Statement Execution:**
   - Dialogue processing → 4.8 (flatten attaches `speaker` per inner statement; listener re-derived per instruction in `execExec`). ✓
   - Assignment → 4.5 (`AssignStmt` writes to listener). ✓
   - Stack operations → 4.6 (`Remember` push listener, `Recall` pop listener → speaker). ✓
   - I/O (Speak/OpenHeart as integers or characters; OpenMind/Listen reads via abstracted reader/writer) → 4.5. ✓
4. **Spec coverage — Control Flow & Trampoline:**
   - GOTO (`Let us proceed to / let us return to`) → 4.7/4.8 (`GotoStmt` → `(pc, true, nil)`; trampoline applies). ✓
   - Conditional `If so` / `If not` → 4.7 (comparison flag + `BranchIfTrue`). ✓
   - Program Counter / index manipulation using semantic registries → 4.7/4.8 (`actLabel`/`sceneLabels` derived from `res.Acts` and the flattened per-act scene registry). ✓
5. **Negative constraint:** No production code is in this plan — only struct field sketches, interface signatures, and numbered algorithmic pseudo steps. ✓
6. **Logic precision:** Each step describes exactly how the value store, stack store, stage, comparison flag, and program counter change. ✓
7. **Test-driven:** Every step ends with explicit table-driven tests and named fixtures (`testdata/interpreter/`). ✓
8. **Context alignment:** Every step lists the spec/taxonomy files to consult and reminds the implementer to update `PROGRESS.md`. ✓
9. **Type consistency:** `instr` struct fields (`stmt`, `speaker`, `sceneRoman`, `actRoman`) are identical across 4.4–4.8. `env` field names (`values`, `stacks`, `stage`, `syms`, `comparison`, `acts`, `sceneLabels`, `actLabel`, `currentActRoman`, `currentSceneRoman`) are consistent across 4.1–4.9. `applyRelation`, `resolveJump`, `listener` helpers are introduced exactly once. ✓
10. **Gate:** `task check` must pass clean at the end of 4.11. ✓

## Execution Handoff

Plan complete. Save this content to `docs/superpowers/plans/2026-07-12-phase4-runtime.md`.

**Required first action (Phase 4):** Step 4.1 — scaffold `internal/runtime/` with the `RuntimeError` type, `env` struct, and `NewEnv` constructor; add a `TestMain` with `logger.Init(logger.LevelDebug)`.

Then proceed through **Steps 4.2 → 4.11** in order, each with its own test gate. Gate the entire phase with `task check` at the end of 4.11.
