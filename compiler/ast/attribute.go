package ast

import (
	"biscuit/compiler/lexer"
	"fmt"
)

// Attribute represents an attribute access expression (obj.attr)
type Attribute struct {
	Object Expr
	Name   lexer.Token

	span lexer.Span
}

func NewAttribute(object Expr, name lexer.Token, span lexer.Span) *Attribute {
	return &Attribute{
		Object: object,
		Name:   name,

		span: span,
	}
}

func (a *Attribute) isExpr() {}

func (a *Attribute) Span() lexer.Span {
	return a.span
}

func (a *Attribute) String() string {
	return fmt.Sprintf("%v.%s", a.Object, a.Name.Lexeme)
}

// Accept calls the VisitAttribute method on the visitor
func (a *Attribute) Accept(visitor Visitor) {
	visitor.VisitAttribute(a)
}
