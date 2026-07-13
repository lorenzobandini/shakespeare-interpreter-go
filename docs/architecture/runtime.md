# Runtime / Evaluator

**Package:** `internal/runtime/`

The runtime executes a semantically validated program by walking a flat instruction
list via integer program counter (PC).

## Execution model

The program is **flattened** during `Execute()` into a single `[]instr` slice.
Each instruction is a closure that mutates the environment. Control flow (goto/if)
is PC-based — branching updates the PC directly.

## Environment

- **Character values** — `map[string]int` (signed integers, initialized to 0)
- **Character stacks** — `map[string][]int` (LIFO per character)
- **Comparison flag** — single `bool` set by the most recent question
- **Stage** — reused from `internal/semantic.Stage`
- **I/O buffers** — stdin reader, stdout/stderr writers

## I/O

| Statement | Direction | Format |
|-----------|-----------|--------|
| `Speak your mind` | output | 1 byte (ASCII) |
| `Open your heart` | output | `%d\n` |
| `Open your mind` | input | 1 byte (0–255) |
| `Listen to your heart` | input | parsed integer |

## Stack operations

- `Remember <expr>` — pushes value of `<expr>` onto listener's stack
- `Recall` — pops listener's stack, assigns to speaker's value (0 if empty)

## Error codes

| Code | Condition |
|------|-----------|
| R001 | Division by zero |
| R002 | Input is not a number (Listen) |
| R003 | Unexpected EOF (OpenMind) |
| R004 | Integer overflow (factorial > 20) |

## REPL compatibility

The REPL uses a **replay-based accumulating buffer model**. Each submission
re-runs the full pipeline on the accumulated buffer. This means infinite loops
are incompatible with the REPL and considered out of scope.
