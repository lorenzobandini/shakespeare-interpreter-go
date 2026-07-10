package parser

import (
	"log/slog"
	"strings"

	"github.com/lorenzobandini/shakespeare-interpreter-go/internal/lexer"
)

// Parser converts a token stream into a typed AST.
type Parser struct {
	tokens     []lexer.Token
	pos        int
	characters map[string]bool
	warnings   []Warning
}

// New creates a Parser for the given token slice.
func New(tokens []lexer.Token) *Parser {
	return &Parser{
		tokens:     tokens,
		pos:        0,
		characters: make(map[string]bool),
	}
}

// Warnings returns collected advisory warnings (S003).
func (p *Parser) Warnings() []Warning {
	return p.warnings
}

func (p *Parser) addWarning(w Warning) {
	p.warnings = append(p.warnings, w)
}

// Cursor helpers.

func (p *Parser) peek() lexer.Token {
	if p.pos >= len(p.tokens) {
		return lexer.Token{Type: lexer.TokenEOF}
	}
	return p.tokens[p.pos]
}

func (p *Parser) peekAt(offset int) lexer.Token {
	if p.pos+offset >= len(p.tokens) {
		return lexer.Token{Type: lexer.TokenEOF}
	}
	return p.tokens[p.pos+offset]
}

func (p *Parser) advance() lexer.Token {
	t := p.peek()
	if t.Type != lexer.TokenEOF {
		p.pos++
	}
	return t
}

func (p *Parser) at(t lexer.TokenType) bool { return p.peek().Type == t }

func (p *Parser) match(t lexer.TokenType) bool {
	if p.at(t) {
		p.advance()
		return true
	}
	return false
}

func (p *Parser) checkWord(value string) bool {
	return p.peek().Type == lexer.TokenWord && lower(p.peek().Lexeme) == value
}

func (p *Parser) matchWord(value string) bool {
	if p.checkWord(value) {
		p.advance()
		return true
	}
	return false
}

func (p *Parser) skipNewlines() {
	for p.at(lexer.TokenNewline) {
		p.advance()
	}
}

func (p *Parser) isEOF() bool { return p.at(lexer.TokenEOF) }

// currentLineCol returns the line/col of the current token (or last token if EOF).
func (p *Parser) currentLineCol() (int, int) {
	t := p.peek()
	if t.Type == lexer.TokenEOF && len(p.tokens) > 0 {
		return p.tokens[len(p.tokens)-1].Line, p.tokens[len(p.tokens)-1].Col
	}
	return t.Line, t.Col
}

// Parse is the top-level entry point.
func (p *Parser) Parse() (*Program, error) {
	slog.Debug("parsing program")
	p.skipNewlines()
	title, err := p.parseTitle()
	if err != nil {
		return nil, err
	}
	p.skipNewlines()
	chars, err := p.parseCharacterDecls()
	if err != nil {
		return nil, err
	}
	p.skipNewlines()
	acts, err := p.parseActs()
	if err != nil {
		return nil, err
	}
	line, col := p.currentLineCol()
	return &Program{
		Title:      title,
		Characters: chars,
		Acts:       acts,
		Warnings:   p.warnings,
		Line:       line,
		Col:        col,
	}, nil
}

// parseTitle: first line ending with '.'. Everything before the '.' is the title text.
func (p *Parser) parseTitle() (Title, error) {
	if p.isEOF() || p.at(lexer.TokenNewline) {
		line, col := p.currentLineCol()
		return Title{}, errMissingTitle(line, col)
	}
	if p.peek().Type != lexer.TokenWord {
		line, col := p.currentLineCol()
		return Title{}, errMissingTitle(line, col)
	}
	firstLine := p.peek().Line
	firstCol := p.peek().Col
	var parts []string
	for !p.isEOF() && !p.at(lexer.TokenNewline) && !p.at(lexer.TokenPeriod) {
		tok := p.advance()
		parts = append(parts, tok.Lexeme)
	}
	if !p.match(lexer.TokenPeriod) {
		return Title{}, errMissingTitle(firstLine, firstCol)
	}
	return Title{Text: strings.Join(parts, " "), Line: firstLine, Col: firstCol}, nil
}

// parseCharacterDecls: loop while next is a WORD and not "act" keyword.
func (p *Parser) parseCharacterDecls() ([]CharacterDecl, error) {
	var decls []CharacterDecl
	for p.peek().Type == lexer.TokenWord && !p.checkWord("act") {
		firstLine := p.peek().Line
		firstCol := p.peek().Col
		name := p.advance().Lexeme
		if !p.match(lexer.TokenComma) {
			return nil, errMissingCharacterDecl(firstLine, firstCol)
		}
		var descParts []string
		for !p.isEOF() && !p.at(lexer.TokenNewline) && !p.at(lexer.TokenPeriod) {
			descParts = append(descParts, p.advance().Lexeme)
		}
		if len(descParts) == 0 || !p.match(lexer.TokenPeriod) {
			return nil, errMissingCharacterDecl(firstLine, firstCol)
		}
		if !isShakespeareCharacter(name) {
			p.addWarning(warnInvalidCharacterName(name, firstLine, firstCol))
		}
		decls = append(decls, CharacterDecl{
			Name:        name,
			Description: strings.Join(descParts, " "),
			Line:        firstLine,
			Col:         firstCol,
		})
		p.characters[lower(name)] = true
		p.skipNewlines()
	}
	if len(decls) == 0 {
		line, col := p.currentLineCol()
		return nil, errMissingCharacterDecl(line, col)
	}
	return decls, nil
}

// parseActs: loop while next is "act" keyword. Sequential numbering enforced.
// Stage state (enterSeen) persists across scenes within an act.
func (p *Parser) parseActs() ([]Act, error) {
	var acts []Act
	expected := 1
	for p.checkWord("act") {
		firstLine := p.peek().Line
		firstCol := p.peek().Col
		p.advance() // consume "act"
		if p.peek().Type != lexer.TokenWord {
			return nil, errInvalidActNumber("", firstLine, firstCol)
		}
		roman := p.advance().Lexeme
		num, ok := parseRoman(roman)
		if !ok {
			return nil, errInvalidActNumber(roman, firstLine, firstCol)
		}
		if num != expected {
			expectedRoman := intToRoman(expected)
			return nil, errActOrder(expectedRoman, roman, firstLine, firstCol)
		}
		expected++
		if !p.match(lexer.TokenColon) {
			if !p.match(lexer.TokenPeriod) {
				return nil, errInvalidActNumber(roman, firstLine, firstCol)
			}
			p.skipNewlines()
			scenes, enterSeen, err := p.parseScenes(roman, false)
			if err != nil {
				return nil, err
			}
			if len(scenes) == 0 {
				return nil, errMissingScene(roman, firstLine, firstCol)
			}
			_ = enterSeen
			acts = append(acts, Act{Number: num, RomanNumeral: roman, Scenes: scenes, Line: firstLine, Col: firstCol})
			continue
		}
		var descParts []string
		for !p.isEOF() && !p.at(lexer.TokenNewline) && !p.at(lexer.TokenPeriod) {
			descParts = append(descParts, p.advance().Lexeme)
		}
		if !p.match(lexer.TokenPeriod) {
			return nil, errInvalidActNumber(roman, firstLine, firstCol)
		}
		p.skipNewlines()
		scenes, _, err := p.parseScenes(roman, false)
		if err != nil {
			return nil, err
		}
		if len(scenes) == 0 {
			return nil, errMissingScene(roman, firstLine, firstCol)
		}
		acts = append(acts, Act{
			Number:       num,
			RomanNumeral: roman,
			Description:  strings.Join(descParts, " "),
			Scenes:       scenes,
			Line:         firstLine,
			Col:          firstCol,
		})
	}
	if len(acts) == 0 {
		line, col := p.currentLineCol()
		return nil, errMissingAct(line, col)
	}
	return acts, nil
}

// parseScenes: loop while next is "scene" keyword. enterSeen persists across scenes.
func (p *Parser) parseScenes(actRoman string, enterSeen bool) ([]Scene, bool, error) {
	var scenes []Scene
	expected := 1
	for p.checkWord("scene") {
		firstLine := p.peek().Line
		firstCol := p.peek().Col
		p.advance() // consume "scene"
		if p.peek().Type != lexer.TokenWord {
			return nil, enterSeen, errInvalidSceneNumber("", firstLine, firstCol)
		}
		roman := p.advance().Lexeme
		num, ok := parseRoman(roman)
		if !ok {
			return nil, enterSeen, errInvalidSceneNumber(roman, firstLine, firstCol)
		}
		if num != expected {
			expectedRoman := intToRoman(expected)
			return nil, enterSeen, errSceneOrder(expectedRoman, roman, firstLine, firstCol)
		}
		expected++
		desc := ""
		if p.match(lexer.TokenColon) {
			var descParts []string
			for !p.isEOF() && !p.at(lexer.TokenNewline) && !p.at(lexer.TokenPeriod) {
				descParts = append(descParts, p.advance().Lexeme)
			}
			desc = strings.Join(descParts, " ")
		}
		if !p.match(lexer.TokenPeriod) {
			return nil, enterSeen, errInvalidSceneNumber(roman, firstLine, firstCol)
		}
		p.skipNewlines()
		stmts, newEnterSeen, err := p.parseStatements(enterSeen)
		if err != nil {
			return nil, enterSeen, err
		}
		enterSeen = newEnterSeen
		scenes = append(scenes, Scene{
			Number:       num,
			RomanNumeral: roman,
			Description:  desc,
			Statements:   stmts,
			Line:         firstLine,
			Col:          firstCol,
		})
	}
	return scenes, enterSeen, nil
}

// parseStatements: collect stage directions and dialogue blocks. enterSeen persists.
func (p *Parser) parseStatements(enterSeen bool) ([]Statement, bool, error) {
	var stmts []Statement
	for {
		p.skipNewlines()
		if p.isEOF() {
			break
		}
		if p.at(lexer.TokenLBracket) {
			stmt, err := p.parseStageDirection()
			if err != nil {
				return nil, enterSeen, err
			}
			if _, ok := stmt.(EnterStmt); ok {
				enterSeen = true
			}
			stmts = append(stmts, stmt)
			continue
		}
		if p.peek().Type == lexer.TokenWord && p.peekAt(1).Type == lexer.TokenColon {
			if !enterSeen {
				return nil, false, errMissingStage(p.peek().Line, p.peek().Col)
			}
			dlg, err := p.parseDialogue()
			if err != nil {
				return nil, enterSeen, err
			}
			stmts = append(stmts, dlg)
			continue
		}
		if p.checkWord("act") || p.checkWord("scene") {
			break
		}
		break
	}
	return stmts, enterSeen, nil
}

// parseStageDirection: [Enter ...], [Exit ...], [Exeunt ...]
func (p *Parser) parseStageDirection() (Statement, error) {
	firstLine := p.peek().Line
	firstCol := p.peek().Col
	p.advance() // consume [
	if p.peek().Type != lexer.TokenWord {
		return nil, errInvalidEnter(firstLine, firstCol)
	}
	keyword := lower(p.advance().Lexeme)
	switch keyword {
	case "enter":
		return p.parseEnter(firstLine, firstCol)
	case "exit":
		return p.parseExit(firstLine, firstCol)
	case "exeunt":
		return p.parseExeunt(firstLine, firstCol)
	default:
		return nil, errInvalidEnter(firstLine, firstCol)
	}
}

func (p *Parser) parseEnter(line, col int) (Statement, error) {
	var chars []string
	if p.peek().Type != lexer.TokenWord {
		return nil, errInvalidEnter(line, col)
	}
	chars = append(chars, p.advance().Lexeme)
	if p.matchWord("and") {
		if p.peek().Type != lexer.TokenWord {
			return nil, errInvalidEnter(line, col)
		}
		chars = append(chars, p.advance().Lexeme)
	}
	if !p.match(lexer.TokenRBracket) {
		return nil, errInvalidEnter(line, col)
	}
	p.skipNewlines()
	return EnterStmt{Characters: chars, Line: line, Col: col}, nil
}

func (p *Parser) parseExit(line, col int) (Statement, error) {
	if p.peek().Type != lexer.TokenWord {
		return nil, errInvalidEnter(line, col)
	}
	name := p.advance().Lexeme
	if p.matchWord("and") {
		return nil, errInvalidEnter(line, col)
	}
	if !p.match(lexer.TokenRBracket) {
		return nil, errInvalidEnter(line, col)
	}
	p.skipNewlines()
	return ExitStmt{Character: name, Line: line, Col: col}, nil
}

func (p *Parser) parseExeunt(line, col int) (Statement, error) {
	var chars []string
	if p.match(lexer.TokenRBracket) {
		p.skipNewlines()
		return ExeuntStmt{Characters: nil, Line: line, Col: col}, nil
	}
	if p.peek().Type != lexer.TokenWord {
		return nil, errInvalidExeunt(line, col)
	}
	chars = append(chars, p.advance().Lexeme)
	for p.matchWord("and") {
		if p.peek().Type != lexer.TokenWord {
			return nil, errInvalidExeunt(line, col)
		}
		chars = append(chars, p.advance().Lexeme)
	}
	if !p.match(lexer.TokenRBracket) {
		return nil, errInvalidExeunt(line, col)
	}
	p.skipNewlines()
	return ExeuntStmt{Characters: chars, Line: line, Col: col}, nil
}

// parseDialogue: Speaker: <statements...>
func (p *Parser) parseDialogue() (Statement, error) {
	firstLine := p.peek().Line
	firstCol := p.peek().Col
	speaker := p.advance().Lexeme
	p.advance() // consume ':'
	p.skipNewlines()
	var stmts []Statement
	for !p.isEOF() {
		if p.at(lexer.TokenLBracket) {
			break
		}
		if p.peek().Type == lexer.TokenWord && p.peekAt(1).Type == lexer.TokenColon {
			break
		}
		if p.checkWord("act") || p.checkWord("scene") {
			break
		}
		stmt, err := p.parseStatement()
		if err != nil {
			return nil, err
		}
		stmts = append(stmts, stmt)
		p.skipNewlines()
	}
	return Dialogue{Speaker: speaker, Statements: stmts, Line: firstLine, Col: firstCol}, nil
}

// parseStatement: dispatch by leading keyword.
func (p *Parser) parseStatement() (Statement, error) {
	if p.peek().Type != lexer.TokenWord {
		line, col := p.currentLineCol()
		return nil, errMissingSpeaker(line, col)
	}
	word := lower(p.peek().Lexeme)
	switch word {
	case "you", "thou":
		return p.parseAssignStmt()
	case "speak":
		return p.parseSpeakStmt()
	case "open":
		return p.parseOpenStmt()
	case "listen":
		return p.parseListenStmt()
	case "remember":
		return p.parseRememberStmt()
	case "recall":
		return p.parseRecallStmt()
	case "am", "is":
		return p.parseQuestionStmt()
	case "if":
		return p.parseIfStmt()
	case "let":
		return p.parseGotoStmt()
	}
	line, col := p.currentLineCol()
	return nil, errMissingSpeaker(line, col)
}

// parseAssignStmt: "You/Thou are/art [as ADJ as] <expr>." or "You/Thou <constant>!".
func (p *Parser) parseAssignStmt() (Statement, error) {
	firstLine := p.peek().Line
	firstCol := p.peek().Col
	p.advance() // consume You/Thou
	target := "you"
	if p.matchWord("are") || p.matchWord("art") {
		simileAdj := ""
		if p.checkWord("as") {
			p.advance() // consume "as"
			p.skipNewlines()
			if p.peek().Type != lexer.TokenWord {
				return nil, errInvalidExpression(firstLine, firstCol)
			}
			simileAdj = p.advance().Lexeme
			p.skipNewlines()
			if !p.matchWord("as") {
				return nil, errInvalidComparative("", firstLine, firstCol)
			}
		}
		expr, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		term, err := p.expectTerminator()
		if err != nil {
			return nil, err
		}
		return AssignStmt{Target: target, SimileAdj: simileAdj, Expr: expr, Terminator: term, Line: firstLine, Col: firstCol}, nil
	}
	// No-copula form: "You <constant>!" or "You <constant>."
	expr, err := p.parseConstant()
	if err != nil {
		return nil, err
	}
	term, err := p.expectTerminator()
	if err != nil {
		return nil, err
	}
	return AssignStmt{Target: target, SimileAdj: "", Expr: expr, Terminator: term, Line: firstLine, Col: firstCol}, nil
}

func (p *Parser) expectTerminator() (string, error) {
	line, col := p.currentLineCol()
	if p.match(lexer.TokenPeriod) {
		return ".", nil
	}
	if p.match(lexer.TokenBang) {
		return "!", nil
	}
	return "", errInvalidExpression(line, col)
}

// parseSpeakStmt: "Speak your/thy mind [.!?]"
func (p *Parser) parseSpeakStmt() (Statement, error) {
	firstLine := p.peek().Line
	firstCol := p.peek().Col
	p.advance() // consume "speak"
	if !p.matchWord("your") && !p.matchWord("thy") {
		return nil, errInvalidExpression(firstLine, firstCol)
	}
	if !p.matchWord("mind") {
		return nil, errInvalidExpression(firstLine, firstCol)
	}
	term, err := p.expectTerminator()
	if err != nil {
		return nil, err
	}
	return SpeakStmt{Terminator: term, Line: firstLine, Col: firstCol}, nil
}

// parseOpenStmt: "Open your/thy heart" (output number) or "Open your/thy mind" (input char).
func (p *Parser) parseOpenStmt() (Statement, error) {
	firstLine := p.peek().Line
	firstCol := p.peek().Col
	p.advance() // consume "open"
	if !p.matchWord("your") && !p.matchWord("thy") {
		return nil, errInvalidExpression(firstLine, firstCol)
	}
	if p.matchWord("heart") {
		term, err := p.expectTerminator()
		if err != nil {
			return nil, err
		}
		return OpenHeartStmt{Terminator: term, Line: firstLine, Col: firstCol}, nil
	}
	if p.matchWord("mind") {
		term, err := p.expectTerminator()
		if err != nil {
			return nil, err
		}
		return OpenMindStmt{Terminator: term, Line: firstLine, Col: firstCol}, nil
	}
	return nil, errInvalidExpression(firstLine, firstCol)
}

// parseListenStmt: "Listen to your/thy heart [.!?]"
func (p *Parser) parseListenStmt() (Statement, error) {
	firstLine := p.peek().Line
	firstCol := p.peek().Col
	p.advance() // consume "listen"
	if !p.matchWord("to") {
		return nil, errInvalidExpression(firstLine, firstCol)
	}
	if !p.matchWord("your") && !p.matchWord("thy") {
		return nil, errInvalidExpression(firstLine, firstCol)
	}
	if !p.matchWord("heart") {
		return nil, errInvalidExpression(firstLine, firstCol)
	}
	term, err := p.expectTerminator()
	if err != nil {
		return nil, err
	}
	return ListenStmt{Terminator: term, Line: firstLine, Col: firstCol}, nil
}

// parseRememberStmt: "Remember <expr>."
func (p *Parser) parseRememberStmt() (Statement, error) {
	firstLine := p.peek().Line
	firstCol := p.peek().Col
	p.advance() // consume "remember"
	expr, err := p.parseExpr()
	if err != nil {
		return nil, errInvalidStackOp("expected expression after 'Remember'", firstLine, firstCol)
	}
	term, err := p.expectTerminator()
	if err != nil {
		return nil, err
	}
	_ = term
	return RememberStmt{Expr: expr, Line: firstLine, Col: firstCol}, nil
}

// parseRecallStmt: "Recall <ignored text>."
func (p *Parser) parseRecallStmt() (Statement, error) {
	firstLine := p.peek().Line
	firstCol := p.peek().Col
	p.advance() // consume "recall"
	var parts []string
	for !p.isEOF() && !p.at(lexer.TokenPeriod) {
		parts = append(parts, p.advance().Lexeme)
	}
	if p.peek().Type != lexer.TokenPeriod {
		return nil, errInvalidStackOp("expected '.' after 'Recall'", firstLine, firstCol)
	}
	p.match(lexer.TokenPeriod)
	return RecallStmt{IgnoredText: strings.Join(parts, " "), Line: firstLine, Col: firstCol}, nil
}

// parseQuestionStmt: "Am I/Is X <comparative> Y?"
func (p *Parser) parseQuestionStmt() (Statement, error) {
	firstLine := p.peek().Line
	firstCol := p.peek().Col
	if p.matchWord("am") {
		if !p.matchWord("i") {
			return nil, errInvalidComparative("", firstLine, firstCol)
		}
		comp, err := p.parseComparative()
		if err != nil {
			return nil, err
		}
		if !p.matchWord("you") && !p.matchWord("thou") {
			return nil, errInvalidComparative("", firstLine, firstCol)
		}
		if !p.match(lexer.TokenQuestion) {
			return nil, errInvalidComparative("", firstLine, firstCol)
		}
		return QuestionStmt{
			Left:        PronounExpr{Ref: "speaker", Line: firstLine, Col: firstCol},
			Comparative: comp,
			Right:       PronounExpr{Ref: "listener", Line: firstLine, Col: firstCol},
			Line:        firstLine,
			Col:         firstCol,
		}, nil
	}
	p.matchWord("is") // consume "is"
	left, err := p.parseExpr()
	if err != nil {
		return nil, err
	}
	comp, err := p.parseComparative()
	if err != nil {
		return nil, err
	}
	right, err := p.parseExpr()
	if err != nil {
		return nil, err
	}
	if !p.match(lexer.TokenQuestion) {
		return nil, errInvalidComparative("", firstLine, firstCol)
	}
	return QuestionStmt{Left: left, Comparative: comp, Right: right, Line: firstLine, Col: firstCol}, nil
}

// parseComparative: 7 forms → 6 relations.
func (p *Parser) parseComparative() (Comparative, error) {
	line, col := p.currentLineCol()
	negated := p.matchWord("not")
	if p.matchWord("as") {
		if p.peek().Type != lexer.TokenWord {
			return Comparative{}, errInvalidComparative("", line, col)
		}
		adj := p.advance().Lexeme
		if !p.matchWord("as") {
			return Comparative{}, errInvalidComparative("", line, col)
		}
		pol := comparativePolarity(adj)
		relation := "equal"
		if pol == "negative" {
			relation = "not_equal"
		}
		if negated {
			if relation == "equal" {
				relation = "not_equal"
			} else {
				relation = "equal"
			}
		}
		return Comparative{Form: "as-as", Adjective: adj, Negated: negated, Relation: relation, Line: line, Col: col}, nil
	}
	if p.peek().Type != lexer.TokenWord {
		return Comparative{}, errInvalidComparative("", line, col)
	}
	adj := p.advance().Lexeme
	if !p.matchWord("than") {
		return Comparative{}, errInvalidComparative(adj, line, col)
	}
	pol := comparativePolarity(adj)
	relation := "greater"
	if pol == "negative" {
		relation = "less"
	}
	if negated {
		if relation == "greater" {
			relation = "less_or_equal"
		} else {
			relation = "greater_or_equal"
		}
	}
	return Comparative{Form: "than", Adjective: adj, Negated: negated, Relation: relation, Line: line, Col: col}, nil
}

// parseIfStmt: "If so/not, let us proceed/return to scene/act X."
func (p *Parser) parseIfStmt() (Statement, error) {
	firstLine := p.peek().Line
	firstCol := p.peek().Col
	p.advance() // consume "if"
	var branchIfTrue bool
	if p.matchWord("so") {
		branchIfTrue = true
	} else if p.matchWord("not") {
		branchIfTrue = false
	} else {
		return nil, errInvalidIf(firstLine, firstCol)
	}
	if !p.match(lexer.TokenComma) {
		return nil, errInvalidIf(firstLine, firstCol)
	}
	if !p.matchWord("let") {
		return nil, errInvalidIf(firstLine, firstCol)
	}
	if !p.matchWord("us") {
		return nil, errInvalidIf(firstLine, firstCol)
	}
	target, targetKind, err := p.parseGotoTarget(firstLine, firstCol)
	if err != nil {
		return nil, err
	}
	return IfStmt{BranchIfTrue: branchIfTrue, Target: target, TargetKind: targetKind, Line: firstLine, Col: firstCol}, nil
}

// parseGotoStmt: "Let us proceed/return to scene/act X."
func (p *Parser) parseGotoStmt() (Statement, error) {
	firstLine := p.peek().Line
	firstCol := p.peek().Col
	p.advance() // consume "let"
	if !p.matchWord("us") {
		return nil, errInvalidIf(firstLine, firstCol)
	}
	target, targetKind, err := p.parseGotoTarget(firstLine, firstCol)
	if err != nil {
		return nil, err
	}
	return GotoStmt{Target: target, TargetKind: targetKind, Line: firstLine, Col: firstCol}, nil
}

func (p *Parser) parseGotoTarget(line, col int) (string, string, error) {
	if !p.matchWord("proceed") && !p.matchWord("return") {
		return "", "", errInvalidIf(line, col)
	}
	if !p.matchWord("to") {
		return "", "", errInvalidIf(line, col)
	}
	var targetKind string
	if p.matchWord("scene") {
		targetKind = "scene"
	} else if p.matchWord("act") {
		targetKind = "act"
	} else {
		return "", "", errInvalidIf(line, col)
	}
	if p.peek().Type != lexer.TokenWord {
		return "", "", errInvalidIf(line, col)
	}
	target := p.advance().Lexeme
	if !p.match(lexer.TokenPeriod) {
		return "", "", errInvalidIf(line, col)
	}
	return target, targetKind, nil
}

// parseExpr: dispatch by leading token. Skips leading newlines.
func (p *Parser) parseExpr() (Expr, error) {
	p.skipNewlines()
	if p.peek().Type != lexer.TokenWord {
		return nil, errInvalidExpression(p.peek().Line, p.peek().Col)
	}
	word := lower(p.peek().Lexeme)
	if word == "the" {
		// Peek past any newlines to find the next significant word.
		// If it's an operator keyword, treat "the" as operator prefix.
		offset := 1
		for p.peekAt(offset).Type == lexer.TokenNewline {
			offset++
		}
		next := p.peekAt(offset)
		if next.Type == lexer.TokenWord && isOpKeyword(lower(next.Lexeme)) {
			return p.parseBinaryOrUnaryOp()
		}
		return p.parseConstant()
	}
	if word == "twice" {
		p.advance()
		operand, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		line, col := operand.Pos()
		return UnaryOpExpr{Op: "twice", Operand: operand, Line: line, Col: col}, nil
	}
	if word == "as" {
		return p.parseSimileInExpr()
	}
	if isSpeakerPronoun(word) {
		line, col := p.peek().Line, p.peek().Col
		p.advance()
		return PronounExpr{Ref: "speaker", Line: line, Col: col}, nil
	}
	if isListenerPronoun(word) {
		line, col := p.peek().Line, p.peek().Col
		p.advance()
		return PronounExpr{Ref: "listener", Line: line, Col: col}, nil
	}
	if p.characters[word] {
		line, col := p.peek().Line, p.peek().Col
		name := p.advance().Lexeme
		return CharRefExpr{Name: name, Line: line, Col: col}, nil
	}
	return p.parseConstant()
}

// parseSimileInExpr: "as <adj> as <expr>" → returns the inner expression.
func (p *Parser) parseSimileInExpr() (Expr, error) {
	line, col := p.currentLineCol()
	p.advance() // consume "as"
	p.skipNewlines()
	if p.peek().Type != lexer.TokenWord {
		return nil, errInvalidComparative("", line, col)
	}
	p.advance() // consume adjective
	p.skipNewlines()
	if !p.matchWord("as") {
		return nil, errInvalidComparative("", line, col)
	}
	return p.parseExpr()
}

// parseBinaryOrUnaryOp: "the <op> of/between <expr> and <expr>" or "the <op> of <expr>".
// Newlines are skipped between tokens (speech spans multiple lines).
func (p *Parser) parseBinaryOrUnaryOp() (Expr, error) {
	line, col := p.currentLineCol()
	p.advance() // consume "the"
	p.skipNewlines()
	if p.peek().Type != lexer.TokenWord {
		return nil, errInvalidExpression(line, col)
	}
	keyword := lower(p.advance().Lexeme)
	switch keyword {
	case "sum":
		return p.parseBinaryTail("sum", "of", line, col)
	case "product":
		return p.parseBinaryTail("product", "of", line, col)
	case "difference":
		return p.parseBinaryTail("difference", "between", line, col)
	case "quotient":
		return p.parseBinaryTail("quotient", "between", line, col)
	case "remainder":
		return p.parseRemainder(line, col)
	case "square":
		if p.matchWord("root") {
			return p.parseUnaryTail("square_root", line, col)
		}
		return p.parseUnaryTail("square", line, col)
	case "cube":
		return p.parseUnaryTail("cube", line, col)
	case "factorial":
		return p.parseUnaryTail("factorial", line, col)
	}
	return nil, errInvalidExpression(line, col)
}

func (p *Parser) parseBinaryTail(op, connector string, line, col int) (Expr, error) {
	p.skipNewlines()
	if !p.matchWord(connector) {
		return nil, errInvalidExpression(line, col)
	}
	p.skipNewlines()
	left, err := p.parseExpr()
	if err != nil {
		return nil, err
	}
	p.skipNewlines()
	if !p.matchWord("and") {
		return nil, errInvalidExpression(line, col)
	}
	p.skipNewlines()
	right, err := p.parseExpr()
	if err != nil {
		return nil, err
	}
	return BinaryOpExpr{Op: op, Left: left, Right: right, Line: line, Col: col}, nil
}

func (p *Parser) parseRemainder(line, col int) (Expr, error) {
	p.skipNewlines()
	if !p.matchWord("of") {
		return nil, errInvalidExpression(line, col)
	}
	p.skipNewlines()
	if !p.matchWord("the") {
		return nil, errInvalidExpression(line, col)
	}
	p.skipNewlines()
	if !p.matchWord("quotient") {
		return nil, errInvalidExpression(line, col)
	}
	p.skipNewlines()
	if !p.matchWord("between") {
		return nil, errInvalidExpression(line, col)
	}
	left, err := p.parseExpr()
	if err != nil {
		return nil, err
	}
	p.skipNewlines()
	if !p.matchWord("and") {
		return nil, errInvalidExpression(line, col)
	}
	right, err := p.parseExpr()
	if err != nil {
		return nil, err
	}
	return BinaryOpExpr{Op: "remainder", Left: left, Right: right, Line: line, Col: col}, nil
}

func (p *Parser) parseUnaryTail(op string, line, col int) (Expr, error) {
	p.skipNewlines()
	if !p.matchWord("of") {
		return nil, errInvalidExpression(line, col)
	}
	operand, err := p.parseExpr()
	if err != nil {
		return nil, err
	}
	return UnaryOpExpr{Op: op, Operand: operand, Line: line, Col: col}, nil
}

// parseConstant: article? adjective* noun. Last non-article word is the noun.
// Newlines within a constant are skipped (speech spans multiple lines).
func (p *Parser) parseConstant() (Expr, error) {
	startLine, startCol := p.peek().Line, p.peek().Col
	adjCount := 0
	noun := ""
	polarity := 1
	hasNoun := false
	for {
		for p.at(lexer.TokenNewline) {
			p.advance()
		}
		if p.peek().Type != lexer.TokenWord {
			break
		}
		w := lower(p.peek().Lexeme)
		if isArticle(w) {
			p.advance()
			continue
		}
		if isPossessive(w) {
			p.advance()
			continue
		}
		if isAdjective(w) {
			adjCount++
			p.advance()
			continue
		}
		// Not an article, possessive, or adjective → it's the noun.
		noun = p.advance().Lexeme
		polarity = nounPolarity(w)
		hasNoun = true
		break
	}
	if !hasNoun {
		return nil, errInvalidExpression(startLine, startCol)
	}
	return ConstExpr{AdjectiveCount: adjCount, Noun: noun, Polarity: polarity, Line: startLine, Col: startCol}, nil
}

// intToRoman converts 1-39 to Roman numeral.
func intToRoman(n int) string {
	if n <= 0 || n > 3999 {
		return ""
	}
	values := []int{1000, 900, 500, 400, 100, 90, 50, 40, 10, 9, 5, 4, 1}
	symbols := []string{"M", "CM", "D", "CD", "C", "XC", "L", "XL", "X", "IX", "V", "IV", "I"}
	var sb strings.Builder
	for i, v := range values {
		for n >= v {
			sb.WriteString(symbols[i])
			n -= v
		}
	}
	return sb.String()
}
