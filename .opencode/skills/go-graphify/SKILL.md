---
name: go-graphify
description: Visualize Go package dependencies and internal architecture. Wraps go mod graph, go list, and optionally goda/graphviz for visual output.
compatibility: opencode
metadata:
  project: shakespeare-interpreter-go
---

## Dependency visualization

### Package dependencies (text)
```bash
go mod graph                          # module-level deps (full tree)
go list -m all                        # direct + indirect module deps
go list -f '{{.Imports}}' ./internal/lexer/...  # imports per package
```

### Internal architecture check
```bash
# Verify unidirectional deps: internal/lexer should NOT import internal/parser
go list -f '{{join .Deps "\n"}}' ./internal/lexer/ | Select-String "parser"  
```

### Visual graph (requires goda + graphviz)
```bash
go install github.com/loov/goda@latest
goda graph ./internal/... | dot -Tpng -o deps.png
```

## When to use

- Before creating a new package: verify it fits the dependency direction
- After refactoring: confirm no circular imports
- During architecture review: visualize the planned pipeline (lexer → parser → ast → runtime)
