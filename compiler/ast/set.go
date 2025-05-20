package ast

import (
	"biscuit/compiler/lexer"
	"fmt"
)

// SetExpr represents a set expression {items}
type SetExpr struct {
	Elements []Expr

	span lexer.Span
}

func NewSetExpr(elements []Expr, span lexer.Span) *SetExpr {
	return &SetExpr{
		Elements: elements,

		span: span,
	}
}

func (s *SetExpr) isExpr() {}

func (s *SetExpr) Span() lexer.Span {
	return s.span
}

func (s *SetExpr) String() string {
	return fmt.Sprintf("{...}")
}

// Accept calls the VisitSetExpr method on the visitor
func (s *SetExpr) Accept(visitor Visitor) {
	visitor.VisitSetExpr(s)
}
