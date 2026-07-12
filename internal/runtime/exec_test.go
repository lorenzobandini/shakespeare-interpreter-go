package runtime

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/lorenzobandini/shakespeare-interpreter-go/internal/lexer"
	"github.com/lorenzobandini/shakespeare-interpreter-go/internal/parser"
	"github.com/lorenzobandini/shakespeare-interpreter-go/internal/semantic"
)

func testEnvWithChars(t *testing.T, charNames ...string) *env {
	t.Helper()
	if len(charNames) < 1 {
		charNames = []string{"Romeo", "Juliet"}
	}
	var sb bytes.Buffer
	sb.WriteString("Test Play.\n")
	for _, name := range charNames {
		fmt.Fprintf(&sb, "%s, a character.\n", name)
	}
	sb.WriteString("Act I: A scene.\nScene I: The only scene.\n[Enter Romeo]\n")
	src := sb.String()
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
	return e
}

func TestExecEnterExitExeunt(t *testing.T) {
	e := testEnvWithChars(t, "Romeo", "Juliet")

	// Enter Romeo
	_, _, err := e.execInstr(instr{
		stmt: parser.EnterStmt{Characters: []string{"Romeo"}, Line: 1, Col: 1},
	})
	if err != nil {
		t.Fatalf("Enter error: %v", err)
	}
	if e.stage.Size() != 1 || !e.stage.Has("Romeo") {
		t.Errorf("expected Romeo on stage, size=%d", e.stage.Size())
	}

	// Enter Juliet
	_, _, err = e.execInstr(instr{
		stmt: parser.EnterStmt{Characters: []string{"Juliet"}, Line: 2, Col: 1},
	})
	if err != nil {
		t.Fatalf("second Enter error: %v", err)
	}
	if e.stage.Size() != 2 {
		t.Errorf("expected size 2, got %d", e.stage.Size())
	}
	if !e.stage.Has("Juliet") {
		t.Error("expected Juliet on stage")
	}

	// Exit Romeo
	_, _, err = e.execInstr(instr{
		stmt: parser.ExitStmt{Character: "Romeo", Line: 3, Col: 1},
	})
	if err != nil {
		t.Fatalf("Exit error: %v", err)
	}
	if e.stage.Size() != 1 || !e.stage.Has("Juliet") {
		t.Errorf("expected Juliet alone, size=%d", e.stage.Size())
	}

	// Exit character not on stage → error discarded (R-D3)
	e.stage.Clear()
	_ = e.stage.Enter([]string{"Romeo", "Juliet"}, e.syms, 0, 0)

	// Exeunt all
	_, _, err = e.execInstr(instr{
		stmt: parser.ExeuntStmt{Characters: nil, Line: 4, Col: 1},
	})
	if err != nil {
		t.Fatalf("Exeunt error: %v", err)
	}
	if e.stage.Size() != 0 {
		t.Errorf("expected empty stage, got %d", e.stage.Size())
	}

	// Exeunt Juliet only
	_ = e.stage.Enter([]string{"Romeo", "Juliet"}, e.syms, 0, 0)
	_, _, err = e.execInstr(instr{
		stmt: parser.ExeuntStmt{Characters: []string{"Juliet"}, Line: 5, Col: 1},
	})
	if err != nil {
		t.Fatalf("Exeunt Juliet error: %v", err)
	}
	if e.stage.Size() != 1 || !e.stage.Has("Romeo") {
		t.Errorf("expected only Romeo, got size=%d", e.stage.Size())
	}

	// Re-enter after backward goto: re-entering Romeo (already on stage) is a no-op (R-D3)
	_, _, err = e.execInstr(instr{
		stmt: parser.EnterStmt{Characters: []string{"Romeo"}, Line: 6, Col: 1},
	})
	if err != nil {
		t.Fatalf("re-enter Romeo error: %v", err)
	}
	if e.stage.Size() != 1 {
		t.Errorf("expected size still 1 after no-op re-enter, got %d", e.stage.Size())
	}
}

func TestExecAssign(t *testing.T) {
	e := testEnvWithChars(t, "Romeo", "Juliet")
	e.stage.Clear()
	_ = e.stage.Enter([]string{"Romeo", "Juliet"}, e.syms, 0, 0)
	// Assign to Juliet via Romeo speaking
	_, _, err := e.execInstr(instr{
		stmt:    parser.AssignStmt{Expr: parser.ConstExpr{AdjectiveCount: 0, Polarity: 1}},
		speaker: "Romeo",
	})
	if err != nil {
		t.Fatalf("Assign error: %v", err)
	}
	if e.values["juliet"] != 1 {
		t.Errorf("values[juliet] = %d; want 1", e.values["juliet"])
	}

	// Assign sum: Juliet = Romeo (3) + flower (1) = 4
	e.values["romeo"] = 3
	_, _, err = e.execInstr(instr{
		stmt: parser.AssignStmt{
			Expr: parser.BinaryOpExpr{
				Op:    "sum",
				Left:  parser.CharRefExpr{Name: "Romeo"},
				Right: parser.ConstExpr{AdjectiveCount: 0, Polarity: 1},
			},
		},
		speaker: "Romeo",
	})
	if err != nil {
		t.Fatalf("Assign sum error: %v", err)
	}
	if e.values["juliet"] != 4 {
		t.Errorf("values[juliet] = %d; want 4", e.values["juliet"])
	}
}

func TestExecSpeak(t *testing.T) {
	var out bytes.Buffer
	e := testEnvWithChars(t, "Romeo", "Juliet")
	e.out = &out
	e.stage.Clear()
	_ = e.stage.Enter([]string{"Romeo", "Juliet"}, e.syms, 0, 0)
	e.values["juliet"] = 72

	_, _, err := e.execInstr(instr{
		stmt:    parser.SpeakStmt{},
		speaker: "Romeo",
	})
	if err != nil {
		t.Fatalf("Speak error: %v", err)
	}
	if out.String() != "H" {
		t.Errorf("got %q; want 'H' (ASCII 72)", out.String())
	}

	out.Reset()
	e.values["juliet"] = 105
	_, _, err = e.execInstr(instr{
		stmt:    parser.SpeakStmt{},
		speaker: "Romeo",
	})
	if err != nil {
		t.Fatalf("Speak error: %v", err)
	}
	if out.String() != "i" {
		t.Errorf("got %q; want 'i' (ASCII 105)", out.String())
	}
}

func TestExecOpenHeart(t *testing.T) {
	var out bytes.Buffer
	e := testEnvWithChars(t, "Romeo", "Juliet")
	e.out = &out
	e.stage.Clear()
	_ = e.stage.Enter([]string{"Romeo", "Juliet"}, e.syms, 0, 0)
	e.values["juliet"] = 42

	_, _, err := e.execInstr(instr{
		stmt:    parser.OpenHeartStmt{},
		speaker: "Romeo",
	})
	if err != nil {
		t.Fatalf("OpenHeart error: %v", err)
	}
	if out.String() != "42\n" {
		t.Errorf("got %q; want '42\\n'", out.String())
	}
}

func TestExecOpenMind(t *testing.T) {
	e := testEnvWithChars(t, "Romeo", "Juliet")
	e.in = bytes.NewBufferString("A")
	e.stage.Clear()
	_ = e.stage.Enter([]string{"Romeo", "Juliet"}, e.syms, 0, 0)

	_, _, err := e.execInstr(instr{
		stmt:    parser.OpenMindStmt{},
		speaker: "Romeo",
	})
	if err != nil {
		t.Fatalf("OpenMind error: %v", err)
	}
	if e.values["juliet"] != 65 {
		t.Errorf("values[juliet] = %d; want 65", e.values["juliet"])
	}

	// EOF
	e.in = &bytes.Buffer{}
	_, _, err = e.execInstr(instr{
		stmt:    parser.OpenMindStmt{Line: 5, Col: 1},
		speaker: "Romeo",
	})
	if err == nil {
		t.Fatal("expected EOF error, got nil")
	}
	re, ok := err.(RuntimeError)
	if !ok || re.Code != "R003" {
		t.Fatalf("expected R003, got %v", err)
	}
}

func TestExecListen(t *testing.T) {
	e := testEnvWithChars(t, "Romeo", "Juliet")
	e.stage.Clear()
	_ = e.stage.Enter([]string{"Romeo", "Juliet"}, e.syms, 0, 0)

	tests := []struct {
		name  string
		input string
		want  int
		err   string
	}{
		{"positive", "123\n", 123, ""},
		{"negative with spaces", "  -7 xyz", -7, ""},
		{"non-numeric", "abc", 0, "R002"},
		{"empty input", "", 0, "R003"},
		{"sign then EOF", "  -", 0, "R002"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e.in = bytes.NewBufferString(tt.input)
			_, _, err := e.execInstr(instr{
				stmt:    parser.ListenStmt{Line: 1, Col: 1},
				speaker: "Romeo",
			})
			if tt.err != "" {
				if err == nil {
					t.Fatalf("expected error %s, got nil", tt.err)
				}
				re, ok := err.(RuntimeError)
				if !ok || re.Code != tt.err {
					t.Fatalf("expected %s, got %v", tt.err, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if e.values["juliet"] != tt.want {
				t.Errorf("values[juliet] = %d; want %d", e.values["juliet"], tt.want)
			}
		})
	}
}

func TestExecSelfTalkIO(t *testing.T) {
	var out bytes.Buffer
	e := testEnvWithChars(t, "Romeo")
	e.out = &out
	e.stage.Clear()
	_ = e.stage.Enter([]string{"Romeo"}, e.syms, 0, 0)
	e.values["romeo"] = 88

	// Self-talk: speaker is Romeo, listener resolves to Romeo (alone on stage)
	_, _, err := e.execInstr(instr{
		stmt:    parser.SpeakStmt{},
		speaker: "Romeo",
	})
	if err != nil {
		t.Fatalf("Self-talk Speak error: %v", err)
	}
	if out.String() != "X" {
		t.Errorf("got %q; want 'X' (ASCII 88)", out.String())
	}
}

func TestRemember(t *testing.T) {
	e := testEnvWithChars(t, "Romeo", "Juliet")
	e.stage.Clear()
	_ = e.stage.Enter([]string{"Romeo", "Juliet"}, e.syms, 0, 0)
	e.values["romeo"] = 3

	// Remember Romeo's value → push 3 onto Juliet's stack
	_, _, err := e.execInstr(instr{
		stmt:    parser.RememberStmt{Expr: parser.CharRefExpr{Name: "Romeo"}},
		speaker: "Romeo",
	})
	if err != nil {
		t.Fatalf("Remember error: %v", err)
	}
	if len(e.stacks["juliet"]) != 1 || e.stacks["juliet"][0] != 3 {
		t.Errorf("stacks[juliet] = %v; want [3]", e.stacks["juliet"])
	}

	// Remember a coward → push -1 onto Juliet's stack
	_, _, err = e.execInstr(instr{
		stmt:    parser.RememberStmt{Expr: parser.ConstExpr{AdjectiveCount: 0, Polarity: -1}},
		speaker: "Romeo",
	})
	if err != nil {
		t.Fatalf("Remember error: %v", err)
	}
	if len(e.stacks["juliet"]) != 2 || e.stacks["juliet"][1] != -1 {
		t.Errorf("stacks[juliet] = %v; want [3, -1]", e.stacks["juliet"])
	}
}

func TestRecall(t *testing.T) {
	e := testEnvWithChars(t, "Romeo", "Juliet")
	e.stage.Clear()
	_ = e.stage.Enter([]string{"Romeo", "Juliet"}, e.syms, 0, 0)
	e.stacks["juliet"] = []int{7, 8}
	e.values["romeo"] = 0

	// Recall → pop 8 from Juliet's stack, assign to Romeo
	_, _, err := e.execInstr(instr{
		stmt:    parser.RecallStmt{},
		speaker: "Romeo",
	})
	if err != nil {
		t.Fatalf("Recall error: %v", err)
	}
	if e.values["romeo"] != 8 {
		t.Errorf("values[romeo] = %d; want 8", e.values["romeo"])
	}
	if len(e.stacks["juliet"]) != 1 || e.stacks["juliet"][0] != 7 {
		t.Errorf("stacks[juliet] = %v; want [7]", e.stacks["juliet"])
	}

	// Recall again → pop 7, assign to Romeo
	_, _, err = e.execInstr(instr{
		stmt:    parser.RecallStmt{},
		speaker: "Romeo",
	})
	if err != nil {
		t.Fatalf("Recall error: %v", err)
	}
	if e.values["romeo"] != 7 {
		t.Errorf("values[romeo] = %d; want 7", e.values["romeo"])
	}
	if len(e.stacks["juliet"]) != 0 {
		t.Errorf("stacks[juliet] = %v; want empty", e.stacks["juliet"])
	}

	// Recall with empty stack → assign 0
	_, _, err = e.execInstr(instr{
		stmt:    parser.RecallStmt{},
		speaker: "Romeo",
	})
	if err != nil {
		t.Fatalf("Recall error: %v", err)
	}
	if e.values["romeo"] != 0 {
		t.Errorf("values[romeo] = %d; want 0", e.values["romeo"])
	}
}

func TestRememberYourself(t *testing.T) {
	e := testEnvWithChars(t, "Romeo", "Juliet")
	e.stage.Clear()
	_ = e.stage.Enter([]string{"Romeo", "Juliet"}, e.syms, 0, 0)
	e.values["juliet"] = 4

	// "Remember yourself." → PronounExpr{Ref:"listener"} → evaluation yields 4
	// → push 4 onto Juliet's (listener's) own stack
	_, _, err := e.execInstr(instr{
		stmt:    parser.RememberStmt{Expr: parser.PronounExpr{Ref: "listener"}},
		speaker: "Romeo",
	})
	if err != nil {
		t.Fatalf("Remember yourself error: %v", err)
	}
	if len(e.stacks["juliet"]) != 1 || e.stacks["juliet"][0] != 4 {
		t.Errorf("stacks[juliet] = %v; want [4]", e.stacks["juliet"])
	}
}

func TestRememberRecallRoundTrip(t *testing.T) {
	e := testEnvWithChars(t, "Romeo", "Juliet")
	e.stage.Clear()
	_ = e.stage.Enter([]string{"Romeo", "Juliet"}, e.syms, 0, 0)

	// Romeo: Remember me. → push Romeo's value (0) onto Juliet's stack
	_, _, err := e.execInstr(instr{
		stmt:    parser.RememberStmt{Expr: parser.PronounExpr{Ref: "speaker"}},
		speaker: "Romeo",
	})
	if err != nil {
		t.Fatalf("Remember me error: %v", err)
	}

	// Juliet: Recall! → pop from Juliet->Romeo's stack? No — listener of Juliet is Romeo
	// So Recall pops from Romeo's stack (listener of Juliet) into Juliet (speaker).
	// But Romeo's stack was never pushed to. So it pops 0.
	// This tests the asymmetry: Remember pushes to listener's stack,
	// Recall pops from listener's stack into speaker.
	e.stacks["juliet"] = []int{42}
	_, _, err = e.execInstr(instr{
		stmt:    parser.RecallStmt{},
		speaker: "Juliet",
	})
	if err != nil {
		t.Fatalf("Recall error: %v", err)
	}
	// Juliet's stack had 42; listener of Juliet is Romeo; pops Romeo's stack (which is empty) → 0
	// Wait no — listener of Juliet is Romeo (since Romeo is the other char on stage)
	// So Recall pops from Romeo's stack into Juliet (speaker)
	// Romeo's stack is empty → val = 0, assigned to Juliet
	if e.values["juliet"] != 0 {
		t.Errorf("values[juliet] = %d; want 0 (empty Romeo stack)", e.values["juliet"])
	}
	_ = 42
}

func TestApplyRelation(t *testing.T) {
	tests := []struct {
		name  string
		rel   string
		left  int
		right int
		want  bool
	}{
		{"equal true", "equal", 5, 5, true},
		{"equal false", "equal", 5, 3, false},
		{"not_equal true", "not_equal", 5, 3, true},
		{"not_equal false", "not_equal", 5, 5, false},
		{"greater true", "greater", 5, 3, true},
		{"greater false", "greater", 3, 5, false},
		{"less true", "less", 3, 5, true},
		{"less false", "less", 5, 3, false},
		{"greater_or_equal true >", "greater_or_equal", 5, 3, true},
		{"greater_or_equal true =", "greater_or_equal", 5, 5, true},
		{"greater_or_equal false", "greater_or_equal", 3, 5, false},
		{"less_or_equal true <", "less_or_equal", 3, 5, true},
		{"less_or_equal true =", "less_or_equal", 5, 5, true},
		{"less_or_equal false", "less_or_equal", 5, 3, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := applyRelation(tt.rel, tt.left, tt.right)
			if got != tt.want {
				t.Errorf("applyRelation(%q, %d, %d) = %v; want %v", tt.rel, tt.left, tt.right, got, tt.want)
			}
		})
	}
}

func TestExecQuestion(t *testing.T) {
	e := testEnvWithChars(t, "Romeo", "Juliet")
	e.stage.Clear()
	_ = e.stage.Enter([]string{"Romeo", "Juliet"}, e.syms, 0, 0)
	e.values["romeo"] = 5
	e.values["juliet"] = 3

	// Am I as good as you? → Romeo's value (5) > Juliet's value (3) → comparison true
	_, _, err := e.execInstr(instr{
		stmt: parser.QuestionStmt{
			Left:        parser.PronounExpr{Ref: "speaker"},
			Comparative: parser.Comparative{Relation: "greater"},
			Right:       parser.PronounExpr{Ref: "listener"},
		},
		speaker: "Romeo",
	})
	if err != nil {
		t.Fatalf("Question error: %v", err)
	}
	if !e.comparison {
		t.Error("expected comparison=true for 5 > 3")
	}

	// Equal question → false
	e.comparison = false
	_, _, err = e.execInstr(instr{
		stmt: parser.QuestionStmt{
			Left:        parser.PronounExpr{Ref: "speaker"},
			Comparative: parser.Comparative{Relation: "equal"},
			Right:       parser.PronounExpr{Ref: "listener"},
		},
		speaker: "Romeo",
	})
	if err != nil {
		t.Fatalf("Question error: %v", err)
	}
	if e.comparison {
		t.Error("expected comparison=false for 5 != 3")
	}
}

func TestExecIf(t *testing.T) {
	e := &env{
		stage:           &semantic.Stage{},
		sceneLabels:     map[string]map[string]int{"i": {"ii": 7}},
		actLabel:        map[string]int{},
		currentActRoman: "i",
	}

	tests := []struct {
		name       string
		comparison bool
		brIfTrue   bool
		target     string
		kind       string
		wantJump   bool
		wantPC     int
	}{
		{"if so matched", true, true, "II", "scene", true, 7},
		{"if so not matched", false, true, "II", "scene", false, 0},
		{"if not matched", false, false, "II", "scene", true, 7},
		{"if not not matched", true, false, "II", "scene", false, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e.comparison = tt.comparison
			pc, jumped, err := e.execInstr(instr{
				stmt: parser.IfStmt{
					BranchIfTrue: tt.brIfTrue,
					Target:       tt.target,
					TargetKind:   tt.kind,
				},
			})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if jumped != tt.wantJump {
				t.Errorf("jumped = %v; want %v", jumped, tt.wantJump)
			}
			if jumped && pc != tt.wantPC {
				t.Errorf("pc = %d; want %d", pc, tt.wantPC)
			}
		})
	}
}

func TestExecGoto(t *testing.T) {
	e := &env{
		stage:           &semantic.Stage{},
		actLabel:        map[string]int{"iii": 12},
		sceneLabels:     map[string]map[string]int{"i": {"ii": 7}},
		currentActRoman: "i",
	}

	// Goto scene II
	pc, jumped, err := e.execInstr(instr{
		stmt: parser.GotoStmt{Target: "II", TargetKind: "scene"},
	})
	if err != nil {
		t.Fatalf("Goto scene error: %v", err)
	}
	if !jumped || pc != 7 {
		t.Errorf("got jumped=%v, pc=%d; want jumped=true, pc=7", jumped, pc)
	}

	// Goto act III
	pc, jumped, err = e.execInstr(instr{
		stmt: parser.GotoStmt{Target: "III", TargetKind: "act"},
	})
	if err != nil {
		t.Fatalf("Goto act error: %v", err)
	}
	if !jumped || pc != 12 {
		t.Errorf("got jumped=%v, pc=%d; want jumped=true, pc=12", jumped, pc)
	}
}

func TestListenerDerivation(t *testing.T) {
	e := testEnvWithChars(t, "Romeo", "Juliet")
	e.stage.Clear()
	_ = e.stage.Enter([]string{"Romeo", "Juliet"}, e.syms, 0, 0)

	listener, ok := e.listener("Romeo")
	if !ok || listener != "Juliet" {
		t.Errorf("listener(Romeo) = %q, %v; want Juliet, true", listener, ok)
	}

	listener, ok = e.listener("Mercutio")
	if listener != "Mercutio" || ok {
		t.Errorf("listener(Mercutio) = %q, %v; want Mercutio, false", listener, ok)
	}

	e2 := testEnvWithChars(t, "Romeo")
	e2.stage.Clear()
	_ = e2.stage.Enter([]string{"Romeo"}, e2.syms, 0, 0)
	listener, ok = e2.listener("Romeo")
	if !ok || listener != "Romeo" {
		t.Errorf("self-talk listener(Romeo) = %q, %v; want Romeo, true", listener, ok)
	}
}
