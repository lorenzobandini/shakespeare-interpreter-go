package runtime

import (
	"bytes"
	"strings"
	"testing"

	"github.com/lorenzobandini/shakespeare-interpreter-go/internal/lexer"
	"github.com/lorenzobandini/shakespeare-interpreter-go/internal/parser"
	"github.com/lorenzobandini/shakespeare-interpreter-go/internal/semantic"
)

func TestFlatten(t *testing.T) {
	src := "Test.\nRomeo, a man.\nJuliet, a woman.\nAct I: First.\nScene I: One.\n[Enter Romeo]\nScene II: Two.\n[Exit Romeo]\nAct II: Second.\nScene I: Three.\n[Enter Romeo and Juliet]\nJuliet:\nSpeak your mind.\n"
	tokens, err := lexer.New(src).ScanTokens()
	if err != nil {
		t.Fatalf("lex error: %v", err)
	}
	prog, err := parser.New(tokens).Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	res := semantic.New("test", prog).Analyze(prog)
	if !res.OK() {
		t.Fatalf("semantic errors: %v", res.Errors)
	}
	e := NewEnv(prog, res, &bytes.Buffer{}, &bytes.Buffer{}, "test")
	e.flatten(prog)

	// Expected instructions: [Enter Romeo], [Exit Romeo], [Enter Romeo and Juliet], Speak
	if len(e.instrs) != 4 {
		t.Fatalf("expected 4 instructions, got %d", len(e.instrs))
	}

	// actLabel["i"] == 0, actLabel["ii"] == 2
	if e.actLabel["i"] != 0 {
		t.Errorf("actLabel[i] = %d; want 0", e.actLabel["i"])
	}
	if e.actLabel["ii"] != 2 {
		t.Errorf("actLabel[ii] = %d; want 2", e.actLabel["ii"])
	}

	// sceneLabels["i"]["i"] == 0, sceneLabels["i"]["ii"] == 1, sceneLabels["ii"]["i"] == 2
	if e.sceneLabels["i"]["i"] != 0 {
		t.Errorf("sceneLabels[i][i] = %d; want 0", e.sceneLabels["i"]["i"])
	}
	if e.sceneLabels["i"]["ii"] != 1 {
		t.Errorf("sceneLabels[i][ii] = %d; want 1", e.sceneLabels["i"]["ii"])
	}
	if e.sceneLabels["ii"]["i"] != 2 {
		t.Errorf("sceneLabels[ii][i] = %d; want 2", e.sceneLabels["ii"]["i"])
	}

	// Check speaker: Enter has empty, Speak has "Juliet"
	if e.instrs[0].speaker != "" {
		t.Errorf("instrs[0].speaker = %q; want empty", e.instrs[0].speaker)
	}
	if _, ok := e.instrs[0].stmt.(parser.EnterStmt); !ok {
		t.Errorf("instrs[0] expected EnterStmt")
	}
	if e.instrs[3].speaker != "Juliet" {
		t.Errorf("instrs[3].speaker = %q; want Juliet", e.instrs[3].speaker)
	}
	if _, ok := e.instrs[3].stmt.(parser.SpeakStmt); !ok {
		t.Errorf("instrs[3] expected SpeakStmt")
	}
}

func TestExecuteLinear(t *testing.T) {
	// Self-talk: Romeo alone on stage, assigns to himself, speaks his value
	src := "Test.\nRomeo, a man.\nAct I: First.\nScene I: One.\n[Enter Romeo]\nRomeo:\nYou are as good as a flower.\nSpeak your mind.\n"
	tokens, err := lexer.New(src).ScanTokens()
	if err != nil {
		t.Fatalf("lex error: %v", err)
	}
	prog, err := parser.New(tokens).Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	res := semantic.New("test", prog).Analyze(prog)
	if !res.OK() {
		t.Fatalf("semantic errors: %v", res.Errors)
	}

	var out bytes.Buffer
	e := NewEnv(prog, res, &bytes.Buffer{}, &out, "test")
	e.flatten(prog)

	err = e.runLoop()
	if err != nil {
		t.Fatalf("runLoop error: %v", err)
	}
	if out.String() != string([]byte{1}) {
		t.Errorf("output = %q; want byte(1)", out.String())
	}
}

func TestExecuteDivZero(t *testing.T) {
	// Romeo speaks; listener is Juliet. Juliet's value is 0 (default).
	// "the quotient between a flower and Juliet" = 1 / 0 → R001
	src := "Test.\nRomeo, a man.\nJuliet, a woman.\nAct I: First.\nScene I: One.\n[Enter Romeo and Juliet]\nRomeo:\nYou are the quotient between a flower and Juliet.\n"
	tokens, err := lexer.New(src).ScanTokens()
	if err != nil {
		t.Fatalf("lex error: %v", err)
	}
	prog, err := parser.New(tokens).Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	res := semantic.New("test", prog).Analyze(prog)
	if !res.OK() {
		t.Fatalf("semantic errors: %v", res.Errors)
	}

	e := NewEnv(prog, res, &bytes.Buffer{}, &bytes.Buffer{}, "test")
	e.flatten(prog)

	err = e.runLoop()
	if err == nil {
		t.Fatal("expected R001 error, got nil")
	}
	re, ok := err.(RuntimeError)
	if !ok || re.Code != "R001" {
		t.Fatalf("expected R001, got %v", err)
	}
	if !strings.Contains(re.Msg, "Act I, Scene I") {
		t.Errorf("message doesn't mention Act/Scene: %s", re.Msg)
	}
}

func TestExecuteEntry(t *testing.T) {
	src := "Test.\nRomeo, a man.\nAct I: First.\nScene I: One.\n[Enter Romeo]\nRomeo:\nYou are as good as a flower.\nSpeak your mind.\n"
	tokens, err := lexer.New(src).ScanTokens()
	if err != nil {
		t.Fatalf("lex error: %v", err)
	}
	prog, err := parser.New(tokens).Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	res := semantic.New("test", prog).Analyze(prog)
	if !res.OK() {
		t.Fatalf("semantic errors: %v", res.Errors)
	}

	var out bytes.Buffer
	err = Execute(prog, res, &bytes.Buffer{}, &out, "test")
	if err != nil {
		t.Fatalf("Execute error: %v", err)
	}
	if out.String() != string([]byte{1}) {
		t.Errorf("output = %q; want byte(1)", out.String())
	}
}

func TestExecuteRejectsUnvalidatedAST(t *testing.T) {
	src := "Test.\nRomeo, a man.\nAct I: First.\nScene I: One.\n[Enter Romeo]\n"
	tokens, err := lexer.New(src).ScanTokens()
	if err != nil {
		t.Fatalf("lex error: %v", err)
	}
	prog, err := parser.New(tokens).Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	// Create a result with a semantic error
	res := semantic.Result{Errors: []semantic.SemanticError{{Code: "M999"}}}
	var out bytes.Buffer
	err = Execute(prog, res, &bytes.Buffer{}, &out, "test")
	if err == nil {
		t.Fatal("expected error for unvalidated AST, got nil")
	}
	if out.Len() > 0 {
		t.Error("expected no output when AST is invalid")
	}
}
