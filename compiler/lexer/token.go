// token.go
//
// Tokens for a (CPython-3.12) Python scanner, suitable for the
// "Crafting Interpreters" architecture but written in Go.
//
// The scanner should emit INDENT / DEDENT *layout* tokens, NEWLINE
// tokens at the ends of logical lines, and EOF when input is finished.
// All keywords are recognised here so the parser can rely on distinct
// TokenType values instead of string comparisons.

package lexer

import "fmt"

// TokenType represents the type of token in Python.
type TokenType int

const (
	// ── single-character punctuation ─────────────────────────────
	LeftParen    TokenType = iota // (
	RightParen                    // )
	LeftBracket                   // [
	RightBracket                  // ]
	LeftBrace                     // {
	RightBrace                    // }
	Comma                         // ,
	Colon                         // :
	Dot                           // .
	Semicolon                     // ;
	Plus                          // +
	Minus                         // -
	Star                          // *
	Slash                         // /
	Percent                       // %
	Pipe                          // |
	Ampersand                     // &
	Caret                         // ^
	Tilde                         // ~
	At                            // @

	// ── one- or two-character operators ─────────────────────────
	Equal               // =
	PlusEqual           // +=
	MinusEqual          // -=
	StarEqual           // *=
	SlashEqual          // /=
	PercentEqual        // %=
	PipeEqual           // |=
	AmpEqual            // &=
	CaretEqual          // ^=
	Arrow               // ->
	AtEqual             // @=
	SlashSlash          // //
	SlashSlashEqual     // //=
	StarStar            // **
	StarStarEqual       // **=
	LessLess            // <<
	GreaterGreater      // >>
	LessLessEqual       // <<=
	GreaterGreaterEqual // >>=
	BangEqual           // !=
	EqualEqual          // ==
	Less                // <
	LessEqual           // <=
	Greater             // >
	GreaterEqual        // >=
	Walrus              // :=
	IsNot               // is not
	NotIn               // not in

	// ── literals & special symbols ──────────────────────────────
	Identifier
	String
	Number
	Ellipsis // ...

	// ── layout / structural tokens ──────────────────────────────
	Newline
	Indent
	Dedent

	// ── keywords (true language keywords, not soft keywords) ────
	And
	As
	Assert
	Async
	Await
	Break
	Class
	Continue
	Def
	Del
	Elif
	Else
	Except
	False // 'False' boolean literal
	Finally
	For
	From
	Global
	If
	Import
	In
	Is
	Lambda
	Match
	None // 'None' singleton
	Nonlocal
	Not
	Or
	Pass
	Raise
	Return
	True // 'True' boolean literal
	Try
	While
	With
	Yield
	Case // soft keyword used inside 'match'
	Type // soft keyword for type aliases

	// biscuit specific tokens
	View
	Component

	EOF
	Illegal
)

// mapping from TokenType values to their string representation, keeping the
// order in sync with the iota declarations above so lookup is O(1).
var tokenTypeNames = [...]string{
	"LeftParen",
	"RightParen",
	"LeftBracket",
	"RightBracket",
	"LeftBrace",
	"RightBrace",
	"Comma",
	"Colon",
	"Dot",
	"Semicolon",
	"Plus",
	"Minus",
	"Star",
	"Slash",
	"Percent",
	"Pipe",
	"Ampersand",
	"Caret",
	"Tilde",
	"At",

	"Equal",
	"PlusEqual",
	"MinusEqual",
	"StarEqual",
	"SlashEqual",
	"PercentEqual",
	"PipeEqual",
	"AmpEqual",
	"CaretEqual",
	"Arrow",
	"AtEqual",
	"SlashSlash",
	"SlashSlashEqual",
	"StarStar",
	"StarStarEqual",
	"LessLess",
	"GreaterGreater",
	"LessLessEqual",
	"GreaterGreaterEqual",
	"BangEqual",
	"EqualEqual",
	"Less",
	"LessEqual",
	"Greater",
	"GreaterEqual",
	"Walrus",
	"IsNot",
	"NotIn",

	"Identifier",
	"String",
	"Number",
	"Ellipsis",

	"Newline",
	"Indent",
	"Dedent",

	"And",
	"As",
	"Assert",
	"Async",
	"Await",
	"Break",
	"Class",
	"Continue",
	"Def",
	"Del",
	"Elif",
	"Else",
	"Except",
	"False",
	"Finally",
	"For",
	"From",
	"Global",
	"If",
	"Import",
	"In",
	"Is",
	"Lambda",
	"Match",
	"None",
	"Nonlocal",
	"Not",
	"Or",
	"Pass",
	"Raise",
	"Return",
	"True",
	"Try",
	"While",
	"With",
	"Yield",
	"Case",
	"Type",

	// biscuit specific tokens
	"View",
	"Component",

	"EOF",
	"Illegal",
}

// String implements fmt.Stringer, returning a human-readable name for the
// token type. If the value is out of range it falls back to the numeric form.
func (tt TokenType) String() string {
	if int(tt) < 0 || int(tt) >= len(tokenTypeNames) {
		return fmt.Sprintf("TokenType(%d)", tt)
	}
	return tokenTypeNames[tt]
}

// Position is a helper type for representing a position in a file.
type Position struct {
	Line   int
	Column int
}

func (p Position) String() string {
	return fmt.Sprintf("L%d:%d", p.Line, p.Column)
}

// Span is a helper type for representing a span of text in a file.
type Span struct {
	Start Position
	End   Position
}

func (s *Span) String() string {
	return fmt.Sprintf("%s-%s", s.Start, s.End)
}

// Token carries full positional information so the analyzer can
// implement precise diagnostics.
type Token struct {
	Type    TokenType
	Lexeme  string
	Literal any // decoded string/number value, or nil
	Span    Span
}

func (t Token) Start() Position {
	return t.Span.Start
}

func (t Token) End() Position {
	return t.Span.End
}

func (t Token) String() string {
	return fmt.Sprintf("%s %q %v", t.Type, t.Lexeme, t.Literal)
}

// Keywords maps the textual form of each keyword to its TokenType.
// Soft keywords ("case", "type") are listed too; when they appear
// in positions where ordinary identifiers are legal, the scanner
// should *not* look them up here so that the parser can choose the
// correct interpretation.
var Keywords = map[string]TokenType{
	"and":      And,
	"as":       As,
	"assert":   Assert,
	"async":    Async,
	"await":    Await,
	"break":    Break,
	"class":    Class,
	"continue": Continue,
	"def":      Def,
	"del":      Del,
	"elif":     Elif,
	"else":     Else,
	"except":   Except,
	"false":    False,
	"finally":  Finally,
	"for":      For,
	"from":     From,
	"global":   Global,
	"if":       If,
	"import":   Import,
	"in":       In,
	"is":       Is,
	"lambda":   Lambda,
	"match":    Match,
	"none":     None,
	"nonlocal": Nonlocal,
	"not":      Not,
	"or":       Or,
	"pass":     Pass,
	"raise":    Raise,
	"return":   Return,
	"true":     True,
	"try":      Try,
	"while":    While,
	"with":     With,
	"yield":    Yield,
	// soft keywords – check context in the parser:
	"case": Case,
	"type": Type,

	// biscuit specific keywords
	"view":      View,
	"component": Component,
}

// IsKeyword reports whether s is a reserved keyword.
func IsKeyword(s string) bool {
	_, ok := Keywords[s]
	return ok
}
