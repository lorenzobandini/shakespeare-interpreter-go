# SPL Error Taxonomy

Categorizing all errors before implementation ensures consistent error messages
and makes it obvious which phase (lexer, parser, semantic, runtime) owns each check.

## Level 1 — Lexical Errors

Produced by `internal/lexer`. Invalid characters or tokens at the character level.

| Code | Name | Example | Message |
|------|------|---------|---------|
| `L001` | UnexpectedCharacter | `Romeo @ Juliet` | `unexpected character '@' at line X, col Y` |
| `L002` | UnterminatedToken | *(if lexer finds incomplete token)* | `unterminated token starting at line X` |

SPL has very few lexical errors — the grammar is mostly words and punctuation.
Most invalid constructs surface at the parser level.

## Level 2 — Syntax Errors

Produced by `internal/parser`. Valid tokens in invalid arrangement.

| Code | Name | Example | Message |
|------|------|---------|---------|
| `S001` | MissingTitle | Program starts without title line + `.` | `expected program title ending with '.'` |
| `S002` | MissingCharacterDecl | `Act I:` with no prior character declarations | `expected at least one character declaration before Act I` |
| `S003` | InvalidCharacterName | `Mario, a plumber.` (not a Shakespeare character) | *warning, not error — spec doesn't enforce this strictly* |
| `S004` | MissingAct | Characters declared but no `Act` found | `expected at least one act` |
| `S005` | InvalidActNumber | `Act 1:` (not Roman numeral) | `expected Roman numeral after 'Act', got '1'` |
| `S006` | ActOrder | `Act III:` after `Act I:` (skips II) | `act numbers must be sequential, expected Act II` |
| `S007` | MissingScene | Act with no `Scene` inside | `expected at least one scene in Act I` |
| `S008` | InvalidSceneNumber | `Scene 1:` (not Roman numeral) | `expected Roman numeral after 'Scene'` |
| `S009` | SceneOrder | `Scene III:` after `Scene I:` | `scene numbers must be sequential` |
| `S010` | MissingEnter | Character speaks without `[Enter ...]` | `character 'Romeo' must enter before speaking` |
| `S011` | InvalidEnter | `[Enter]` with no character | `expected character name after 'Enter'` |
| `S012` | InvalidExeunt | `[Exeunt Romeo]` — Exeunt takes no args | `'Exeunt' must stand alone, use 'Exit' for single character` |
| `S013` | MissingStage | Dialogue without `[Enter ...]` | `expected [Enter ...] before dialogue` |
| `S014` | MissingSpeaker | Line without `Name:` prefix | `expected character name followed by ':'` |
| `S015` | InvalidExpression | `You are the sum of` (incomplete) | `expected expression after 'sum of'` |
| `S016` | InvalidIf | `If so, let us proceed to scene` (missing scene) | `expected scene number after 'proceed to scene'` |

## Level 3 — Semantic Errors

Produced by the semantic analyzer (`internal/semantic/` or validator pass).
Valid syntax but invalid meaning.

| Code | Name | Example | Message |
|------|------|---------|---------|
| `M001` | UndefinedCharacter | `[Enter Banquo]` but Banquo not declared | `character 'Banquo' is not declared` |
| `M002` | TooManyOnStage | `[Enter Romeo, Juliet, Hamlet]` | `too many characters on stage (max 2), got 3` |
| `M003` | StageOverflow | `[Enter Romeo]` when two already on stage | `cannot enter: stage is full (Romeo, Juliet already on stage)` |
| `M004` | CharacterNotOnStage | Speaking to Romeo when Romeo not on stage | `character 'Romeo' is not on stage` |
| `M005` | ExitNotOnStage | `[Exit Hamlet]` but Hamlet not on stage | `character 'Hamlet' is not on stage` |
| `M006` | UndefinedScene | `let us proceed to scene V` (scene V doesn't exist) | `scene V is not defined` |
| `M007` | SelfReferenceEnter | `[Enter Romeo and Romeo]` | `cannot enter the same character twice` |
| `M008` | NoSceneInAct | Act exists but contains zero scenes | `Act I has no scenes` |

## Level 4 — Runtime Errors

Produced by `internal/runtime` during execution.

| Code | Name | Example | Message |
|------|------|---------|---------|
| `R001` | DivisionByZero | `the quotient between Hamlet and a coward` (Juliet = 0) | `division by zero at Act I, Scene II` |
| `R002` | InputNotANumber | `Listen to your heart.` but stdin is "hello" | `expected a number, got 'hello'` |
| `R003` | InputEOF | `Open your mind.` but stdin is exhausted | `unexpected end of input` |
| `R004` | IntegerOverflow | Value exceeds int bounds | *depends on implementation — Go ints wrap, may not error* |

## Error Format

All errors follow this format:

```
error[<CODE>]: <message>
  --> <file>:<line>:<col>
```

Example:
```
error[M001]: character 'Banquo' is not declared
  --> program.shpl:3:1
```

## Severity

- **Lexical/Syntax errors** → stop pipeline, no further processing
- **Semantic errors** → stop pipeline after analysis, no execution
- **Runtime errors** → stop execution with error output to stderr
