package lexer

import (
	"fmt"
	"log/slog"
	"strings"
)

// Lexer scans SPL source bytes into tokens.
type Lexer struct {
	source      []byte
	pos         int
	line        int
	col         int
	inBracket   bool
	bracketLine int
	bracketCol  int
}

// New creates a Lexer for the given source string.
func New(src string) *Lexer {
	if len(src) >= 3 && src[:3] == "\xEF\xBB\xBF" {
		src = src[3:]
	}
	return &Lexer{
		source: []byte(src),
		line:   1,
		col:    1,
	}
}

// ScanTokens scans the entire source, returning tokens (including trailing EOF).
// Returns an error on the first lexical error.
func (l *Lexer) ScanTokens() ([]Token, error) {
	var tokens []Token
	for {
		tok, err := l.ScanToken()
		if err != nil {
			return nil, err
		}
		if tok.Type == "" {
			continue
		}
		tokens = append(tokens, tok)
		if tok.Type == TokenEOF {
			return tokens, nil
		}
	}
}

// ScanToken scans and returns a single token. An empty Type signals
// "skip this" (whitespace) — the caller should not append it.
func (l *Lexer) ScanToken() (Token, error) {
	if l.inBracket && l.pos >= len(l.source) {
		return Token{}, LexError{
			Code: "L002",
			Line: l.bracketLine,
			Col:  l.bracketCol,
			Msg:  fmt.Sprintf("unterminated stage direction starting at line %d", l.bracketLine),
		}
	}
	if l.pos >= len(l.source) {
		tok := Token{Type: TokenEOF, Line: l.line, Col: l.col}
		slog.Debug("emitted EOF", "line", tok.Line, "col", tok.Col)
		return tok, nil
	}
	c := l.source[l.pos]
	switch {
	case c == ' ' || c == '\t' || c == '\r':
		l.advance()
		return Token{Type: ""}, nil
	case c == '\n':
		tok := Token{Type: TokenNewline, Lexeme: "\n", Line: l.line, Col: l.col}
		l.line++
		l.advance()
		l.col = 1
		slog.Debug("emitted NEWLINE", "line", tok.Line, "col", tok.Col)
		return tok, nil
	case isControlChar(c):
		return Token{}, LexError{
			Code: "L001",
			Line: l.line,
			Col:  l.col,
			Msg:  fmt.Sprintf("unexpected control character 0x%02X at line %d, col %d", c, l.line, l.col),
		}
	case c == '.':
		return l.single(TokenPeriod, "."), nil
	case c == ',':
		return l.single(TokenComma, ","), nil
	case c == ':':
		return l.single(TokenColon, ":"), nil
	case c == '!':
		return l.single(TokenBang, "!"), nil
	case c == '?':
		return l.single(TokenQuestion, "?"), nil
	case c == '[':
		l.inBracket = true
		l.bracketLine = l.line
		l.bracketCol = l.col
		return l.single(TokenLBracket, "["), nil
	case c == ']':
		l.inBracket = false
		return l.single(TokenRBracket, "]"), nil
	default:
		return l.scanWord()
	}
}

func (l *Lexer) single(tt TokenType, lex string) Token {
	tok := Token{Type: tt, Lexeme: lex, Line: l.line, Col: l.col}
	l.advance()
	slog.Debug("emitted token", "type", tok.Type, "lexeme", tok.Lexeme, "line", tok.Line, "col", tok.Col)
	return tok
}

func (l *Lexer) scanWord() (Token, error) {
	startLine, startCol := l.line, l.col
	var sb strings.Builder
	for l.pos < len(l.source) {
		c := l.source[l.pos]
		if isWordChar(c) {
			sb.WriteByte(c)
			l.advance()
		} else {
			break
		}
	}
	lex := sb.String()
	tok := Token{Type: TokenWord, Lexeme: lex, Line: startLine, Col: startCol}
	slog.Debug("emitted WORD", "lexeme", lex, "line", startLine, "col", startCol)
	return tok, nil
}

func (l *Lexer) advance() byte {
	c := l.source[l.pos]
	l.pos++
	l.col++
	return c
}

func isWordChar(c byte) bool {
	if c <= 0x20 || c == 0x7F {
		return false
	}
	switch c {
	case '.', ',', ':', '!', '?', '[', ']', '\n', '\r':
		return false
	}
	return true
}

func isControlChar(c byte) bool {
	if c == '\t' || c == '\n' || c == '\r' {
		return false
	}
	if c <= 0x08 {
		return true
	}
	if c == 0x0B || c == 0x0C {
		return true
	}
	if c >= 0x0E && c <= 0x1F {
		return true
	}
	if c == 0x7F {
		return true
	}
	return false
}
