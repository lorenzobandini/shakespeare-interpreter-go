package semantic

import (
	"os"
	"strings"
	"testing"

	"github.com/lorenzobandini/shakespeare-interpreter-go/internal/lexer"
	"github.com/lorenzobandini/shakespeare-interpreter-go/internal/logger"
	"github.com/lorenzobandini/shakespeare-interpreter-go/internal/parser"
)

func TestMain(m *testing.M) {
	logger.Init(logger.LevelDebug)
	m.Run()
}

func TestSemanticErrorFormat(t *testing.T) {
	tests := []struct {
		name string
		err  SemanticError
		want string
	}{
		{
			name: "M001 default filename",
			err:  errUndefinedCharacter("Banquo", 3, 1),
			want: "error[M001]: character 'Banquo' is not declared\n  --> input:3:1",
		},
		{
			name: "M001 with filename",
			err:  SemanticError{Code: "M001", Msg: "character 'Banquo' is not declared", Line: 3, Col: 1, Filename: "play.spl"},
			want: "error[M001]: character 'Banquo' is not declared\n  --> play.spl:3:1",
		},
		{
			name: "M002",
			err:  errTooManyOnStage(3, 5, 7),
			want: "error[M002]: too many characters on stage (max 2), got 3\n  --> input:5:7",
		},
		{
			name: "M003",
			err:  errStageOverflow([]string{"Romeo", "Juliet"}, 10, 3),
			want: "error[M003]: cannot enter: stage is full (Romeo, Juliet already on stage)\n  --> input:10:3",
		},
		{
			name: "M004",
			err:  errCharacterNotOnStage("Romeo", 15, 1),
			want: "error[M004]: character 'Romeo' is not on stage\n  --> input:15:1",
		},
		{
			name: "M005",
			err:  errExitNotOnStage("Hamlet", 20, 5),
			want: "error[M005]: character 'Hamlet' is not on stage\n  --> input:20:5",
		},
		{
			name: "M006 scene with context",
			err:  errUndefinedScene("V", "scene", "I", 25, 3),
			want: "error[M006]: scene V is not defined in Act I\n  --> input:25:3",
		},
		{
			name: "M006 act without context",
			err:  errUndefinedScene("IX", "act", "", 30, 1),
			want: "error[M006]: act IX is not defined\n  --> input:30:1",
		},
		{
			name: "M007 self-enter",
			err:  errSelfReferenceEnter(35, 10),
			want: "error[M007]: cannot enter the same character twice\n  --> input:35:10",
		},
		{
			name: "M007 already on stage",
			err:  errAlreadyOnStage("Romeo", 36, 5),
			want: "error[M007]: cannot enter 'Romeo': already on stage\n  --> input:36:5",
		},
		{
			name: "M008",
			err:  errNoSceneInAct("I", 40, 1),
			want: "error[M008]: Act I has no scenes\n  --> input:40:1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.err.Error()
			if got != tt.want {
				t.Errorf("got:\n%s\nwant:\n%s", got, tt.want)
			}
		})
	}
}

func TestNewSymbolTable(t *testing.T) {
	decls := []parser.CharacterDecl{
		{Name: "Romeo", Description: "a young man", Line: 2, Col: 1},
		{Name: "JULIET", Description: "a young woman", Line: 3, Col: 1},
		{Name: "Hamlet", Description: "a prince", Line: 4, Col: 1},
	}
	st, errs := newSymbolTable(decls)
	if len(errs) != 0 {
		t.Fatalf("expected no errors, got %v", errs)
	}
	if !st.Has("romeo") {
		t.Error("expected Has(romeo)=true")
	}
	if !st.Has("juliet") {
		t.Error("expected Has(juliet)=true")
	}
	if !st.Has("hamlet") {
		t.Error("expected Has(hamlet)=true")
	}
	if st.Has("banquo") {
		t.Error("expected Has(banquo)=false")
	}
	sym, ok := st.Get("romeo")
	if !ok || sym.Name != "Romeo" {
		t.Errorf("expected Romeo, got %+v", sym)
	}
	sym, ok = st.Get("juliet")
	if !ok || sym.Name != "JULIET" {
		t.Errorf("expected JULIET, got %+v", sym)
	}
}

func TestNewSymbolTableDuplicate(t *testing.T) {
	decls := []parser.CharacterDecl{
		{Name: "Romeo", Description: "a man", Line: 2, Col: 1},
		{Name: "ROMEO", Description: "again", Line: 3, Col: 1},
	}
	_, errs := newSymbolTable(decls)
	if len(errs) != 1 {
		t.Fatalf("expected 1 error for duplicate, got %d: %v", len(errs), errs)
	}
	if errs[0].Code != "M001" {
		t.Errorf("expected M001, got %s", errs[0].Code)
	}
}

func TestCheckDeclared(t *testing.T) {
	decls := []parser.CharacterDecl{
		{Name: "Romeo", Description: "a man", Line: 2, Col: 1},
		{Name: "Juliet", Description: "a woman", Line: 3, Col: 1},
	}
	prog := &parser.Program{Characters: decls}
	a := New("test", prog)
	if err := a.checkDeclared("Romeo", 5, 1); err != nil {
		t.Errorf("expected nil for 'Romeo', got %v", err)
	}
	if err := a.checkDeclared("romeo", 5, 1); err != nil {
		t.Errorf("expected nil for 'romeo' (case-insensitive), got %v", err)
	}
	if err := a.checkDeclared("Banquo", 10, 3); err == nil {
		t.Error("expected M001 for 'Banquo', got nil")
	} else if err.Code != "M001" {
		t.Errorf("expected M001, got %s", err.Code)
	} else if !strings.Contains(err.Msg, "Banquo") {
		t.Errorf("expected message to mention Banquo, got %q", err.Msg)
	}
}

func TestStageEnter(t *testing.T) {
	syms, _ := newSymbolTable([]parser.CharacterDecl{
		{Name: "Romeo", Description: "a man"},
		{Name: "Juliet", Description: "a woman"},
		{Name: "Hamlet", Description: "a prince"},
	})
	tests := []struct {
		name     string
		initial  []string
		enter    []string
		wantSize int
		wantErrs int
	}{
		{"single enter", nil, []string{"Romeo"}, 1, 0},
		{"two enter", nil, []string{"Romeo", "Juliet"}, 2, 0},
		{"self-enter duplicate", nil, []string{"Romeo", "Romeo"}, 0, 1},
		{"already on stage", []string{"Romeo"}, []string{"Romeo"}, 1, 1},
		{"overflow when full", []string{"Romeo", "Juliet"}, []string{"Hamlet"}, 2, 1},
		{"too many in one enter", nil, []string{"Romeo", "Juliet", "Hamlet"}, 0, 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Stage{}
			s.names = append(s.names, tt.initial...)
			errs := s.Enter(tt.enter, syms, 1, 1)
			if len(errs) != tt.wantErrs {
				t.Errorf("got %d errors, want %d: %v", len(errs), tt.wantErrs, errs)
			}
			if s.Size() != tt.wantSize {
				t.Errorf("stage size: got %d, want %d", s.Size(), tt.wantSize)
			}
		})
	}
}

func TestStageExitExeunt(t *testing.T) {
	syms, _ := newSymbolTable([]parser.CharacterDecl{
		{Name: "Romeo", Description: "a man"},
		{Name: "Juliet", Description: "a woman"},
		{Name: "Hamlet", Description: "a prince"},
	})
	t.Run("exit named", func(t *testing.T) {
		s := &Stage{names: []string{"Romeo"}}
		errs := s.Exit("Romeo", syms, 1, 1)
		if len(errs) != 0 {
			t.Fatalf("unexpected errors: %v", errs)
		}
		if s.Size() != 0 {
			t.Errorf("size: got %d, want 0", s.Size())
		}
	})
	t.Run("exit not on stage", func(t *testing.T) {
		s := &Stage{names: []string{"Romeo"}}
		errs := s.Exit("Hamlet", syms, 1, 1)
		if len(errs) != 1 || errs[0].Code != "M005" {
			t.Fatalf("expected M005, got %v", errs)
		}
	})
	t.Run("exeunt all", func(t *testing.T) {
		s := &Stage{names: []string{"Romeo", "Juliet"}}
		errs := s.Exeunt(nil, syms, 1, 1)
		if len(errs) != 0 {
			t.Fatalf("unexpected errors: %v", errs)
		}
		if s.Size() != 0 {
			t.Errorf("size: got %d, want 0", s.Size())
		}
	})
	t.Run("exeunt named", func(t *testing.T) {
		s := &Stage{names: []string{"Romeo", "Juliet"}}
		errs := s.Exeunt([]string{"Juliet"}, syms, 1, 1)
		if len(errs) != 0 {
			t.Fatalf("unexpected errors: %v", errs)
		}
		if s.Size() != 1 || s.names[0] != "Romeo" {
			t.Errorf("expected [Romeo], got %v", s.names)
		}
	})
	t.Run("exeunt absent", func(t *testing.T) {
		s := &Stage{names: []string{"Romeo"}}
		errs := s.Exeunt([]string{"Hamlet"}, syms, 1, 1)
		if len(errs) != 1 || errs[0].Code != "M005" {
			t.Fatalf("expected M005, got %v", errs)
		}
	})
}

func TestListenerResolution(t *testing.T) {
	t.Run("two on stage", func(t *testing.T) {
		s := &Stage{names: []string{"Romeo", "Juliet"}}
		listener, ok := s.Listener("Romeo")
		if !ok || listener != "Juliet" {
			t.Errorf("expected Juliet, got %q (ok=%v)", listener, ok)
		}
	})
	t.Run("self talk", func(t *testing.T) {
		s := &Stage{names: []string{"Romeo"}}
		listener, ok := s.Listener("Romeo")
		if !ok || listener != "Romeo" {
			t.Errorf("expected Romeo (self), got %q (ok=%v)", listener, ok)
		}
	})
	t.Run("empty stage", func(t *testing.T) {
		s := &Stage{}
		_, ok := s.Listener("Romeo")
		if ok {
			t.Error("expected ok=false for empty stage")
		}
	})
}

// -- Helpers for fixture-based integration tests --

func analyzeFile(t *testing.T, path string) Result {
	t.Helper()
	src, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	tokens, err := lexer.New(string(src)).ScanTokens()
	if err != nil {
		t.Fatal(err)
	}
	prog, err := parser.New(tokens).Parse()
	if err != nil {
		t.Fatalf("parse error for %s: %v", path, err)
	}
	a := New(path, prog)
	return a.Analyze(prog)
}

func codeSet(errs []SemanticError) map[string]bool {
	s := make(map[string]bool, len(errs))
	for _, e := range errs {
		s[e.Code] = true
	}
	return s
}

// -- Integration tests --

func TestAnalyzeEmptyScenesDefensive(t *testing.T) {
	prog := &parser.Program{
		Title: parser.Title{Text: "Test"},
		Characters: []parser.CharacterDecl{
			{Name: "Romeo", Description: "a man", Line: 2, Col: 1},
		},
		Acts: []parser.Act{
			{Number: 1, RomanNumeral: "I", Line: 5, Col: 1},
		},
	}
	a := New("test", prog)
	res := a.Analyze(prog)
	if len(res.Errors) != 1 || res.Errors[0].Code != "M008" {
		t.Fatalf("expected one M008 error, got %v", res.Errors)
	}
}

func TestAnalyzeWalksMultipleActsWithPersistence(t *testing.T) {
	res := analyzeFile(t, "../../testdata/semantic/primes-persistence.spl")
	if !res.OK() {
		t.Fatalf("expected zero errors, got %v", res.Errors)
	}
}

func TestAnalyzeHelloWorld(t *testing.T) {
	res := analyzeFile(t, "../../testdata/semantic/hello.spl")
	if !res.OK() {
		t.Fatalf("expected zero errors, got %v", res.Errors)
	}
}

func TestAnalyzeMinimalValid(t *testing.T) {
	res := analyzeFile(t, "../../testdata/semantic/minimal-valid.spl")
	if !res.OK() {
		t.Fatalf("expected zero errors, got %v", res.Errors)
	}
}

func TestAnalyzeSelfTalk(t *testing.T) {
	res := analyzeFile(t, "../../testdata/semantic/self-talk.spl")
	if !res.OK() {
		t.Fatalf("expected zero errors, got %v", res.Errors)
	}
}

func TestAnalyzeOffStageValueRead(t *testing.T) {
	res := analyzeFile(t, "../../testdata/semantic/off-stage-value-read.spl")
	if !res.OK() {
		t.Fatalf("expected zero errors (D2: off-stage value reads allowed), got %v", res.Errors)
	}
}

func TestAnalyzeValidOperations(t *testing.T) {
	res := analyzeFile(t, "../../testdata/semantic/valid-operations.spl")
	if !res.OK() {
		t.Fatalf("expected zero errors, got %v", res.Errors)
	}
}

func TestAnalyzeTruthMachine(t *testing.T) {
	res := analyzeFile(t, "../../testdata/semantic/truth-machine.spl")
	if !res.OK() {
		t.Fatalf("expected zero errors, got %v", res.Errors)
	}
}

func TestM001UndeclaredEnter(t *testing.T) {
	res := analyzeFile(t, "../../testdata/semantic/m001-undeclared-enter.spl")
	if res.OK() {
		t.Fatal("expected errors, got none")
	}
	codes := codeSet(res.Errors)
	if !codes["M001"] {
		t.Fatalf("expected M001 in errors, got %v", res.Errors)
	}
}

func TestM001UndeclaredSpeaker(t *testing.T) {
	res := analyzeFile(t, "../../testdata/semantic/m001-undeclared-speaker.spl")
	if res.OK() {
		t.Fatal("expected errors, got none")
	}
	codes := codeSet(res.Errors)
	if !codes["M001"] {
		t.Fatalf("expected M001 in errors, got %v", res.Errors)
	}
}

func TestM003StageOverflow(t *testing.T) {
	res := analyzeFile(t, "../../testdata/semantic/m003-stage-overflow.spl")
	if res.OK() {
		t.Fatal("expected errors, got none")
	}
	codes := codeSet(res.Errors)
	if !codes["M003"] {
		t.Fatalf("expected M003 in errors, got %v", res.Errors)
	}
}

func TestM004SpeakerNotOnStage(t *testing.T) {
	res := analyzeFile(t, "../../testdata/semantic/m004-speaker-not-on-stage.spl")
	if res.OK() {
		t.Fatal("expected errors, got none")
	}
	codes := codeSet(res.Errors)
	if !codes["M004"] {
		t.Fatalf("expected M004 in errors, got %v", res.Errors)
	}
}

func TestM004EmptyStageCrossAct(t *testing.T) {
	res := analyzeFile(t, "../../testdata/semantic/m004-empty-stage-cross-act.spl")
	if res.OK() {
		t.Fatal("expected errors, got none")
	}
	codes := codeSet(res.Errors)
	if !codes["M004"] {
		t.Fatalf("expected M004 in errors, got %v", res.Errors)
	}
}

func TestM005ExitNotOnStage(t *testing.T) {
	res := analyzeFile(t, "../../testdata/semantic/m005-exit-not-on-stage.spl")
	if res.OK() {
		t.Fatal("expected errors, got none")
	}
	codes := codeSet(res.Errors)
	if !codes["M005"] {
		t.Fatalf("expected M005 in errors, got %v", res.Errors)
	}
}

func TestM006UndefinedScene(t *testing.T) {
	res := analyzeFile(t, "../../testdata/semantic/m006-undefined-scene.spl")
	if res.OK() {
		t.Fatal("expected errors, got none")
	}
	codes := codeSet(res.Errors)
	if !codes["M006"] {
		t.Fatalf("expected M006 in errors, got %v", res.Errors)
	}
}

func TestM006UndefinedAct(t *testing.T) {
	res := analyzeFile(t, "../../testdata/semantic/m006-undefined-act.spl")
	if res.OK() {
		t.Fatal("expected errors, got none")
	}
	codes := codeSet(res.Errors)
	if !codes["M006"] {
		t.Fatalf("expected M006 in errors, got %v", res.Errors)
	}
}

func TestM007SelfEnter(t *testing.T) {
	res := analyzeFile(t, "../../testdata/semantic/m007-self-enter.spl")
	if res.OK() {
		t.Fatal("expected errors, got none")
	}
	codes := codeSet(res.Errors)
	if !codes["M007"] {
		t.Fatalf("expected M007 in errors, got %v", res.Errors)
	}
}

func TestAnalyzeMultipleErrors(t *testing.T) {
	res := analyzeFile(t, "../../testdata/semantic/multiple-errors.spl")
	if res.OK() {
		t.Fatal("expected multiple errors, got none")
	}
	codes := codeSet(res.Errors)
	if !codes["M001"] {
		t.Errorf("expected M001 in error set, got %v", res.Errors)
	}
	if !codes["M003"] {
		t.Errorf("expected M003 in error set, got %v", res.Errors)
	}
}

func TestGotoCrossActScene(t *testing.T) {
	res := analyzeFile(t, "../../testdata/semantic/goto-cross-act-scene.spl")
	if res.OK() {
		t.Fatal("expected some errors (scene III doesn't exist in Act II), got none")
	}
	codes := codeSet(res.Errors)
	if !codes["M006"] {
		t.Fatalf("expected M006 in errors, got %v", res.Errors)
	}
}

func TestAnalyzeExprUnit(t *testing.T) {
	syms, _ := newSymbolTable([]parser.CharacterDecl{
		{Name: "Romeo", Description: "a man"},
		{Name: "Juliet", Description: "a woman"},
	})
	a := &Analyzer{symbols: syms, stage: Stage{names: []string{}}}
	ctx := dialogCtx{speaker: "Romeo", listener: "Juliet"}

	// CharRefExpr undeclared → M001
	a.analyzeExpr(parser.CharRefExpr{Name: "Banquo", Line: 1, Col: 1}, ctx)
	if len(a.errs) != 1 || a.errs[0].Code != "M001" {
		t.Fatalf("expected M001 for undeclared CharRef, got %v", a.errs)
	}
	a.errs = nil

	// CharRefExpr declared but off-stage → no error (D2)
	a.analyzeExpr(parser.CharRefExpr{Name: "Romeo", Line: 1, Col: 1}, ctx)
	if len(a.errs) != 0 {
		t.Fatalf("expected no error for declared off-stage CharRef (D2), got %v", a.errs)
	}

	// BinaryOpExpr with valid nested expressions
	a.analyzeExpr(parser.BinaryOpExpr{
		Op:    "sum",
		Left:  parser.CharRefExpr{Name: "Romeo", Line: 1, Col: 1},
		Right: parser.ConstExpr{Noun: "flower", Polarity: 1, Line: 1, Col: 1},
	}, ctx)
	if len(a.errs) != 0 {
		t.Fatalf("expected no errors for valid BinaryOpExpr, got %v", a.errs)
	}

	// UnaryOpExpr with valid operand
	a.analyzeExpr(parser.UnaryOpExpr{
		Op:      "square",
		Operand: parser.CharRefExpr{Name: "Romeo", Line: 1, Col: 1},
	}, ctx)
	if len(a.errs) != 0 {
		t.Fatalf("expected no errors for valid UnaryOpExpr, got %v", a.errs)
	}
}

func TestBuildRegistries(t *testing.T) {
	acts := []parser.Act{
		{Number: 1, RomanNumeral: "I", Scenes: []parser.Scene{
			{Number: 1, RomanNumeral: "I"},
			{Number: 2, RomanNumeral: "II"},
		}},
		{Number: 2, RomanNumeral: "II", Scenes: []parser.Scene{
			{Number: 1, RomanNumeral: "I"},
			{Number: 3, RomanNumeral: "III"},
		}},
	}
	actReg := buildActRegistry(acts)
	if _, ok := actReg.Resolve("ii"); !ok {
		t.Error("expected actReg to resolve 'ii'")
	}
	if a, ok := actReg.Resolve("ii"); ok && a.RomanNumeral != "II" {
		t.Errorf("expected Act II, got %s", a.RomanNumeral)
	}
	if _, ok := actReg.Resolve("iii"); ok {
		t.Error("expected actReg not to resolve 'iii' (doesn't exist)")
	}

	// Per-act scene scoping: Act I has Scene II, Act II does NOT have Scene II
	sceneReg1 := buildSceneRegistry(&acts[0])
	if _, ok := sceneReg1.Resolve("ii"); !ok {
		t.Error("expected Act I scene reg to resolve 'ii'")
	}
	sceneReg2 := buildSceneRegistry(&acts[1])
	if _, ok := sceneReg2.Resolve("ii"); ok {
		t.Error("expected Act II scene reg NOT to resolve 'ii' (per-act scoping)")
	}
	if _, ok := sceneReg2.Resolve("iii"); !ok {
		t.Error("expected Act II scene reg to resolve 'iii'")
	}
}
