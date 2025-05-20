package ast

import (
	"biscuit/compiler/lexer"
	"fmt"
)

// ListExpr represents a list expression [items]
type ListExpr struct {
	Elements []Expr

	span lexer.Span
}

func NewListExpr(elements []Expr, span lexer.Span) *ListExpr {
	return &ListExpr{
		Elements: elements,

		span: span,
	}
}

func (l *ListExpr) isExpr() {}

func (l *ListExpr) Span() lexer.Span {
	return l.span
}

func (l *ListExpr) String() string {
	return fmt.Sprintf("[...]")
}

// Accept calls the VisitListExpr method on the visitor
func (l *ListExpr) Accept(visitor Visitor) {
	visitor.VisitListExpr(l)
}
