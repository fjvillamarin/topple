package ast

import (
	"biscuit/compiler/lexer"
)

// ListExpr represents a list expression [items]
type ListExpr struct {
	Elements []Expr

	Span lexer.Span
}

func NewListExpr(elements []Expr, Span lexer.Span) *ListExpr {
	return &ListExpr{
		Elements: elements,
		Span:     Span,
	}
}

func (l *ListExpr) isExpr() {}

func (l *ListExpr) GetSpan() lexer.Span {
	return l.Span
}

func (l *ListExpr) String() string {
	return "[...]"
}

// Accept calls the VisitListExpr method on the visitor
func (l *ListExpr) Accept(visitor Visitor) {
	visitor.VisitListExpr(l)
}
