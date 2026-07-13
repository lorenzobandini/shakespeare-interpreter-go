package runtime

import (
	"fmt"
	"io"
	"log/slog"
	"strconv"
	"strings"

	"github.com/lorenzobandini/shakespeare-interpreter-go/internal/parser"
)

type instr struct {
	stmt       parser.Statement
	speaker    string
	sceneRoman string
	actRoman   string
}

func (e *env) execInstr(i instr) (jumpPC int, jumped bool, err error) {
	listener, _ := e.listener(i.speaker)
	_ = listener

	switch s := i.stmt.(type) {
	case parser.EnterStmt:
		_ = e.stage.Enter(s.Characters, e.syms, s.Line, s.Col)
		slog.Debug("stage enter", "chars", s.Characters, "size", e.stage.Size())
		return 0, false, nil

	case parser.ExitStmt:
		_ = e.stage.Exit(s.Character, e.syms, s.Line, s.Col)
		slog.Debug("stage exit", "char", s.Character, "size", e.stage.Size())
		return 0, false, nil

	case parser.ExeuntStmt:
		_ = e.stage.Exeunt(s.Characters, e.syms, s.Line, s.Col)
		slog.Debug("stage exeunt", "chars", s.Characters, "size", e.stage.Size())
		return 0, false, nil

	case parser.AssignStmt:
		listener, ok := e.listener(i.speaker)
		if !ok {
			return 0, false, RuntimeError{Code: "R000", Line: s.Line, Col: s.Col, Msg: "internal: speaker not on stage"}
		}
		v, err := e.eval(s.Expr, i.speaker, listener)
		if err != nil {
			return 0, false, err
		}
		e.values[strings.ToLower(listener)] = v
		slog.Debug("assign", "char", listener, "value", v)
		return 0, false, nil

	case parser.SpeakStmt:
		listener, ok := e.listener(i.speaker)
		if !ok {
			return 0, false, RuntimeError{Code: "R000", Line: s.Line, Col: s.Col, Msg: "internal: speaker not on stage"}
		}
		v := e.values[strings.ToLower(listener)]
		_, err := e.out.Write([]byte{byte(v & 0xFF)})
		if err != nil {
			return 0, false, err
		}
		slog.Debug("speak", "char", listener, "value", v)
		return 0, false, nil

	case parser.OpenHeartStmt:
		listener, ok := e.listener(i.speaker)
		if !ok {
			return 0, false, RuntimeError{Code: "R000", Line: s.Line, Col: s.Col, Msg: "internal: speaker not on stage"}
		}
		v := e.values[strings.ToLower(listener)]
		_, err := fmt.Fprintf(e.out, "%d\n", v)
		if err != nil {
			return 0, false, err
		}
		slog.Debug("openheart", "char", listener, "value", v)
		return 0, false, nil

	case parser.OpenMindStmt:
		listener, ok := e.listener(i.speaker)
		if !ok {
			return 0, false, RuntimeError{Code: "R000", Line: s.Line, Col: s.Col, Msg: "internal: speaker not on stage"}
		}
		var buf [1]byte
		n, err := e.in.Read(buf[:])
		if err == io.EOF || n == 0 {
			return 0, false, errInputEOF(s.Line, s.Col)
		}
		if err != nil {
			return 0, false, errInputEOF(s.Line, s.Col)
		}
		e.values[strings.ToLower(listener)] = int(buf[0])
		slog.Debug("openmind", "char", listener, "value", int(buf[0]))
		return 0, false, nil

	case parser.RememberStmt:
		listener, ok := e.listener(i.speaker)
		if !ok {
			return 0, false, RuntimeError{Code: "R000", Line: s.Line, Col: s.Col, Msg: "internal: speaker not on stage"}
		}
		v, err := e.eval(s.Expr, i.speaker, listener)
		if err != nil {
			return 0, false, err
		}
		key := strings.ToLower(listener)
		e.stacks[key] = append(e.stacks[key], v)
		slog.Debug("remember", "char", listener, "value", v)
		return 0, false, nil

	case parser.RecallStmt:
		listener, ok := e.listener(i.speaker)
		if !ok {
			return 0, false, RuntimeError{Code: "R000", Line: s.Line, Col: s.Col, Msg: "internal: speaker not on stage"}
		}
		stack := e.stacks[strings.ToLower(listener)]
		var v int
		if len(stack) > 0 {
			v = stack[len(stack)-1]
			e.stacks[strings.ToLower(listener)] = stack[:len(stack)-1]
		}
		e.values[strings.ToLower(i.speaker)] = v
		slog.Debug("recall", "from", listener, "to", i.speaker, "value", v)
		return 0, false, nil

	case parser.ListenStmt:
		listener, ok := e.listener(i.speaker)
		if !ok {
			return 0, false, RuntimeError{Code: "R000", Line: s.Line, Col: s.Col, Msg: "internal: speaker not on stage"}
		}
		val, err := e.readInt()
		if err != nil {
			return 0, false, err
		}
		e.values[strings.ToLower(listener)] = val
		slog.Debug("listen", "char", listener, "value", val)
		return 0, false, nil

	case parser.QuestionStmt:
		listener, ok := e.listener(i.speaker)
		if !ok {
			return 0, false, RuntimeError{Code: "R000", Line: s.Line, Col: s.Col, Msg: "internal: speaker not on stage"}
		}
		leftV, err := e.eval(s.Left, i.speaker, listener)
		if err != nil {
			return 0, false, err
		}
		rightV, err := e.eval(s.Right, i.speaker, listener)
		if err != nil {
			return 0, false, err
		}
		e.comparison = applyRelation(s.Comparative.Relation, leftV, rightV)
		slog.Debug("question", "relation", s.Comparative.Relation, "left", leftV, "right", rightV, "result", e.comparison)
		return 0, false, nil

	case parser.IfStmt:
		matched := e.comparison == s.BranchIfTrue
		if !matched {
			return 0, false, nil
		}
		pc, err := e.resolveJump(s.Target, s.TargetKind)
		if err != nil {
			return 0, false, err
		}
		return pc, true, nil

	case parser.GotoStmt:
		pc, err := e.resolveJump(s.Target, s.TargetKind)
		if err != nil {
			return 0, false, err
		}
		return pc, true, nil

	default:
		return 0, false, RuntimeError{
			Code: "R000",
			Msg:  "internal: not yet implemented",
		}
	}
}

func applyRelation(relation string, left, right int) bool {
	switch relation {
	case "equal":
		return left == right
	case "not_equal":
		return left != right
	case "greater":
		return left > right
	case "less":
		return left < right
	case "greater_or_equal":
		return left >= right
	case "less_or_equal":
		return left <= right
	default:
		return false
	}
}

func (e *env) resolveJump(target, kind string) (int, error) {
	lt := strings.ToLower(target)
	switch kind {
	case "scene":
		actKey := strings.ToLower(e.currentActRoman)
		scenes, ok := e.sceneLabels[actKey]
		if !ok {
			return 0, RuntimeError{Code: "R000", Line: 0, Col: 0, Msg: "internal: no scene labels for current act"}
		}
		pc, ok := scenes[lt]
		if !ok {
			return 0, RuntimeError{Code: "R000", Line: 0, Col: 0, Msg: "internal: scene label not found"}
		}
		return pc, nil
	case "act":
		pc, ok := e.actLabel[lt]
		if !ok {
			return 0, RuntimeError{Code: "R000", Line: 0, Col: 0, Msg: "internal: act label not found"}
		}
		return pc, nil
	default:
		return 0, RuntimeError{Code: "R000", Line: 0, Col: 0, Msg: "internal: unknown jump kind"}
	}
}

func (e *env) listener(speaker string) (string, bool) {
	return e.stage.Listener(speaker)
}

func (e *env) readInt() (int, error) {
	var buf [1]byte
	var digits []byte
	sign := 1
	sawSign := false
	sawDigit := false

	// Skip leading whitespace
	for {
		n, err := e.in.Read(buf[:])
		if err == io.EOF || n == 0 {
			if sawDigit {
				break
			}
			if sawSign {
				return 0, errInputNotANumber("", 0, 0)
			}
			return 0, errInputEOF(0, 0)
		}
		if err != nil {
			return 0, errInputEOF(0, 0)
		}
		b := buf[0]
		if b == ' ' || b == '\t' || b == '\r' || b == '\n' {
			if sawDigit || sawSign {
				break
			}
			continue
		}
		if !sawDigit && !sawSign && (b == '+' || b == '-') {
			if b == '-' {
				sign = -1
			}
			sawSign = true
			continue
		}
		if b >= '0' && b <= '9' {
			digits = append(digits, b)
			sawDigit = true
			continue
		}
		// Non-digit, non-whitespace before any digit
		if !sawDigit {
			got := string(b)
			return 0, errInputNotANumber(got, 0, 0)
		}
		break
	}

	if !sawDigit {
		return 0, errInputEOF(0, 0)
	}

	v, err := strconv.Atoi(string(digits))
	if err != nil {
		return 0, RuntimeError{Code: "R000", Msg: fmt.Sprintf("internal: parse number: %v", err)}
	}
	return v * sign, nil
}
