package lexer

import "testing"

func TestTokenString(t *testing.T) {
	tests := []struct {
		name string
		tok  Token
		want string
	}{
		{"word", Token{TokenWord, "Romeo", 3, 5}, "{WORD Romeo 3:5}"},
		{"eof", Token{TokenEOF, "", 1, 1}, "{EOF  1:1}"},
		{"period", Token{TokenPeriod, ".", 2, 10}, "{PERIOD . 2:10}"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.tok.String()
			if got != tt.want {
				t.Errorf("String() = %q, want %q", got, tt.want)
			}
		})
	}
}
