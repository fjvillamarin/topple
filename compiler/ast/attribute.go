package ast

import (
	"fmt"
	"sylfie/compiler/lexer"
)

// Attribute represents an attribute access expression (obj.attr)
type Attribute struct {
	Object Expr
	Name   lexer.Token

	Span lexer.Span
}

func (a *Attribute) isExpr() {}

func (a *Attribute) GetSpan() lexer.Span {
	return a.Span
}

func (a *Attribute) String() string {
	return fmt.Sprintf("%v.%s", a.Object, a.Name.Lexeme)
}

// Accept calls the VisitAttribute method on the visitor
func (a *Attribute) Accept(visitor Visitor) {
	visitor.VisitAttribute(a)
}
