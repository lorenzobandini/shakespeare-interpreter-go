package semantic

import (
	"log/slog"
	"strings"
)

// Stage tracks which characters are currently on stage (0-2).
type Stage struct {
	names []string
}

func (s *Stage) Clear() {
	s.names = nil
}

func (s *Stage) Has(name string) bool {
	lower := strings.ToLower(name)
	for _, n := range s.names {
		if strings.ToLower(n) == lower {
			return true
		}
	}
	return false
}

func (s *Stage) Size() int {
	return len(s.names)
}

func (s *Stage) OnStage() []string {
	snap := make([]string, len(s.names))
	copy(snap, s.names)
	return snap
}

// Enter adds characters to the stage, returning any semantic errors.
// Order: M002 (too many) → M001 (undeclared) → M007 (duplicate/on-stage) → M003 (overflow).
func (s *Stage) Enter(chars []string, syms SymbolTable, line, col int) []SemanticError {
	var errs []SemanticError
	tooMany := len(chars) > 2
	if tooMany {
		errs = append(errs, errTooManyOnStage(len(chars), line, col))
	}

	seen := make(map[string]bool, len(chars))
	for _, c := range chars {
		key := strings.ToLower(c)
		if seen[key] {
			errs = append(errs, errSelfReferenceEnter(c, line, col))
			continue
		}
		seen[key] = true

		if !syms.Has(key) {
			errs = append(errs, errUndefinedCharacter(c, line, col))
			continue
		}

		if s.Has(c) {
			errs = append(errs, errAlreadyOnStage(c, line, col))
			continue
		}
	}

	// Overflow check: only unless M002 already fired (M002 dominates).
	if !tooMany {
		currentSize := s.Size()
		validNew := 0
		for _, c := range chars {
			key := strings.ToLower(c)
			if !s.Has(c) && syms.Has(key) {
				validNew++
			}
		}
		if currentSize+validNew > 2 {
			errs = append(errs, errStageOverflow("", s.OnStage(), line, col))
		}
	}

	if len(errs) > 0 {
		return errs
	}

	for _, c := range chars {
		if syms.Has(strings.ToLower(c)) {
			s.names = append(s.names, c)
		}
	}
	slog.Debug("stage enter", "chars", chars, "size", s.Size())
	return nil
}

// Exit removes a single character from stage.
func (s *Stage) Exit(name string, syms SymbolTable, line, col int) []SemanticError {
	var errs []SemanticError

	key := strings.ToLower(name)
	if !syms.Has(key) {
		errs = append(errs, errUndefinedCharacter(name, line, col))
		return errs
	}

	if !s.Has(name) {
		errs = append(errs, errExitNotOnStage(name, line, col))
		return errs
	}

	for i, n := range s.names {
		if strings.ToLower(n) == key {
			s.names = append(s.names[:i], s.names[i+1:]...)
			break
		}
	}
	slog.Debug("stage exit", "name", name, "size", s.Size())
	return nil
}

// Exeunt removes characters from stage. chars==nil means exit all.
func (s *Stage) Exeunt(chars []string, syms SymbolTable, line, col int) []SemanticError {
	if chars == nil {
		s.Clear()
		slog.Debug("stage exeunt all")
		return nil
	}
	var errs []SemanticError
	for _, c := range chars {
		key := strings.ToLower(c)
		if !syms.Has(key) {
			errs = append(errs, errUndefinedCharacter(c, line, col))
			continue
		}
		if !s.Has(c) {
			errs = append(errs, errExitNotOnStage(c, line, col))
			continue
		}
		for i, n := range s.names {
			if strings.ToLower(n) == key {
				s.names = append(s.names[:i], s.names[i+1:]...)
				break
			}
		}
	}
	if len(errs) > 0 {
		return errs
	}
	slog.Debug("stage exeunt named", "chars", chars, "size", s.Size())
	return nil
}

// Listener returns the other character on stage (if any), or the speaker if alone.
// Returns ok=false if speaker is not on stage (caller already emitted M004).
func (s *Stage) Listener(speaker string) (string, bool) {
	lowerSpeaker := strings.ToLower(speaker)
	var others []string
	for _, n := range s.names {
		if strings.ToLower(n) != lowerSpeaker {
			others = append(others, n)
		}
	}
	if len(others) == 1 {
		return others[0], true
	}
	if len(others) == 0 && len(s.names) > 0 {
		return speaker, true
	}
	return speaker, false
}
