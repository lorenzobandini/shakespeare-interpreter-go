package runtime

import (
	"fmt"
	"io"
	"strings"

	"github.com/lorenzobandini/shakespeare-interpreter-go/internal/parser"
	"github.com/lorenzobandini/shakespeare-interpreter-go/internal/semantic"
)

type env struct {
	values            map[string]int
	stacks            map[string][]int
	stage             *semantic.Stage
	syms              semantic.SymbolTable
	comparison        bool
	in                io.Reader
	out               io.Writer
	acts              semantic.ActRegistry
	sceneLabels       map[string]map[string]int
	actLabel          map[string]int
	instrs            []instr
	currentActRoman   string
	currentSceneRoman string
	filename          string
}

// NewEnv creates a runtime environment from a parsed program and semantic result.
func NewEnv(prog *parser.Program, res semantic.Result, in io.Reader, out io.Writer, filename string) *env {
	e := &env{
		values:      make(map[string]int),
		stacks:      make(map[string][]int),
		stage:       &semantic.Stage{},
		syms:        res.Symbols,
		in:          in,
		out:         out,
		acts:        res.Acts,
		sceneLabels: make(map[string]map[string]int),
		actLabel:    make(map[string]int),
		filename:    filename,
	}
	for _, c := range prog.Characters {
		key := strings.ToLower(c.Name)
		e.values[key] = 0
		e.stacks[key] = nil
	}
	return e
}

// Execute runs the full SPL program: flatten to instructions then evaluate.
func Execute(prog *parser.Program, res semantic.Result, in io.Reader, out io.Writer, filename string) error {
	if !res.OK() {
		return fmt.Errorf("semantic analysis failed: %d error(s)", len(res.Errors))
	}
	e := NewEnv(prog, res, in, out, filename)
	e.flatten(prog)
	return e.runLoop()
}
