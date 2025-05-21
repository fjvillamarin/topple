package ast

import (
	"biscuit/compiler/lexer"
	"fmt"
)

// TypeAlias represents a 'type' statement.
type TypeAlias struct {
	Name   lexer.Token
	Params []Expr // Type parameters as expressions
	Value  Expr   // The type value expression

	Span lexer.Span
}

func (t *TypeAlias) isStmt() {}

func (t *TypeAlias) GetSpan() lexer.Span {
	return t.Span
}

// Accept calls the VisitTypeAlias method on the visitor
func (t *TypeAlias) Accept(visitor Visitor) {
	visitor.VisitTypeAlias(t)
}

func (t *TypeAlias) String() string {
	return fmt.Sprintf("TypeAlias(%s)", t.Name.Lexeme)
}
