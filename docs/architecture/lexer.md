# Lexer

**Package:** `internal/lexer/`

The lexer converts a `.shpl` source string into a stream of 10 token types.

## Token types

| Token | Example |
|-------|---------|
| `EOF` | end of input |
| `NEWLINE` | `\n` |
| `WORD` | `Romeo`, `Act`, `Enter`, `flower` |
| `PERIOD` | `.` |
| `COMMA` | `,` |
| `COLON` | `:` |
| `BANG` | `!` |
| `QUESTION` | `?` |
| `LBRACKET` | `[` |
| `RBRACKET` | `]` |

The lexer is deliberately **dumb** — it emits generic `WORD` for all text tokens.
Classification into character names, nouns, adjectives, verbs, etc. happens in the
parser using the dictionary.

## Key behavior

- **Case-insensitive** classification; original lexeme preserved on token.
- **Newlines** are significant (sentence boundaries) and emitted as `NEWLINE`.
- **Whitespace** (spaces, tabs, `\r`) is skipped.
- **Stage directions** `[...]` produce `LBRACKET`/`RBRACKET` tokens around interior `WORD`s.

## Error codes

| Code | Condition |
|------|-----------|
| L001 | Control character in source (0x00–0x08, 0x0B, 0x0C, 0x0E–0x1F, 0x7F) |
| L002 | Unterminated `[` (no matching `]`) |

## Usage

```go
tokens, err := lexer.ScanTokens(source)
for _, tok := range tokens {
    fmt.Printf("%s:%d:%d %s\n", tok.Type, tok.Line, tok.Col, tok.Lexeme)
}
```
