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
