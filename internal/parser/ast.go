package parser

import "fmt"

// Node is implemented by all AST nodes. Pos returns the 1-indexed line and column.
type Node interface {
	Pos() (line, col int)
	String() string
}

// Statement is the marker interface for all dialogue-line statements.
type Statement interface {
	Node
	stmtNode()
}

// Expr is the marker interface for all expression nodes.
type Expr interface {
	Node
	exprNode()
}

// Program is the root AST node.
type Program struct {
	Title      Title           `json:"title"`
	Characters []CharacterDecl `json:"characters"`
	Acts       []Act           `json:"acts"`
	Warnings   []Warning       `json:"warnings"`
	Line       int             `json:"line"`
	Col        int             `json:"col"`
}

func (p *Program) Pos() (int, int) { return p.Line, p.Col }
func (p *Program) String() string {
	return fmt.Sprintf("Program{title=%q, chars=%d, acts=%d, warnings=%d}",
		p.Title.Text, len(p.Characters), len(p.Acts), len(p.Warnings))
}

// Title is the play title (first line, everything before the terminating '.').
type Title struct {
	Text string `json:"text"`
	Line int    `json:"line"`
	Col  int    `json:"col"`
}

func (t Title) Pos() (int, int) { return t.Line, t.Col }
func (t Title) String() string  { return fmt.Sprintf("Title{%q}", t.Text) }

// CharacterDecl is a Dramatis Personae entry.
type CharacterDecl struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Line        int    `json:"line"`
	Col         int    `json:"col"`
}

func (c CharacterDecl) Pos() (int, int) { return c.Line, c.Col }
func (c CharacterDecl) String() string {
	return fmt.Sprintf("CharacterDecl{name=%q, desc=%q}", c.Name, c.Description)
}

// Act is an act header.
type Act struct {
	Number       int     `json:"number"`
	RomanNumeral string  `json:"roman_numeral"`
	Description  string  `json:"description"`
	Scenes       []Scene `json:"scenes"`
	Line         int     `json:"line"`
	Col          int     `json:"col"`
}

func (a Act) Pos() (int, int) { return a.Line, a.Col }
func (a Act) String() string {
	return fmt.Sprintf("Act{%s, scenes=%d}", a.RomanNumeral, len(a.Scenes))
}

// Scene is a scene header + its statements.
type Scene struct {
	Number       int         `json:"number"`
	RomanNumeral string      `json:"roman_numeral"`
	Description  string      `json:"description"`
	Statements   []Statement `json:"statements"`
	Line         int         `json:"line"`
	Col          int         `json:"col"`
}

func (s Scene) Pos() (int, int) { return s.Line, s.Col }
func (s Scene) String() string {
	return fmt.Sprintf("Scene{%s, stmts=%d}", s.RomanNumeral, len(s.Statements))
}

// EnterStmt: [Enter A] or [Enter A and B]
type EnterStmt struct {
	Characters []string `json:"characters"`
	Line       int      `json:"line"`
	Col        int      `json:"col"`
}

func (s EnterStmt) Pos() (int, int) { return s.Line, s.Col }
func (s EnterStmt) String() string {
	return fmt.Sprintf("EnterStmt{chars=%v}", s.Characters)
}
func (EnterStmt) stmtNode() {}

// ExitStmt: [Exit A]
type ExitStmt struct {
	Character string `json:"character"`
	Line      int    `json:"line"`
	Col       int    `json:"col"`
}

func (s ExitStmt) Pos() (int, int) { return s.Line, s.Col }
func (s ExitStmt) String() string  { return fmt.Sprintf("ExitStmt{char=%q}", s.Character) }
func (ExitStmt) stmtNode()         {}

// ExeuntStmt: [Exeunt] (nil = all) or [Exeunt A and B]
type ExeuntStmt struct {
	Characters []string `json:"characters"`
	Line       int      `json:"line"`
	Col        int      `json:"col"`
}

func (s ExeuntStmt) Pos() (int, int) { return s.Line, s.Col }
func (s ExeuntStmt) String() string {
	return fmt.Sprintf("ExeuntStmt{chars=%v}", s.Characters)
}
func (ExeuntStmt) stmtNode() {}

// Dialogue is a speaker turn containing multiple statements.
type Dialogue struct {
	Speaker    string      `json:"speaker"`
	Statements []Statement `json:"statements"`
	Line       int         `json:"line"`
	Col        int         `json:"col"`
}

func (d Dialogue) Pos() (int, int) { return d.Line, d.Col }
func (d Dialogue) String() string {
	return fmt.Sprintf("Dialogue{speaker=%q, stmts=%d}", d.Speaker, len(d.Statements))
}
func (Dialogue) stmtNode() {}

// AssignStmt: "You are [as ADJ as] <expr>." or "You <constant>!" (no copula).
type AssignStmt struct {
	Target     string `json:"target"`
	SimileAdj  string `json:"simile_adj"`
	Expr       Expr   `json:"expr"`
	Terminator string `json:"terminator"`
	Line       int    `json:"line"`
	Col        int    `json:"col"`
}

func (s AssignStmt) Pos() (int, int) { return s.Line, s.Col }
func (s AssignStmt) String() string {
	return fmt.Sprintf("AssignStmt{target=%q, simile=%q, expr=%s, term=%q}",
		s.Target, s.SimileAdj, s.Expr, s.Terminator)
}
func (AssignStmt) stmtNode() {}

// SpeakStmt: "Speak your/thy mind." — output ASCII of listener.
type SpeakStmt struct {
	Terminator string `json:"terminator"`
	Line       int    `json:"line"`
	Col        int    `json:"col"`
}

func (s SpeakStmt) Pos() (int, int) { return s.Line, s.Col }
func (s SpeakStmt) String() string  { return fmt.Sprintf("SpeakStmt{term=%q}", s.Terminator) }
func (SpeakStmt) stmtNode()         {}

// OpenHeartStmt: "Open your/thy heart." — output numeric of listener.
type OpenHeartStmt struct {
	Terminator string `json:"terminator"`
	Line       int    `json:"line"`
	Col        int    `json:"col"`
}

func (s OpenHeartStmt) Pos() (int, int) { return s.Line, s.Col }
func (s OpenHeartStmt) String() string  { return fmt.Sprintf("OpenHeartStmt{term=%q}", s.Terminator) }
func (OpenHeartStmt) stmtNode()         {}

// OpenMindStmt: "Open your/thy mind." — read ASCII char into listener.
type OpenMindStmt struct {
	Terminator string `json:"terminator"`
	Line       int    `json:"line"`
	Col        int    `json:"col"`
}

func (s OpenMindStmt) Pos() (int, int) { return s.Line, s.Col }
func (s OpenMindStmt) String() string  { return fmt.Sprintf("OpenMindStmt{term=%q}", s.Terminator) }
func (OpenMindStmt) stmtNode()         {}

// ListenStmt: "Listen to your/thy heart." — read number into listener.
type ListenStmt struct {
	Terminator string `json:"terminator"`
	Line       int    `json:"line"`
	Col        int    `json:"col"`
}

func (s ListenStmt) Pos() (int, int) { return s.Line, s.Col }
func (s ListenStmt) String() string  { return fmt.Sprintf("ListenStmt{term=%q}", s.Terminator) }
func (ListenStmt) stmtNode()         {}

// QuestionStmt: "Am I/Is X <comparative> Y?" — sets comparison flag.
type QuestionStmt struct {
	Left        Expr        `json:"left"`
	Comparative Comparative `json:"comparative"`
	Right       Expr        `json:"right"`
	Line        int         `json:"line"`
	Col         int         `json:"col"`
}

func (s QuestionStmt) Pos() (int, int) { return s.Line, s.Col }
func (s QuestionStmt) String() string {
	return fmt.Sprintf("QuestionStmt{%s %s %s}", s.Left, s.Comparative, s.Right)
}
func (QuestionStmt) stmtNode() {}

// Comparative represents a comparative phrase (7 forms → 6 relations).
type Comparative struct {
	Form      string `json:"form"`
	Adjective string `json:"adjective"`
	Negated   bool   `json:"negated"`
	Relation  string `json:"relation"`
	Line      int    `json:"line"`
	Col       int    `json:"col"`
}

func (c Comparative) Pos() (int, int) { return c.Line, c.Col }
func (c Comparative) String() string {
	neg := ""
	if c.Negated {
		neg = "not "
	}
	if c.Form == "as-as" {
		return fmt.Sprintf("%sas %s as(%s)", neg, c.Adjective, c.Relation)
	}
	return fmt.Sprintf("%s%s than(%s)", neg, c.Adjective, c.Relation)
}

// IfStmt: "If so/not, let us proceed/return to scene/act X." (conditional).
type IfStmt struct {
	BranchIfTrue bool   `json:"branch_if_true"`
	Target       string `json:"target"`
	TargetKind   string `json:"target_kind"`
	Line         int    `json:"line"`
	Col          int    `json:"col"`
}

func (s IfStmt) Pos() (int, int) { return s.Line, s.Col }
func (s IfStmt) String() string {
	branch := "so"
	if !s.BranchIfTrue {
		branch = "not"
	}
	return fmt.Sprintf("IfStmt{branch=%s, target=%s %s}", branch, s.TargetKind, s.Target)
}
func (IfStmt) stmtNode() {}

// GotoStmt: "Let us proceed/return to scene/act X." (unconditional).
type GotoStmt struct {
	Target     string `json:"target"`
	TargetKind string `json:"target_kind"`
	Line       int    `json:"line"`
	Col        int    `json:"col"`
}

func (s GotoStmt) Pos() (int, int) { return s.Line, s.Col }
func (s GotoStmt) String() string {
	return fmt.Sprintf("GotoStmt{target=%s %s}", s.TargetKind, s.Target)
}
func (GotoStmt) stmtNode() {}

// RememberStmt: "Remember <expr>." — push onto listener's stack.
type RememberStmt struct {
	Expr Expr `json:"expr"`
	Line int  `json:"line"`
	Col  int  `json:"col"`
}

func (s RememberStmt) Pos() (int, int) { return s.Line, s.Col }
func (s RememberStmt) String() string  { return fmt.Sprintf("RememberStmt{expr=%s}", s.Expr) }
func (RememberStmt) stmtNode()         {}

// RecallStmt: "Recall <ignored text>." — pop listener's stack into speaker.
type RecallStmt struct {
	IgnoredText string `json:"ignored_text"`
	Line        int    `json:"line"`
	Col         int    `json:"col"`
}

func (s RecallStmt) Pos() (int, int) { return s.Line, s.Col }
func (s RecallStmt) String() string  { return fmt.Sprintf("RecallStmt{ignored=%q}", s.IgnoredText) }
func (RecallStmt) stmtNode()         {}

// ConstExpr: article? adjective* noun. Value = Polarity × 2^AdjectiveCount.
type ConstExpr struct {
	AdjectiveCount int    `json:"adjective_count"`
	Noun           string `json:"noun"`
	Polarity       int    `json:"polarity"`
	Line           int    `json:"line"`
	Col            int    `json:"col"`
}

func (e ConstExpr) Pos() (int, int) { return e.Line, e.Col }
func (e ConstExpr) String() string {
	return fmt.Sprintf("Const{adj=%d, noun=%q, pol=%+d}", e.AdjectiveCount, e.Noun, e.Polarity)
}
func (ConstExpr) exprNode() {}

// CharRefExpr: reference to a declared character.
type CharRefExpr struct {
	Name string `json:"name"`
	Line int    `json:"line"`
	Col  int    `json:"col"`
}

func (e CharRefExpr) Pos() (int, int) { return e.Line, e.Col }
func (e CharRefExpr) String() string  { return fmt.Sprintf("CharRef{%s}", e.Name) }
func (CharRefExpr) exprNode()         {}

// PronounExpr: "speaker" or "listener" reference.
type PronounExpr struct {
	Ref  string `json:"ref"`
	Line int    `json:"line"`
	Col  int    `json:"col"`
}

func (e PronounExpr) Pos() (int, int) { return e.Line, e.Col }
func (e PronounExpr) String() string  { return fmt.Sprintf("Pronoun{%s}", e.Ref) }
func (PronounExpr) exprNode()         {}

// BinaryOpExpr: sum, difference, product, quotient, remainder.
type BinaryOpExpr struct {
	Op    string `json:"op"`
	Left  Expr   `json:"left"`
	Right Expr   `json:"right"`
	Line  int    `json:"line"`
	Col   int    `json:"col"`
}

func (e BinaryOpExpr) Pos() (int, int) { return e.Line, e.Col }
func (e BinaryOpExpr) String() string {
	return fmt.Sprintf("BinOp{%s, %s, %s}", e.Op, e.Left, e.Right)
}
func (BinaryOpExpr) exprNode() {}

// UnaryOpExpr: square, cube, square_root, factorial, twice.
type UnaryOpExpr struct {
	Op      string `json:"op"`
	Operand Expr   `json:"operand"`
	Line    int    `json:"line"`
	Col     int    `json:"col"`
}

func (e UnaryOpExpr) Pos() (int, int) { return e.Line, e.Col }
func (e UnaryOpExpr) String() string {
	return fmt.Sprintf("UnaryOp{%s, %s}", e.Op, e.Operand)
}
func (UnaryOpExpr) exprNode() {}
