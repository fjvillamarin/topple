package ast

import (
	"biscuit/compiler/lexer"
	"fmt"
)

type Argument struct {
	Name         *Name
	Value        Expr
	IsStar       bool
	IsDoubleStar bool

	Span lexer.Span
}

func (a *Argument) isExpr() {}

func (a *Argument) GetSpan() lexer.Span {
	return a.Span
}

func (a *Argument) Accept(visitor Visitor) {
	visitor.VisitArgument(a)
}

func (a *Argument) String() string {
	return fmt.Sprintf("Argument(%s, %s)", a.Name, a.Value)
}
