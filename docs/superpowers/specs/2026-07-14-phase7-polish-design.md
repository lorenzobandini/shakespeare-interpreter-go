# Phase 7 — Polish & Infrastructure

Date: 2026-07-14

## Objective

Close the project with 6 maintenance tasks (Batch 1) + 1 structural rename (Batch 2, separate PR).

## Batch 1 — Maintenance & Fix (this session)

### 1. Docker vuln fix
- `Dockerfile:34`: `alpine:3.19` → `alpine:3.21`
- Verification: `docker build --target=check -t shpl:check .` passes, 0 vulnerabilities

### 2. Playground markdown + favicon
- `docs/playground/index.md:6`: remove `{ .md-button .md-button--primary }`, plain link
- `docs/playground/editor.html`: add `<link rel="icon">` in `<head>`

### 3. REPL quickParse fix
- `cmd/shpl/main.go:534-540`: In unconsumed token loop, treat as `ErrIncomplete`:
  - Word followed by COMMA (`,`): potential char declaration
  - `[Aa]ct` / `[Ss]cene` + Roman numeral: potential act/scene header
  - Block ending with PERIOD: potential title or declaration
- Tests: Add cases to `TestTryQuickParse_ErrorLines`

### 4. Dependabot config
- New file `.github/dependabot.yml`: gomod + docker + github-actions, weekly

### 5. Version ASCII art
- `cmd/shpl/main.go`: Add `888` ASCII art before version text
- Update `TestVersionCommand`

### 6. Minimal GoDoc
- One-line doc comments on public exported types/functions in `internal/*`

## Batch 2 — Rename SHPL → SPL (separate PR)

### 7. Structural rename
- `cmd/shpl/` → `cmd/spl/`, binary `spl.exe`, WASM `spl.wasm`
- Update all docs, CI, Taskfile, Cobra `Use: "spl"`
- ~20 files affected, isolated PR for review safety

## Verification
- `task check` passes after each commit
- REPL manual: `go build -o bin/spl.exe cmd/shpl/main.go && ./bin/spl.exe repl`
- WASM build: `GOOS=js GOARCH=wasm go build -o docs/assets/wasm/shpl.wasm cmd/wasm/main.go`
