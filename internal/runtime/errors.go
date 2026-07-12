package runtime

import "fmt"

type RuntimeError struct {
	Code, Msg string
	Line, Col int
	Filename  string
}

func (e RuntimeError) Error() string {
	f := e.Filename
	if f == "" {
		f = "input"
	}
	return fmt.Sprintf("error[%s]: %s\n  --> %s:%d:%d", e.Code, e.Msg, f, e.Line, e.Col)
}

func errDivisionByZero(actRoman, sceneRoman string, line, col int) RuntimeError {
	return RuntimeError{
		Code: "R001", Line: line, Col: col,
		Msg: fmt.Sprintf("division by zero at Act %s, Scene %s", actRoman, sceneRoman),
	}
}

func errInputNotANumber(got string, line, col int) RuntimeError {
	return RuntimeError{
		Code: "R002", Line: line, Col: col,
		Msg: fmt.Sprintf("expected a number, got '%s'", got),
	}
}

func errInputEOF(line, col int) RuntimeError {
	return RuntimeError{
		Code: "R003", Line: line, Col: col,
		Msg: "unexpected end of input",
	}
}

func errIntegerOverflow(op string, line, col int) RuntimeError {
	return RuntimeError{
		Code: "R004", Line: line, Col: col,
		Msg: fmt.Sprintf("integer overflow in '%s'", op),
	}
}
