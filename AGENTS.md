# AGENTS.md

## Project status

Phase 0 scaffolding — **no interpreter logic exists yet**:
- `cmd/shpl/main.go` is a stub (`fmt.Println` only). Cobra is **not** wired up, even though `go.mod` lists `spf13/cobra` (and `pflag`, `mousetrap`) as indirect deps. Don't assume the CLI exists.
- `internal/lexer`, `internal/logger` are minimal stubs.
- `testdata/{lexer,parser,interpreter}` are empty placeholder dirs for future `.shpl` fixtures.

Target architecture (from the project plan, not yet implemented): Cobra CLI with subcommands `run`, `ast`, `tokens`, `repl`, `version`, `about`; global `--debug`/`--trace` flags; pipeline `lexer → parser → ast → runtime`; `internal/` packages with unidirectional dependencies; structured logging via `log/slog`.

The authoritative SPL language reference is `docs/SPL_SPECIFICATION.md`. Read it before implementing any lexer/parser/ast/runtime logic.
The error taxonomy (`docs/ERROR_TAXONOMY.md`) defines error codes (L001, S001, M001, R001) — use these for consistent error reporting across all phases.

## Agent workflow

1. **Plan first** (Planning Agent step) — outline the strategy, model packages with Mermaid diagrams, get approval before writing code.
2. **Scaffold** — before importing or calling a library, fetch current docs via `ctx7` (`npx ctx7@latest library <name> "<query>"`). Don't rely on training data for API signatures or version-specific config.
3. **Implement** — write code, write tests, keep commits atomic and conventional.
4. **Gate** — run `task check` before considering any unit of work complete. The CI runs exactly this.
5. **Track** — after each meaningful chunk of work, update `PROGRESS.md` with what was done, what was decided, and what remains. This guards against context loss in long sessions. When context is near full, update PROGRESS.md before the next turn so the fresh session can pick up cleanly.
6. **Commit** — each atomic change gets its own Conventional Commits message. Small, frequent commits.
7. **Review** — before declaring a phase complete, do a self-review pass: re-read changed files, verify `task check` passes, confirm tests cover the change.

## Skills

Use these skills during development:
- `find-docs` — fetches current library documentation (replaces stale training data)
- `find-skills` — discovers and installs new skills as the project grows
- `golang-patterns` — idiomatic Go patterns, best practices, and conventions
- `golang-testing` — table-driven tests, subtests, benchmarks, fuzzing, coverage
- `security-review` — authentication, user input, secrets, API endpoints: checklist and patterns
- Various superpowers skills (brainstorming, planning, reviewing, etc.) — loaded via plugin

Additional skills can be installed via the `find-skills` skill when new needs arise.
Plugins (`opencode.json`): `@dietrichgebert/ponytail` (token-efficient ruleset), `superpowers` (comprehensive agent skills).

## Commands

All quality gates run through Task. The single pre-commit / CI gate is:

```
task check     # fmt → lint → vuln → test (deps run in that order, then echo)
```

Both Lefthook (`lefthook.yml`, serial) and CI (`.github/workflows/ci.yml`) run exactly `task check` — prefer it over running steps individually.

- `task fmt` — `gofmt -s -w .` + `goimports -w .`
- `task lint` — `golangci-lint run ./...`
- `task vuln` — `govulncheck ./...`
- `task test` — `go test -v -race -coverprofile=coverage.out ./...`
- `task build` — `go build -o bin/shpl.exe cmd/shpl/main.go` (Windows target; depends on `fmt`)

Single test / package (always pass `-race`, the suite uses it):

```
go test -race -run TestName ./cmd/shpl/...
go test -race ./internal/lexer/...
```

`coverage.out` is gitignored via `*.out` — don't commit it.

## Toolchain requirements

- **Go 1.26.5** is pinned in `go.mod`, `Dockerfile`, and CI. Install exactly this version.
- **golangci-lint v2** — `.golangci.yaml` uses `default: none` and enables only `errcheck, govet, ineffassign, staticcheck, unused, bodyclose`. To add a linter, edit `.golangci.yaml` (v2 schema, `linters.enable`); don't assume defaults apply. Timeout 3m.
- **Task 3.x**, **govulncheck**, **goimports** must be on PATH.

## Logger gotcha

`internal/logger.Init(level)` must be called before any `slog.Debug` output appears — e.g. `internal/lexer.ScanToken` logs at debug level and is silent until then. Uses `slog.NewTextHandler` to stdout. Levels: `LevelInfo` (default), `LevelDebug`.

## Conventions

- **Design-first**: before writing execution logic, outline a step-by-step plan ("Planning Agent" step) and model package boundaries / AST / class structure with Mermaid diagrams. Don't jump straight to code.
- **Conventional Commits** are required for all commit messages.
- **GoDoc comments** stay inline next to code. Language specification lives in `docs/SPL_SPECIFICATION.md`. Keep `README.md` thin.
- `testdata/{lexer,parser,interpreter}` is the canonical home for `.shpl` fixtures used by table-driven and snapshot tests.
- **PROGRESS.md** in the repo root tracks completed work, open decisions, and remaining tasks. Update it after each meaningful unit of work.

## Docker

`Dockerfile` builds a static linux binary (`CGO_ENABLED=0 GOOS=linux`, `-ldflags="-s -w"`) on `golang:1.26.5-alpine`, runtime `alpine:3.19`, entry `./shpl --help`. Use this for a portable linux build; use `task build` for the local Windows `.exe`.
