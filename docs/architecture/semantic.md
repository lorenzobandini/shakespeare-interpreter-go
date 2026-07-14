# Semantic Analysis

**Package:** `internal/semantic/`

Validates the AST for meaning-level correctness — syntax is fine, but the program
doesn't make sense in SPL terms.

## Components

- **SymbolTable** — case-insensitive map of declared characters
- **ActRegistry / SceneRegistry** — per-act scene tracking for goto targets
- **Stage manager** — tracks who is on stage (max 2)

## Validation rules

| Code | Rule |
|------|------|
| M001 | Character referenced but not declared |
| M002 | Too many characters in a single Enter |
| M003 | Stage overflow (2 already on stage, trying to enter another) |
| M004 | Speaker not on stage, or listener unreachable |
| M005 | Exit/Exeunt targeting character not on stage |
| M006 | Goto/If target scene or act doesn't exist |
| M007 | Self-reference Enter (same name twice, or re-enter on-stage) |
| M008 | Act with zero scenes (defensive) |

## Decisions

- Stage persists across act boundaries (confirmed by `primes.spl`).
- Character values can be read off-stage (confirmed by canonical Hello World).
- All errors collected in one pass — no fail-fast.
- Type-switch dispatch, no Visitor interface.

## Usage

```go
res := semantic.Analyze(program)
if !res.OK() {
    for _, err := range res.Errors {
        fmt.Println(err)
    }
}
```
