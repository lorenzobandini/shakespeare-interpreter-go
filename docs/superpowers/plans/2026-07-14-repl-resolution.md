# Plan: REPL Resolution (2026-07-14)

## Objective

Fix the interactive REPL so multi-line, incremental, and pasted-block entry works without spurious S002/S006/S007/S009/L002/M004 errors. All changes are confined to `cmd/shpl/main.go` REPL state management; the lexer/parser/semantic/runtime pipeline (`runPipeline`) is untouched. The replay-based accumulating-buffer model is retained (PROGRESS.md Phase 5 D1). The root defect is premature skeleton injection and mis-ordered auto-declaration splicing; secondary defects are heuristic incomplete-input detection and brittle rollback.

---

## Feasibility analysis 

**Verdict: VIABLE.** The failures in the logs are not caused by SPL grammatical incompatibility — they are localized bugs in the REPL's buffer/skeleton splicing logic. The core pipeline (`runPipeline`, lexer, parser, semantic, runtime) is sound and shared with `run`, which works flawlessly.

**Root cause of every logged failure** (traced via `cmd/shpl/main.go:300-385` `replayBlock` + `:307-316`):

The REPL auto-prepends a skeleton `"The REPL Session.\n\nAct I: The REPL Session.\nScene I: The REPL Session.\n\n"` into `rs.buffer` on the first submit *whenever the user's block lacks an `Act`/`Scene` header* (`hasActOrScene`, `main.go:287`). This is the bug:

1. **Log 1 & 2**: User submits `The Truth Machine.` alone. Skeleton (containing `Act I:`) is prepended *before* the title. Parser sees `Act I` with zero preceding character declarations → **S002**. Same for a lone `Romeo, a young man.` block — skeleton's Act I precedes the declaration.
2. **Log 4 (Stack Test)**: First submit is the title alone → S002 (skeleton before title). Subsequent submits append to a buffer whose prefix is already the *skeleton + title*, so the user's `Act I:` now conflicts with the skeleton's `Act I` (S006/S009 surface as the "expected II" variant later).
3. **Log 5 (Hello World)**: `Act II:` submitted *without a preceding `Scene I:` in the same block* → skeleton/accumulated buffer has Act II with no scene → **S007**. Then `Scene I:` after Act II exists → **S009** (parser expects Scene II, the user had already established Scene I in an earlier block).

**Why a clean REPL is architecturally compatible with SPL:**

- SPL top-level order is fixed (`Title → Chars → Act I → Scene I → body`), but the *accumulating-buffer replay model* (PROGRESS.md D1) sidesteps this: each submit re-runs the whole pipeline on the grown buffer. The only requirement is that, *after each successful submit*, `rs.buffer.String()` parses as a complete valid program. That is achievable with **phase-aware skeleton injection** (inject `Act I/Scene I` only after the user has finished the declaration phase) and **correct auto-declaration splice ordering** (insert user chars + auto-detected chars *before* Act I, never after).
- Runtime state persistence is *not* a blocker: `runtime.Execute` rebuilds `env` from scratch on each replay (`env.go:30`), re-deriving values/stacks/stage by re-executing the prefix. This is the documented O(n²) trade-off (PROGRESS.md T1) and is acceptable at human-typing scale. It requires no refactor of `internal/runtime` (unexported) — keeping the main interpreter untouched, per AGENTS.md.
- Jump mechanics (goto/if to scene labels) work under replay because the full program is always present in the buffer; labels resolve against the complete registry each replay.
- The O(n²) cost and infinite-loop hang are documented out-of-scope trade-offs (PROGRESS.md T2), not feasibility blockers.

**Why the alternative (true resumable `env` snapshotting) was rejected:** it requires refactoring `internal/runtime/env.go` (unexported `env`, `flatten`, trampoline PC) to support snapshot/restore across submits. That crosses the package boundary the AGENTS.md design protects and risks the main `run` path. The replay model avoids it entirely.

---

## Step 5.1: Phase-tracking state machine in `replState`

**Goal:** Replace the boolean `skeletonBuilt` flag with an explicit phase enum so the REPL knows whether the user has finished the declaration preamble and entered the act/scene body.

**Context Files to Reference:** `docs/spl/specification.md` §"Program Structure", `cmd/shpl/main.go` (replState struct ~:239, replayBlock ~:300).

**Implementation Details:**
- Add a `phase` field to `replState` of a new unexported type, e.g. `type replPhase uint8` with constants `phaseTitle`, `phaseChars`, `phaseBody`, `phaseClosed`.
- Initial phase: `phaseTitle`. Transitions:
  - `phaseTitle` → `phaseChars` when a line matching the character-declaration shape (`WORD , <non-empty description> .`) is seen, OR when any act header is seen (skip straight to `phaseBody`).
  - `phaseChars` → `phaseBody` when the first `Act <Roman>:` header is appended.
  - `phaseBody` → `phaseClosed` when `[Exeunt]` with no names is submitted (optional, informational only).
- `:reset` resets `phase` to `phaseTitle`.
- Phase is derived from the *pending block* (the local `block` plus the already-committed `rs.buffer`), never from speculative parse results.
- Define one helper `classifyLine(line string) replLineKind` returning `{titleLine, charDeclLine, actHeader, sceneHeader, stageDir, dialogue, blank, other}`. The classification uses cheap substring/regex checks, NOT full parsing. This is the single source of truth for phase transitions.

**Error Handling:** `classifyLine` must never panic on malformed input; `other` is the safe fallback (keeps block growing).

**Testing & Verification:** Unit test `TestClassifyLine` (table-driven) and `TestReplPhaseTransitions` feeding scripted lines through an in-memory REPL and asserting `rs.phase` after each submit.

**Documentation Update:** Add a "Phase model" subsection to PROGRESS.md Phase 5 decisions.

---

## Step 5.2: Lazy, correctly-placed skeleton injection

**Goal:** Eliminate the S002 avalanche. The skeleton (`Act I: … / Scene I: …`) must be injected only when the user transitions into `phaseBody`, and it must be spliced into `rs.buffer` *after* all character declarations — never before them. Users who lack character declarations before their first act header receive an honest S002 with buffer rollback — no phantom characters invented by the REPL.

**Context Files to Reference:** `cmd/shpl/main.go` replayBlock ~:307-316, hasActOrScene ~:287; `docs/spl/specification.md` §"Character Declarations".

**Implementation Details:**
- Remove the unconditional `if !rs.skeletonBuilt` prepend at `main.go:307-316`.
- New injection rule, applied once per session, at the first submit where `phase == phaseBody`:
  1. Compute the split index `declEnd` = byte offset in `rs.buffer` immediately after the last character-declaration line (scan by line, using `classifyLine`).
  2. Construct the skeleton string `skeleton = "Act I: The REPL Session.\nScene I: The REPL Session.\n\n"`.
  3. Rebuild the buffer as `rs.buffer[:declEnd] + skeleton + rs.buffer[declEnd:]`.
  4. If the user's own block already contains an `Act <Roman>:` header, do NOT inject the skeleton — the user is providing their own act structure. (This preserves the "type the whole program" workflow that already works, per log 3.)
- Do **not** inject the skeleton if the pending block transitions to `phaseBody` without any preceding character declarations in `rs.buffer`. In this case, submit the buffer as-is; S002 is the genuine, instructive error. Roll back `rs.buffer` to the pre-submit checkpoint, report `error[S002]`, and continue the session at the next `spl>` prompt.
- Rationale for no synthetic character: inventing a character the user's own text never references breaks the "faithful SPL playground, no invented behavior" principle established for this REPL's design and risks silent collisions (user later declares a differently-described `Romeo`). The auto-declaration in Step 5.3 is scoped to names the user's *own text* already references — that is automating requested boilerplate, which is distinct.
- `:reset` clears the injected skeleton byte range as part of the buffer reset.

**Error Handling:** If the splice would place `Act I` before a character declaration that the user typed in an *earlier* committed block, prefer the user's ordering: insert the skeleton at the *end* of the committed character-declaration block. Never reorder user text. The S002 rollback path is the same as any pipeline error: undo the buffer addition, report, prompt again.

**Testing & Verification:** `TestRepl_SkeletonInjectionOrder` — feed title, then two char decls across separate submits, then `Act I:` in a third submit; assert the final `rs.buffer` parses cleanly via a direct `runPipeline` call in the test (no S002). Add a second case: title + `Act I:` with no char decl → asserts S002 reported once and `rs.buffer` rolled back (empty), confirming the REPL neither masks the error nor injects a phantom character. Regression test that mirrors *each* of the 5 user logs, asserting no S002/S006/S007/S009.

**Documentation Update:** Update PROGRESS.md D4 (skeleton semantics) to describe phase-gated injection and the no-synthetic-char invariant.

---

## Step 5.3: Correct auto-declaration splice ordering

**Goal:** Auto-detected character declarations (from `[Enter]`/`[Exit]`/`[Exeunt]`/speaker prefixes) must be inserted *before* `Act I`, merged with the user's explicit declarations, deduplicated case-insensitively, and never inserted after an act header.

**Context Files to Reference:** `cmd/shpl/main.go` extractChars ~:254, auto-decl splicing ~:321-348; `internal/semantic/symbol_table.go` (M001 source).

**Implementation Details:**
- Change `extractChars` to return an ordered slice (not just a map) preserving first-seen order, so spliced declarations have deterministic positions.
- During the submit that transitions to/in `phaseBody`, compute `autoDecls = extractChars(block) ∪ extractChars(rs.buffer)` minus names already present in the explicit declaration region of `rs.buffer`.
- Splice point: the byte offset *immediately before* the skeleton's `Act I:` (or the user's first `Act` header). Use the `declEnd` index from Step 5.2.
- Emit each missing declaration as `"<Name>, a character.\n"` (single canonical description; the description text is free-form per spec and unused at runtime).
- Dedup case-insensitively against both explicit declarations and already-spliced auto declarations.
- Re-run splicing on every submit (declarations may appear in later blocks via new `[Enter X]` entries); it is idempotent because of the dedup step.
- **Scope boundary:** this step only invents declarations for names the user's text *already references* in stage directions or speaker prefixes. It does not invent characters from nothing — that is Step 5.2's explicit boundary.

**Error Handling:** If a name first appears as a speaker prefix but is never entered, the auto-declaration still splices it; the resulting M004 (speaker not on stage) is the real semantic error and is reported normally — do not suppress it.

**Testing & Verification:** `TestRepl_AutoDeclarationMidProgram` — start with a body that uses only `Romeo`, then in a later block submit `[Enter Juliet]` + `Juliet: You are a flower.`; assert no M001/M004 and that `Juliet` appears declared before `Act I` in the buffer.

**Documentation Update:** Refine PROGRESS.md D5 with the ordered/deduped splice semantics and the scope boundary.

---

## Step 5.4: Robust incomplete-input detection (replace `tryQuickParse` heuristics)

**Goal:** Stop classifying genuine syntax errors as `ErrIncomplete` and stop reporting `ErrIncomplete` as a fatal error. The REPL must keep growing the block only when the partial block *plausibly* needs more input; otherwise report the genuine error immediately (but without breaking the session).

**Context Files to Reference:** `cmd/shpl/main.go` tryQuickParse ~:505-560, ErrIncomplete ~; `internal/parser/errors.go` (S-code set); `internal/lexer/lexer.go` L002.

**Implementation Details:**
- Redefine `ErrIncomplete` to mean *only*: (a) L002 unterminated bracket at EOF, (b) a trivially short block (< 1 non-blank line) the phase model has not yet classified, (c) an open multi-line expression (heuristic: the last sentence lacks a terminating `.`/`!`/`?` AND the line count since the last terminator is below a small cap, e.g. 40 lines — guard against unbounded growth).
- Remove the `strings.Contains(s, "S013"|"S006"|"S009")` reclassification block and the `!p.Done()` blanket `ErrIncomplete`. These two heuristics are what produced the false negatives (e.g. log 5's `Act II` with no scene being silently swallowed) AND the false S002 surfacing on submit.
- For `tryQuickParse`'s genuine-error path: print the error to stderr immediately and **truncate the block back to the checkpoint** (`main.go:666-667` already does this) — but do NOT touch `rs.buffer`. The session continues at the next `spl>` prompt with the bad line discarded. This is the existing rollback path; keep it.
- On the final blank-line submit, run the full pipeline. If it errors, roll back `rs.buffer` to the checkpoint (existing behavior at `main.go:362-373`) and report; the session continues. No change needed to that path beyond Steps 5.2/5.3 making the spliced buffer actually valid.

**Error Handling:** A genuine parse error during the quick-pass is reported to the user but the REPL stays alive (`rs.buffer` untouched). A genuine error at submit time rolls back the buffer (so the next prompt reflects the last-known-good state) and reports. Neither terminates the loop.

**Testing & Verification:** Extend `TestTryQuickParse_ErrorLines` (`main_test.go:381`) — currently all 12 cases expect `ErrIncomplete`. Split into two tables: genuine-error cases (assert a non-nil non-`ErrIncomplete` error) and incomplete cases. Add cases for: `Act II:` with no scene (genuine S007), `Scene I:` after a prior committed `Scene I` (genuine S009), `[Enter` unterminated (incomplete L002).

**Documentation Update:** Replace PROGRESS.md D1's "lenient quickParse" note with the precise incomplete criteria.

---

## Step 5.5: Rollback correctness and output-delta invariant

**Goal:** Guarantee that a submit which fails *after* the pipeline has already written partial output to `captureOut` does not corrupt the `lastOutputLen` cursor, and document the deterministic prefix invariant under which the delta slice is valid.

**Context Files to Reference:** `cmd/shpl/main.go` replayBlock ~:359-383, singleReader ~:400-464; PROGRESS.md D6.

**Implementation Details:**
- Wrap `captureOut` in a counting writer; record `bytesWrittenThisSubmit`. On the success path, slice `delta = captureOut.Bytes()[rs.lastOutputLen:]` then set `rs.lastOutputLen = captureOut.Len()` (existing). On the error path, **do not** advance `rs.lastOutputLen`; also roll back `singleReader.recorded`/`cursor` to the pre-submit checkpoint (existing) so the next replay does not double-read runtime stdin.
- Add a class-level invariant comment on `replState` stating the precondition the delta slice relies on: re-execution of `rs.buffer[:oldLen]` must reproduce `captureout[:rs.lastOutputLen]` byte-for-byte. This holds because (i) the lexer/parser/semantic are pure functions of source, (ii) the runtime is deterministic given source + recorded stdin, and (iii) the buffer only grows on the success path. Any future feature that violates (ii) (e.g. wall-clock reads) must be gated out of the REPL.
- Add a defensive guard: if, on the success path, `bytes.Equal(captureOut.Bytes()[:rs.lastOutputLen], expectedPrefix)` would be too expensive, at minimum assert `captureOut.Len() >= rs.lastOutputLen` and on violation, treat as an error (roll back, report "internal: output prefix violated"). Cheap and catches nondeterminism early.

**Error Handling:** A violated prefix invariant is an internal error: roll back the buffer, report `error: internal output-prefix invariant violated`, and continue. Do not emit the delta slice.

**Testing & Verification:** `TestRepl_PartialOutputRollback` — craft a block whose first pipeline stage writes output then a later stage errors; assert `rs.lastOutputLen` is unchanged and the next successful submit's delta is correct. `TestRepl_StdinReplayDeterminism` (extend existing `main_test.go:521`) — two submits with the same recorded stdin produce identical prefixes.

**Documentation Update:** Add D8 to PROGRESS.md Phase 5: "Output-prefix invariant and rollback contract".

---

## Step 5.6: REPL test fixtures and experimental reproductions

**Goal:** Lock the 5 user-reported failure logs as regression tests so the fixes can never regress.

**Context Files to Reference:** `cmd/shpl/main_test.go` (existing REPL tests ~:260-521); `testdata/` (currently has no repl dir).

**Implementation Details:**
- Add `testdata/repl/` with one `.txt` fixture per user log (logs 1, 2, 3, 4, 5) containing the *exact* scripted stdin sequence (blank lines preserved as `\n` sentinels).
- Add a table-driven test `TestRepl_UserLogRegressions` that streams each fixture through an in-memory `singleReader` + `replState` and asserts: (a) no `S002` ever appears, (b) the stack-test log (log 3/4) reproduces `1\n0\n0\n` exactly, (c) the hello-world log reaches the `[Exeunt]` close with zero unrecoverable errors, (d) the truth-machine log accepts the full program across multiple submits.
- Add a helper `feedScript(t *testing.T, script string) (output string, stderr string)` that drives the REPL input loop with a pre-loaded `strings.NewReader(script)` as both SPL stdin and the command stream, returning captured out/err. Reuse `singleReader` so runtime stdin semantics are preserved.

**Error Handling:** The test asserts that genuine errors (where the logs show them) are reported exactly once per occurrence and do not terminate the loop.

**Testing & Verification:** `go test -race -run TestRepl_UserLogRegressions ./cmd/shpl/...`. Then `task check`.

**Documentation Update:** Append a "REPL regression fixtures" note to PROGRESS.md Phase 5; reference this plan file from the Phase 7 polish entry.

---

## Step 5.7: Final verification gate

**Goal:** Confirm the entire fix passes the project's single CI gate and that `run` is unaffected.

**Context Files to Reference:** `AGENTS.md` (Commands section), `Taskfile.yaml`.

**Implementation Details:**
- Run `task check` (fmt → lint → vuln → test). All must pass.
- Smoke-run the 5 user logs manually against `./bin/shpl.exe repl` and confirm the behavior matches the test assertions.
- Confirm `./bin/shpl.exe run testdata/runtime/stack.shpl` and the canonical hello-world still produce identical output (no pipeline change expected — guard only).

**Error Handling:** N/A (verification step).

**Testing & Verification:** `task check` green; manual smoke log signed off in PROGRESS.md.

**Documentation Update:** Mark the REPL resolution complete in PROGRESS.md; reference this plan and list the user logs as resolved. Also explicitly flag `:softreset` as a candidate future enhancement (clears the post-declaration buffer, keeps declared characters, for iterative scene editing) with rationale: defer until the base declaration→act→scene→body flow is proven solid via the Step 5.6 regression fixtures. Do not implement this cycle.
