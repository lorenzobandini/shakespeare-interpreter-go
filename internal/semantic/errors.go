package semantic

import (
	"fmt"
	"strings"
)

// SemanticError represents a semantic analysis error with an M-code.
type SemanticError struct {
	Code, Msg string
	Line, Col int
	Filename  string
}

func (e SemanticError) Error() string {
	f := e.Filename
	if f == "" {
		f = "input"
	}
	return fmt.Sprintf("error[%s]: %s\n  --> %s:%d:%d", e.Code, e.Msg, f, e.Line, e.Col)
}

func errUndefinedCharacter(name string, line, col int) SemanticError {
	return SemanticError{
		Code: "M001", Line: line, Col: col,
		Msg: fmt.Sprintf("character '%s' is not declared", name),
	}
}

func errTooManyOnStage(got int, line, col int) SemanticError {
	return SemanticError{
		Code: "M002", Line: line, Col: col,
		Msg: fmt.Sprintf("too many characters on stage (max 2), got %d", got),
	}
}

func errStageOverflow(onStage []string, line, col int) SemanticError {
	return SemanticError{
		Code: "M003", Line: line, Col: col,
		Msg: fmt.Sprintf("cannot enter: stage is full (%s already on stage)", strings.Join(onStage, ", ")),
	}
}

func errCharacterNotOnStage(name string, line, col int) SemanticError {
	return SemanticError{
		Code: "M004", Line: line, Col: col,
		Msg: fmt.Sprintf("character '%s' is not on stage", name),
	}
}

func errExitNotOnStage(name string, line, col int) SemanticError {
	return SemanticError{
		Code: "M005", Line: line, Col: col,
		Msg: fmt.Sprintf("character '%s' is not on stage", name),
	}
}

func errUndefinedScene(target, kind, ctx string, line, col int) SemanticError {
	noun := kind
	if noun == "" {
		noun = "target"
	}
	if ctx != "" {
		return SemanticError{
			Code: "M006", Line: line, Col: col,
			Msg: fmt.Sprintf("%s %s is not defined in Act %s", noun, target, ctx),
		}
	}
	return SemanticError{
		Code: "M006", Line: line, Col: col,
		Msg: fmt.Sprintf("%s %s is not defined", noun, target),
	}
}

func errSelfReferenceEnter(line, col int) SemanticError {
	return SemanticError{
		Code: "M007", Line: line, Col: col,
		Msg: "cannot enter the same character twice",
	}
}

func errAlreadyOnStage(name string, line, col int) SemanticError {
	return SemanticError{
		Code: "M007", Line: line, Col: col,
		Msg: fmt.Sprintf("cannot enter '%s': already on stage", name),
	}
}

func errNoSceneInAct(actRoman string, line, col int) SemanticError {
	return SemanticError{
		Code: "M008", Line: line, Col: col,
		Msg: fmt.Sprintf("Act %s has no scenes", actRoman),
	}
}
