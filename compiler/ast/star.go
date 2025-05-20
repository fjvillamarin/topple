package ast

import (
	"biscuit/compiler/lexer"
	"fmt"
)

// StarExpr represents a star expression (*expr)
type StarExpr struct {
	Expr Expr

	span lexer.Span
}

func NewStarExpr(expr Expr, span lexer.Span) *StarExpr {
	return &StarExpr{
		Expr: expr,

		span: span,
	}
}

func (s *StarExpr) isExpr() {}

func (s *StarExpr) Span() lexer.Span {
	return s.span
}

func (s *StarExpr) String() string {
	return fmt.Sprintf("*%v", s.Expr)
}

// Accept calls the VisitStarExpr method on the visitor
func (s *StarExpr) Accept(visitor Visitor) {
	visitor.VisitStarExpr(s)
}
