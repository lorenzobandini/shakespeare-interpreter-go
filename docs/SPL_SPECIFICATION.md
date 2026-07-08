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
