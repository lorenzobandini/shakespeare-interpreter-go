---
name: go-test
description: Run and write Go tests for this project. Knows test patterns, snapshot tests, table-driven tests, and fixture locations.
compatibility: opencode
metadata:
  project: shakespeare-interpreter-go
---

## Test commands

```bash
task test                              # full suite: go test -v -race -coverprofile=coverage.out ./...
go test -race -run TestName ./cmd/shpl/...
go test -race ./internal/lexer/...
```

Always pass `-race` — the suite uses the race detector.

## Test patterns

### Table-driven tests
```go
func TestLexer(t *testing.T) {
    tests := []struct {
        name  string
        input string
        want  []Token
    }{
        {"hello world", "Romeo: Speak your mind.", []Token{...}},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := Lex(tt.input)
            if !reflect.DeepEqual(got, tt.want) {
                t.Errorf("got %v, want %v", got, tt.want)
            }
        })
    }
}
```

### Snapshot / fixture tests
- Fixtures live in `testdata/{lexer,parser,interpreter}/`.
- Load with `os.ReadFile("testdata/lexer/hello.shpl")`.
- Compare output with golden files or inline expected strings.

## Coverage
`coverage.out` is gitignored (`*.out`). Generate HTML: `go tool cover -html=coverage.out`.
