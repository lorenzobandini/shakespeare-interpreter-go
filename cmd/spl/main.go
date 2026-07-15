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
	debugFlag    bool
	traceFlag    bool
	maxStepsFlag int
)

// Build info — set via ldflags.
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

// rootCmd is the base command.
var rootCmd = &cobra.Command{
	Use:   "spl",
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
			"spl %s\ncommit: %s\ndate:   %s\n", version, commit, date)
		return err
	},
}

// ---------------------------------------------------------------------------
// about subcommand
// ---------------------------------------------------------------------------

const aboutASCII = `
 .d8888b.  8888888b.  888      
d88P  Y88b 888   Y88b 888      
Y88b.      888    888 888      
 "Y888b.   888   d88P 888      
    "Y88b. 8888888P"  888      
      "888 888        888      
Y88b  d88P 888        888      
 "Y8888P"  888        88888888 
`

const aboutText = "Shakespeare Programming Language (SPL)\n" +
	"Original language design: Karl Hasselström and Jon Åslund (2001)\n" +
	"\n" +
	"Reference implementation: zmbc/shakespearelang\n" +
	"This implementation: https://github.com/lorenzobandini/shakespeare-interpreter-go\n" +
	"\n" +
	"Go version: 1.26.5"

var aboutCmd = &cobra.Command{
	Use:   "about",
	Short: "About the Shakespeare Programming Language",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		_, err := fmt.Fprintf(cmd.OutOrStdout(), "%s\n%s\n", aboutASCII, aboutText)
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
		return runPipeline(string(src), cmd.InOrStdin(), cmd.OutOrStdout(), args[0], cmd.ErrOrStderr(), maxStepsFlag)
	},
}

// ---------------------------------------------------------------------------
// Shared pipeline runner (used by runCmd and the REPL)
// ---------------------------------------------------------------------------

func runPipeline(src string, in io.Reader, out io.Writer, filename string, traceOut io.Writer, maxSteps int) error {
	if traceFlag {
		_, _ = fmt.Fprintln(traceOut, "--- TOKENS ---")
	}
	tokens, err := lexer.New(src).ScanTokens()
	if err != nil {
		return fmt.Errorf("%w", err)
	}
	if traceFlag {
		_, _ = fmt.Fprintln(traceOut, "--- AST ---")
	}
	prog, err := parser.New(tokens).Parse()
	if err != nil {
		return fmt.Errorf("%w", err)
	}
	if traceFlag {
		_, _ = fmt.Fprintln(traceOut, "--- SEMANTIC ---")
	}
	res := semantic.New(filename, prog).Analyze()
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
	if err := runtime.ExecuteWithLimit(prog, res, in, out, filename, maxSteps); err != nil {
		return fmt.Errorf("%w", err)
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

// Phases for the REPL state machine.
type replPhase uint8

const (
	phaseTitle replPhase = iota
	phaseChars
	phaseBody
	phaseClosed
)

// Classification of individual lines for phase transitions.
type replLineKind uint8

const (
	lineBlank replLineKind = iota
	lineTitle
	lineCharDecl
	lineActHeader
	lineSceneHeader
	lineStageDir
	lineDialogue
	lineOther
)

// classifyLine classifies a single line of SPL text using cheap
// substring checks (no full parsing).
func classifyLine(line string) replLineKind {
	trimmed := strings.TrimSpace(line)
	if trimmed == "" {
		return lineBlank
	}
	lower := strings.ToLower(trimmed)
	if strings.HasPrefix(lower, "act ") {
		return lineActHeader
	}
	if strings.HasPrefix(lower, "scene ") {
		return lineSceneHeader
	}
	if trimmed[0] == '[' && trimmed[len(trimmed)-1] == ']' {
		return lineStageDir
	}
	if colIdx := strings.Index(trimmed, ":"); colIdx > 0 && colIdx < 60 {
		before := trimmed[:colIdx]
		if !strings.Contains(before, " ") && !strings.Contains(before, ",") {
			return lineDialogue
		}
	}
	if commaIdx := strings.Index(trimmed, ","); commaIdx > 0 && commaIdx < 60 {
		before := trimmed[:commaIdx]
		if !strings.Contains(before, " ") && !strings.Contains(before, ":") {
			return lineCharDecl
		}
	}
	if strings.HasSuffix(trimmed, ".") || strings.HasSuffix(trimmed, "!") || strings.HasSuffix(trimmed, "?") {
		return lineTitle
	}
	return lineOther
}

// derivePhase determines the REPL phase from accumulated text.
func derivePhase(text string) replPhase {
	p := phaseTitle
	for _, line := range strings.Split(text, "\n") {
		switch classifyLine(line) {
		case lineCharDecl:
			if p == phaseTitle {
				p = phaseChars
			}
		case lineActHeader:
			p = phaseBody
		}
	}
	return p
}

// hasBodyContent returns true if the text contains stage directions
// or dialogue lines (i.e., needs a program structure around it).
func hasBodyContent(text string) bool {
	for _, line := range strings.Split(text, "\n") {
		switch classifyLine(line) {
		case lineStageDir, lineDialogue:
			return true
		}
	}
	return false
}

// extractDeclaredNames returns character names from explicit declarations
// in text (lines matching the "Name, ..." pattern).
func extractDeclaredNames(text string) []string {
	var names []string
	for _, line := range strings.Split(text, "\n") {
		if classifyLine(line) == lineCharDecl {
			if commaIdx := strings.Index(line, ","); commaIdx > 0 {
				names = append(names, strings.TrimSpace(line[:commaIdx]))
			}
		}
	}
	return names
}

// hasTitle returns true if the text contains a title-like line.
func hasTitle(text string) bool {
	for _, line := range strings.Split(text, "\n") {
		if classifyLine(line) == lineTitle {
			return true
		}
	}
	return false
}

// replState holds all mutable state across REPL turns.
//
// Invariant: re-execution of rs.buffer[:oldLen] must reproduce
// captureOut[:rs.lastOutputLen] byte-for-byte. This holds because:
// (i) the lexer/parser/semantic are pure functions of source text,
// (ii) the runtime is deterministic given source + recorded stdin,
// (iii) the buffer only grows on the success path.
type replState struct {
	out           io.Writer
	err           io.Writer
	buffer        bytes.Buffer
	declared      map[string]bool
	skeletonBuilt bool
	phase         replPhase

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

// hasActOrScene reports whether the text contains "act" or "scene" at the
// start of a line (outside of stage directions), indicating the user is
// providing their own program structure.
func hasActOrScene(text string) bool {
	for _, line := range strings.Split(text, "\n") {
		trimmed := strings.TrimSpace(line)
		lower := strings.ToLower(trimmed)
		if strings.HasPrefix(lower, "act ") || strings.HasPrefix(lower, "scene ") {
			return true
		}
	}
	return false
}

// replayBlock appends text, auto-declares, runs the pipeline, handles output
// slicing and rollback.
func (rs *replState) replayBlock(input string, sr *singleReader) error {
	bufCheck := rs.buffer.Len()
	recCheck := sr.recordCheckpoint()
	prevSkeletonBuilt := rs.skeletonBuilt

	fullText := rs.buffer.String() + input + "\n"

	// ---- Phase-gated skeleton injection (Step 5.2) ----
	// Inject skeleton once, on the first submit where input has body-level
	// content but no user-provided act/scene structure.
	// Order: title (skeleton if missing) → existing buffer → Act I/Scene I
	if !rs.skeletonBuilt && !hasActOrScene(fullText) && hasBodyContent(input) {
		var titlePart string
		if !hasTitle(rs.buffer.String()) {
			titlePart = "The REPL Session.\n\n"
		}
		actPart := "Act I: The REPL Session.\nScene I: The REPL Session.\n\n"
		existing := rs.buffer.String()
		rs.buffer.Reset()
		rs.buffer.WriteString(titlePart)
		rs.buffer.WriteString(existing)
		rs.buffer.WriteString(actPart)
		rs.skeletonBuilt = true
	}

	// ---- Auto-declaration splice (Step 5.3) ----
	newChars := extractChars(input)
	var newlyDeclared []string
	if len(newChars) > 0 && rs.skeletonBuilt {
		// Populate declared from existing char decls in the buffer to
		// avoid duplicating user-provided declarations.
		for _, name := range extractDeclaredNames(rs.buffer.String()) {
			rs.declared[strings.ToLower(name)] = true
		}

		var decls strings.Builder
		for _, name := range newChars {
			key := strings.ToLower(name)
			if !rs.declared[key] {
				rs.declared[key] = true
				newlyDeclared = append(newlyDeclared, key)
				_, _ = fmt.Fprintf(&decls, "%s, a REPL character.\n", name)
			}
		}
		if decls.Len() > 0 {
			declText := decls.String()
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

	// Append input to buffer.
	rs.buffer.WriteString(input)
	rs.buffer.WriteString("\n")

	// Update phase.
	newPhase := derivePhase(rs.buffer.String())
	prevPhase := rs.phase
	rs.phase = newPhase

	// Pre-body submits accumulate in the buffer without validation.
	// Only run the pipeline once we reach body phase.
	if newPhase < phaseBody {
		return nil
	}

	// Start replay mode so the pipeline sees recorded bytes first, then
	// fresh stdin.
	sr.replayMode()

	// Run the full pipeline on captured output.
	var captureOut bytes.Buffer
	src := rs.buffer.String()
	err := runPipeline(src, sr, &captureOut, "repl", rs.traceOut, 0)
	if err != nil {
		// Rollback buffer, recorded inputs, skeleton, and phase.
		rs.buffer.Truncate(bufCheck)
		sr.rollbackRecorded(recCheck)
		for _, key := range newlyDeclared {
			delete(rs.declared, key)
		}
		rs.skeletonBuilt = prevSkeletonBuilt
		rs.phase = prevPhase
		return fmt.Errorf("error: %v", err)
	}

	// Show output delta.
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
// input — unterminated brackets, unclosed expressions — rather than
// containing a genuine syntax error.
var ErrIncomplete = fmt.Errorf("incomplete input")

// tryQuickParse runs a quick heuristic check on the pending block to decide
// whether the REPL should keep accumulating (return ErrIncomplete) or report
// the error immediately (any other non-nil error). Returns nil if the block
// looks complete enough to submit.
//
// Criteria for ErrIncomplete:
//  1. L002 (unterminated bracket) at EOF.
//  2. Bracket token imbalance (more `[` than `]`).
//  3. Open multi-line expression: the last non-blank line lacks a terminating
//     . / ! / ?, and the line count since the last terminator is ≤ 40 (guard
//     against unbounded growth).
//
// All other errors (S-codes, M-codes, etc.) are treated as genuine and
// reported immediately — the session continues at the next prompt.
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
	// 2. Bracket token imbalance.
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

	// 3. Open multi-line expression: block doesn't end with a sentence
	//    terminator and hasn't grown beyond the heuristic cap.
	if isOpenExpression(block) {
		return ErrIncomplete
	}
	return nil
}

// isOpenExpression returns true when the block looks like an incomplete
// sentence: the last non-blank line does not end with . / ! / ?, and the
// line count since the last terminator is ≤ 40.
func isOpenExpression(text string) bool {
	lines := strings.Split(text, "\n")
	lastContent := ""
	for i := len(lines) - 1; i >= 0; i-- {
		if trimmed := strings.TrimSpace(lines[i]); trimmed != "" {
			lastContent = trimmed
			break
		}
	}
	if lastContent == "" {
		return false
	}
	switch lastContent[len(lastContent)-1] {
	case '.', '!', '?':
		return false
	}
	// Count non-blank lines since the last terminator.
	count := 0
	for i := len(lines) - 1; i >= 0; i-- {
		trimmed := strings.TrimSpace(lines[i])
		if trimmed == "" {
			continue
		}
		count++
		if last := trimmed[len(trimmed)-1]; last == '.' || last == '!' || last == '?' {
			return false // found a terminator on a previous line
		}
	}
	return count <= 40
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
		defer signal.Stop(sigCh)

		var block bytes.Buffer

		for {
			select {
			case <-sigCh:
				_, _ = fmt.Fprintln(rs.err, "^C")
				return nil
			default:
			}
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
					rs.phase = phaseTitle
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
	rootCmd.PersistentFlags().IntVar(&maxStepsFlag, "max-steps", 1000000, "maximum execution steps (0 = unlimited)")
	rootCmd.Version = version
	rootCmd.AddCommand(tokensCmd, astCmd, runCmd, versionCmd, aboutCmd, replCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
