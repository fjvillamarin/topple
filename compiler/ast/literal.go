package ast

import (
	"biscuit/compiler/lexer"
	"fmt"
)

// Literal represents a literal value (number, string, etc.).
type Literal struct {
	Token lexer.Token
	Value any

	span lexer.Span
}

func NewLiteral(token lexer.Token, value any, span lexer.Span) *Literal {
	return &Literal{
		Token: token,
		Value: value,

		span: span,
	}
}

func (l *Literal) isExpr() {}

func (l *Literal) Span() lexer.Span {
	return l.span
}

func (l *Literal) String() string {
	return fmt.Sprintf("%v", l.Value)
}

// Accept calls the VisitLiteral method on the visitor
func (l *Literal) Accept(visitor Visitor) {
	visitor.VisitLiteral(l)
}
