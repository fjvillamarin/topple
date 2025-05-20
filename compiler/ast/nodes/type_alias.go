package nodes

import (
	"biscuit/compiler/lexer"
	"fmt"
)

// TypeAlias represents a 'type' statement.
type TypeAlias struct {
	Name   lexer.Token
	Params []Expr // Type parameters as expressions
	Value  Expr   // The type value expression

	span lexer.Span
}

func NewTypeAlias(name lexer.Token, params []Expr, value Expr, span lexer.Span) *TypeAlias {
	return &TypeAlias{
		Name:   name,
		Params: params,
		Value:  value,
		span:   span,
	}
}

func (t *TypeAlias) isStmt() {}

func (t *TypeAlias) Span() lexer.Span {
	return t.span
}

// Accept calls the VisitTypeAlias method on the visitor
func (t *TypeAlias) Accept(visitor Visitor) {
	visitor.VisitTypeAlias(t)
}

func (t *TypeAlias) String() string {
	return fmt.Sprintf("TypeAlias(%s)", t.Name.Lexeme)
}
