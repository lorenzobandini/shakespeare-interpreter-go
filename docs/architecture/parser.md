# Parser

**Package:** `internal/parser/`

The parser converts a token stream from the lexer into an Abstract Syntax Tree (AST)
using recursive descent.

## AST node types

| Node | Purpose |
|------|---------|
| `Program` | Root — title, character declarations, acts |
| `Title` | Program title line |
| `CharacterDecl` | Dramatis Personae entry |
| `Act` | Act with Roman numeral + scenes |
| `Scene` | Scene with Roman numeral + statements |
| `EnterStmt` / `ExitStmt` / `ExeuntStmt` | Stage directions |
| `Dialogue` | Speaker + list of statements |
| `AssignStmt` | `You are <expr>.` or `You <constant>!` |
| `SpeakStmt` / `OpenHeartStmt` / `OpenMindStmt` / `ListenStmt` | I/O |
| `QuestionStmt` | `Am I <comparative> you?` |
| `IfStmt` / `GotoStmt` | Control flow |
| `RememberStmt` / `RecallStmt` | Stack operations |
| `ConstExpr` / `CharRefExpr` / `PronounExpr` | Value expressions |
| `BinaryOpExpr` / `UnaryOpExpr` / `Comparative` | Operators |

## Dictionary

The parser includes a curated dictionary of ~80 words classifying nouns (positive,
negative, neutral), adjectives, comparatives, pronouns, articles, and Shakespeare
character names. Unknown words default to positive noun (value +1) or unknown
comparative positive polarity.

## Recursive descent

Each grammar production has a dedicated `parse*` method:

```
Parse()        → parseTitle → parseCharacters → parseActs → parseActs
parseActs()    → parseAct   → parseScenes → parseSceneStatements
parseExpressions() → parseConst, parseBinary, parseUnary, etc.
```

## Error codes (S001–S018)

See [Error Taxonomy](../spl/errors.md) for the full list of syntax error codes,
including expected messages and examples.
