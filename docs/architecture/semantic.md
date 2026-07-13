# Semantic Analysis

**Package:** `internal/semantic/`

The semantic analyzer validates the AST for meaning-level correctness —
things that are syntactically valid but semantically invalid.

## Components

- **SymbolTable** — case-insensitive map of declared character names and their metadata
- **ActRegistry / SceneRegistry** — per-act scene tracking for goto/if target validation
- **Stage manager** — tracks which characters are on stage (max 2), handles Enter/Exit/Exeunt

## Validation rules

| Code | Rule |
|------|------|
| M001 | Character referenced in stage direction or dialogue not declared |
| M002 | Too many characters in single Enter (defensive — parser caps at 2) |
| M003 | Stage overflow (2 already on stage, attempting to enter another) |
| M004 | Speaker not on stage, or listener not reachable |
| M005 | Exit/Exeunt targeting character not on stage |
| M006 | Goto/If target scene or act does not exist |
| M007 | Self-reference Enter (same name twice, or re-enter already-on-stage) |
| M008 | Act with zero scenes (defensive) |

## Key decisions

- **D1**: Stage persists across act boundaries (confirmed by `primes.spl` canonical fixture).
- **D2**: Character value reads allowed off-stage (confirmed by canonical Hello World).
- **D3**: All errors collected in one pass (linter-style), no fail-fast.
- **D4**: Type-switch dispatch, no formal Visitor interface.

## Usage

```go
res := semantic.Analyze(program)
if !res.OK() {
    for _, err := range res.Errors {
        fmt.Println(err)
    }
}
```
