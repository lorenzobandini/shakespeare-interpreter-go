# Shakespeare Programming Language (SPL) Specification

Reference extracted from the [SPL manual](http://shakespearelang.com/) and the [Esolang wiki](https://esolangs.org/wiki/Shakespeare).

## Program Structure

An SPL program is a play script divided into:

### 1. Title
First line of the program. Any characters except newline and `.`, terminated by `.`.
```
Romeo and Juliet: A Tragic Computation.
```

### 2. Character Declarations (Dramatis Personae)
```
Name, description.
```
- Names must be characters from Shakespeare plays.
- At least one non-whitespace character required in the description.
- Characters are variables (signed integers, initialized to 0).
```
Romeo, a young man with a remarkable patience.
Juliet, a likewise young woman of remarkable grace.
```

### 3. Acts
At least one act required. Act numbers use Roman numerals, starting at I.
```
Act I: Description of the act.
Act II: Another act.
```

### 4. Scenes
At least one scene per act. Scenes are goto labels.
```
Scene I: Description of the scene.
Scene II: Another scene.
```

## Stage Management

Characters must be on stage before they can speak or be referenced.
Maximum two characters on stage at any time.

```
[Enter Romeo]                     Enter single character
[Enter Romeo and Juliet]          Enter two characters
[Exit Romeo]                      Remove one character
[Exeunt]                          Remove all characters
```

## Variables and Values

Characters hold signed integer values, initialized to 0.

### Constant Expressions
Built from nouns (value 1 or -1) and adjectives (double the value):
- **Positive nouns**: any noun with positive connotation → 1
  - E.g., `flower`, `hero`, `king`, `angel`, `summer's day`
- **Negative nouns**: any noun with negative connotation → -1
  - E.g., `coward`, `liar`, `fool`, `pig`
- **Neutral nouns**: nouns with neither connotation → 1 (default)

Each adjective doubles the value:
```
flower          = 1
red flower      = 2
red hot flower  = 4   (2 adjectives: red, hot)
coward          = -1
big coward      = -2
```

### Similes
`as adjective as` — evaluates to 0 but maintains grammatical flow:
```
You are as lovely as a summer's day.       → character = 0 (simile alone)
You are as lovely as Romeo.                → character = value of Romeo
```

### Operations
```
the sum of A and B              → A + B
the difference between A and B  → A - B   (order matters!)
the product of A and B          → A * B
the quotient between A and B    → A / B   (integer division)
the remainder of the quotient between A and B  → A % B
the square of A                 → A * A
the cube of A                   → A * A * A
twice A                         → A * 2
```

Operations can be nested:
```
the sum of Romeo and the difference between Juliet and a flower
```

## Assignment

The general form: `You are <expression>.`
```
You are a vile coward.                      → value = -4  (vile = 2 adj, coward = -1)
You are as brave as Hamlet.                 → value = Hamlet's value
You are the sum of Romeo and a flower.      → value = Romeo + 1
```

### Pronoun Reference
- `yourself` / `thyself` — refers to character being spoken to (must be on stage)
- Character name — refers to any declared character

## I/O Commands

Commands spoken by one character to another:

| Command | Meaning |
|---------|---------|
| `Speak your mind.` / `Speak thy mind.` | Output ASCII char of character being spoken to |
| `Open your heart.` / `Open thy heart.` | Output numeric value of character being spoken to |
| `Open your mind.` / `Open thy mind.` | Read ASCII char from input, store in character being spoken to |
| `Listen to your heart.` / `Listen to thy heart.` | Read number from input, store in character being spoken to |

## Control Flow

### Conditional Jumps
```
Am I better than you?           → Is speaker > listener?
Am I as good as you?            → Is speaker == listener?
Am I worse than you?            → Is speaker < listener?
```

To act on the comparison:
```
Is X better than Y?
If so, let us proceed to scene III.    → jump if X > Y
If not, let us proceed to scene IV.    → jump if X <= Y
```

### Goto
```
Let us proceed to scene I.
Let us return to scene II.
```

## Grammar Notes

- `you` and `thou` (and their forms `your`/`thy`, `yourself`/`thyself`) are interchangeable.
- Sentences must maintain Shakespearean grammar.
- Dialogue lines are assigned to a speaker with `Name:` prefix.

## Computational Class

SPL is Turing-complete. Every SPL program can be converted from brainfuck (see [Brain2Speare](https://github.com/mjdarby/Brain2Speare)).

## Examples

### Hello World
```
The Infamous Hello World Program.

Romeo, a young man with a remarkable patience.
Juliet, a likewise young woman of remarkable grace.
Ophelia, a remarkable woman much in dispute with Hamlet.
Hamlet, the flatterer of Andersen Insulting A/S.

Act I: Hamlet's insults and flattery.
Scene I: The insulting of Romeo.

[Enter Hamlet and Romeo]

Hamlet:
 You lying stupid fatherless big smelly half-witted coward!
 You are as stupid as the difference between a handsome rich brave
 hero and thyself! Speak your mind!
 You are as brave as the sum of your fat little stuffed misused dusty
 old rotten codpiece and a beautiful fair warm peaceful sunny summer's
 day. You are as healthy as the difference between the sum of the
 sweetest reddest rose and my father and yourself! Speak your mind!
 You are as cowardly as the sum of yourself and the difference
 between a big mighty proud kingdom and a horse. Speak your mind.
 Speak your mind!

[Exit Romeo]

Scene II: The praising of Juliet.

[Enter Juliet]

Hamlet:
 Thou art as sweet as the sum of the sum of Romeo and his horse and his
 black cat! Speak thy mind!

[Exit Juliet]

Scene III: The praising of Ophelia.

[Enter Ophelia]

Hamlet:
 Thou art as beautiful as the difference between Romeo and the square
 of a huge green peaceful tree. Speak thy mind!
 Thou art as lovely as the product of a large rural town and my amazing
 bottomless embroidered purse. Speak thy mind!
 Thou art as loving as the product of the bluest clearest sweetest sky
 and the sum of a squirrel and a white horse. Thou art as beautiful as
 the difference between Juliet and thyself. Speak thy mind!

[Exeunt Ophelia and Hamlet]

Act II: Behind Hamlet's back.
Scene I: Romeo and Juliet's conversation.

[Enter Romeo and Juliet]

Romeo:
 Speak your mind. You are as worried as the sum of yourself and the
 difference between my small smooth hamster and my nose. Speak your mind!

Juliet:
 Speak YOUR mind! You are as bad as Hamlet! You are as small as the
 difference between the square of the difference between my little pony
 and your big hairy hound and the cube of your sorry little
 codpiece. Speak your mind!

[Exit Romeo]

Scene II: Juliet and Ophelia's conversation.

[Enter Ophelia]

Juliet:
 Thou art as good as the quotient between Romeo and the sum of a small
 furry animal and a leech. Speak your mind!

Ophelia:
 Thou art as disgusting as the quotient between Romeo and twice the
 difference between a mistletoe and an oozing infected blister! Speak your mind!

[Exeunt]
```

### Truth Machine
```
The Truth Machine.

Romeo, a young man with a remarkable patience.
Juliet, a likewise young woman of remarkable grace.

Act I: The Truth.
Scene I: The Initialization.

[Enter Romeo and Juliet]

Scene II: The Looping Heart.

Romeo:
 Listen to your heart! Open your heart!

Juliet:
 Am I better than you?

Romeo:
 If so, let us proceed to scene II.

[Exeunt]
```

---

## Canonical Grammar (vetted)

> **Note:** This section was added after cross-checking the prose above against the
> canonical SPL reference (https://grokipedia.com/page/Shakespeare_Programming_Language,
> https://esolangs.org/wiki/Shakespeare). It corrects several errors and omissions
> in the original prose. The parser and AST in `internal/parser/` implement this
> vetted grammar. Existing prose is preserved as-is for historical reference.

### Corrected and extended grammar

#### Assignment forms (extended)
The original prose documents only `You are <expression>.`. Canonical SPL also
supports a **no-copula form** where a bare constant is assigned after `You`/`Thou`,
terminated by `!` (or `.`). This is the "insult/praise" form used throughout the
canonical Hello World Program:
```
You lying stupid fatherless big smelly half-witted coward!
```
This is equivalent to `You are a [6 adjectives] coward!` — a constant with 6
adjectives and a negative noun.

#### Simile semantics (corrected)
The original prose states: "`as adjective as` — evaluates to 0 but maintains
grammatical flow." This is **wrong**. Similes evaluate to the referenced value:
```
You are as stupid as the difference between a handsome rich brave
 hero and thyself!      → value = the difference (NOT 0)
You are as brave as Hamlet.                       → value = Hamlet's value
You are as lovely as a summer's day.              → value of the constant
```
The adjective in `as <adj> as <value>` is grammatical filler (discarded for
evaluation in expression context; captured in `AssignStmt.SimileAdj` for
completeness).

#### Constant evaluation (corrected example)
The original prose example `a vile coward` → -4 is arithmetically wrong.
Each adjective doubles the magnitude once; the noun's sign provides the polarity.
- `flower`         = +1 (positive noun, 0 adjectives)
- `red flower`     = +2 (1 adjective × positive noun)
- `red hot flower` = +4 (2 adjectives × positive noun)
- `coward`         = -1 (negative noun, 0 adjectives)
- `big coward`     = -2 (1 adjective × negative noun)
- `a vile coward`  = -2 (1 adjective `vile` × negative noun `coward`)

Value = `noun_polarity × 2^adjective_count`.

#### Stage management — Exeunt (corrected)
The original prose says `[Exeunt]` exits all on stage. The canonical Hello World
Program also uses `[Exeunt Ophelia and Hamlet]` to exit specific named characters.
Both forms are valid:
- `[Exeunt]` — exits all characters currently on stage
- `[Exeunt A]` — exits character A
- `[Exeunt A and B]` — exits characters A and B (matches the canonical Hello World)

#### Stack operations (added — missing from original prose)
Each character has a LIFO stack. The current value is the top of the stack; an
empty stack yields 0. Two operations manipulate the stack:
- **`Remember <expr>.`** — pushes the value of `<expr>` onto the **listener's** stack.
  Examples: `Remember me.` (pushes speaker's value), `Remember yourself.` (pushes
  listener's value onto its own stack), `Remember the sum of Romeo and a flower.`
- **`Recall <ignored text>.`** — pops the top of the **listener's** stack and assigns
  it as the **speaker's** new value (0 if empty). The text after `Recall` is
  semantically ignored dramatic filler: `Recall your tragic fate.`

#### Additional unary operations (added)
The original prose lists only `square`, `cube`, and `twice`. Canonical SPL also
supports:
- `the square root of A` — integer square root
- `the factorial of A` — factorial

#### Seven comparative forms (corrected)
The original prose lists only 3 comparative forms (`better than`, `as good as`,
`worse than`). Canonical SPL supports **7 syntactic forms** mapping to **6
relations**, determined by adjective polarity (positive/negative) + presence of
`not` + `as...as` vs `...than`:

| # | Form | Adjective polarity | Negated | Relation |
|---|------|-------------------|---------|----------|
| 1 | `as <pos> as` | positive | no | equal (`==`) |
| 2 | `as <neg> as` | negative | no | not_equal (`!=`) |
| 3 | `<pos> than` | positive | no | greater (`>`) |
| 4 | `<neg> than` | negative | no | less (`<`) |
| 5 | `not as <pos> as` | positive | yes | not_equal (`!=`) |
| 6 | `not <pos> than` | positive | yes | less_or_equal (`<=`) |
| 7 | `not <neg> than` | negative | yes | greater_or_equal (`>=`) |

Adjective polarity: `good/better/big/large/fair/brave/sweet/...` = positive;
`bad/worse/small/poor/ugly/...` = negative. Unknown comparative → positive (default).

#### Extended pronoun reference (corrected)
The original prose mentions only `yourself`/`thyself`. Canonical SPL has a full
pronoun system:
- **Possessive (ignored like articles in constants):** `my`, `thy`, `your`, `his`,
  `her`, `mine`, `thine`
- **Reflexive / self-reference:**
  - `me`, `myself` → speaker's value
  - `thyself`, `yourself` → listener's value
  - `you`, `thou` → listener (in assignment target and question operand contexts)
  - `I` → speaker (in question contexts)
- Example: `the sum of the sweetest reddest rose and my father and yourself!`
  - `the sweetest reddest rose` = +4 (positive noun `rose`, 2 adjectives)
  - `my father` = +1 (possessive `my` ignored; noun `father` = +1)
  - `yourself` = listener's current value

#### Question and branch forms (corrected)
The original prose shows `Is X better than Y?` with arbitrary X/Y. Canonical SPL
primarily uses the fixed frame `Am I <comparative> you?` (speaker vs. listener).
The `Is X <comparative> Y?` form is also supported in the parser. After a
question, the response may be:
- `If so, let us proceed to scene/act X.` — jump if question is true
- `If not, let us proceed to scene/act X.` — jump if question is false
- `If so, let us return to scene/act X.` — backward jump (true)
- `If not, let us return to scene/act X.` — backward jump (false)
- `Let us proceed to scene/act X.` — unconditional goto
- `Let us return to scene/act X.` — unconditional goto

#### Character names (clarified)
Character names in the Dramatics Personae are **single words** in the canonical
examples (Romeo, Juliet, Hamlet, Ophelia, etc.). Multi-word names like "Sir Andrew"
are a possible v2 extension; v1 assumes single-word names.

#### Lexical character policy (clarified)
The lexer recognizes only structural punctuation (`.` `,` `:` `!` `?` `[` `]`),
newlines, and EOF. Everything else (including `@`, `/`, digits, `A/S`) folds into
`WORD` tokens. This allows character descriptions to contain free-text punctuation
(e.g., `Hamlet, the flatterer of Andersen Insulting A/S.`). L001 fires only for
genuinely invalid control characters.
