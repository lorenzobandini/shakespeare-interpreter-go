# AGENTS.md

## Project status

Phase 0 scaffolding — **no interpreter logic exists yet**:
- `cmd/shpl/main.go` is a stub (`fmt.Println` only). Cobra is **not** wired up, even though `go.mod` lists `spf13/cobra` (and `pflag`, `mousetrap`) as indirect deps. Don't assume the CLI exists.
- `internal/lexer`, `internal/logger` are minimal stubs.
- `testdata/{lexer,parser,interpreter}` are empty placeholder dirs for future `.shpl` fixtures.

Target architecture (from the project plan, not yet implemented): Cobra CLI with subcommands `run`, `ast`, `tokens`, `repl`, `version`, `about`; global `--debug`/`--trace` flags; pipeline `lexer → parser → ast → runtime`; `internal/` packages with unidirectional dependencies; structured logging via `log/slog`.

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
- **GoDoc comments** stay inline next to code. Deep technical / grammar docs live on a separate `docs` branch (GitHub Pages), **not** in the repo root — keep `README.md` thin.
- `testdata/{lexer,parser,interpreter}` is the canonical home for `.shpl` fixtures used by table-driven and snapshot tests.

## Docker

`Dockerfile` builds a static linux binary (`CGO_ENABLED=0 GOOS=linux`, `-ldflags="-s -w"`) on `golang:1.26.5-alpine`, runtime `alpine:3.19`, entry `./shpl --help`. Use this for a portable linux build; use `task build` for the local Windows `.exe`.
