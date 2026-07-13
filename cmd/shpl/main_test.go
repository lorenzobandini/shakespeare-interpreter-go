package main

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/lorenzobandini/shakespeare-interpreter-go/internal/logger"
)

func TestMain(m *testing.M) {
	logger.Init(logger.LevelInfo)
	m.Run()
}

// resetGlobalFlags clears flag state that leaks between tests.
func resetGlobalFlags() {
	debugFlag = false
	traceFlag = false
}

// ---------------------------------------------------------------------------
// tokens subcommand
// ---------------------------------------------------------------------------

func TestTokensCommand(t *testing.T) {
	resetGlobalFlags()
	buf := &bytes.Buffer{}
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"tokens", "../../testdata/lexer/minimal.shpl"})
	defer rootCmd.SetArgs([]string{})
	defer rootCmd.SetOut(nil)

	if err := rootCmd.Execute(); err != nil {
		t.Fatal(err)
	}
	output := buf.String()
	if !strings.Contains(output, "{WORD") {
		t.Errorf("output missing WORD tokens:\n%s", output)
	}
	if !strings.Contains(output, "{EOF") {
		t.Errorf("output missing EOF token:\n%s", output)
	}
}

func TestTokensCommand_FileNotFound(t *testing.T) {
	resetGlobalFlags()
	buf := &bytes.Buffer{}
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"tokens", "nonexistent.shpl"})
	defer rootCmd.SetArgs([]string{})
	defer rootCmd.SetOut(nil)

	if err := rootCmd.Execute(); err == nil {
		t.Fatal("expected error for nonexistent file, got nil")
	}
}

func TestTokensCommand_LexerError(t *testing.T) {
	resetGlobalFlags()
	buf := &bytes.Buffer{}
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"tokens", "../../testdata/lexer/bad-char.shpl"})
	defer rootCmd.SetArgs([]string{})
	defer rootCmd.SetOut(nil)

	if err := rootCmd.Execute(); err == nil {
		t.Fatal("expected lexer error, got nil")
	}
}

// ---------------------------------------------------------------------------
// ast subcommand
// ---------------------------------------------------------------------------

func TestASTCommand(t *testing.T) {
	resetGlobalFlags()
	buf := &bytes.Buffer{}
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"ast", "../../testdata/lexer/minimal.shpl"})
	defer rootCmd.SetArgs([]string{})
	defer rootCmd.SetOut(nil)

	if err := rootCmd.Execute(); err != nil {
		t.Fatal(err)
	}
	output := buf.String()
	if !strings.Contains(output, `"title"`) {
		t.Errorf("output missing title field:\n%s", output)
	}
	if !strings.Contains(output, `"acts"`) {
		t.Errorf("output missing acts field:\n%s", output)
	}
}

func TestASTCommand_FileNotFound(t *testing.T) {
	resetGlobalFlags()
	buf := &bytes.Buffer{}
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"ast", "nonexistent.shpl"})
	defer rootCmd.SetArgs([]string{})
	defer rootCmd.SetOut(nil)

	if err := rootCmd.Execute(); err == nil {
		t.Fatal("expected error for nonexistent file, got nil")
	}
}

// ---------------------------------------------------------------------------
// run subcommand
// ---------------------------------------------------------------------------

func TestRunCommand(t *testing.T) {
	resetGlobalFlags()
	buf := &bytes.Buffer{}
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"run", "../../testdata/semantic/self-talk.shpl"})
	defer rootCmd.SetArgs([]string{})
	defer rootCmd.SetOut(nil)

	if err := rootCmd.Execute(); err != nil {
		t.Fatal(err)
	}
}

func TestRunCommand_FileNotFound(t *testing.T) {
	resetGlobalFlags()
	buf := &bytes.Buffer{}
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"run", "nonexistent.shpl"})
	defer rootCmd.SetArgs([]string{})
	defer rootCmd.SetOut(nil)

	if err := rootCmd.Execute(); err == nil {
		t.Fatal("expected error for nonexistent file, got nil")
	}
}

// ---------------------------------------------------------------------------
// version subcommand
// ---------------------------------------------------------------------------

func TestVersionCommand(t *testing.T) {
	resetGlobalFlags()
	buf := &bytes.Buffer{}
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"version"})
	defer rootCmd.SetArgs([]string{})
	defer rootCmd.SetOut(nil)

	if err := rootCmd.Execute(); err != nil {
		t.Fatal(err)
	}
	output := buf.String()
	if !strings.HasPrefix(output, "shpl ") {
		t.Errorf("output should start with 'shpl ': %s", output)
	}
	if !strings.Contains(output, "dev") {
		t.Errorf("output should contain version 'dev': %s", output)
	}
}

// ---------------------------------------------------------------------------
// about subcommand
// ---------------------------------------------------------------------------

func TestAboutCommand(t *testing.T) {
	resetGlobalFlags()
	buf := &bytes.Buffer{}
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"about"})
	defer rootCmd.SetArgs([]string{})
	defer rootCmd.SetOut(nil)

	if err := rootCmd.Execute(); err != nil {
		t.Fatal(err)
	}
	output := buf.String()
	if !strings.Contains(output, "Shakespeare Programming Language") {
		t.Errorf("output missing title:\n%s", output)
	}
	if !strings.Contains(output, "github.com/lorenzobandini") {
		t.Errorf("output missing repo:\n%s", output)
	}
}

// ---------------------------------------------------------------------------
// --trace flag
// ---------------------------------------------------------------------------

func TestTraceFlag_Tokens(t *testing.T) {
	resetGlobalFlags()
	outBuf := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	rootCmd.SetOut(outBuf)
	rootCmd.SetErr(errBuf)
	rootCmd.SetArgs([]string{"--trace", "tokens", "../../testdata/lexer/minimal.shpl"})
	defer rootCmd.SetArgs([]string{})
	defer rootCmd.SetOut(nil)
	defer rootCmd.SetErr(nil)

	if err := rootCmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(errBuf.String(), "--- TOKENS ---") {
		t.Errorf("stderr missing stage marker:\n%s", errBuf.String())
	}
}

func TestTraceFlag_Run(t *testing.T) {
	resetGlobalFlags()
	outBuf := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	rootCmd.SetOut(outBuf)
	rootCmd.SetErr(errBuf)
	rootCmd.SetArgs([]string{"--trace", "run", "../../testdata/semantic/self-talk.shpl"})
	defer rootCmd.SetArgs([]string{})
	defer rootCmd.SetOut(nil)
	defer rootCmd.SetErr(nil)

	if err := rootCmd.Execute(); err != nil {
		t.Fatal(err)
	}
	stderr := errBuf.String()
	for _, marker := range []string{"--- TOKENS ---", "--- AST ---", "--- SEMANTIC ---", "--- EXECUTE ---"} {
		if !strings.Contains(stderr, marker) {
			t.Errorf("stderr missing marker %q:\n%s", marker, stderr)
		}
	}
}

func TestTraceFlag_MarkersNotOnDebug(t *testing.T) {
	resetGlobalFlags()
	errBuf := &bytes.Buffer{}
	rootCmd.SetErr(errBuf)
	rootCmd.SetArgs([]string{"--debug", "tokens", "../../testdata/lexer/minimal.shpl"})
	defer rootCmd.SetArgs([]string{})
	defer rootCmd.SetErr(nil)

	if err := rootCmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(errBuf.String(), "--- TOKENS ---") {
		t.Errorf("--debug should not emit stage markers, got:\n%s", errBuf.String())
	}
}

// ---------------------------------------------------------------------------
// REPL subcommand
// ---------------------------------------------------------------------------

func TestRepl_BasicInteraction(t *testing.T) {
	resetGlobalFlags()
	input := "[Enter Romeo]\nRomeo: You are as good as a flower!\n\n:quit\n"
	inBuf := strings.NewReader(input)
	outBuf := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}

	rootCmd.SetIn(inBuf)
	rootCmd.SetOut(outBuf)
	rootCmd.SetErr(errBuf)
	rootCmd.SetArgs([]string{"repl"})
	defer rootCmd.SetIn(nil)
	defer rootCmd.SetOut(nil)
	defer rootCmd.SetErr(nil)
	defer rootCmd.SetArgs([]string{})

	if err := rootCmd.Execute(); err != nil {
		t.Fatal(err)
	}
}

func TestRepl_AutoDeclaration(t *testing.T) {
	resetGlobalFlags()
	input := "[Enter Hamlet]\nHamlet: Speak your mind!\n\n:quit\n"
	inBuf := strings.NewReader(input)
	outBuf := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}

	rootCmd.SetIn(inBuf)
	rootCmd.SetOut(outBuf)
	rootCmd.SetErr(errBuf)
	rootCmd.SetArgs([]string{"repl"})
	defer rootCmd.SetIn(nil)
	defer rootCmd.SetOut(nil)
	defer rootCmd.SetErr(nil)
	defer rootCmd.SetArgs([]string{})

	if err := rootCmd.Execute(); err != nil {
		t.Fatal(err)
	}
}

func TestRepl_BufferRollbackOnError(t *testing.T) {
	resetGlobalFlags()
	input := "[Enter Juliet]\nJuliet: You are as good as a flower!\n\nthis is not valid SPL\n\nJuliet: Speak your mind!\n\n:quit\n"
	inBuf := strings.NewReader(input)
	outBuf := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}

	rootCmd.SetIn(inBuf)
	rootCmd.SetOut(outBuf)
	rootCmd.SetErr(errBuf)
	rootCmd.SetArgs([]string{"repl"})
	defer rootCmd.SetIn(nil)
	defer rootCmd.SetOut(nil)
	defer rootCmd.SetErr(nil)
	defer rootCmd.SetArgs([]string{})

	if err := rootCmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(errBuf.String(), "error:") {
		t.Errorf("expected error message on stderr for invalid SPL:\n%s", errBuf.String())
	}
}

func TestRepl_TraceFlag(t *testing.T) {
	resetGlobalFlags()
	input := "[Enter Romeo]\nRomeo: Open your heart!\n\n:quit\n"
	inBuf := strings.NewReader(input)
	outBuf := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}

	rootCmd.SetIn(inBuf)
	rootCmd.SetOut(outBuf)
	rootCmd.SetErr(errBuf)
	rootCmd.SetArgs([]string{"--trace", "repl"})
	defer rootCmd.SetIn(nil)
	defer rootCmd.SetOut(nil)
	defer rootCmd.SetErr(nil)
	defer rootCmd.SetArgs([]string{})

	if err := rootCmd.Execute(); err != nil {
		t.Fatal(err)
	}
	stderr := errBuf.String()
	for _, marker := range []string{"--- TOKENS ---", "--- AST ---", "--- SEMANTIC ---", "--- EXECUTE ---"} {
		if !strings.Contains(stderr, marker) {
			t.Errorf("stderr missing marker %q:\n%s", marker, stderr)
		}
	}
}

// ---------------------------------------------------------------------------
// --help flag (smoke test: all subcommands registered)
// ---------------------------------------------------------------------------

func TestHelpOutput_ListsAllSubcommands(t *testing.T) {
	resetGlobalFlags()
	buf := &bytes.Buffer{}
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"--help"})
	defer rootCmd.SetArgs([]string{})
	defer rootCmd.SetOut(nil)

	if err := rootCmd.Execute(); err != nil {
		t.Fatal(err)
	}
	help := buf.String()
	for _, cmd := range []string{"tokens", "ast", "run", "version", "about", "repl"} {
		if !strings.Contains(help, cmd) {
			t.Errorf("help output missing subcommand %q:\n%s", cmd, help)
		}
	}
}

// TestTryQuickParse_ErrorLines verifies that a genuine syntax error on a
// single input line is detected immediately (returned as a real error) rather
// than being classified as "incomplete" (ErrIncomplete).
func TestTryQuickParse_ErrorLines(t *testing.T) {
	tests := []struct {
		name    string
		block   string
		wantErr bool   // true = genuine error, false = incomplete
		desc    string // what the test verifies
	}{
		{
			name:    "unclosed bracket is incomplete",
			block:   "[Enter Romeo",
			wantErr: false, // INCOMPLETE: caught by step 1 (L002) or step 2 (balance)
			desc:    "premature-EOF mid-construct — should wait for more input",
		},
		{
			name:    "valid bracket then trailing garbage is genuine error",
			block:   "[Enter Romeo] xyz\n",
			wantErr: true, // GENUINE ERROR: caught by step 4 (unconsumed token "xyz")
			desc:    "valid construct followed by trailing garbage on same line — should error immediately",
		},
		{
			name:    "valid standalone stage direction is incomplete",
			block:   "[Enter Romeo]\n",
			wantErr: false, // INCOMPLETE: step 5 (M001 — REPL auto-declares)
			desc:    "valid complete construct, but Romeo undeclared — needs more input",
		},
		{
			name:    "dialogue with undeclared speaker is incomplete",
			block:   "Romeo: You are a flower.\n",
			wantErr: false, // INCOMPLETE: step 5 (M001)
			desc:    "valid dialogue, speaker undeclared — REPL auto-declares on submit",
		},
		{
			name:    "garbage alone is genuine error",
			block:   "fsdgsdg\n",
			wantErr: true, // GENUINE ERROR: step 4 (unconsumed token)
			desc:    "stray word outside any construct — should error immediately",
		},
		{
			name:    "at-sign garbage is genuine error",
			block:   "@invalid\n",
			wantErr: true, // GENUINE ERROR: step 4 (unconsumed token)
			desc:    "stray word with special char — should error immediately",
		},
		{
			name:    "empty line is fine",
			block:   "",
			wantErr: false,
			desc:    "empty block — no error",
		},
		{
			name:    "exit with trailing garbage is genuine error",
			block:   "[Exit Romeo] garbage\n",
			wantErr: true, // GENUINE ERROR: step 4 ("garbage" unconsumed)
			desc:    "valid Exit followed by trailing word — should error immediately",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tryQuickParse(tt.block)
			if tt.wantErr && (err == nil || err == ErrIncomplete) {
				t.Errorf("[%s] expected a genuine error, got %v", tt.desc, err)
			}
			if !tt.wantErr && err != nil && err != ErrIncomplete {
				t.Errorf("[%s] expected incomplete (ErrIncomplete), got %v", tt.desc, err)
			}
		})
	}
}

// TestSingleReader_SharedBuffering verifies that the singleReader correctly
// handles both SPL text lines and runtime input from the same *bufio.Reader,
// simulating a piped stdin scenario where all data arrives at once.
// The runtime reads ONE BYTE at a time (buf [1]byte in readInt), so the test
// reads one byte at a time to match production behavior.
func TestSingleReader_SharedBuffering(t *testing.T) {
	input := "line one\nline two\n\n42\nmore text\n\n:quit\n"
	sr := newSingleReader(strings.NewReader(input), io.Discard)

	// Read SPL lines (simulating the REPL loop).
	line1, err := sr.readLine()
	if err != nil || line1 != "line one" {
		t.Fatalf("expected 'line one', got %q (err=%v)", line1, err)
	}
	line2, err := sr.readLine()
	if err != nil || line2 != "line two" {
		t.Fatalf("expected 'line two', got %q (err=%v)", line2, err)
	}
	blank, err := sr.readLine()
	if err != nil || blank != "" {
		t.Fatalf("expected '', got %q (err=%v)", blank, err)
	}

	// Simulate pipeline execution: runtime reads stdin one byte at a time
	// (matching readInt's buf [1]byte).
	sr.replayMode()
	one := make([]byte, 1)
	var gotBuf []byte
	for i := 0; i < 3; i++ { // read "42\n" = 3 bytes
		n, err := sr.Read(one)
		if err != nil {
			t.Fatalf("replay Read byte %d: %v", i, err)
		}
		gotBuf = append(gotBuf, one[:n]...)
	}
	got := string(gotBuf)
	if got != "42\n" {
		t.Fatalf("expected '42\\n', got %q", got)
	}
}

// TestRepl_RollbackStdinRecords verifies that when a replay fails, the
// recorded-stdin log is rolled back to its pre-submission state, so that a
// subsequent valid submission replays the correct (not shifted) stdin values.
//
// Sequence:
//  1. Submit [Enter R] + R: Listen + blank line → pipeline reads "42\n" from stdin
//  2. Submit invalid SPL → pipeline fails → buffer + recorded-stdin roll back
//  3. Submit R: Open your heart! + blank line → pipeline replays "42\n" from
//     recorded-stdin → outputs "42"
func TestRepl_RollbackStdinRecording(t *testing.T) {
	resetGlobalFlags()
	// Block 1: Enter + Listen (reads runtime input "42\n" from stdin)
	// Block 2: invalid SPL (causes failure + rollback)
	// Block 3: OpenHeart (replays "42" from recorded-stdin, outputs "42")
	//
	// The "42\n" between the first blank line and the invalid text is
	// consumed by the runtime's Listen statement on the first pipeline run.
	input := fmt.Sprintf(
		"[Enter Romeo]\nRomeo: Listen to your heart!\n\n" +
			"42\n" +
			"this is not valid SPL text\n\n" +
			"Romeo: Open your heart!\n\n" +
			":quit\n",
	)
	inBuf := strings.NewReader(input)
	outBuf := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}

	rootCmd.SetIn(inBuf)
	rootCmd.SetOut(outBuf)
	rootCmd.SetErr(errBuf)
	rootCmd.SetArgs([]string{"repl"})
	defer rootCmd.SetIn(nil)
	defer rootCmd.SetOut(nil)
	defer rootCmd.SetErr(nil)
	defer rootCmd.SetArgs([]string{})

	if err := rootCmd.Execute(); err != nil {
		t.Fatal(err)
	}
	// stderr should contain the error from the invalid SPL submission.
	if !strings.Contains(errBuf.String(), "error:") {
		t.Errorf("expected error on stderr after invalid SPL:\n%s", errBuf.String())
	}
	// stdout should contain "42" from the successful OpenHeart replay.
	if !strings.Contains(outBuf.String(), "42") {
		t.Errorf("expected stdout to contain '42' after rollback and replay:\nstdout=%q\nstderr=%q",
			outBuf.String(), errBuf.String())
	}
}
