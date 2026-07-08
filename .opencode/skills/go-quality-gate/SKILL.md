---
name: go-quality-gate
description: Enforce that task check passes before any unit of work is considered complete. Knows the full pipeline: fmt, lint, vuln, test.
compatibility: opencode
metadata:
  project: shakespeare-interpreter-go
---

## Quality gate

Before declaring any unit of work complete, run:

```bash
task check
```

This runs sequentially: `fmt` → `lint` → `vuln` → `test`.

## What each step does

| Step | Command | What it catches |
|------|---------|-----------------|
| `fmt` | `gofmt -s -w .` + `goimports -w .` | Formatting, import organization |
| `lint` | `golangci-lint run ./...` | Bugs, dead code, unclosed bodies |
| `vuln` | `govulncheck ./...` | Known CVEs in dependencies |
| `test` | `go test -v -race -coverprofile=coverage.out ./...` | Test failures, race conditions |

## Rules

1. Run `task check` after every meaningful code change.
2. If `task check` fails, fix the failures before continuing.
3. If a check modifies files (e.g., `fmt` rewrites code), re-run `task check`.
4. Never commit code that doesn't pass `task check`.
5. Both Lefthook pre-commit and CI run exactly this — it must pass.
