# REPL

The REPL (`shpl repl`) provides an interactive environment for writing and testing
SPL programs line by line.

## Replay-based Accumulating Buffer Model

Each line you type is appended to an accumulating buffer. On submission (blank line),
the full buffer is re-executed through the entire pipeline: lex → parse → analyze →
execute.

This means every submission sees the complete program so far, enabling valid programs
that span multiple submissions. The tradeoff is O(n²) complexity on accumulated
buffer size — negligible at human typing scale.

## Auto-declaration

The REPL automatically detects character names used in `[Enter]`, `[Exit]`,
`[Exeunt]`, and dialogue speaker prefixes. It inserts character declarations before
Act I on your behalf.

Characters explicitly declared by you in the input are detected and not duplicated.

## Skeleton structure

On the first submission, the REPL prepends:

```text
The REPL Session.

Act I: The REPL Session.
Scene I: The REPL Session.
```

This guarantees a valid SPL structure. You provide stage directions, dialogue,
and expressions within it.

## Example session

```text
shpl repl
input> [Enter Romeo and Juliet]
input> Juliet: Open your heart!
input> Romeo: You are a flower!
input>
--- EXECUTE ---
0
```

Errors in any submission roll back the buffer so you can correct and retry.

## Limitations

- **Infinite loops** (e.g., an unterminated truth-machine) are incompatible with
  the replay model.
- `--trace` flag works in the REPL for debugging pipeline stages.
