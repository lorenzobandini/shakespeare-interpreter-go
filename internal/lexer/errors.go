package lexer

import "fmt"

// LexError represents a lexical error with a taxonomy error code, source
// position, and human-readable message. Implements the error interface.
type LexError struct {
	Code string
	Line int
	Col  int
	Msg  string
}

// Error formats the error per the SPL error taxonomy:
//
//	error[L001]: <message>
//	  --> input:3:1
func (e LexError) Error() string {
	return fmt.Sprintf("error[%s]: %s\n  --> input:%d:%d", e.Code, e.Msg, e.Line, e.Col)
}
