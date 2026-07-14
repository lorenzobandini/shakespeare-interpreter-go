# Runtime / Evaluator

**Package:** `internal/runtime/`

Walks a flat instruction list via integer program counter. The entire program is
flattened to `[]instr` during `Execute()` — each instruction is a closure that
mutates the environment.

## Environment

- **Character values** — `map[string]int`, initialized to 0
- **Character stacks** — `map[string][]int`, LIFO per character
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

- `Remember <expr>` — push value onto listener's stack
- `Recall` — pop listener's stack, assign to speaker's value (0 if empty)

## Error codes

| Code | Condition |
|------|-----------|
| R001 | Division by zero |
| R002 | Input is not a number (Listen) |
| R003 | Unexpected EOF (OpenMind) |
| R004 | Integer overflow (factorial > 20) |

## REPL compatibility

The REPL uses a replay-based accumulating buffer. Each submission re-runs the
full pipeline, which means infinite loops are out of scope for the REPL.
