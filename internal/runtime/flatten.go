package runtime

import (
	"strings"

	"github.com/lorenzobandini/shakespeare-interpreter-go/internal/parser"
)

func (e *env) flatten(prog *parser.Program) {
	e.instrs = nil
	e.actLabel = make(map[string]int)
	e.sceneLabels = make(map[string]map[string]int)

	for _, act := range prog.Acts {
		actRoman := strings.ToLower(act.RomanNumeral)
		if _, exists := e.actLabel[actRoman]; !exists {
			e.actLabel[actRoman] = len(e.instrs)
		}

		scenes := make(map[string]int)
		for _, scene := range act.Scenes {
			sceneRoman := strings.ToLower(scene.RomanNumeral)
			scenes[sceneRoman] = len(e.instrs)

			for _, stmt := range scene.Statements {
				switch s := stmt.(type) {
				case parser.EnterStmt, parser.ExitStmt, parser.ExeuntStmt:
					e.instrs = append(e.instrs, instr{
						stmt:       stmt,
						speaker:    "",
						sceneRoman: scene.RomanNumeral,
						actRoman:   act.RomanNumeral,
					})
				case parser.Dialogue:
					for _, inner := range s.Statements {
						e.instrs = append(e.instrs, instr{
							stmt:       inner,
							speaker:    s.Speaker,
							sceneRoman: scene.RomanNumeral,
							actRoman:   act.RomanNumeral,
						})
					}
				}
			}
		}
		e.sceneLabels[actRoman] = scenes
	}
}

func (e *env) runLoop() error {
	pc := 0
	for pc < len(e.instrs) {
		i := e.instrs[pc]

		if i.actRoman != e.currentActRoman {
			e.currentActRoman = i.actRoman
		}
		if i.sceneRoman != e.currentSceneRoman {
			e.currentSceneRoman = i.sceneRoman
		}

		next, jumped, err := e.execInstr(i)
		if err != nil {
			return err
		}
		if jumped {
			pc = next
		} else {
			pc++
		}
	}
	return nil
}
