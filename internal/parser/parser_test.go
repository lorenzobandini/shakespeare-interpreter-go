package parser

import (
	"testing"

	"github.com/lorenzobandini/shakespeare-interpreter-go/internal/lexer"
	"github.com/lorenzobandini/shakespeare-interpreter-go/internal/logger"
)

func TestMain(m *testing.M) {
	logger.Init(logger.LevelDebug)
	m.Run()
}

func lex(t *testing.T, path string) []lexer.Token {
	t.Helper()
	src, err := readFile(path)
	if err != nil {
		t.Fatal(err)
	}
	tokens, err := lexer.New(string(src)).ScanTokens()
	if err != nil {
		t.Fatal(err)
	}
	return tokens
}

func parseTokens(t *testing.T, path string) *Program {
	t.Helper()
	tokens := lex(t, path)
	prog, err := New(tokens).Parse()
	if err != nil {
		t.Fatal(err)
	}
	return prog
}

func TestParseHelloWorld(t *testing.T) {
	prog := parseTokens(t, "../../testdata/lexer/hello.shpl")
	if prog.Title.Text != "The Infamous Hello World Program" {
		t.Errorf("title: got %q", prog.Title.Text)
	}
	if len(prog.Characters) != 4 {
		t.Errorf("characters: got %d, want 4", len(prog.Characters))
	}
	if len(prog.Acts) != 2 {
		t.Errorf("acts: got %d, want 2", len(prog.Acts))
	}
}

func TestParseTruthMachine(t *testing.T) {
	prog := parseTokens(t, "../../testdata/lexer/truth-machine.shpl")
	if prog.Title.Text != "The Truth Machine" {
		t.Errorf("title: got %q", prog.Title.Text)
	}
	if len(prog.Acts) != 1 {
		t.Errorf("acts: got %d, want 1", len(prog.Acts))
	}
}

func TestParseMinimal(t *testing.T) {
	prog := parseTokens(t, "../../testdata/lexer/minimal.shpl")
	if len(prog.Acts) != 1 {
		t.Errorf("acts: got %d", len(prog.Acts))
	}
}

func TestParseErrorNoTitle(t *testing.T) {
	tokens, err := lexer.New("").ScanTokens()
	if err != nil {
		t.Fatal(err)
	}
	_, err = New(tokens).Parse()
	if err == nil {
		t.Fatal("expected S001 error")
	}
	pe, ok := err.(ParseError)
	if !ok || pe.Code != "S001" {
		t.Errorf("got %v, want S001", err)
	}
}
