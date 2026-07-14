# Lexer

**Package:** `internal/lexer/`

Turns SPL source text into a stream of tokens. Deliberately simple — it emits
generic `WORD` for every word and leaves classification (character names, nouns,
adjectives, verbs) to the parser.

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

## Behaviour

- Case-insensitive matching; preserves the original lexeme on each token.
- Newlines are significant (they delimit sentences) and emitted as `NEWLINE`.
- Stage directions `[...]` bracket their contents.
- Whitespace (spaces, tabs, `\r`) is skipped.

## Error codes

| Code | Condition |
|------|-----------|
| L001 | Control character in source (0x00–0x08, 0x0B, 0x0C, 0x0E–0x1F, 0x7F) |
| L002 | Unterminated `[` |

## Usage

```go
tokens, err := lexer.ScanTokens(source)
for _, tok := range tokens {
    fmt.Printf("%s:%d:%d %s\n", tok.Type, tok.Line, tok.Col, tok.Lexeme)
}
```
