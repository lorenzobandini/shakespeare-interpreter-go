package runtime

import (
	"bytes"
	"errors"
	"flag"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/lorenzobandini/shakespeare-interpreter-go/internal/lexer"
	"github.com/lorenzobandini/shakespeare-interpreter-go/internal/parser"
	"github.com/lorenzobandini/shakespeare-interpreter-go/internal/semantic"
)

var update = flag.Bool("update", false, "update golden files")

func findSource(name string) string {
	base := "../../testdata/runtime"
	candidates := []string{name + ".shpl"}
	for i := len(name) - 1; i >= 0; i-- {
		if name[i] == '-' {
			candidates = append(candidates, name[:i]+".shpl")
		}
	}
	for _, c := range candidates {
		p := filepath.Join(base, c)
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return filepath.Join(base, name+".shpl")
}

func runFixture(t *testing.T, name, stdin string) (string, error) {
	t.Helper()
	srcPath := findSource(name)
	srcBytes, err := os.ReadFile(srcPath)
	if err != nil {
		t.Fatalf("read %s.shpl: %v", name, err)
	}
	tokens, err := lexer.New(string(srcBytes)).ScanTokens()
	if err != nil {
		t.Fatalf("lex %s: %v", name, err)
	}
	prog, err := parser.New(tokens).Parse()
	if err != nil {
		t.Fatalf("parse %s: %v", name, err)
	}
	res := semantic.New(name+".shpl", prog).Analyze(prog)
	if !res.OK() {
		t.Fatalf("semantic %s: %v", name, res.Errors)
	}
	var in io.Reader = &bytes.Buffer{}
	if stdin != "" {
		in = strings.NewReader(stdin)
	}
	out := &bytes.Buffer{}
	err = Execute(prog, res, in, out, name+".shpl")
	return out.String(), err
}

func TestGoldens(t *testing.T) {
	tests := []struct {
		name    string
		stdin   string
		wantErr string
	}{
		{name: "hello"},
		{name: "branch"},
		{name: "stack"},
		{name: "io-ascii", stdin: "X"},
		{name: "truth-machine", stdin: "0\n"},
		{name: "truth-machine-1", stdin: "1\n", wantErr: "R003"},
		{name: "divzero", wantErr: "R001"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := runFixture(t, tt.name, tt.stdin)

			goldenPath := filepath.Join("../../testdata/runtime", tt.name+".golden")

			if *update {
				if err := os.WriteFile(goldenPath, []byte(got), 0644); err != nil {
					t.Fatalf("write golden %s: %v", goldenPath, err)
				}
				return
			}

			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf("expected error %s, got nil", tt.wantErr)
				}
				var re RuntimeError
				if errors.As(err, &re) {
					if re.Code != tt.wantErr {
						t.Fatalf("expected error code %s, got %s: %v", tt.wantErr, re.Code, err)
					}
				}
				return
			}

			if err != nil {
				t.Fatalf("Execute error: %v", err)
			}

			want, err := os.ReadFile(goldenPath)
			if err != nil {
				t.Fatalf("read golden %s: %v (got output %q)", goldenPath, err, got)
			}
			if got != string(want) {
				t.Errorf("output mismatch for %s:\ngot:  %q\nwant: %q", tt.name, got, string(want))
			}
		})
	}
}
