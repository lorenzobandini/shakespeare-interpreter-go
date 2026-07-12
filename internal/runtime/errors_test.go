package runtime

import (
	"bytes"
	"testing"

	"github.com/lorenzobandini/shakespeare-interpreter-go/internal/lexer"
	"github.com/lorenzobandini/shakespeare-interpreter-go/internal/parser"
	"github.com/lorenzobandini/shakespeare-interpreter-go/internal/semantic"
)

func TestRuntimeErrorFormat(t *testing.T) {
	tests := []struct {
		name string
		err  RuntimeError
		want string
	}{
		{
			name: "R001 default filename",
			err:  errDivisionByZero("I", "II", 3, 7),
			want: "error[R001]: division by zero at Act I, Scene II\n  --> input:3:7",
		},
		{
			name: "R001 with filename",
			err:  RuntimeError{Code: "R001", Msg: "division by zero at Act I, Scene II", Line: 3, Col: 7, Filename: "play.shpl"},
			want: "error[R001]: division by zero at Act I, Scene II\n  --> play.shpl:3:7",
		},
		{
			name: "R002",
			err:  errInputNotANumber("hello", 5, 3),
			want: "error[R002]: expected a number, got 'hello'\n  --> input:5:3",
		},
		{
			name: "R003",
			err:  errInputEOF(8, 1),
			want: "error[R003]: unexpected end of input\n  --> input:8:1",
		},
		{
			name: "R004",
			err:  errIntegerOverflow("factorial", 10, 4),
			want: "error[R004]: integer overflow in 'factorial'\n  --> input:10:4",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.err.Error()
			if got != tt.want {
				t.Errorf("got %q; want %q", got, tt.want)
			}
		})
	}
}

func TestNewEnv(t *testing.T) {
	src := "Hello, World.\nRomeo, a young man.\nAct I: A scene.\nScene I: The only scene.\n[Enter Romeo]\n"
	tokens, err := lexer.New(src).ScanTokens()
	if err != nil {
		t.Fatalf("lex error: %v", err)
	}
	prog, err := parser.New(tokens).Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	res := semantic.New("test.shpl", prog).Analyze(prog)
	if !res.OK() {
		t.Fatalf("semantic errors: %v", res.Errors)
	}

	var inBuf bytes.Buffer
	outBuf := &bytes.Buffer{}
	e := NewEnv(prog, res, &inBuf, outBuf, "test.shpl")

	if v, ok := e.values["romeo"]; !ok || v != 0 {
		t.Errorf("values[romeo] = %d, %v; want 0, true", v, ok)
	}
	if len(e.stacks["romeo"]) != 0 {
		t.Errorf("stacks[romeo] = %v; want nil", e.stacks["romeo"])
	}
	if e.stage.Size() != 0 {
		t.Errorf("stage.Size() = %d; want 0", e.stage.Size())
	}
	if e.filename != "test.shpl" {
		t.Errorf("filename = %q; want 'test.shpl'", e.filename)
	}
	if e.acts == nil {
		t.Error("acts is nil")
	}
}
