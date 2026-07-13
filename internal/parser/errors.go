package parser

import "fmt"

// ParseError represents a syntax error with a taxonomy error code.
type ParseError struct {
	Code string
	Line int
	Col  int
	Msg  string
}

// Error formats per the SPL error taxonomy:
//
//	error[S001]: <message>
//	  --> input:1:1
func (e ParseError) Error() string {
	return fmt.Sprintf("error[%s]: %s\n  --> input:%d:%d", e.Code, e.Msg, e.Line, e.Col)
}

func errMissingTitle(line, col int) ParseError {
	return ParseError{Code: "S001", Line: line, Col: col, Msg: "expected program title ending with '.'"}
}

func errMissingCharacterDecl(line, col int) ParseError {
	return ParseError{Code: "S002", Line: line, Col: col, Msg: "expected at least one character declaration before first act"}
}

func errMissingAct(line, col int) ParseError {
	return ParseError{Code: "S004", Line: line, Col: col, Msg: "expected at least one act"}
}

func errInvalidActNumber(got string, line, col int) ParseError {
	return ParseError{Code: "S005", Line: line, Col: col, Msg: fmt.Sprintf("expected Roman numeral after 'Act', got '%s'", got)}
}

func errActOrder(expected string, line, col int) ParseError {
	return ParseError{Code: "S006", Line: line, Col: col, Msg: fmt.Sprintf("act numbers must be sequential, expected %s", expected)}
}

func errMissingScene(actRoman string, line, col int) ParseError {
	return ParseError{Code: "S007", Line: line, Col: col, Msg: fmt.Sprintf("expected at least one scene in Act %s", actRoman)}
}

func errInvalidSceneNumber(got string, line, col int) ParseError {
	return ParseError{Code: "S008", Line: line, Col: col, Msg: fmt.Sprintf("expected Roman numeral after 'Scene', got '%s'", got)}
}

func errSceneOrder(expected string, line, col int) ParseError {
	return ParseError{Code: "S009", Line: line, Col: col, Msg: fmt.Sprintf("scene numbers must be sequential, expected %s", expected)}
}

func errInvalidEnter(line, col int) ParseError {
	return ParseError{Code: "S011", Line: line, Col: col, Msg: "expected character name after stage direction"}
}

func errInvalidExeunt(line, col int) ParseError {
	return ParseError{Code: "S012", Line: line, Col: col, Msg: "expected character name or ']' after 'Exeunt'"}
}

func errMissingStage(line, col int) ParseError {
	return ParseError{Code: "S013", Line: line, Col: col, Msg: "expected [Enter ...] before dialogue"}
}

func errMissingSpeaker(line, col int) ParseError {
	return ParseError{Code: "S014", Line: line, Col: col, Msg: "expected character name followed by ':'"}
}

func errInvalidExpression(line, col int) ParseError {
	return ParseError{Code: "S015", Line: line, Col: col, Msg: "expected expression"}
}

func errInvalidIf(line, col int) ParseError {
	return ParseError{Code: "S016", Line: line, Col: col, Msg: "expected 'If so' or 'If not' followed by 'let us proceed/return to scene/act N'"}
}

func errInvalidComparative(got string, line, col int) ParseError {
	return ParseError{Code: "S017", Line: line, Col: col, Msg: fmt.Sprintf("expected comparative phrase (e.g., 'as good as', 'better than'), got '%s'", got)}
}

func errInvalidStackOp(msg string, line, col int) ParseError {
	return ParseError{Code: "S018", Line: line, Col: col, Msg: msg}
}

// Warning is a non-fatal advisory (S003 only). Collected alongside the AST.
type Warning struct {
	Code string
	Line int
	Col  int
	Msg  string
}

func (w Warning) String() string {
	return fmt.Sprintf("warning[%s]: %s at line %d, col %d", w.Code, w.Msg, w.Line, w.Col)
}

func warnInvalidCharacterName(name string, line, col int) Warning {
	return Warning{Code: "S003", Line: line, Col: col, Msg: fmt.Sprintf("character name '%s' is not a known Shakespeare character", name)}
}
