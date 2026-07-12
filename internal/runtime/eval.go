package runtime

import (
	"fmt"
	"log/slog"
	"math"
	"strings"

	"github.com/lorenzobandini/shakespeare-interpreter-go/internal/parser"
)

func (e *env) eval(expr parser.Expr, speaker, listener string) (int, error) {
	switch ex := expr.(type) {
	case parser.ConstExpr:
		value := ex.Polarity * (1 << uint(ex.AdjectiveCount))
		slog.Debug("const", "noun", ex.Noun, "value", value)
		return value, nil

	case parser.CharRefExpr:
		key := strings.ToLower(ex.Name)
		v, ok := e.values[key]
		if !ok {
			return 0, RuntimeError{
				Code: "R000", Line: ex.Line, Col: ex.Col,
				Msg: fmt.Sprintf("internal: character '%s' not in value store", ex.Name),
			}
		}
		return v, nil

	case parser.PronounExpr:
		switch ex.Ref {
		case "speaker":
			return e.values[strings.ToLower(speaker)], nil
		case "listener":
			return e.values[strings.ToLower(listener)], nil
		default:
			return 0, RuntimeError{
				Code: "R000", Line: ex.Line, Col: ex.Col,
				Msg: fmt.Sprintf("internal: unknown pronoun ref '%s'", ex.Ref),
			}
		}

	case parser.BinaryOpExpr:
		left, err := e.eval(ex.Left, speaker, listener)
		if err != nil {
			return 0, err
		}
		right, err := e.eval(ex.Right, speaker, listener)
		if err != nil {
			return 0, err
		}
		switch ex.Op {
		case "sum":
			return left + right, nil
		case "difference":
			return left - right, nil
		case "product":
			return left * right, nil
		case "quotient":
			if right == 0 {
				return 0, errDivisionByZero(e.currentActRoman, e.currentSceneRoman, ex.Line, ex.Col)
			}
			return left / right, nil
		case "remainder":
			if right == 0 {
				return 0, errDivisionByZero(e.currentActRoman, e.currentSceneRoman, ex.Line, ex.Col)
			}
			return left % right, nil
		default:
			return 0, RuntimeError{
				Code: "R000", Line: ex.Line, Col: ex.Col,
				Msg: fmt.Sprintf("internal: unknown binary op '%s'", ex.Op),
			}
		}

	case parser.UnaryOpExpr:
		v, err := e.eval(ex.Operand, speaker, listener)
		if err != nil {
			return 0, err
		}
		switch ex.Op {
		case "square":
			return v * v, nil
		case "cube":
			return v * v * v, nil
		case "square_root":
			if v < 0 {
				return 0, RuntimeError{
					Code: "R000", Line: ex.Line, Col: ex.Col,
					Msg: "internal: square root of negative number",
				}
			}
			return int(math.Sqrt(float64(v))), nil
		case "factorial":
			if v < 0 {
				return 0, RuntimeError{
					Code: "R000", Line: ex.Line, Col: ex.Col,
					Msg: "internal: factorial of negative number",
				}
			}
			if v > 20 {
				return 0, errIntegerOverflow("factorial", ex.Line, ex.Col)
			}
			result := 1
			for i := 2; i <= v; i++ {
				result *= i
			}
			return result, nil
		case "twice":
			return v * 2, nil
		default:
			return 0, RuntimeError{
				Code: "R000", Line: ex.Line, Col: ex.Col,
				Msg: fmt.Sprintf("internal: unknown unary op '%s'", ex.Op),
			}
		}

	default:
		return 0, RuntimeError{
			Code: "R000",
			Msg:  fmt.Sprintf("internal: unsupported expr %T", expr),
		}
	}
}
