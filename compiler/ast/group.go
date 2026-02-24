package ast

import (
	"fmt"
	"github.com/fjvillamarin/topple/compiler/lexer"
)

// GroupExpr represents a parenthesized expression (expr)
type GroupExpr struct {
	Expression Expr

	Span lexer.Span
}

func (g *GroupExpr) isExpr() {}

func (g *GroupExpr) GetSpan() lexer.Span {
	return g.Span
}

func (g *GroupExpr) String() string {
	return fmt.Sprintf("(%v)", g.Expression)
}

// Accept calls the VisitGroupExpr method on the visitor
func (g *GroupExpr) Accept(visitor Visitor) {
	visitor.VisitGroupExpr(g)
}
