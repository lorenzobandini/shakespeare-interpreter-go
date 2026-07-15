//go:build js && wasm

package main

import (
	"bytes"
	"fmt"
	"strings"
	"syscall/js"

	"github.com/lorenzobandini/shakespeare-interpreter-go/internal/lexer"
	"github.com/lorenzobandini/shakespeare-interpreter-go/internal/parser"
	"github.com/lorenzobandini/shakespeare-interpreter-go/internal/runtime"
	"github.com/lorenzobandini/shakespeare-interpreter-go/internal/semantic"
)

func main() {
	js.Global().Set("executeSPL", js.FuncOf(execute))
	<-make(chan struct{})
}

func execute(this js.Value, args []js.Value) any {
	if len(args) < 1 {
		return map[string]any{"error": "missing source code argument"}
	}
	source := args[0].String()
	input := ""
	if len(args) > 1 {
		input = args[1].String()
	}

	tokens, err := lexer.New(source).ScanTokens()
	if err != nil {
		return map[string]any{"error": fmt.Sprintf("lexer error: %v", err)}
	}

	prog, err := parser.New(tokens).Parse()
	if err != nil {
		return map[string]any{"error": fmt.Sprintf("parser error: %v", err)}
	}

	res := semantic.New("input.spl", prog).Analyze()
	if !res.OK() {
		var b strings.Builder
		for _, e := range res.Errors {
			b.WriteString(e.Error())
			b.WriteString("\n")
		}
		return map[string]any{"error": strings.TrimSuffix(b.String(), "\n")}
	}

	var stdout bytes.Buffer
	in := strings.NewReader(input)
	if err := runtime.Execute(prog, res, in, &stdout, "input.spl"); err != nil {
		return map[string]any{"output": stdout.String(), "error": err.Error()}
	}
	return map[string]any{"output": stdout.String()}
}
