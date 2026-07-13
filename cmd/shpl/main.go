package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/signal"
	"regexp"
	"strings"

	"github.com/spf13/cobra"

	"github.com/lorenzobandini/shakespeare-interpreter-go/internal/lexer"
	"github.com/lorenzobandini/shakespeare-interpreter-go/internal/logger"
	"github.com/lorenzobandini/shakespeare-interpreter-go/internal/parser"
	"github.com/lorenzobandini/shakespeare-interpreter-go/internal/runtime"
	"github.com/lorenzobandini/shakespeare-interpreter-go/internal/semantic"
)

// Global flags.
var (
	debugFlag bool
	traceFlag bool
)

// Build info — set via ldflags.
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

// rootCmd is the base command.
var rootCmd = &cobra.Command{
	Use:   "shpl",
	Short: "Shakespeare Programming Language interpreter",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		switch {
		case traceFlag:
			logger.Init(logger.LevelDebug)
		case debugFlag:
			logger.Init(logger.LevelDebug)
		default:
			logger.Init(logger.LevelInfo)
		}
	},
}

// ---------------------------------------------------------------------------
// version subcommand
// ---------------------------------------------------------------------------

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		_, err := fmt.Fprintf(cmd.OutOrStdout(),
			"shpl %s (commit %s, built %s)\n", version, commit, date)
		return err
	},
}

// ---------------------------------------------------------------------------
// about subcommand
// ---------------------------------------------------------------------------

const aboutText = `Shakespeare Programming Language (SPL)
Original language design: Karl Hasselström and Jon Åslund (2001)

Reference implementation: zmbc/shakespearelang
This implementation: https://github.com/lorenzobandini/shakespeare-interpreter-go

Phase 1-4: Lexer, Parser, Semantic Analysis, Runtime
Phase 5: CLI Integration (Cobra)
Go version: 1.26.5
`

var aboutCmd = &cobra.Command{
	Use:   "about",
	Short: "About the Shakespeare Programming Language",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		_, err := fmt.Fprint(cmd.OutOrStdout(), aboutText)
		return err
	},
}

// ---------------------------------------------------------------------------
// tokens subcommand
// ---------------------------------------------------------------------------

var tokensCmd = &cobra.Command{
	Use:   "tokens <file>",
	Short: "Tokenize an SPL source file and print the token stream",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		src, err := os.ReadFile(args[0])
		if err != nil {
			return fmt.Errorf("cannot read file %q: %w", args[0], err)
		}
		if traceFlag {
			_, _ = fmt.Fprintln(cmd.ErrOrStderr(), "--- TOKENS ---")
		}
		tokens, err := lexer.New(string(src)).ScanTokens()
		if err != nil {
			return fmt.Errorf("%v", err)
		}
		out := cmd.OutOrStdout()
		for _, tok := range tokens {
			if _, err := fmt.Fprintln(out, tok.String()); err != nil {
				return err
			}
		}
		return nil
	},
}

// ---------------------------------------------------------------------------
// ast subcommand
// ---------------------------------------------------------------------------

var astCmd = &cobra.Command{
	Use:   "ast <file>",
	Short: "Parse an SPL source file and print the AST as JSON",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		src, err := os.ReadFile(args[0])
		if err != nil {
			return fmt.Errorf("cannot read file %q: %w", args[0], err)
		}
		if traceFlag {
			_, _ = fmt.Fprintln(cmd.ErrOrStderr(), "--- TOKENS ---")
		}
		tokens, err := lexer.New(string(src)).ScanTokens()
		if err != nil {
			return fmt.Errorf("%v", err)
		}
		if traceFlag {
			_, _ = fmt.Fprintln(cmd.ErrOrStderr(), "--- AST ---")
		}
		prog, err := parser.New(tokens).Parse()
		if err != nil {
			return fmt.Errorf("%v", err)
		}
		out, err := json.MarshalIndent(prog, "", "  ")
		if err != nil {
			return err
		}
		_, err = fmt.Fprintln(cmd.OutOrStdout(), string(out))
		return err
	},
}

// ---------------------------------------------------------------------------
// run subcommand
// ---------------------------------------------------------------------------

var runCmd = &cobra.Command{
	Use:   "run <file>",
	Short: "Execute an SPL source file",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		src, err := os.ReadFile(args[0])
		if err != nil {
			return fmt.Errorf("cannot read file %q: %w", args[0], err)
		}
		return runPipeline(string(src), os.Stdin, cmd.OutOrStdout(), args[0], cmd.ErrOrStderr())
	},
}

// ---------------------------------------------------------------------------
// Shared pipeline runner (used by runCmd and the REPL)
// ---------------------------------------------------------------------------

func runPipeline(src string, in io.Reader, out io.Writer, filename string, traceOut io.Writer) error {
	if traceFlag {
		_, _ = fmt.Fprintln(traceOut, "--- TOKENS ---")
	}
	tokens, err := lexer.New(src).ScanTokens()
	if err != nil {
		return fmt.Errorf("%v", err)
	}
	if traceFlag {
		_, _ = fmt.Fprintln(traceOut, "--- AST ---")
	}
	prog, err := parser.New(tokens).Parse()
	if err != nil {
		return fmt.Errorf("%v", err)
	}
	if traceFlag {
		_, _ = fmt.Fprintln(traceOut, "--- SEMANTIC ---")
	}
	res := semantic.New(filename, prog).Analyze(prog)
	if !res.OK() {
		var b strings.Builder
		for _, e := range res.Errors {
			b.WriteString(e.Error())
			b.WriteString("\n")
		}
		return fmt.Errorf("%s", strings.TrimSuffix(b.String(), "\n"))
	}
	if traceFlag {
		_, _ = fmt.Fprintln(traceOut, "--- EXECUTE ---")
	}
	if err := runtime.Execute(prog, res, in, out, filename); err != nil {
		return fmt.Errorf("%v", err)
	}
	return nil
}

// ---------------------------------------------------------------------------
// REPL subcommand — Replay-based Accumulating Buffer Model
// ---------------------------------------------------------------------------

// Patterns for auto-declaration of characters.
var (
	enterRe   = regexp.MustCompile(`\[Enter\s+([^]]+?)\]`)
	exitRe    = regexp.MustCompile(`\[Exit\s+([^]]+?)\]`)
	exeuntRe  = regexp.MustCompile(`\[Exeunt\s+([^]]+?)\]`)
	speakerRe = regexp.MustCompile(`^([A-Za-z]\w*):`)
)

// replState holds all mutable state across REPL turns.
type replState struct {
	out           io.Writer
	err           io.Writer
	buffer        bytes.Buffer
	declared      map[string]bool
	skeletonBuilt bool

	// Output slicing.
	lastOutputLen int

	// Trace output writer (usually cmd.ErrOrStderr()).
	traceOut io.Writer
}

// extractChars extracts potential character names from a block of SPL text.
func extractChars(text string) []string {
	seen := map[string]bool{}
	var chars []string
	add := func(name string) {
		name = strings.TrimSpace(name)
		if name == "" || seen[strings.ToLower(name)] {
			return
		}
		seen[strings.ToLower(name)] = true
		chars = append(chars, name)
	}
	for _, m := range enterRe.FindAllStringSubmatch(text, -1) {
		for _, part := range strings.Split(m[1], " and ") {
			add(part)
		}
	}
	for _, m := range exitRe.FindAllStringSubmatch(text, -1) {
		add(m[1])
	}
	for _, m := range exeuntRe.FindAllStringSubmatch(text, -1) {
		for _, part := range strings.Split(m[1], " and ") {
			add(part)
		}
	}
	for _, m := range speakerRe.FindAllStringSubmatch(text, -1) {
		add(m[1])
	}
	return chars
}

// replayBlock appends text, auto-declares, runs the pipeline, handles output
// slicing and rollback.
func (rs *replState) replayBlock(input string, sr *singleReader) error {
	bufCheck := rs.buffer.Len()
	recCheck := sr.recordCheckpoint()

	// 1. Always insert skeleton on the first submission so that auto-declared
	//    characters have a valid SPL structure to live in.
	if !rs.skeletonBuilt {
		rs.buffer.WriteString("The REPL Session.\n\nAct I: The REPL Session.\nScene I: The REPL Session.\n\n")
		rs.skeletonBuilt = true
	}

	// 2. Auto-declare newly discovered characters.
	newChars := extractChars(input)
	if len(newChars) > 0 {
		var decls strings.Builder
		for _, name := range newChars {
			key := strings.ToLower(name)
			if !rs.declared[key] {
				rs.declared[key] = true
				_, _ = fmt.Fprintf(&decls, "%s, a REPL character.\n", name)
			}
		}
		if decls.Len() > 0 {
			declText := decls.String()
			// Insert declarations after Title, before Act I.
			buf := rs.buffer.Bytes()
			actIdx := bytes.Index(bytes.ToLower(buf), []byte("act i"))
			if actIdx >= 0 {
				var newBuf bytes.Buffer
				newBuf.Write(buf[:actIdx])
				newBuf.WriteString(declText)
				newBuf.WriteString("\n")
				newBuf.Write(buf[actIdx:])
				rs.buffer = newBuf
			} else {
				rs.buffer.WriteString(declText)
				rs.buffer.WriteString("\n")
			}
		}
	}

	// 3. Append input to buffer.
	rs.buffer.WriteString(input)
	rs.buffer.WriteString("\n")

	// 4. Start replay mode so the pipeline sees recorded bytes first, then
	//    fresh stdin.
	sr.replayMode()

	// 5. Run the full pipeline on captured output.
	var captureOut bytes.Buffer
	src := rs.buffer.String()
	err := runPipeline(src, sr, &captureOut, "repl", rs.traceOut)
	if err != nil {
		// Rollback buffer and recorded inputs.
		rs.buffer.Truncate(bufCheck)
		sr.rollbackRecorded(recCheck)
		for _, name := range newChars {
			delete(rs.declared, strings.ToLower(name))
		}
		if bufCheck == 0 {
			rs.skeletonBuilt = false
		}
		return fmt.Errorf("error: %v", err)
	}

	// 6. Show output delta.
	full := captureOut.String()
	if len(full) > rs.lastOutputLen {
		delta := full[rs.lastOutputLen:]
		if _, err := fmt.Fprint(rs.out, delta); err != nil {
			return err
		}
	}
	rs.lastOutputLen = len(full)
	return nil
}

const replHelp = `REPL commands:
  :quit, :exit   Exit the REPL
  :help          Show this help
  :reset         Reset the session (clear buffer, declarations, recordings)

Enter SPL text (dialogue, stage directions) and submit with a blank line.
`

// singleReader is the single stdin reader for the entire REPL session. Both
// SPL text line-reading and runtime input-reading go through the same
// *bufio.Reader, so piped stdin is never split between two competing readers.
// SPL text reads are NOT recorded (they live in the accumulating buffer).
// Runtime input reads ARE recorded for deterministic replay.
type singleReader struct {
	br       *bufio.Reader
	recorded []byte
	cursor   int
	promptW  io.Writer
	inPrompt bool
}

func newSingleReader(r io.Reader, promptW io.Writer) *singleReader {
	return &singleReader{br: bufio.NewReader(r), promptW: promptW}
}

// readLine reads one SPL text line (up to \n) without recording it.
// Handles both \n (Unix/Mac) and \r\n (Windows) line endings.
func (sr *singleReader) readLine() (string, error) {
	line, err := sr.br.ReadString('\n')
	line = strings.TrimSuffix(line, "\r\n")
	line = strings.TrimSuffix(line, "\n")
	if err == io.EOF && len(line) > 0 {
		return line, nil
	}
	if err != nil {
		return "", err
	}
	return line, nil
}

// Read implements io.Reader for runtime input. It first serves previously
// recorded bytes (deterministic replay), then reads from the shared
// bufio.Reader, recording new bytes as they arrive.
func (sr *singleReader) Read(p []byte) (int, error) {
	if sr.cursor < len(sr.recorded) {
		n := copy(p, sr.recorded[sr.cursor:])
		sr.cursor += n
		return n, nil
	}
	if !sr.inPrompt {
		_, _ = fmt.Fprint(sr.promptW, "input> ")
		sr.inPrompt = true
	}
	n, err := sr.br.Read(p)
	if n > 0 {
		sr.recorded = append(sr.recorded, p[:n]...)
		sr.cursor += n
	}
	return n, err
}

// replayMode resets the recorded cursor to 0 so that the next Read serves
// recorded bytes again (for a fresh pipeline replay).
func (sr *singleReader) replayMode() {
	sr.cursor = 0
	sr.inPrompt = false
}

// recordCheckpoint returns the current recorded length for rollback.
func (sr *singleReader) recordCheckpoint() int {
	return len(sr.recorded)
}

// rollbackRecorded truncates the recorded slice to the given checkpoint.
func (sr *singleReader) rollbackRecorded(cp int) {
	sr.recorded = sr.recorded[:cp]
	sr.cursor = cp
}

// ErrIncomplete is returned by tryQuickParse when the block merely needs more
// input (unclosed bracket, undeclared characters, missing stage entrance)
// rather than containing a genuine syntax error.
var ErrIncomplete = fmt.Errorf("incomplete input")

// tryQuickParse runs the lexer and parser on block (wrapped in a minimal
// skeleton) to decide whether the REPL should keep accumulating (incomplete)
// or report the error immediately (genuine syntax error).
//
// Decision algorithm (in order):
//
//  1. LEX block directly.
//     - L002 (unterminated bracket)          → INCOMPLETE (unclosed `[`)
//     - any other lex error                  → GENUINE ERROR
//
//  2. BRACKET BALANCE on lexed tokens.
//     - more LBRACKET than RBRACKET          → INCOMPLETE (unclosed `[`,
//     backstop if lexer L002
//     didn't fire)
//
//  3. FULL PARSE with skeleton (char X + [Enter X]).
//     - S013 (missing Enter before dialogue) → INCOMPLETE (user hasn't
//     entered chars yet)
//     - any other parse error                → GENUINE ERROR
//
//  4. UNCONSUMED TOKENS (parser.Done()).
//     - parser ignored trailing words        → GENUINE ERROR
//     ("fsdgsdg", "@invalid", "[Exit] xyz")
//
//  5. SEMANTIC ANALYSIS on parsed program.
//     - M001 (undeclared character)          → INCOMPLETE (REPL
//     auto-declares on submit)
//     - any other semantic error             → GENUINE ERROR
//
// Steps 1-2 are structurally separate from step 4: premature EOF (unclosed
// bracket) is caught before parsing even begins, while trailing garbage that
// passes the parser's lax statement loop is caught after parsing. They are
// NOT proxies for each other.
func tryQuickParse(block string) error {
	if block == "" {
		return nil
	}

	// 1. Lex for structural completeness (unclosed brackets).
	tokens, err := lexer.New(block).ScanTokens()
	if err != nil {
		if strings.Contains(err.Error(), "L002") {
			return ErrIncomplete
		}
		return err
	}
	balance := 0
	for _, t := range tokens {
		if t.Type == lexer.TokenLBracket {
			balance++
		}
		if t.Type == lexer.TokenRBracket {
			balance--
		}
	}
	if balance > 0 {
		return ErrIncomplete
	}

	// 2. Wrap in a minimal program so the parser + analyzer reach the user's
	//    actual input rather than failing early at S002 or S013.
	src := "The REPL Session.\n\nX, a character.\n\nAct I: The REPL Session.\nScene I: The REPL Session.\n[Enter X]\n\n" + block + "\n"
	tokens2, err := lexer.New(src).ScanTokens()
	if err != nil {
		return err
	}
	p := parser.New(tokens2)
	prog, err := p.Parse()
	if err != nil {
		s := err.Error()
		// S013 is expected when the block has dialogue but the Enter hasn't
		// been typed yet.
		if strings.Contains(s, "S013") {
			return ErrIncomplete
		}
		return err
	}
	// Check for unconsumed tokens: the parser silently ignores stray words
	// outside of stage directions and dialogue (e.g. "fsdgsdg" or "@invalid").
	// Only top-level words (depth == 0) are considered — words inside valid
	// `[...]` constructs like [Enter Romeo] are skipped via bracket-depth
	// tracking.
	if !p.Done() {
		tokens3, _ := lexer.New(block).ScanTokens()
		depth := 0
		for _, tok := range tokens3 {
			switch tok.Type {
			case lexer.TokenLBracket:
				depth++
				continue
			case lexer.TokenRBracket:
				depth--
				continue
			}
			if depth > 0 {
				continue
			}
			if tok.Type == lexer.TokenWord &&
				!strings.EqualFold(tok.Lexeme, "enter") &&
				!strings.EqualFold(tok.Lexeme, "exit") &&
				!strings.EqualFold(tok.Lexeme, "exeunt") {
				return fmt.Errorf("error[S014]: unexpected text '%s'", tok.Lexeme)
			}
		}
	}
	// Run the semantic analyser so that M001 (undeclared character) and
	// similar semantic-level errors are surfaced immediately.
	res := semantic.New("(preview)", prog).Analyze(prog)
	if !res.OK() {
		s := res.Errors[0].Error()
		// M001 (undeclared character) is expected when the user hasn't
		// typed a declaration yet — the REPL auto-declares on submission.
		if strings.Contains(s, "M001") {
			return ErrIncomplete
		}
		return fmt.Errorf("%v", res.Errors[0])
	}
	return nil
}

// replCmd starts the interactive REPL.
var replCmd = &cobra.Command{
	Use:   "repl",
	Short: "Start an interactive SPL REPL",
	Long: `Start an interactive Read-Eval-Print Loop for the Shakespeare
Programming Language.

Enter SPL dialogue or stage directions line by line. Submit your input
with a blank line. The REPL maintains state across submissions.`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		rs := &replState{
			out:      cmd.OutOrStdout(),
			err:      cmd.ErrOrStderr(),
			declared: map[string]bool{},
			traceOut: cmd.ErrOrStderr(),
		}
		sr := newSingleReader(cmd.InOrStdin(), rs.err)

		// Handle SIGINT gracefully.
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, os.Interrupt)
		done := make(chan struct{}, 1)
		go func() {
			select {
			case <-sigCh:
				_, _ = fmt.Fprintln(rs.err, "^C")
				os.Exit(0)
			case <-done:
			}
		}()
		defer close(done)

		var block bytes.Buffer

		for {
			if block.Len() == 0 {
				_, _ = fmt.Fprint(rs.err, "spl> ")
			} else {
				_, _ = fmt.Fprint(rs.err, "...> ")
			}

			line, err := sr.readLine()
			if err != nil {
				if err == io.EOF {
					break
				}
				return err
			}

			// Meta-commands (only at block start).
			if block.Len() == 0 && strings.HasPrefix(line, ":") {
				switch line {
				case ":quit", ":exit":
					return nil
				case ":help":
					_, _ = fmt.Fprint(rs.err, replHelp)
					continue
				case ":reset":
					rs.buffer.Reset()
					sr.recorded = nil
					sr.cursor = 0
					rs.lastOutputLen = 0
					rs.declared = map[string]bool{}
					rs.skeletonBuilt = false
					continue
				default:
					_, _ = fmt.Fprintf(rs.err, "unknown command: %s\n", line)
					continue
				}
			}

			// Blank line submits the accumulated block.
			if strings.TrimSpace(line) == "" {
				if block.Len() == 0 {
					continue
				}
				input := block.String()
				block.Reset()
				if err := rs.replayBlock(input, sr); err != nil {
					_, _ = fmt.Fprintln(rs.err, err)
				}
				continue
			}

			// Save checkpoint so we can rollback on quick-parse failure.
			cp := block.Len()
			block.WriteString(line)
			block.WriteString("\n")

			// Try a quick parse to detect genuine syntax errors
			// immediately rather than waiting for a blank-line submit.
			if err := tryQuickParse(block.String()); err != nil && err != ErrIncomplete {
				block.Truncate(cp)
				_, _ = fmt.Fprintf(rs.err, "error: %v\n", err)
			}
		}
		return nil
	},
}

// ---------------------------------------------------------------------------
// init & main
// ---------------------------------------------------------------------------

func init() {
	rootCmd.PersistentFlags().BoolVar(&debugFlag, "debug", false, "enable debug logging")
	rootCmd.PersistentFlags().BoolVar(&traceFlag, "trace", false, "enable pipeline tracing")
	rootCmd.Version = version
	rootCmd.AddCommand(tokensCmd, astCmd, runCmd, versionCmd, aboutCmd, replCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
