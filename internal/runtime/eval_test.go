package runtime

import (
	"testing"

	"github.com/lorenzobandini/shakespeare-interpreter-go/internal/parser"
)

func TestEvalConst(t *testing.T) {
	tests := []struct {
		name string
		adj  int
		pol  int
		want int
	}{
		{"flower (0, +1)", 0, 1, 1},
		{"red flower (1, +1)", 1, 1, 2},
		{"red hot flower (2, +1)", 2, 1, 4},
		{"coward (0, -1)", 0, -1, -1},
		{"big coward (1, -1)", 1, -1, -2},
		{"a vile coward (1, -1)", 1, -1, -2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &env{values: map[string]int{}, stacks: map[string][]int{}}
			expr := parser.ConstExpr{AdjectiveCount: tt.adj, Polarity: tt.pol, Noun: "test"}
			got, err := e.eval(expr, "", "")
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("got %d; want %d", got, tt.want)
			}
		})
	}
}

func TestEvalCharRef(t *testing.T) {
	e := &env{
		values: map[string]int{"romeo": 3, "juliet": 5},
		stacks: map[string][]int{},
	}
	tests := []struct {
		name string
		ref  string
		want int
	}{
		{"Romeo", "Romeo", 3},
		{"JULIET (case insensitive)", "JULIET", 5},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := e.eval(parser.CharRefExpr{Name: tt.ref}, "", "")
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("got %d; want %d", got, tt.want)
			}
		})
	}
}

func TestEvalCharRefDefensive(t *testing.T) {
	e := &env{values: map[string]int{}, stacks: map[string][]int{}}
	_, err := e.eval(parser.CharRefExpr{Name: "Ghost", Line: 2, Col: 3}, "", "")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	re, ok := err.(RuntimeError)
	if !ok || re.Code != "R000" {
		t.Fatalf("expected R000, got %v", err)
	}
}

func TestEvalPronoun(t *testing.T) {
	e := &env{
		values: map[string]int{"romeo": 3, "juliet": 5},
		stacks: map[string][]int{},
	}
	tests := []struct {
		name     string
		ref      string
		speaker  string
		listener string
		want     int
	}{
		{"speaker = Romeo", "speaker", "Romeo", "Juliet", 3},
		{"listener = Juliet", "listener", "Romeo", "Juliet", 5},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := e.eval(parser.PronounExpr{Ref: tt.ref}, tt.speaker, tt.listener)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("got %d; want %d", got, tt.want)
			}
		})
	}
}

func TestEvalPronounDefensive(t *testing.T) {
	e := &env{values: map[string]int{}, stacks: map[string][]int{}}
	_, err := e.eval(parser.PronounExpr{Ref: "other", Line: 1, Col: 1}, "", "")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	re, ok := err.(RuntimeError)
	if !ok || re.Code != "R000" {
		t.Fatalf("expected R000, got %v", err)
	}
}

func TestEvalBinary(t *testing.T) {
	e := &env{
		values:            map[string]int{},
		stacks:            map[string][]int{},
		currentActRoman:   "I",
		currentSceneRoman: "II",
	}
	tests := []struct {
		name    string
		op      string
		left    parser.Expr
		right   parser.Expr
		want    int
		wantErr string
	}{
		{"sum", "sum", parser.ConstExpr{AdjectiveCount: 0, Polarity: 3}, parser.ConstExpr{AdjectiveCount: 0, Polarity: 7}, 10, ""},
		{"difference", "difference", parser.ConstExpr{AdjectiveCount: 0, Polarity: 10}, parser.ConstExpr{AdjectiveCount: 0, Polarity: 3}, 7, ""},
		{"product", "product", parser.ConstExpr{AdjectiveCount: 0, Polarity: 4}, parser.ConstExpr{AdjectiveCount: 0, Polarity: 5}, 20, ""},
		{"quotient", "quotient", parser.ConstExpr{AdjectiveCount: 0, Polarity: 10}, parser.ConstExpr{AdjectiveCount: 0, Polarity: 3}, 3, ""},
		{"remainder", "remainder", parser.ConstExpr{AdjectiveCount: 0, Polarity: 10}, parser.ConstExpr{AdjectiveCount: 0, Polarity: 3}, 1, ""},
		{"div by zero", "quotient", parser.ConstExpr{AdjectiveCount: 0, Polarity: 5}, parser.ConstExpr{AdjectiveCount: 0, Polarity: 0}, 0, "R001"},
		{"mod by zero", "remainder", parser.ConstExpr{AdjectiveCount: 0, Polarity: 5}, parser.ConstExpr{AdjectiveCount: 0, Polarity: 0}, 0, "R001"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := e.eval(parser.BinaryOpExpr{Op: tt.op, Left: tt.left, Right: tt.right}, "", "")
			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf("expected error %s, got nil", tt.wantErr)
				}
				re, ok := err.(RuntimeError)
				if !ok || re.Code != tt.wantErr {
					t.Fatalf("expected %s, got %v", tt.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("got %d; want %d", got, tt.want)
			}
		})
	}
}

func TestEvalUnary(t *testing.T) {
	e := &env{values: map[string]int{}, stacks: map[string][]int{}}
	tests := []struct {
		name    string
		op      string
		operand parser.Expr
		want    int
		wantErr string
	}{
		{"square", "square", parser.ConstExpr{AdjectiveCount: 0, Polarity: 5}, 25, ""},
		{"cube", "cube", parser.ConstExpr{AdjectiveCount: 0, Polarity: 3}, 27, ""},
		{"square_root", "square_root", parser.ConstExpr{AdjectiveCount: 0, Polarity: 16}, 4, ""},
		{"factorial 5", "factorial", parser.ConstExpr{AdjectiveCount: 0, Polarity: 5}, 120, ""},
		{"twice", "twice", parser.ConstExpr{AdjectiveCount: 0, Polarity: 7}, 14, ""},
		{"factorial 21 overflows", "factorial", parser.ConstExpr{AdjectiveCount: 0, Polarity: 21}, 0, "R004"},
		{"sqrt negative", "square_root", parser.ConstExpr{AdjectiveCount: 0, Polarity: -1}, 0, "R000"},
		{"factorial negative", "factorial", parser.ConstExpr{AdjectiveCount: 0, Polarity: -5}, 0, "R000"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := e.eval(parser.UnaryOpExpr{Op: tt.op, Operand: tt.operand}, "", "")
			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf("expected error %s, got nil", tt.wantErr)
				}
				re, ok := err.(RuntimeError)
				if !ok || re.Code != tt.wantErr {
					t.Fatalf("expected %s, got %v", tt.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("got %d; want %d", got, tt.want)
			}
		})
	}
}

func TestEvalNesting(t *testing.T) {
	e := &env{
		values: map[string]int{"romeo": 3, "juliet": 5},
		stacks: map[string][]int{},
	}
	// sum of Romeo and difference between Juliet and flower → 3 + (5 - 1) = 7
	expr := parser.BinaryOpExpr{
		Op:   "sum",
		Left: parser.CharRefExpr{Name: "Romeo"},
		Right: parser.BinaryOpExpr{
			Op:    "difference",
			Left:  parser.CharRefExpr{Name: "Juliet"},
			Right: parser.ConstExpr{AdjectiveCount: 0, Polarity: 1},
		},
	}
	got, err := e.eval(expr, "", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != 7 {
		t.Errorf("got %d; want 7", got)
	}
}

func TestEvalDefensive(t *testing.T) {
	e := &env{values: map[string]int{}, stacks: map[string][]int{}}
	_, err := e.eval(parser.CharRefExpr{Name: "Ghost", Line: 1, Col: 1}, "", "")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	re, ok := err.(RuntimeError)
	if !ok || re.Code != "R000" {
		t.Fatalf("expected R000, got %v", err)
	}
}
