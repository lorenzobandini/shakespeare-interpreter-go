package semantic

import (
	"log/slog"
	"strings"

	"github.com/lorenzobandini/shakespeare-interpreter-go/internal/parser"
)

// Analyzer performs semantic analysis on a parsed SPL AST.
type Analyzer struct {
	filename   string
	symbols    SymbolTable
	acts       ActRegistry
	currentAct *parser.Act
	sceneReg   SceneRegistry
	stage      Stage
	errs       []SemanticError
}

// Result holds the output of semantic analysis.
type Result struct {
	Symbols SymbolTable
	Acts    ActRegistry
	Errors  []SemanticError
}

func (r Result) OK() bool {
	return len(r.Errors) == 0
}

// New creates an Analyzer for the given program.
func New(filename string, prog *parser.Program) *Analyzer {
	symbols, symErrs := newSymbolTable(prog.Characters)
	acts := buildActRegistry(prog.Acts)
	return &Analyzer{
		filename: filename,
		symbols:  symbols,
		acts:     acts,
		errs:     symErrs,
	}
}

func (a *Analyzer) checkDeclared(name string, line, col int) *SemanticError {
	key := strings.ToLower(name)
	if !a.symbols.Has(key) {
		err := errUndefinedCharacter(name, line, col)
		return &err
	}
	slog.Debug("symbol lookup", "name", name, "found", true)
	return nil
}

// Analyze runs the full semantic analysis pass on the program.
func (a *Analyzer) Analyze(prog *parser.Program) Result {
	for i := range prog.Acts {
		act := &prog.Acts[i]
		a.currentAct = act
		slog.Debug("act begin", "roman", act.RomanNumeral)

		// D1: stage persists across acts — do NOT clear.

		if len(act.Scenes) == 0 {
			a.errs = append(a.errs, errNoSceneInAct(act.RomanNumeral, act.Line, act.Col))
		}

		a.sceneReg = buildSceneRegistry(act)
		for si := range act.Scenes {
			a.analyzeSceneStatements(act.Scenes[si].Statements)
		}
	}
	return Result{Symbols: a.symbols, Acts: a.acts, Errors: a.errs}
}

// analyzeSceneStatements is the level-1 dispatch: stage directions and dialogues.
func (a *Analyzer) analyzeSceneStatements(stmts []parser.Statement) {
	for _, stmt := range stmts {
		switch s := stmt.(type) {
		case parser.EnterStmt:
			a.errs = append(a.errs, a.stage.Enter(s.Characters, a.symbols, s.Line, s.Col)...)
		case parser.ExitStmt:
			a.errs = append(a.errs, a.stage.Exit(s.Character, a.symbols, s.Line, s.Col)...)
		case parser.ExeuntStmt:
			a.errs = append(a.errs, a.stage.Exeunt(s.Characters, a.symbols, s.Line, s.Col)...)
		case parser.Dialogue:
			a.analyzeDialogue(s)
		}
	}
}

// dialogCtx carries the speaker and resolved listener for a dialogue turn.
type dialogCtx struct {
	speaker, listener string
}

func (a *Analyzer) analyzeDialogue(d parser.Dialogue) {
	if err := a.checkDeclared(d.Speaker, d.Line, d.Col); err != nil {
		a.errs = append(a.errs, *err)
	}
	if !a.stage.Has(d.Speaker) {
		a.errs = append(a.errs, errCharacterNotOnStage(d.Speaker, "speaker", d.Line, d.Col))
	}

	listener, ok := a.stage.Listener(d.Speaker)
	if !ok {
		listener = d.Speaker
	}
	slog.Debug("dialogue", "speaker", d.Speaker, "listener", listener)
	ctx := dialogCtx{speaker: d.Speaker, listener: listener}

	for _, s := range d.Statements {
		a.analyzeStatement(s, ctx)
	}
}

func (a *Analyzer) analyzeStatement(stmt parser.Statement, ctx dialogCtx) {
	switch s := stmt.(type) {
	case parser.AssignStmt:
		a.analyzeExpr(s.Expr, ctx)
	case parser.SpeakStmt:
	case parser.OpenHeartStmt:
	case parser.OpenMindStmt:
	case parser.ListenStmt:
	case parser.RememberStmt:
		a.analyzeExpr(s.Expr, ctx)
	case parser.RecallStmt:
	case parser.QuestionStmt:
		a.analyzeExpr(s.Left, ctx)
		a.analyzeExpr(s.Right, ctx)
	case parser.IfStmt:
		a.resolveGoto(s.Target, s.TargetKind, s.Line, s.Col)
	case parser.GotoStmt:
		a.resolveGoto(s.Target, s.TargetKind, s.Line, s.Col)
	}
}

func (a *Analyzer) analyzeExpr(e parser.Expr, ctx dialogCtx) {
	switch ex := e.(type) {
	case parser.ConstExpr:
	case parser.CharRefExpr:
		if err := a.checkDeclared(ex.Name, ex.Line, ex.Col); err != nil {
			a.errs = append(a.errs, *err)
		}
		slog.Debug("charref", "name", ex.Name, "on_stage", a.stage.Has(ex.Name))
	case parser.PronounExpr:
	case parser.BinaryOpExpr:
		a.analyzeExpr(ex.Left, ctx)
		a.analyzeExpr(ex.Right, ctx)
	case parser.UnaryOpExpr:
		a.analyzeExpr(ex.Operand, ctx)
	}
}

func (a *Analyzer) resolveGoto(target, kind string, line, col int) {
	switch kind {
	case "scene":
		if _, ok := a.sceneReg.Resolve(strings.ToLower(target)); !ok {
			a.errs = append(a.errs, errUndefinedScene(target, "scene", a.currentAct.RomanNumeral, line, col))
		}
	case "act":
		if _, ok := a.acts.Resolve(strings.ToLower(target)); !ok {
			a.errs = append(a.errs, errUndefinedScene(target, "act", "", line, col))
		}
	}
}
