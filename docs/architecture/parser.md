# Parser

**Package:** `internal/parser/`

Converts a token stream into an AST using recursive descent.

## AST node types

| Node | Purpose |
|------|---------|
| `Program` | Root — title, character declarations, acts |
| `Title` | Program title |
| `CharacterDecl` | Dramatis Personae entry |
| `Act` | Act with Roman numeral + scenes |
| `Scene` | Scene with Roman numeral + statements |
| `EnterStmt` / `ExitStmt` / `ExeuntStmt` | Stage directions |
| `Dialogue` | Speaker + statements |
| `AssignStmt` | `You are <expr>.` or `You <constant>!` |
| `SpeakStmt` / `OpenHeartStmt` / `OpenMindStmt` / `ListenStmt` | I/O |
| `QuestionStmt` | `Am I <comparative> you?` |
| `IfStmt` / `GotoStmt` | Control flow |
| `RememberStmt` / `RecallStmt` | Stack operations |
| `ConstExpr` / `CharRefExpr` / `PronounExpr` | Value expressions |
| `BinaryOpExpr` / `UnaryOpExpr` / `Comparative` | Operators |

## Dictionary

~80 curated words classifying nouns (positive, negative, neutral), adjectives,
comparatives, pronouns, articles, and Shakespeare character names. Unknown words
default to positive noun (value +1) or positive comparative polarity.

## Parse flow

```
Parse()        → title → characters → acts
parseActs()    → per-act: scenes → scene statements
parseExpressions() → constants → binary → unary → pronouns
```

## Error codes

See [Error Taxonomy](../spl/errors.md) for S001–S018.
