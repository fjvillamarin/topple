package ast

import (
	"biscuit/compiler/lexer"
	"fmt"
)

// GroupExpr represents a parenthesized expression (expr)
type GroupExpr struct {
	Expression Expr

	span lexer.Span
}

func NewGroupExpr(expression Expr, span lexer.Span) *GroupExpr {
	return &GroupExpr{
		Expression: expression,

		span: span,
	}
}

func (g *GroupExpr) isExpr() {}

func (g *GroupExpr) Span() lexer.Span {
	return g.span
}

func (g *GroupExpr) String() string {
	return fmt.Sprintf("(%v)", g.Expression)
}

// Accept calls the VisitGroupExpr method on the visitor
func (g *GroupExpr) Accept(visitor Visitor) {
	visitor.VisitGroupExpr(g)
}
