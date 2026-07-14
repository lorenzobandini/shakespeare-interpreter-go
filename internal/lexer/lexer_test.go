package lexer

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/lorenzobandini/shakespeare-interpreter-go/internal/logger"
)

func TestMain(m *testing.M) {
	logger.Init(logger.LevelDebug)
	m.Run()
}

func TestScanTokensWhitespace(t *testing.T) {
	tokens, err := New("  \t\n  ").ScanTokens()
	if err != nil {
		t.Fatal(err)
	}
	if len(tokens) != 2 {
		t.Fatalf("expected 2 tokens, got %d: %v", len(tokens), tokens)
	}
	if tokens[0].Type != TokenNewline {
		t.Errorf("token 0: got %v, want NEWLINE", tokens[0])
	}
	if tokens[1].Type != TokenEOF {
		t.Errorf("token 1: got %v, want EOF", tokens[1])
	}
}

func TestScanTokensControlCharL001(t *testing.T) {
	_, err := New("\x07").ScanTokens()
	if err == nil {
		t.Fatal("expected L001 error")
	}
	lex, ok := err.(LexError)
	if !ok {
		t.Fatalf("expected LexError, got %T", err)
	}
	if lex.Code != "L001" {
		t.Errorf("code: got %q, want L001", lex.Code)
	}
}

func TestScanTokensPunctuation(t *testing.T) {
	tests := []struct {
		input string
		want  []TokenType
	}{
		{".", []TokenType{TokenPeriod}},
		{",", []TokenType{TokenComma}},
		{":", []TokenType{TokenColon}},
		{"!", []TokenType{TokenBang}},
		{"?", []TokenType{TokenQuestion}},
		{"[]", []TokenType{TokenLBracket, TokenRBracket}},
		{".:,!?", []TokenType{TokenPeriod, TokenColon, TokenComma, TokenBang, TokenQuestion}},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			tokens, err := New(tt.input).ScanTokens()
			if err != nil {
				t.Fatal(err)
			}
			if len(tokens) != len(tt.want)+1 {
				t.Fatalf("expected %d tokens + EOF, got %d: %v", len(tt.want), len(tokens), tokens)
			}
			for i, typ := range tt.want {
				if tokens[i].Type != typ {
					t.Errorf("token %d: got %v, want %v", i, tokens[i].Type, typ)
				}
			}
			if tokens[len(tokens)-1].Type != TokenEOF {
				t.Errorf("last token: got %v, want EOF", tokens[len(tokens)-1])
			}
		})
	}
}

func TestScanTokensWordSimple(t *testing.T) {
	tokens, err := New("Romeo").ScanTokens()
	if err != nil {
		t.Fatal(err)
	}
	if len(tokens) != 2 {
		t.Fatalf("expected 2 tokens, got %d", len(tokens))
	}
	if tokens[0].Type != TokenWord || tokens[0].Lexeme != "Romeo" {
		t.Errorf("token 0: got %v, want WORD Romeo", tokens[0])
	}
}

func TestScanTokensWordApostropheHyphen(t *testing.T) {
	tokens, err := New("summer's half-witted").ScanTokens()
	if err != nil {
		t.Fatal(err)
	}
	if len(tokens) != 3 {
		t.Fatalf("expected 3 tokens, got %d", len(tokens))
	}
	if tokens[0].Lexeme != "summer's" {
		t.Errorf("token 0: got %q, want summer's", tokens[0].Lexeme)
	}
	if tokens[1].Lexeme != "half-witted" {
		t.Errorf("token 1: got %q, want half-witted", tokens[1].Lexeme)
	}
}

func TestScanTokensWordFreeText(t *testing.T) {
	tokens, err := New("A/S").ScanTokens()
	if err != nil {
		t.Fatal(err)
	}
	if len(tokens) != 2 {
		t.Fatalf("expected 2 tokens, got %d", len(tokens))
	}
	if tokens[0].Lexeme != "A/S" {
		t.Errorf("token 0: got %q, want A/S", tokens[0].Lexeme)
	}
}

func TestScanTokensAtSignFoldsToWord(t *testing.T) {
	tokens, err := New("a @ b").ScanTokens()
	if err != nil {
		t.Fatal(err)
	}
	// Expected: WORD a, WORD @, WORD b, EOF
	if len(tokens) != 4 {
		t.Fatalf("expected 4 tokens, got %d: %v", len(tokens), tokens)
	}
	if tokens[0].Lexeme != "a" || tokens[1].Lexeme != "@" || tokens[2].Lexeme != "b" {
		t.Errorf("unexpected tokens: %v", tokens)
	}
}

func TestScanTokensWordBeforePunctuation(t *testing.T) {
	tokens, err := New("Romeo,").ScanTokens()
	if err != nil {
		t.Fatal(err)
	}
	if len(tokens) != 3 {
		t.Fatalf("expected 3 tokens, got %d", len(tokens))
	}
	if tokens[0].Type != TokenWord || tokens[0].Lexeme != "Romeo" {
		t.Errorf("token 0: got %v, want WORD Romeo", tokens[0])
	}
	if tokens[1].Type != TokenComma {
		t.Errorf("token 1: got %v, want COMMA", tokens[1])
	}
}

func TestScanTokensMixedCasePreserved(t *testing.T) {
	tokens, err := New("Speak YOUR mind").ScanTokens()
	if err != nil {
		t.Fatal(err)
	}
	if len(tokens) != 4 {
		t.Fatalf("expected 4 tokens, got %d", len(tokens))
	}
	if tokens[0].Lexeme != "Speak" || tokens[1].Lexeme != "YOUR" || tokens[2].Lexeme != "mind" {
		t.Errorf("case not preserved: %v", tokens)
	}
}

func TestScanTokensEOF(t *testing.T) {
	tokens, err := New("Romeo").ScanTokens()
	if err != nil {
		t.Fatal(err)
	}
	if len(tokens) != 2 {
		t.Fatalf("expected 2 tokens, got %d", len(tokens))
	}
	if tokens[1].Type != TokenEOF {
		t.Errorf("token 1: got %v, want EOF", tokens[1])
	}
}

func TestScanTokensEOFEmpty(t *testing.T) {
	tokens, err := New("").ScanTokens()
	if err != nil {
		t.Fatal(err)
	}
	if len(tokens) != 1 {
		t.Fatalf("expected 1 token, got %d", len(tokens))
	}
	if tokens[0].Type != TokenEOF {
		t.Errorf("token 0: got %v, want EOF", tokens[0])
	}
}

func TestScanTokensLineColTracking(t *testing.T) {
	tokens, err := New("a\nb").ScanTokens()
	if err != nil {
		t.Fatal(err)
	}
	if len(tokens) != 4 {
		t.Fatalf("expected 4 tokens, got %d", len(tokens))
	}
	if tokens[0].Line != 1 || tokens[0].Col != 1 {
		t.Errorf("token 0: got %d:%d, want 1:1", tokens[0].Line, tokens[0].Col)
	}
	if tokens[1].Type != TokenNewline || tokens[1].Line != 1 || tokens[1].Col != 2 {
		t.Errorf("newline: got %v, want NEWLINE 1:2", tokens[1])
	}
	if tokens[2].Type != TokenWord || tokens[2].Line != 2 || tokens[2].Col != 1 {
		t.Errorf("token 2: got %v, want WORD 2:1", tokens[2])
	}
	if tokens[3].Type != TokenEOF || tokens[3].Line != 2 || tokens[3].Col != 2 {
		t.Errorf("token 3: got %v, want EOF 2:2", tokens[3])
	}
}

func TestScanTokensL002UnterminatedBracket(t *testing.T) {
	_, err := New("[Enter Romeo").ScanTokens()
	if err == nil {
		t.Fatal("expected L002 error")
	}
	lex, ok := err.(LexError)
	if !ok {
		t.Fatalf("expected LexError, got %T", err)
	}
	if lex.Code != "L002" {
		t.Errorf("code: got %q, want L002", lex.Code)
	}
}

func TestScanTokensL002NotTriggered(t *testing.T) {
	tokens, err := New("[Enter Romeo]").ScanTokens()
	if err != nil {
		t.Fatal(err)
	}
	if len(tokens) != 5 {
		t.Fatalf("expected 5 tokens, got %d", len(tokens))
	}
	if tokens[0].Type != TokenLBracket {
		t.Errorf("token 0: got %v, want LBRACKET", tokens[0])
	}
	if tokens[4].Type != TokenEOF {
		t.Errorf("token 4: got %v, want EOF", tokens[4])
	}
}

func TestScanFixtureHello(t *testing.T) {
	tokens, err := readFixture("../../testdata/lexer/hello.spl")
	if err != nil {
		t.Fatal(err)
	}
	if len(tokens) == 0 {
		t.Fatal("no tokens")
	}
	if tokens[len(tokens)-1].Type != TokenEOF {
		t.Fatal("missing EOF")
	}
}

func TestScanFixtureTruthMachine(t *testing.T) {
	tokens, err := readFixture("../../testdata/lexer/truth-machine.spl")
	if err != nil {
		t.Fatal(err)
	}
	if len(tokens) == 0 {
		t.Fatal("no tokens")
	}
}

func TestScanFixtureMinimal(t *testing.T) {
	tokens, err := readFixture("../../testdata/lexer/minimal.spl")
	if err != nil {
		t.Fatal(err)
	}
	if len(tokens) == 0 {
		t.Fatal("no tokens")
	}
}

func TestScanFixtureBadCharL001(t *testing.T) {
	_, err := readFixture("../../testdata/lexer/bad-char.spl")
	if err == nil {
		t.Fatal("expected L001 error")
	}
	lex, ok := err.(LexError)
	if !ok {
		t.Fatalf("expected LexError, got %T", err)
	}
	if lex.Code != "L001" {
		t.Errorf("code: got %q, want L001", lex.Code)
	}
}

func TestGoldenSnapshots(t *testing.T) {
	fixtures := []string{"hello", "truth-machine", "minimal"}
	for _, name := range fixtures {
		t.Run(name, func(t *testing.T) {
			srcPath := filepath.Join("../../testdata/lexer", name+".spl")
			goldenPath := filepath.Join("../../testdata/lexer", name+".golden.txt")
			src, err := os.ReadFile(srcPath)
			if err != nil {
				t.Fatal(err)
			}
			tokens, err := New(string(src)).ScanTokens()
			if err != nil {
				t.Fatal(err)
			}
			var sb strings.Builder
			for _, tok := range tokens {
				sb.WriteString(tok.String())
				sb.WriteByte('\n')
			}
			got := sb.String()
			existing, err := os.ReadFile(goldenPath)
			if err != nil {
				if os.IsNotExist(err) {
					if writeErr := os.WriteFile(goldenPath, []byte(got), 0o644); writeErr != nil {
						t.Fatal(writeErr)
					}
					t.Fatalf("golden file %s created; inspect and re-run", goldenPath)
				}
				t.Fatal(err)
			}
			if got != string(existing) {
				t.Errorf("golden mismatch for %s\n--- got ---\n%s\n--- want ---\n%s", name, got, string(existing))
			}
		})
	}
}

func readFixture(path string) ([]Token, error) {
	src, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return New(string(src)).ScanTokens()
}
