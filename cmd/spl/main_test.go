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
	rootCmd.SetArgs([]string{"tokens", "../../testdata/lexer/minimal.spl"})
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
	rootCmd.SetArgs([]string{"tokens", "nonexistent.spl"})
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
	rootCmd.SetArgs([]string{"tokens", "../../testdata/lexer/bad-char.spl"})
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
	rootCmd.SetArgs([]string{"ast", "../../testdata/lexer/minimal.spl"})
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
	rootCmd.SetArgs([]string{"ast", "nonexistent.spl"})
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
	rootCmd.SetArgs([]string{"run", "../../testdata/semantic/self-talk.spl"})
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
	rootCmd.SetArgs([]string{"run", "nonexistent.spl"})
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
	if !strings.Contains(output, "spl ") {
		t.Errorf("output should contain 'spl ': %s", output)
	}
	if !strings.Contains(output, "commit:") {
		t.Errorf("output should contain 'commit:': %s", output)
	}
	if !strings.Contains(output, "date:") {
		t.Errorf("output should contain 'date:': %s", output)
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
	if !strings.Contains(output, "888") {
		t.Errorf("output missing ASCII art:\n%s", output)
	}
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
	rootCmd.SetArgs([]string{"--trace", "tokens", "../../testdata/lexer/minimal.spl"})
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
	rootCmd.SetArgs([]string{"--trace", "run", "../../testdata/semantic/self-talk.spl"})
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
	rootCmd.SetArgs([]string{"--debug", "tokens", "../../testdata/lexer/minimal.spl"})
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

func TestRepl_AcceptsAllLinesAndSubmits(t *testing.T) {
	resetGlobalFlags()
	input := "[Enter Juliet]\nJuliet: You are as good as a flower!\n\nJuliet: Speak your mind!\n\n:quit\n"
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
	// Verify the REPL completed without crash or unexpected error.
	if strings.Contains(errBuf.String(), "panic") {
		t.Errorf("unexpected panic in output:\n%s", errBuf.String())
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

// TestTryQuickParse_ErrorLines verifies the quick-parse pass flags only
// structural issues (unclosed brackets, open expressions) as incomplete.
// All other text passes through to be caught or accepted on submit.
func TestTryQuickParse_ErrorLines(t *testing.T) {
	t.Run("incomplete", func(t *testing.T) {
		tests := []struct {
			name  string
			block string
			desc  string
		}{
			{
				name:  "unclosed bracket",
				block: "[Enter Romeo",
				desc:  "L002 unclosed bracket — wait for more input",
			},
			{
				name:  "trailing garbage after valid bracket",
				block: "[Enter Romeo] xyz\n",
				desc:  "trailing text that doesn't end with terminator",
			},
			{
				name:  "standalone stage direction",
				block: "[Enter Romeo]\n",
				desc:  "bracket ends with ] — not a terminator",
			},
			{
				name:  "garbage alone",
				block: "fsdgsdg\n",
				desc:  "garbage without terminator",
			},
			{
				name:  "at-sign garbage",
				block: "@invalid\n",
				desc:  "garbage without terminator",
			},
			{
				name:  "exit with trailing garbage",
				block: "[Exit Romeo] garbage\n",
				desc:  "trailing garbage without terminator",
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				err := tryQuickParse(tt.block)
				if err != ErrIncomplete {
					t.Errorf("[%s] expected ErrIncomplete, got %v", tt.desc, err)
				}
			})
		}
	})

	t.Run("no_error", func(t *testing.T) {
		tests := []struct {
			name  string
			block string
			desc  string
		}{
			{
				name:  "empty block",
				block: "",
				desc:  "empty — no error",
			},
			{
				name:  "dialogue with undeclared speaker",
				block: "Romeo: You are a flower.\n",
				desc:  "terminated with . — passes through to submit pipeline",
			},
			{
				name:  "character declaration",
				block: "Romeo, a young man.\n",
				desc:  "terminated with . — passes through",
			},
			{
				name:  "complete act header",
				block: "Act I: The Act.\n",
				desc:  "terminated with . — passes through",
			},
			{
				name:  "complete scene header",
				block: "Scene I: The Scene.\n",
				desc:  "terminated with . — passes through",
			},
			{
				name:  "title-like prose",
				block: "The Branch Test.\n",
				desc:  "terminated with . — passes through",
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				err := tryQuickParse(tt.block)
				if err != nil {
					t.Errorf("[%s] expected no error, got %v", tt.desc, err)
				}
			})
		}
	})
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

// TestRepl_StdinReplayAcrossSubmits verifies that recorded stdin (from
// Listen/OpenMind in a previous submission) is correctly replayed when the
// full pipeline runs again on a subsequent submission. The "this is not
// valid SPL text" between blocks is silently ignored by the parser (no AST
// effect), but the pipeline replay still serves the recorded "42\n" to
// OpenHeart.
func TestRepl_StdinReplayAcrossSubmits(t *testing.T) {
	resetGlobalFlags()
	// Block 1: Enter + Listen (reads runtime input "42\n" from stdin)
	// Runtime stdin data: "42\n"
	// Block 2: garbage text (parsed into nothing, no effect)
	// Block 3: OpenHeart (replays "42" from recorded-stdin, outputs "42")
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
	// stdout should contain "42" from the OpenHeart replay.
	if !strings.Contains(outBuf.String(), "42") {
		t.Errorf("expected stdout to contain '42' after stdin replay:\nstdout=%q\nstderr=%q",
			outBuf.String(), errBuf.String())
	}
}

// ---------------------------------------------------------------------------
// classifyLine
// ---------------------------------------------------------------------------

func TestClassifyLine(t *testing.T) {
	tests := []struct {
		line string
		want replLineKind
	}{
		{"", lineBlank},
		{"  ", lineBlank},
		{"Act I: The Act.", lineActHeader},
		{"Act II: More drama.", lineActHeader},
		{"Scene I: The Scene.", lineSceneHeader},
		{"Scene V: Another.", lineSceneHeader},
		{"Romeo, a young man.", lineCharDecl},
		{"Juliet, a woman.", lineCharDecl},
		{"[Enter Romeo]", lineStageDir},
		{"[Exit Juliet]", lineStageDir},
		{"[Exeunt]", lineStageDir},
		{"[Enter Romeo and Juliet]", lineStageDir},
		{"Romeo: You are a flower!", lineDialogue},
		{"Juliet: Speak your mind!", lineDialogue},
		{"The Truth Machine.", lineTitle},
		{"The Branch Test.", lineTitle},
		{"fsdgsdg", lineOther},
		{"@invalid", lineOther},
	}
	for _, tt := range tests {
		t.Run(tt.line, func(t *testing.T) {
			got := classifyLine(tt.line)
			if got != tt.want {
				t.Errorf("classifyLine(%q) = %v, want %v", tt.line, got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// derivePhase
// ---------------------------------------------------------------------------

func TestDerivePhase(t *testing.T) {
	tests := []struct {
		text string
		want replPhase
	}{
		{"", phaseTitle},
		{"The Truth Machine.", phaseTitle},
		{"Romeo, a young man.", phaseChars},
		{"The Truth Machine.\nRomeo, a young man.", phaseChars},
		{"Juliet, a woman.", phaseChars},
		{"Act I: The Act.", phaseBody},
		{"The Truth Machine.\nAct I: The Act.", phaseBody},
		{"[Enter Romeo]", phaseTitle},
		{"Romeo: You are a flower!", phaseTitle},
		{"Romeo, a young man.\nAct I: The Truth.", phaseBody},
	}
	for _, tt := range tests {
		t.Run(tt.text, func(t *testing.T) {
			got := derivePhase(tt.text)
			if got != tt.want {
				t.Errorf("derivePhase(%q) = %v, want %v", tt.text, got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// isOpenExpression
// ---------------------------------------------------------------------------

func TestIsOpenExpression(t *testing.T) {
	tests := []struct {
		text string
		want bool
	}{
		{"", false},
		{".", false},
		{"foo", true},
		{"foo.", false},
		{"foo!", false},
		{"foo?", false},
		{"[Enter Romeo]", true},
		{"Romeo: You are a flower!", false},
		{"[Enter Romeo]\nRomeo:", true},
		{"Act I: The Act.\nScene I: The Scene.", false},
	}
	for _, tt := range tests {
		t.Run(tt.text, func(t *testing.T) {
			got := isOpenExpression(tt.text)
			if got != tt.want {
				t.Errorf("isOpenExpression(%q) = %v, want %v", tt.text, got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// REPL skeleton injection ordering
// ---------------------------------------------------------------------------

func TestRepl_SkeletonInjectionOrder(t *testing.T) {
	resetGlobalFlags()
	input := "The Truth Machine.\n\nRomeo, a young man.\n\n[Enter Romeo]\nRomeo: Open your heart!\n\n:quit\n"
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
	if strings.Contains(errBuf.String(), "S002") {
		t.Errorf("unexpected S002 (skeleton injected before declarations):\nstderr=%q", errBuf.String())
	}
}

func TestRepl_NoSkeletonWhenUserProvidesActs(t *testing.T) {
	resetGlobalFlags()
	input := "The Truth Machine.\n\nRomeo, a young man.\n\nAct I: The Truth.\nScene I: The Init.\n[Enter Romeo]\nRomeo: Open your heart!\n\n:quit\n"
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
	if strings.Contains(errBuf.String(), "S002") {
		t.Errorf("unexpected S002 with user-provided acts:\nstderr=%q", errBuf.String())
	}
}

func TestRepl_MissingCharDeclBeforeActI_ReportsS002(t *testing.T) {
	resetGlobalFlags()
	input := "The Truth Machine.\n\nAct I: The Truth.\nScene I: The Scene.\n[Enter Romeo]\n\n:quit\n"
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
	if !strings.Contains(errBuf.String(), "S002") {
		t.Errorf("expected S002 for missing char decl before Act I:\nstderr=%q", errBuf.String())
	}
}

// ---------------------------------------------------------------------------
// Auto-declaration of new characters in later submits (Step 5.3)
// ---------------------------------------------------------------------------

func TestRepl_AutoDeclarationMidProgram(t *testing.T) {
	resetGlobalFlags()
	// First submit: body content references Romeo (auto-declared on skeleton).
	// Second submit: introduces Juliet (auto-declared without duplicating Romeo).
	input := "[Enter Romeo]\nRomeo: Open your heart!\n\n[Enter Juliet]\nJuliet: Open your heart!\n\n:quit\n"
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
	if strings.Contains(errBuf.String(), "M001") {
		t.Errorf("unexpected M001 — characters not auto-declared:\nstderr=%q", errBuf.String())
	}
	if strings.Contains(errBuf.String(), "S002") {
		t.Errorf("unexpected S002:\nstderr=%q", errBuf.String())
	}
}

// ---------------------------------------------------------------------------
// Partial output rollback (Step 5.5)
// ---------------------------------------------------------------------------

func TestRepl_PartialOutputRollback(t *testing.T) {
	resetGlobalFlags()
	// First submit succeeds and produces output.
	// Second submit is invalid and should roll back without advancing the cursor.
	input := "[Enter Romeo]\nRomeo: Open your heart!\n\n" +
		"garbage that will fail\n\n" +
		"Romeo: Open your heart!\n\n:quit\n"
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
	if strings.Contains(outBuf.String(), "panic") || strings.Contains(errBuf.String(), "panic") {
		t.Errorf("unexpected panic in output:\nstdout=%q\nstderr=%q", outBuf.String(), errBuf.String())
	}
	// The third submit's OpenHeart should produce output (cursor preserved).
	if !strings.Contains(outBuf.String(), "0\n") {
		t.Errorf("expected '0\\n' from OpenHeart:\nstdout=%q\nstderr=%q", outBuf.String(), errBuf.String())
	}
}

// ---------------------------------------------------------------------------
// User-log regression tests (Step 5.6)
// ---------------------------------------------------------------------------

func TestRepl_UserLogRegressions(t *testing.T) {
	resetGlobalFlags()

	t.Run("title-only-no-s002", func(t *testing.T) {
		// Log 1: title-only accumulates silently (no pipeline until body)
		inBuf := strings.NewReader("The Truth Machine.\n\nHello: Open your heart!\n\n:quit\n")
		outBuf := &bytes.Buffer{}
		errBuf := &bytes.Buffer{}
		runREPL(t, inBuf, outBuf, errBuf)
		if strings.Contains(errBuf.String(), "S002") {
			t.Errorf("title-only got S002:\nstderr=%q", errBuf.String())
		}
	})

	t.Run("char-decl-and-body", func(t *testing.T) {
		// Log 2: char decl + body content should work via skeleton
		inBuf := strings.NewReader("Romeo, a young man.\n\n[Enter Romeo]\nRomeo: Open your heart!\n\n:quit\n")
		outBuf := &bytes.Buffer{}
		errBuf := &bytes.Buffer{}
		runREPL(t, inBuf, outBuf, errBuf)
		if strings.Contains(errBuf.String(), "S002") {
			t.Errorf("char-decl-body got S002:\nstderr=%q", errBuf.String())
		}
	})

	t.Run("stack-test-body-only", func(t *testing.T) {
		// Log 3/4: Body-only submit triggers skeleton + auto-declaration
		inBuf := strings.NewReader("[Enter Romeo]\nRomeo: Open your heart!\n\n:quit\n")
		outBuf := &bytes.Buffer{}
		errBuf := &bytes.Buffer{}
		runREPL(t, inBuf, outBuf, errBuf)
		if strings.Contains(errBuf.String(), "S002") {
			t.Errorf("stack-test got S002:\nstderr=%q", errBuf.String())
		}
	})

	t.Run("act-without-scene-reports-s007", func(t *testing.T) {
		// Log 5: Act II without Scene I should report S007
		inBuf := strings.NewReader("The Hello Test.\n\nRomeo, a young man.\n\nAct I: The Start.\nScene I: The Scene.\n[Enter Romeo]\nRomeo: You are a flower!\n\nAct II: The Second.\n\n:quit\n")
		outBuf := &bytes.Buffer{}
		errBuf := &bytes.Buffer{}
		runREPL(t, inBuf, outBuf, errBuf)
		if !strings.Contains(errBuf.String(), "S007") {
			t.Errorf("expected S007 for Act II without Scene I:\nstderr=%q", errBuf.String())
		}
	})
}

// runREPL is a helper that drives the REPL with the given input/output buffers.
func runREPL(t *testing.T, inBuf io.Reader, outBuf, errBuf *bytes.Buffer) {
	t.Helper()
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
