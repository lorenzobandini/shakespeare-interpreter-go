package semantic

import (
	"log/slog"
	"strings"

	"github.com/lorenzobandini/shakespeare-interpreter-go/internal/parser"
)

// CharacterSymbol stores metadata for a declared character.
type CharacterSymbol struct {
	Name, Description string
	Line, Col         int
}

// SymbolTable maps lowercased character names to their symbols.
type SymbolTable struct {
	chars map[string]CharacterSymbol
}

func newSymbolTable(decls []parser.CharacterDecl) (SymbolTable, []SemanticError) {
	chars := make(map[string]CharacterSymbol, len(decls))
	var errs []SemanticError
	for _, d := range decls {
		key := strings.ToLower(d.Name)
		if _, dup := chars[key]; dup {
			errs = append(errs, errUndefinedCharacter(d.Name, d.Line, d.Col))
			continue
		}
		chars[key] = CharacterSymbol{
			Name:        d.Name,
			Description: d.Description,
			Line:        d.Line,
			Col:         d.Col,
		}
		slog.Debug("symbol inserted", "name", d.Name, "line", d.Line)
	}
	return SymbolTable{chars: chars}, errs
}

func (s SymbolTable) Has(lowerName string) bool {
	_, ok := s.chars[lowerName]
	return ok
}

func (s SymbolTable) Get(lowerName string) (CharacterSymbol, bool) {
	sym, ok := s.chars[lowerName]
	return sym, ok
}

// ActRegistry maps lowercased Roman numerals to Acts.
type ActRegistry map[string]*parser.Act

func buildActRegistry(acts []parser.Act) ActRegistry {
	r := make(ActRegistry, len(acts))
	for i := range acts {
		key := strings.ToLower(acts[i].RomanNumeral)
		r[key] = &acts[i]
	}
	return r
}

func (r ActRegistry) Resolve(lowerRoman string) (*parser.Act, bool) {
	a, ok := r[lowerRoman]
	return a, ok
}

// SceneRegistry maps lowercased Roman numerals to Scenes within one act.
type SceneRegistry map[string]*parser.Scene

func buildSceneRegistry(act *parser.Act) SceneRegistry {
	r := make(SceneRegistry, len(act.Scenes))
	for i := range act.Scenes {
		key := strings.ToLower(act.Scenes[i].RomanNumeral)
		r[key] = &act.Scenes[i]
	}
	return r
}

func (r SceneRegistry) Resolve(lowerRoman string) (*parser.Scene, bool) {
	s, ok := r[lowerRoman]
	return s, ok
}
