package ast

import (
	"sylfie/compiler/lexer"
)

// SetExpr represents a set expression {items}
type SetExpr struct {
	Elements []Expr

	Span lexer.Span
}

func (s *SetExpr) isExpr() {}

func (s *SetExpr) GetSpan() lexer.Span {
	return s.Span
}

func (s *SetExpr) String() string {
	return "{...}"
}

// Accept calls the VisitSetExpr method on the visitor
func (s *SetExpr) Accept(visitor Visitor) {
	visitor.VisitSetExpr(s)
}
