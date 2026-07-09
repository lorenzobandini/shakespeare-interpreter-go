package lexer

import "fmt"

// TokenType identifies the kind of lexical token emitted by the lexer.
//
// The lexer is deliberately dumb: it recognizes only structural punctuation
// (. , : ! ? [ ]), newlines, and EOF. Everything else folds into TokenWord.
type TokenType string

const (
	TokenEOF      TokenType = "EOF"
	TokenNewline  TokenType = "NEWLINE"
	TokenWord     TokenType = "WORD"
	TokenPeriod   TokenType = "PERIOD"
	TokenComma    TokenType = "COMMA"
	TokenColon    TokenType = "COLON"
	TokenBang     TokenType = "BANG"
	TokenQuestion TokenType = "QUESTION"
	TokenLBracket TokenType = "LBRACKET"
	TokenRBracket TokenType = "RBRACKET"
)

// Token is a single lexical token with its source position.
//
// Line and Col are 1-indexed. Col resets to 1 at each newline.
// Lexeme preserves the original source text (case and all) so the parser
// can normalize for classification while debug output stays faithful.
type Token struct {
	Type   TokenType
	Lexeme string
	Line   int
	Col    int
}

// String returns a compact debug representation: "{TYPE Lexeme line:col}".
// EOF renders as "{EOF  line:col}" (empty lexeme).
func (t Token) String() string {
	return fmt.Sprintf("{%s %s %d:%d}", t.Type, t.Lexeme, t.Line, t.Col)
}
