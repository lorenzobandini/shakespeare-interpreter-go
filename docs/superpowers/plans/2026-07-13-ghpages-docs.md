# GitHub Pages Documentation Platform — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Stand up an MkDocs + Material documentation site at `https://lorenzobandini.github.io/shakespeare-interpreter-go/`, deployed via GitHub Actions on push to `main`.

**Architecture:** Single `mkdocs.yml` at repo root with `docs_dir: docs`. Material slate theme with tabs + sections nav. Exclude `docs/superpowers/` (development artifacts). Deploy via `mkdocs gh-deploy --force`.

**Tech Stack:** MkDocs, mkdocs-material, mkdocs-minify-plugin, Python 3.x, GitHub Actions.

## Global Constraints

- `pip install -r requirements-docs.txt` must succeed with no errors.
- All pages must be referenced in `mkdocs.yml` nav (no orphan pages).
- Every nav entry must resolve to an existing `.md` file inside `docs/`.
- `mkdocs build --strict` must succeed with zero warnings.
- CI workflow must build on PR (no deploy) and build+deploy on `main` push.
- Existing `docs/superpowers/` directory must be excluded from the published site.
- `AGENTS.md` path references (`docs/SPL_SPECIFICATION.md`, `docs/ERROR_TAXONOMY.md`) must be updated to `docs/spl/` equivalents.
- `docs/.gitkeep` must be excluded from the published site.

---

## File Structure

```
Create:
├── requirements-docs.txt
├── mkdocs.yml
├── docs/index.md
├── docs/getting-started/installation.md
├── docs/getting-started/usage.md
├── docs/architecture/overview.md
├── docs/architecture/lexer.md
├── docs/architecture/parser.md
├── docs/architecture/semantic.md
├── docs/architecture/runtime.md
├── docs/cli/commands.md
├── docs/cli/repl.md
├── docs/contributing.md
├── docs/about.md
├── docs/assets/stylesheets/extra.css
├── .github/workflows/docs.yml
├── docs/spl/.gitkeep                  # temp placeholder until move

Move:
├── docs/SPL_SPECIFICATION.md  →  docs/spl/specification.md
├── docs/ERROR_TAXONOMY.md     →  docs/spl/errors.md

Modify:
├── AGENTS.md                          # update paths
├── PROGRESS.md                        # mark phase complete, update paths
```

---

### Task 1: Scaffold — root configs + directory structure

**Files:**
- Create: `requirements-docs.txt`
- Create: `mkdocs.yml`
- Create: `docs/assets/stylesheets/extra.css`
- Create: `docs/spl/.gitkeep` (temp, removed in Task 2)

**Interfaces:**
- Consumes: nothing
- Produces: `mkdocs.yml` with full nav tree (pages don't exist yet; build will fail until Task 5 — that's intentional)

- [ ] **Step 1: Create `requirements-docs.txt`**

```text
mkdocs-material~=9.6
mkdocs-minify-plugin~=0.8
```

- [ ] **Step 2: Run pip to verify deps install**

Run: `pip install -r requirements-docs.txt`
Expected: both packages install cleanly.

- [ ] **Step 3: Create directory structure**

Run:
```powershell
New-Item -ItemType Directory -Path "docs/getting-started" -Force
New-Item -ItemType Directory -Path "docs/spl" -Force
New-Item -ItemType Directory -Path "docs/architecture" -Force
New-Item -ItemType Directory -Path "docs/cli" -Force
New-Item -ItemType Directory -Path "docs/assets/stylesheets" -Force
New-Item -ItemType Directory -Path "docs/assets/images" -Force
New-Item -ItemType Directory -Path ".github/workflows" -Force
```

- [ ] **Step 4: Create empty `docs/assets/stylesheets/extra.css`**

Run: `New-Item -ItemType File -Path "docs/assets/stylesheets/extra.css" -Force`

- [ ] **Step 5: Create `docs/spl/.gitkeep` (temp)**

Run: `New-Item -ItemType File -Path "docs/spl/.gitkeep" -Force`

Note: .gitkeep prevents git from dropping empty dirs. Gets excluded from site via `exclude_docs`.

- [ ] **Step 6: Write `mkdocs.yml`**

```yaml
site_name: Shakespeare Interpreter
site_description: A Go-based interpreter for the Shakespeare Programming Language
site_url: https://lorenzobandini.github.io/shakespeare-interpreter-go
repo_url: https://github.com/lorenzobandini/shakespeare-interpreter-go
repo_name: lorenzobandini/shakespeare-interpreter-go
edit_uri: edit/main/docs/

theme:
  name: material
  palette:
    scheme: slate
    primary: indigo
    accent: deep orange
  features:
    - navigation.tabs
    - navigation.sections
    - toc.integrate
    - content.code.copy

plugins:
  - search
  - minify:
      minify_html: true

markdown_extensions:
  - admonition
  - pymdownx.superfences
  - pymdownx.tabbed:
      alternate_style: true
  - pymdownx.highlight
  - pymdownx.inlinehilite
  - toc:
      permalink: true

nav:
  - Home: index.md
  - Getting Started:
    - Installation: getting-started/installation.md
    - Usage: getting-started/usage.md
  - SPL Language:
    - Specification: spl/specification.md
    - Error Taxonomy: spl/errors.md
  - Architecture:
    - Pipeline Overview: architecture/overview.md
    - Lexer: architecture/lexer.md
    - Parser: architecture/parser.md
    - Semantic Analysis: architecture/semantic.md
    - Runtime: architecture/runtime.md
  - CLI Reference:
    - Commands: cli/commands.md
    - REPL: cli/repl.md
  - Contributing: contributing.md
  - About: about.md

extra:
  social:
    - icon: fontawesome/brands/github
      link: https://github.com/lorenzobandini/shakespeare-interpreter-go

exclude_docs:
  - superpowers/
  - .gitkeep
```

- [ ] **Step 7: Commit**

```bash
git add requirements-docs.txt mkdocs.yml docs/spl/.gitkeep docs/assets/stylesheets/extra.css docs/getting-started docs/architecture docs/cli docs/assets/images .github/workflows
git commit -m "chore: scaffold MkDocs directory structure and mkdocs.yml"
```

---

### Task 2: Migrate existing SPL documentation

**Files:**
- Move: `docs/SPL_SPECIFICATION.md` → `docs/spl/specification.md`
- Move: `docs/ERROR_TAXONOMY.md` → `docs/spl/errors.md`
- Delete: `docs/spl/.gitkeep` (no longer needed)

**Interfaces:**
- Consumes: `docs/spl/.gitkeep` (deleted), `mkdocs.yml` (nav already references `spl/specification.md` and `spl/errors.md`)
- Produces: populated `docs/spl/` directory

- [ ] **Step 1: Move SPL_SPECIFICATION.md**

Run:
```powershell
Move-Item -Path "docs/SPL_SPECIFICATION.md" -Destination "docs/spl/specification.md" -Force
```

- [ ] **Step 2: Move ERROR_TAXONOMY.md**

Run:
```powershell
Move-Item -Path "docs/ERROR_TAXONOMY.md" -Destination "docs/spl/errors.md" -Force
```

- [ ] **Step 3: Remove temp .gitkeep**

Run: `Remove-Item -Path "docs/spl/.gitkeep" -Force`

- [ ] **Step 4: Verify moved files render**

Run: `mkdocs build --strict`
Expected: FAILS — nav references `index.md`, `getting-started/installation.md`, etc. that don't exist yet. But no errors about `spl/specification.md` or `spl/errors.md`.

- [ ] **Step 5: Commit**

```bash
git add docs/spl/specification.md docs/spl/errors.md docs/SPL_SPECIFICATION.md docs/ERROR_TAXONOMY.md
git commit -m "feat: migrate SPL spec and error taxonomy into MkDocs tree"
```

---

### Task 3: Write Home + Getting Started pages

**Files:**
- Create: `docs/index.md`
- Create: `docs/getting-started/installation.md`
- Create: `docs/getting-started/usage.md`

**Interfaces:**
- Consumes: nothing beyond scaffold
- Produces: first navigable pages

- [ ] **Step 1: Write `docs/index.md`**

```markdown
# Shakespeare Interpreter

A **Go-based interpreter** for the [Shakespeare Programming Language](http://shakespearelang.com/) (SPL),
a Turing-complete esoteric language where programs read like Elizabethan plays.

## Features

- **Full pipeline** — lexer, parser, semantic analyzer, runtime interpreter
- **Complete SPL support** — all canonical operations, control flow, I/O, stack
- **Cobra CLI** — `run`, `tokens`, `ast`, `repl`, `version`, `about` subcommands
- **Debugging** — `--debug` and `--trace` flags for development
- **Portable** — cross-platform Go binary, Docker image available

## Quick start

```sh
task build
./bin/shpl.exe run examples/hello.shpl
```

For detailed instructions see [Installation](getting-started/installation.md) and [Usage](getting-started/usage.md).

## Project status

| Phase | Status |
|-------|--------|
| Scaffolding | ✅ |
| Lexer | ✅ |
| Parser | ✅ |
| Semantic Analysis | ✅ |
| Runtime / Evaluator | ✅ |
| CLI Integration | ✅ |
| **Documentation** | ✅ **this site** |
```

- [ ] **Step 2: Write `docs/getting-started/installation.md`**

```markdown
# Installation

## Prerequisites

- **Go 1.26.5** ([download](https://go.dev/dl/))
- **Task 3.x** — `go install github.com/go-task/task/v3/cmd/task@latest`
- **golangci-lint v2** — `go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest`
- **govulncheck** — `go install golang.org/x/vuln/cmd/govulncheck@latest`
- **goimports** — `go install golang.org/x/tools/cmd/goimports@latest`

## Build from source

```sh
git clone https://github.com/lorenzobandini/shakespeare-interpreter-go.git
cd shakespeare-interpreter-go
task build
./bin/shpl.exe --help
```

The `task build` command compiles a Windows `.exe`. For other platforms, use `go build` directly:

```sh
go build -o bin/shpl ./cmd/shpl/...
```

## Docker

A portable Linux binary is built via Docker:

```sh
docker build -t shpl .
docker run --rm shpl --help
```

## Verify installation

```sh
./bin/shpl.exe version
```

Should print a version string.
```

- [ ] **Step 3: Write `docs/getting-started/usage.md`**

```markdown
# Usage

The CLI provides six subcommands. Global flags `--debug` and `--trace` are available on all commands.

## `shpl run <file>`

Execute a `.shpl` source file through the full pipeline: lex → parse → analyze → execute.

```sh
./bin/shpl.exe run examples/hello.shpl
```

With tracing:

```sh
./bin/shpl.exe --trace run examples/hello.shpl
```

## `shpl tokens <file>`

Lex the source file and print the token stream.

```sh
./bin/shpl.exe tokens examples/hello.shpl
```

## `shpl ast <file>`

Lex and parse the source file, then print the AST as nested JSON.

```sh
./bin/shpl.exe ast examples/hello.shpl
```

## `shpl repl`

Launch an interactive REPL session. Each submission is accumulated and replayed
through the full pipeline. Characters are auto-declared on first use.

```sh
./bin/shpl.exe repl
```

Type SPL dialogue line by line. See [REPL](../cli/repl.md) for details.

## `shpl version`

Print the build version (injected via `-ldflags` at build time).

## `shpl about`

Print credits and licensing information.

## Pipeline overview

```text
Source (.shpl)  →  Lexer  →  Token stream  →  Parser  →  AST
                                                     ↓
                                              Semantic Analyzer
                                                     ↓
                                              Runtime / Execute
```

Each stage validates its input and reports errors in a consistent format:

```
error[<CODE>]: <message>
  --> <file>:<line>:<col>
```

See [Error Taxonomy](../spl/errors.md) for all error codes.
```

- [ ] **Step 4: Build and verify**

Run: `mkdocs build --strict`
Expected: still fails (architecture + cli pages missing). But `index.md`, `getting-started/installation.md`, `getting-started/usage.md` should compile without errors.

Run: `mkdocs build 2>&1`
Expected: shows warnings/errors only about missing architecture/* and cli/* pages.

- [ ] **Step 5: Commit**

```bash
git add docs/index.md docs/getting-started/installation.md docs/getting-started/usage.md
git commit -m "docs: add home and getting started pages"
```

---

### Task 4: Write Architecture pages

**Files:**
- Create: `docs/architecture/overview.md`
- Create: `docs/architecture/lexer.md`
- Create: `docs/architecture/parser.md`
- Create: `docs/architecture/semantic.md`
- Create: `docs/architecture/runtime.md`

- [ ] **Step 1: Write `docs/architecture/overview.md`**

```markdown
# Architecture Overview

The interpreter follows a four-stage pipeline: **Lexer → Parser → Semantic Analyzer → Runtime**.

```text
Source (.shpl)
    │
    ▼
┌─────────────┐
│   Lexer     │  → Token stream ([]Token)
│             │    Errors: L001, L002
└─────────────┘
    │
    ▼
┌─────────────┐
│   Parser    │  → Abstract Syntax Tree (*ast.Program)
│             │    Errors: S001–S018
└─────────────┘
    │
    ▼
┌─────────────┐
│  Semantic   │  → Validated AST + SymbolTable + Stage state
│  Analyzer   │    Errors: M001–M008
└─────────────┘
    │
    ▼
┌─────────────┐
│  Runtime    │  → Program output (stdout) + exit code
│  (Evaluate) │    Errors: R001–R004
└─────────────┘
```

## Package dependency graph

```
cmd/shpl/          ← Cobra CLI entry points
    │
    ▼
internal/lexer/    ← Token scanner
internal/logger/   ← Structured slog handler
    │
    ▼
internal/parser/   ← Recursive descent parser, AST definitions, dictionary
    │
    ▼
internal/semantic/ ← Symbol table, stage manager, validation
    │
    ▼
internal/runtime/  ← Environment, instruction trampoline, I/O
```

Dependencies flow downward only. No package imports from a higher layer.

## Key design decisions

- **Recursive descent parser** — one method per grammar production, no parser generator.
- **Type-switch dispatch** in semantic analyzer and runtime (no Visitor pattern — YAGNI).
- **Trampoline execution** — program flattened to `[]instr`, integer PC, no nested traversal.
- **Replay-based REPL** — full pipeline rerun on each submission (O(n²) at human scale).
```

- [ ] **Step 2: Write `docs/architecture/lexer.md`**

```markdown
# Lexer

**Package:** `internal/lexer/`

The lexer converts a `.shpl` source string into a stream of 9 token types.

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
```

- [ ] **Step 3: Write `docs/architecture/parser.md`**

```markdown
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
```

- [ ] **Step 4: Write `docs/architecture/semantic.md`**

```markdown
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
```

- [ ] **Step 5: Write `docs/architecture/runtime.md`**

```markdown
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
```

- [ ] **Step 6: Build and verify**

Run: `mkdocs build --strict`
Expected: still fails (cli/*.md, contributing.md, about.md missing). Architecture pages compile without errors.

- [ ] **Step 7: Commit**

```bash
git add docs/architecture/overview.md docs/architecture/lexer.md docs/architecture/parser.md docs/architecture/semantic.md docs/architecture/runtime.md
git commit -m "docs: add architecture pages"
```

---

### Task 5: Write CLI reference + Contributing + About pages

**Files:**
- Create: `docs/cli/commands.md`
- Create: `docs/cli/repl.md`
- Create: `docs/contributing.md`
- Create: `docs/about.md`

- [ ] **Step 1: Write `docs/cli/commands.md`**

```markdown
# CLI Commands

All commands are subcommands of `shpl`. Use `shpl --help` for an overview.

## Global flags

| Flag | Effect |
|------|--------|
| `--debug` | Enable debug-level logging (`slog.LevelDebug`) |
| `--trace` | Enable debug logging + pipeline stage markers on stderr |

## `shpl run <file>`

Execute a `.shpl` file through the full pipeline.

```sh
./bin/shpl.exe run examples/hello.shpl
```

Exit codes: 0 on success, 1 on any error (lexical, syntax, semantic, or runtime).

## `shpl tokens <file>`

Print the token stream from lexing.

```sh
./bin/shpl.exe tokens examples/hello.shpl
```

Output format: `TYPE:LN:COL Lexeme` on each line.

## `shpl ast <file>`

Print the AST as indented JSON.

```sh
./bin/shpl.exe ast examples/hello.shpl
```

## `shpl repl`

Launch an interactive REPL. See [REPL](repl.md) for details.

## `shpl version`

Print the build version.

## `shpl about`

Print credits and license information.

## Error format

All errors use this format:

```
error[<CODE>]: <message>
  --> <file>:<line>:<col>
```
```

- [ ] **Step 2: Write `docs/cli/repl.md`**

```markdown
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
```

- [ ] **Step 3: Write `docs/contributing.md`**

```markdown
# Contributing

## Development workflow

1. **Fork** the repository and create a feature branch.
2. **Make changes** — follow existing code conventions.
3. **Run the quality gate** before committing:

```sh
task check
```

This runs: `gofmt -s` → `goimports` → `golangci-lint` → `govulncheck` → `go test -race`.

4. **Write tests** — the project uses table-driven tests with golden file snapshots.
   See existing tests in `internal/lexer/`, `internal/parser/`, `internal/semantic/`,
   and `internal/runtime/` for patterns.

5. **Commit** using [Conventional Commits](https://www.conventionalcommits.org/):
   `feat:`, `fix:`, `docs:`, `chore:`, `refactor:`, `test:`.

## Project structure

```
cmd/shpl/             — CLI entry points (Cobra)
internal/lexer/       — Token scanner
internal/parser/      — Recursive descent parser + AST + dictionary
internal/semantic/    — Semantic analysis + symbol table + stage manager
internal/runtime/     — Interpreter evaluator
internal/logger/      — Structured logger (slog)
testdata/             — .shpl fixture files for table-driven tests
docs/                 — Documentation (this site)
```

## Code conventions

- **Go 1.26.5** — pinned in `go.mod`, Dockerfile, CI.
- **golangci-lint v2** — config at `.golangci.yaml`. Enable/disable linters there.
- **No interfaces with one implementation** — YAGNI.
- **No unrequested abstractions** — factories, visitors, or config for values that never change.

## Pre-commit hooks

The project uses Lefthook. Install it and the pre-commit hook runs `task check`
automatically.
```

- [ ] **Step 4: Write `docs/about.md`**

```markdown
# About

## Project

Shakespeare Interpreter is a Go-based interpreter for the
[Shakespeare Programming Language](http://shakespearelang.com/) (SPL),
created by Karl Hasselström and Jon Åslund in 2001.

SPL is a Turing-complete esoteric language where programs resemble
Elizabethan plays. Characters are variables, dialogue is arithmetic,
and stage directions control program flow.

## License

This project is licensed under the [GNU General Public License v3.0](https://github.com/lorenzobandini/shakespeare-interpreter-go/blob/main/LICENSE).

## Author

Built by [Lorenzo Bandini](https://github.com/lorenzobandini).

## Related

- [SPL at Esolang Wiki](https://esolangs.org/wiki/Shakespeare)
- [Original SPL site](http://shakespearelang.com/)
- [Brain2Speare](https://github.com/mjdarby/Brain2Speare) — brainfuck to SPL compiler
```

- [ ] **Step 5: Build and verify — should pass cleanly**

Run: `mkdocs build --strict`
Expected: zero warnings, exit code 0. Check for generated `site/` directory.

- [ ] **Step 6: Commit**

```bash
git add docs/cli/commands.md docs/cli/repl.md docs/contributing.md docs/about.md
git commit -m "docs: add CLI reference, contributing, and about pages"
```

---

### Task 6: CI/CD Workflow

**Files:**
- Create: `.github/workflows/docs.yml`

- [ ] **Step 1: Write `.github/workflows/docs.yml`**

```yaml
name: Docs
on:
  push:
    branches: [main]
    paths:
      - 'docs/**'
      - 'mkdocs.yml'
      - 'requirements-docs.txt'
  pull_request:
    branches: [main]
    paths:
      - 'docs/**'
      - 'mkdocs.yml'
      - 'requirements-docs.txt'
jobs:
  build-and-deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-python@v5
        with:
          python-version: '3.x'
      - name: Install dependencies
        run: pip install -r requirements-docs.txt
      - name: Build (strict)
        run: mkdocs build --strict
      - name: Deploy to gh-pages
        if: github.ref == 'refs/heads/main'
        run: mkdocs gh-deploy --force
```

- [ ] **Step 2: Validate YAML syntax**

Run: `python -c "import yaml; yaml.safe_load(open('.github/workflows/docs.yml'))"`
Expected: no error.

- [ ] **Step 3: Commit**

```bash
git add .github/workflows/docs.yml
git commit -m "ci: add GitHub Actions workflow for docs build and deploy"
```

---

### Task 7: Update AGENTS.md and PROGRESS.md

**Files:**
- Modify: `AGENTS.md` (lines 12-13, 77 — path updates)
- Modify: `PROGRESS.md` (mark phase complete, update paths)

- [ ] **Step 1: Update path references in `AGENTS.md`**

Edit `AGENTS.md`:
- Line 12: `` `docs/SPL_SPECIFICATION.md` `` → `` `docs/spl/specification.md` ``
- Line 13: `` `docs/ERROR_TAXONOMY.md` `` → `` `docs/spl/errors.md` ``
- Line 77: `` `docs/SPL_SPECIFICATION.md` `` → `` `docs/spl/specification.md` ``

- [ ] **Step 2: Update `PROGRESS.md` — mark phase complete and update paths**

Edit `PROGRESS.md`. Change line 13 from:
`- [x] `docs/SPL_SPECIFICATION.md` — canonical SPL language reference`
to:
`- [x] `docs/spl/specification.md` — canonical SPL language reference`

Change line 20 from:
`- [x] **Dependency**: `docs/SPL_SPECIFICATION.md`.`
to:
`- [x] **Dependency**: `docs/spl/specification.md`.`

Add a new section after Phase 5:

```markdown
## Phase 6 — Documentation Platform ✅

Stand up official MkDocs + Material documentation site deployed to GitHub Pages.

- [x] `mkdocs.yml` with Material slate theme, nav tree, GitHub integration
- [x] Migrate `SPL_SPECIFICATION.md` and `ERROR_TAXONOMY.md` into MkDocs tree
- [x] Content pages: Home, Getting Started, Architecture, CLI, Contributing, About
- [x] CI/CD workflow: build + deploy on push to `main`
- [x] Local verification: `mkdocs build --strict` passes
```

Also update the "Future Phases" section header from:

```
## Future Phases (post-v1)
```

to:

```
## Future Phases (post-v1)
```

(Keep the remaining future phases as-is: Language extensions, LSP server, WASM build, Performance profiling.)

- [ ] **Step 3: Verify `task check` still passes**

Run: `task check`
Expected: green across all gates (fmt, lint, vuln, test). The doc move shouldn't affect Go compilation — nothing imports the `.md` files.

- [ ] **Step 4: Commit**

```bash
git add AGENTS.md PROGRESS.md
git commit -m "docs: update paths and mark Phase 6 (docs platform) complete"
```

---

### Task 8: Final verification

- [ ] **Step 1: Strict build**

Run: `mkdocs build --strict`
Expected: zero warnings, exit code 0.

- [ ] **Step 2: Check nav completeness**

Run:
```powershell
$navPages = Select-String -Path "mkdocs.yml" -Pattern "\.md" | ForEach-Object { $_ -replace '.*?(\S+\.md).*', '$1' }
$existingPages = Get-ChildItem -Recurse -Filter "*.md" -Path "docs" | Where-Object { $_.FullName -notmatch "superpowers" } | ForEach-Object { $_.Name }
Write-Host "All nav pages exist: $($navPages.Count) pages in nav"
```
Expected: no orphan pages (every `.md` under `docs/` excluding `superpowers/` is in nav).

- [ ] **Step 3: Check exclude_docs**

Verify `site/` does not contain `superpowers/`:
```powershell
if (Test-Path "site/superpowers") { Write-Host "FAIL: superpowers/ leaked into site" } else { Write-Host "PASS: superpowers/ excluded" }
```
Expected: PASS.

- [ ] **Step 4: Serve and smoke-test**

Run: `mkdocs serve`
Expected: server starts on `127.0.0.1:8000`. Visit in browser, confirm:
- Home page renders
- Navigation tabs visible (Getting Started, SPL Language, Architecture, CLI Reference)
- Dark theme (slate) active
- Search works
- All subpages render without 404s
- GitHub edit links on each page resolve correctly

Stop server with Ctrl+C.

- [ ] **Step 5: Final commit if any fixes needed**

```bash
git add -A
git commit -m "chore: final verification fixes"
```
