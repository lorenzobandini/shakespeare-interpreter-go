package parser

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
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
	return parseTokensWithWarnings(t, path)
}

func parseTokensWithWarnings(t *testing.T, path string) *Program {
	t.Helper()
	tokens := lex(t, path)
	prog, err := New(tokens).Parse()
	if err != nil {
		t.Fatal(err)
	}
	return prog
}

// parseExprStr lexes and parses a single expression from a string within a dialogue context.
func parseExprStr(t *testing.T, input string) Expr {
	t.Helper()
	src := "The Test.\n\nRomeo, a man.\nJuliet, a woman.\n\nAct I: Test.\nScene I: Test.\n\n[Enter Romeo and Juliet]\n\nJuliet:\nYou are " + input + ".\n\n[Exeunt]\n"
	tokens, err := lexer.New(src).ScanTokens()
	if err != nil {
		t.Fatal(err)
	}
	p := New(tokens)
	prog, err := p.Parse()
	if err != nil {
		t.Fatalf("parse error for input %q: %v", input, err)
	}
	if len(prog.Acts) == 0 || len(prog.Acts[0].Scenes) == 0 {
		t.Fatal("no scenes")
	}
	stmts := prog.Acts[0].Scenes[0].Statements
	dlg := findDialogue(stmts)
	if dlg == nil {
		t.Fatalf("no dialogue found for %q; stmts=%d", input, len(stmts))
	}
	if len(dlg.Statements) == 0 {
		t.Fatalf("no statements in dialogue for %q", input)
	}
	as, ok := dlg.Statements[0].(AssignStmt)
	if !ok {
		t.Fatalf("expected AssignStmt at stmts[0], got %T (type: %T)", dlg.Statements[0], dlg.Statements[0])
	}
	return as.Expr
}

// findDialogue finds the first Dialogue node in a list of statements.
func findDialogue(stmts []Statement) *Dialogue {
	for _, s := range stmts {
		if d, ok := s.(Dialogue); ok {
			return &d
		}
	}
	return nil
}

// parseStmtStr lexes and parses a single statement from a string.
func parseStmtStr(t *testing.T, input string) Statement {
	t.Helper()
	src := "The Test.\n\nRomeo, a man.\n\nAct I: Test.\nScene I: Test.\n\n[Enter Romeo]\n\nRomeo:\n" + input + "\n\n[Exeunt]\n"
	tokens, err := lexer.New(src).ScanTokens()
	if err != nil {
		t.Fatal(err)
	}
	p := New(tokens)
	prog, err := p.Parse()
	if err != nil {
		t.Fatalf("parse error for %q: %v", input, err)
	}
	stmts := prog.Acts[0].Scenes[0].Statements
	dlg := findDialogue(stmts)
	if dlg == nil {
		t.Fatalf("no dialogue in stmts (len=%d) for %q", len(stmts), input)
	}
	if len(dlg.Statements) == 0 {
		t.Fatalf("no statements in dialogue for %q", input)
	}
	return dlg.Statements[0]
}

// -- Existing canonical fixture tests --

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

func TestParseCrossActPersistence(t *testing.T) {
	prog := parseTokens(t, "../../testdata/parser/cross-act-persistence.shpl")
	if len(prog.Acts) != 2 {
		t.Fatalf("acts: got %d, want 2", len(prog.Acts))
	}
	if len(prog.Acts[0].Scenes) != 1 {
		t.Errorf("Act I scenes: got %d, want 1", len(prog.Acts[0].Scenes))
	}
	if len(prog.Acts[1].Scenes) != 1 {
		t.Errorf("Act II scenes: got %d, want 1", len(prog.Acts[1].Scenes))
	}
}

// -- Expression unit tests --

func TestParseConstantSimple(t *testing.T) {
	expr := parseExprStr(t, "a flower")
	c, ok := expr.(ConstExpr)
	if !ok {
		t.Fatalf("expected ConstExpr, got %T", expr)
	}
	if c.AdjectiveCount != 0 || c.Noun != "flower" || c.Polarity != 1 {
		t.Errorf("got ConstExpr{adj=%d, noun=%q, pol=%d}", c.AdjectiveCount, c.Noun, c.Polarity)
	}
}

func TestParseConstantWithAdjectives(t *testing.T) {
	expr := parseExprStr(t, "a red hot flower")
	c, ok := expr.(ConstExpr)
	if !ok {
		t.Fatalf("expected ConstExpr, got %T", expr)
	}
	if c.AdjectiveCount != 2 || c.Noun != "flower" || c.Polarity != 1 {
		t.Errorf("got ConstExpr{adj=%d, noun=%q, pol=%d}", c.AdjectiveCount, c.Noun, c.Polarity)
	}
}

func TestParseConstantNegativeNoun(t *testing.T) {
	expr := parseExprStr(t, "a coward")
	c, ok := expr.(ConstExpr)
	if !ok {
		t.Fatalf("expected ConstExpr, got %T", expr)
	}
	if c.AdjectiveCount != 0 || c.Noun != "coward" || c.Polarity != -1 {
		t.Errorf("got ConstExpr{adj=%d, noun=%q, pol=%d}", c.AdjectiveCount, c.Noun, c.Polarity)
	}
}

func TestParseConstantVileCoward(t *testing.T) {
	expr := parseExprStr(t, "a vile coward")
	c, ok := expr.(ConstExpr)
	if !ok {
		t.Fatalf("expected ConstExpr, got %T", expr)
	}
	if c.AdjectiveCount != 1 || c.Noun != "coward" || c.Polarity != -1 {
		t.Errorf("got ConstExpr{adj=%d, noun=%q, pol=%d}", c.AdjectiveCount, c.Noun, c.Polarity)
	}
}

func TestParseConstantSummersDay(t *testing.T) {
	expr := parseExprStr(t, "summer's day")
	c, ok := expr.(ConstExpr)
	if !ok {
		t.Fatalf("expected ConstExpr, got %T", expr)
	}
	if c.AdjectiveCount != 1 || c.Noun != "day" || c.Polarity != 1 {
		t.Errorf("got ConstExpr{adj=%d, noun=%q, pol=%d}", c.AdjectiveCount, c.Noun, c.Polarity)
	}
}

func TestParseConstantUnknownNoun(t *testing.T) {
	expr := parseExprStr(t, "a chrysanthemum")
	c, ok := expr.(ConstExpr)
	if !ok {
		t.Fatalf("expected ConstExpr, got %T", expr)
	}
	if c.AdjectiveCount != 0 || c.Noun != "chrysanthemum" || c.Polarity != 1 {
		t.Errorf("got ConstExpr{adj=%d, noun=%q, pol=%d}", c.AdjectiveCount, c.Noun, c.Polarity)
	}
}

func TestParseCharRef(t *testing.T) {
	expr := parseExprStr(t, "Romeo")
	cr, ok := expr.(CharRefExpr)
	if !ok {
		t.Fatalf("expected CharRefExpr, got %T", expr)
	}
	if cr.Name != "Romeo" {
		t.Errorf("got CharRef{%q}", cr.Name)
	}
}

func TestParsePronounSpeaker(t *testing.T) {
	for _, tok := range []string{"me", "myself"} {
		t.Run(tok, func(t *testing.T) {
			expr := parseExprStr(t, tok)
			pr, ok := expr.(PronounExpr)
			if !ok {
				t.Fatalf("expected PronounExpr, got %T", expr)
			}
			if pr.Ref != "speaker" {
				t.Errorf("got Pronoun{%q}", pr.Ref)
			}
		})
	}
}

func TestParsePronounListener(t *testing.T) {
	for _, tok := range []string{"thyself", "yourself"} {
		t.Run(tok, func(t *testing.T) {
			expr := parseExprStr(t, tok)
			pr, ok := expr.(PronounExpr)
			if !ok {
				t.Fatalf("expected PronounExpr, got %T", expr)
			}
			if pr.Ref != "listener" {
				t.Errorf("got Pronoun{%q}", pr.Ref)
			}
		})
	}
}

// -- Binary operation unit tests --

func TestParseSum(t *testing.T) {
	expr := parseExprStr(t, "the sum of Romeo and a flower")
	b, ok := expr.(BinaryOpExpr)
	if !ok {
		t.Fatalf("expected BinaryOpExpr, got %T", expr)
	}
	if b.Op != "sum" {
		t.Errorf("op=%q", b.Op)
	}
}

func TestParseDifference(t *testing.T) {
	expr := parseExprStr(t, "the difference between Romeo and Juliet")
	b, ok := expr.(BinaryOpExpr)
	if !ok {
		t.Fatalf("expected BinaryOpExpr, got %T", expr)
	}
	if b.Op != "difference" {
		t.Errorf("op=%q", b.Op)
	}
}

func TestParseProduct(t *testing.T) {
	expr := parseExprStr(t, "the product of a flower and a flower")
	b, ok := expr.(BinaryOpExpr)
	if !ok {
		t.Fatalf("expected BinaryOpExpr, got %T", expr)
	}
	if b.Op != "product" {
		t.Errorf("op=%q", b.Op)
	}
}

func TestParseQuotient(t *testing.T) {
	expr := parseExprStr(t, "the quotient between a flower and a flower")
	b, ok := expr.(BinaryOpExpr)
	if !ok {
		t.Fatalf("expected BinaryOpExpr, got %T", expr)
	}
	if b.Op != "quotient" {
		t.Errorf("op=%q", b.Op)
	}
}

func TestParseRemainder(t *testing.T) {
	expr := parseExprStr(t, "the remainder of the quotient between a flower and a flower")
	b, ok := expr.(BinaryOpExpr)
	if !ok {
		t.Fatalf("expected BinaryOpExpr, got %T", expr)
	}
	if b.Op != "remainder" {
		t.Errorf("op=%q", b.Op)
	}
}

func TestParseNestedOps(t *testing.T) {
	expr := parseExprStr(t, "the sum of Romeo and the difference between Juliet and a flower")
	b, ok := expr.(BinaryOpExpr)
	if !ok {
		t.Fatalf("expected BinaryOpExpr, got %T", expr)
	}
	if b.Op != "sum" {
		t.Errorf("op=%q", b.Op)
	}
}

// -- Unary operation unit tests --

func TestParseSquare(t *testing.T) {
	expr := parseExprStr(t, "the square of a flower")
	u, ok := expr.(UnaryOpExpr)
	if !ok {
		t.Fatalf("expected UnaryOpExpr, got %T", expr)
	}
	if u.Op != "square" {
		t.Errorf("op=%q", u.Op)
	}
}

func TestParseSquareRoot(t *testing.T) {
	expr := parseExprStr(t, "the square root of a flower")
	u, ok := expr.(UnaryOpExpr)
	if !ok {
		t.Fatalf("expected UnaryOpExpr, got %T", expr)
	}
	if u.Op != "square_root" {
		t.Errorf("op=%q", u.Op)
	}
}

func TestParseCube(t *testing.T) {
	expr := parseExprStr(t, "the cube of a flower")
	u, ok := expr.(UnaryOpExpr)
	if !ok {
		t.Fatalf("expected UnaryOpExpr, got %T", expr)
	}
	if u.Op != "cube" {
		t.Errorf("op=%q", u.Op)
	}
}

func TestParseFactorial(t *testing.T) {
	expr := parseExprStr(t, "the factorial of a flower")
	u, ok := expr.(UnaryOpExpr)
	if !ok {
		t.Fatalf("expected UnaryOpExpr, got %T", expr)
	}
	if u.Op != "factorial" {
		t.Errorf("op=%q", u.Op)
	}
}

func TestParseTwice(t *testing.T) {
	expr := parseExprStr(t, "twice a flower")
	u, ok := expr.(UnaryOpExpr)
	if !ok {
		t.Fatalf("expected UnaryOpExpr, got %T", expr)
	}
	if u.Op != "twice" {
		t.Errorf("op=%q", u.Op)
	}
}

// -- Statement unit tests --

func TestParseAssignCopula(t *testing.T) {
	stmt := parseStmtStr(t, "You are as good as a flower.")
	as, ok := stmt.(AssignStmt)
	if !ok {
		t.Fatalf("expected AssignStmt, got %T", stmt)
	}
	if as.Target != "you" || as.SimileAdj != "good" || as.Terminator != "." {
		t.Errorf("got AssignStmt{target=%q, simile=%q, term=%q}", as.Target, as.SimileAdj, as.Terminator)
	}
}

func TestParseAssignNoCopula(t *testing.T) {
	stmt := parseStmtStr(t, "You lying stupid fatherless big smelly half-witted coward!")
	as, ok := stmt.(AssignStmt)
	if !ok {
		t.Fatalf("expected AssignStmt, got %T", stmt)
	}
	if as.Target != "you" || as.SimileAdj != "" || as.Terminator != "!" {
		t.Errorf("got AssignStmt{target=%q, simile=%q, term=%q}", as.Target, as.SimileAdj, as.Terminator)
	}
	c, ok := as.Expr.(ConstExpr)
	if !ok {
		t.Fatalf("expected ConstExpr, got %T", as.Expr)
	}
	if c.AdjectiveCount != 6 || c.Noun != "coward" || c.Polarity != -1 {
		t.Errorf("got ConstExpr{adj=%d, noun=%q, pol=%d}", c.AdjectiveCount, c.Noun, c.Polarity)
	}
}

func TestParseSpeak(t *testing.T) {
	stmt := parseStmtStr(t, "Speak your mind.")
	s, ok := stmt.(SpeakStmt)
	if !ok {
		t.Fatalf("expected SpeakStmt, got %T", stmt)
	}
	if s.Terminator != "." {
		t.Errorf("term=%q", s.Terminator)
	}
}

func TestParseSpeakThy(t *testing.T) {
	stmt := parseStmtStr(t, "Speak thy mind!")
	s, ok := stmt.(SpeakStmt)
	if !ok {
		t.Fatalf("expected SpeakStmt, got %T", stmt)
	}
	if s.Terminator != "!" {
		t.Errorf("term=%q", s.Terminator)
	}
}

func TestParseOpenHeart(t *testing.T) {
	stmt := parseStmtStr(t, "Open your heart.")
	_, ok := stmt.(OpenHeartStmt)
	if !ok {
		t.Fatalf("expected OpenHeartStmt, got %T", stmt)
	}
}

func TestParseOpenMind(t *testing.T) {
	stmt := parseStmtStr(t, "Open your mind.")
	_, ok := stmt.(OpenMindStmt)
	if !ok {
		t.Fatalf("expected OpenMindStmt, got %T", stmt)
	}
}

func TestParseListen(t *testing.T) {
	stmt := parseStmtStr(t, "Listen to your heart!")
	_, ok := stmt.(ListenStmt)
	if !ok {
		t.Fatalf("expected ListenStmt, got %T", stmt)
	}
}

func TestParseRemember(t *testing.T) {
	stmt := parseStmtStr(t, "Remember me.")
	r, ok := stmt.(RememberStmt)
	if !ok {
		t.Fatalf("expected RememberStmt, got %T", stmt)
	}
	pr, ok := r.Expr.(PronounExpr)
	if !ok {
		t.Fatalf("expected PronounExpr, got %T", r.Expr)
	}
	if pr.Ref != "speaker" {
		t.Errorf("got Pronoun{%q}", pr.Ref)
	}
}

func TestParseRecall(t *testing.T) {
	stmt := parseStmtStr(t, "Recall your tragic fate.")
	r, ok := stmt.(RecallStmt)
	if !ok {
		t.Fatalf("expected RecallStmt, got %T", stmt)
	}
	if r.IgnoredText != "your tragic fate" {
		t.Errorf("ignored=%q", r.IgnoredText)
	}
}

func TestParseRecallEmpty(t *testing.T) {
	stmt := parseStmtStr(t, "Recall.")
	r, ok := stmt.(RecallStmt)
	if !ok {
		t.Fatalf("expected RecallStmt, got %T", stmt)
	}
	if r.IgnoredText != "" {
		t.Errorf("ignored=%q", r.IgnoredText)
	}
}

// -- Comparative unit tests --

func TestParseComparativeForms(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		relation string
		form     string
		negated  bool
	}{
		{"as pos as", "as good as", "equal", "as-as", false},
		{"as neg as", "as bad as", "not_equal", "as-as", false},
		{"pos than", "better than", "greater", "than", false},
		{"neg than", "worse than", "less", "than", false},
		{"not as pos as", "not as good as", "not_equal", "as-as", true},
		{"not pos than", "not better than", "less_or_equal", "than", true},
		{"not neg than", "not worse than", "greater_or_equal", "than", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			src := "The Test.\n\nRomeo, a man.\nJuliet, a woman.\n\nAct I: Test.\nScene I: Test.\n\n[Enter Romeo and Juliet]\n\nJuliet:\nAm I " + tt.input + " you?\n\n[Exeunt]\n"
			tokens, err := lexer.New(src).ScanTokens()
			if err != nil {
				t.Fatal(err)
			}
			p := New(tokens)
			prog, err := p.Parse()
			if err != nil {
				t.Fatalf("parse error: %v", err)
			}
			stmts := prog.Acts[0].Scenes[0].Statements
			dlg := findDialogue(stmts)
			if dlg == nil {
				t.Fatal("expected dialogue")
			}
			q, ok := dlg.Statements[0].(QuestionStmt)
			if !ok {
				t.Fatalf("expected QuestionStmt, got %T", dlg.Statements[0])
			}
			if q.Comparative.Relation != tt.relation {
				t.Errorf("relation: got %q, want %q", q.Comparative.Relation, tt.relation)
			}
			if q.Comparative.Form != tt.form {
				t.Errorf("form: got %q, want %q", q.Comparative.Form, tt.form)
			}
			if q.Comparative.Negated != tt.negated {
				t.Errorf("negated: got %v, want %v", q.Comparative.Negated, tt.negated)
			}
		})
	}
}

func TestParseQuestionIs(t *testing.T) {
	src := "The Test.\n\nRomeo, a man.\nJuliet, a woman.\n\nAct I: Test.\nScene I: Test.\n\n[Enter Romeo and Juliet]\n\nJuliet:\nIs Romeo as good as Juliet?\n\n[Exeunt]\n"
	tokens, err := lexer.New(src).ScanTokens()
	if err != nil {
		t.Fatal(err)
	}
	p := New(tokens)
	prog, err := p.Parse()
	if err != nil {
		t.Fatal(err)
	}
	stmts := prog.Acts[0].Scenes[0].Statements
	dlg := findDialogue(stmts)
	if dlg == nil {
		t.Fatal("expected dialogue")
	}
	q, ok := dlg.Statements[0].(QuestionStmt)
	if !ok {
		t.Fatalf("expected QuestionStmt, got %T", dlg.Statements[0])
	}
	if q.Comparative.Relation != "equal" {
		t.Errorf("relation: got %q", q.Comparative.Relation)
	}
}

// -- Control flow unit tests --

func TestParseIfSo(t *testing.T) {
	stmt := parseStmtStr(t, "If so, let us proceed to scene II.")
	is, ok := stmt.(IfStmt)
	if !ok {
		t.Fatalf("expected IfStmt, got %T", stmt)
	}
	if !is.BranchIfTrue || is.Target != "II" || is.TargetKind != "scene" {
		t.Errorf("got IfStmt{branch=%v, target=%s %s}", is.BranchIfTrue, is.TargetKind, is.Target)
	}
}

func TestParseIfNot(t *testing.T) {
	stmt := parseStmtStr(t, "If not, let us proceed to scene I.")
	is, ok := stmt.(IfStmt)
	if !ok {
		t.Fatalf("expected IfStmt, got %T", stmt)
	}
	if is.BranchIfTrue || is.Target != "I" || is.TargetKind != "scene" {
		t.Errorf("got IfStmt{branch=%v, target=%s %s}", is.BranchIfTrue, is.TargetKind, is.Target)
	}
}

func TestParseGotoProceed(t *testing.T) {
	stmt := parseStmtStr(t, "Let us proceed to scene II.")
	gs, ok := stmt.(GotoStmt)
	if !ok {
		t.Fatalf("expected GotoStmt, got %T", stmt)
	}
	if gs.Target != "II" || gs.TargetKind != "scene" {
		t.Errorf("got GotoStmt{target=%s %s}", gs.TargetKind, gs.Target)
	}
}

func TestParseGotoReturn(t *testing.T) {
	stmt := parseStmtStr(t, "Let us return to act I.")
	gs, ok := stmt.(GotoStmt)
	if !ok {
		t.Fatalf("expected GotoStmt, got %T", stmt)
	}
	if gs.Target != "I" || gs.TargetKind != "act" {
		t.Errorf("got GotoStmt{target=%s %s}", gs.TargetKind, gs.Target)
	}
}

// -- Golden snapshot tests --

func TestGoldenSnapshots(t *testing.T) {
	fixtures := []string{
		"../../testdata/lexer/minimal.shpl",
		"../../testdata/lexer/hello.shpl",
		"../../testdata/lexer/truth-machine.shpl",
		"../../testdata/parser/arithmetic.shpl",
		"../../testdata/parser/stack.shpl",
		"../../testdata/parser/conditionals.shpl",
	}
	for _, path := range fixtures {
		name := strings.TrimSuffix(filepath.Base(path), ".shpl")
		t.Run(name, func(t *testing.T) {
			tokens := lex(t, path)
			p := New(tokens)
			prog, err := p.Parse()
			if err != nil {
				t.Fatalf("parse error for %s: %v", path, err)
			}
			got, err := json.MarshalIndent(prog, "", "  ")
			if err != nil {
				t.Fatal(err)
			}
			goldenPath := filepath.Join(filepath.Dir(path), name+".golden.json")
			if _, err := os.Stat(goldenPath); os.IsNotExist(err) {
				if err := os.WriteFile(goldenPath, got, 0644); err != nil {
					t.Fatal(err)
				}
				t.Fatalf("golden file %s created; inspect and re-run", goldenPath)
			}
			want, err := readFile(goldenPath)
			if err != nil {
				t.Fatal(err)
			}
			if string(got) != string(want) {
				t.Errorf("golden mismatch for %s", name)
			}
		})
	}
}

// -- Error tests --

func TestParseErrors(t *testing.T) {
	tests := []struct {
		name string
		path string
		code string
	}{
		// S001 is tested in TestParseErrorS001Empty (empty input)
		{"no characters", "../../testdata/parser/error-no-chars.shpl", "S002"},
		{"bad act num", "../../testdata/parser/error-bad-act-num.shpl", "S005"},
		{"bad scene num", "../../testdata/parser/error-bad-scene-num.shpl", "S008"},
		{"no enter", "../../testdata/parser/error-no-enter.shpl", "S013"},
		{"bad expression", "../../testdata/parser/error-bad-expr.shpl", "S015"},
		{"bad if", "../../testdata/parser/error-bad-if.shpl", "S016"},
		{"bad comparative", "../../testdata/parser/error-bad-comparative.shpl", "S017"},
		{"bad stack", "../../testdata/parser/error-bad-stack.shpl", "S018"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokens := lex(t, tt.path)
			_, err := New(tokens).Parse()
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			pe, ok := err.(ParseError)
			if !ok {
				t.Fatalf("expected ParseError, got %T: %v", err, err)
			}
			if pe.Code != tt.code {
				t.Errorf("expected code %s, got %s: %v", tt.code, pe.Code, err)
			}
		})
	}
}

func TestParseErrorS001Empty(t *testing.T) {
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

func TestWarningIsNotError(t *testing.T) {
	src := "The Test.\n\nCoriolanus, a Roman general.\n\nAct I: Test.\nScene I: Test.\n\n[Enter Coriolanus]\n\nCoriolanus:\nSpeak your mind.\n\n[Exeunt]\n"
	tokens, err := lexer.New(src).ScanTokens()
	if err != nil {
		t.Fatal(err)
	}
	prog, err := New(tokens).Parse()
	if err != nil {
		t.Fatal(err)
	}
	if len(prog.Warnings) == 0 {
		t.Fatal("expected a warning for non-Shakespeare character 'Coriolanus'")
	}
	w := prog.Warnings[0]
	if w.Code != "S003" {
		t.Errorf("expected S003 warning, got %s: %s", w.Code, w.Msg)
	}
}
