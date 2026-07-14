# REPL

The REPL (`shpl repl`) provides an interactive environment for writing and testing
SPL programs line by line.

## Phase model

The REPL tracks which part of the program you are writing:

| Phase | What it accepts | Pipeline runs? |
|---|---|---|
| title | Program title | No (accumulates silently) |
| characters | Character declarations | No (accumulates silently) |
| body | Acts, scenes, stage directions, dialogue | Yes |
| closed | After `[Exeunt]` (informational) | Yes |

Pre-body text accumulates silently in the buffer. The full pipeline (lex → parse →
analyze → execute) only runs once you submit body content — stage directions or
dialogue.

## Workflow

### Recommended approach

Submit your program in logical sections:

```text
spl> The Truth Machine.           ← title
...>
spl> Romeo, a young man.          ← character declarations
...> Juliet, a likewise young woman.
...>
spl> [Enter Romeo and Juliet]     ← body: triggers pipeline
...> Romeo: Listen to your heart!
...>
input> 42                         ← runtime stdin (if program reads input)
spl> Juliet: Open your heart!
...>
spl> :quit
```

### Body-first (simpler for small programs)

For short programs, type everything in one submission. The REPL provides the
program structure (title + Act I + Scene I) automatically:

```text
spl> [Enter Romeo]
...> Romeo: Open your heart!
...>                             ← blank line submits
0
spl> :quit
```

Characters like `Romeo` are auto-declared before Act I.

### Providing your own acts

If your input contains an `Act` or `Scene` header, the REPL skips skeleton injection
and uses your structure directly:

```text
spl> The Branch Test.
...>
spl> Romeo, a young man.
...>
spl> Act I: The Branch.
...> Scene I: The Decision.
...> [Enter Romeo]
...> Romeo: Open your heart!
...>
--- EXECUTE ---
0
spl> :quit
```

## Auto-declaration

The REPL automatically declares characters it finds in your text:

- Names in `[Enter X]`, `[Exit X]`, `[Exeunt X]`
- Dialogue speakers (`Name: ...`)

Declarations are inserted before Act I, deduplicated against both your explicit
declarations and earlier auto-declarations. Example:

```text
spl> [Enter Juliet]               ← Juliet auto-declared before Act I
...> Juliet: Open your heart!
...>
0
```

If you later introduce a new character, it is auto-declared in the next submission
without duplicating existing declarations:

```text
spl> [Enter Romeo]                ← Romeo auto-declared, Juliet still present
...> Romeo: Open your heart!
...>
0
```

## Skeleton injection

When you submit body content without act/scene headers, the REPL injects a
skeleton once:

```text
[REPL Session Title if missing]
[your accumulated text — title, declarations]
Act I: The REPL Session.
Scene I: The REPL Session.
```

- The **title** part goes at the buffer start (only if you haven't provided one)
- Your existing **content** stays in the middle
- **Act I / Scene I** goes at the end

If you haven't declared any characters before the body, the pipeline reports
`S002` (genuine error) and rolls back — the REPL does not invent phantom
characters.

## Meta-commands

Enter these at the start of a block (first column, no leading spaces):

| Command | Effect |
|---------|--------|
| `:quit` / `:exit` | Exit the REPL |
| `:help` | Show help text |
| `:reset` | Clear buffer, declarations, and recorded stdin |

## Error handling

Errors in any submission roll back the buffer to its previous valid state. The REPL
reports the error and continues:

```text
spl> Romeo, a young man.
...>
spl> Act I: The Act.              ← no declarations before Act I
...>
error: error[S002]: expected at least one character declaration before first act
spl>                               ← buffer rolled back, try again
```

The prompt returns immediately; you can correct and resubmit.

## Limitations

- **Infinite loops** (e.g., an unterminated truth-machine) are incompatible with
  the replay model.
- `--trace` flag works in the REPL for debugging pipeline stages.
- Programs that span multiple submissions have O(n²) replay cost on the
  accumulated buffer size — negligible at human typing scale.
