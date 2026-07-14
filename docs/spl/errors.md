# SPL Error Taxonomy

Categorizing all errors before implementation ensures consistent error messages
and makes it obvious which phase (lexer, parser, semantic, runtime) owns each check.

## Level 1 — Lexical Errors

Produced by `internal/lexer`. Invalid characters or tokens at the character level.

| Code | Name | Example | Message |
|------|------|---------|---------|
| `L001` | UnexpectedCharacter | source contains a control character (e.g. NUL, BEL) | `unexpected control character 0x%02X at line X, col Y` |
| `L002` | UnterminatedToken | `[Enter Romeo` (no closing `]`) | `unterminated stage direction starting at line X` |

SPL has very few lexical errors — the grammar is mostly words and punctuation.
Most invalid constructs surface at the parser level. **L001 fires only for control
characters** (0x00–0x08, 0x0B, 0x0C, 0x0E–0x1F, 0x7F). Other printable characters
that are not structural punctuation fold into `WORD` tokens and surface as S-code
parse errors. This allows character descriptions to contain free-text punctuation
(e.g., `A/S` in `Hamlet, the flatterer of Andersen Insulting A/S.`).

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
| `S012` | InvalidExeunt | `[Exeunt ,]` (malformed Exeunt args) | `expected character name or ']' after 'Exeunt'` |
| `S013` | MissingStage | Dialogue without `[Enter ...]` | `expected [Enter ...] before dialogue` |
| `S014` | MissingSpeaker | Line without `Name:` prefix | `expected character name followed by ':'` |
| `S015` | InvalidExpression | `You are the sum of` (incomplete) | `expected expression` |
| `S016` | InvalidIf | `If so, let us proceed to scene` (missing scene) | `expected scene or act number after 'proceed to'` |
| `S017` | InvalidComparative | `Am I good you?` (missing `as`/`than`) | `expected comparative phrase (e.g., 'as good as', 'better than')` |
| `S018` | InvalidStackOp | `Remember .` (no expression) or `Recall your fate` (no `.`) | `expected expression after 'Remember'` / `expected '.' after 'Recall'` |

> **Note on S012 (Exeunt):** Both bare `[Exeunt]` (exits all characters on stage) and
> `[Exeunt A and B]` (exits the named characters) are valid SPL. S012 fires only for
> *malformed* Exeunt syntax (e.g., `[Exeunt ,]` or `[Exeunt and]`). This is the
> canonical SPL behavior confirmed against the official Hello World Program, which
> contains `[Exeunt Ophelia and Hamlet]`.

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
| `R004` | IntegerOverflow | Value exceeds int bounds | `integer overflow in 'factorial'` |

**Runtime notes:**
- R001 fires on both `quotient` and `remainder` binary operations when the divisor evaluates to zero.
- R004 currently fires only for `factorial` of values > 20 (the one operation where overflow is detectable and catastrophic). Other overflows silently wrap per R-D6.

## Error Format

All errors follow this format:

```
error[<CODE>]: <message>
  --> <file>:<line>:<col>
```

Example:
```
error[M001]: character 'Banquo' is not declared
  --> program.spl:3:1
```

## Severity

- **Lexical/Syntax errors** → stop pipeline, no further processing
- **Semantic errors** → stop pipeline after analysis, no execution
- **Runtime errors** → stop execution with error output to stderr
