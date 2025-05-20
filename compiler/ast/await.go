package ast

import (
	"biscuit/compiler/lexer"
	"fmt"
)

// AwaitExpr represents an await expression (await expr)
type AwaitExpr struct {
	Expr Expr

	span lexer.Span
}

func NewAwaitExpr(expr Expr, span lexer.Span) *AwaitExpr {
	return &AwaitExpr{
		Expr: expr,

		span: span,
	}
}

func (a *AwaitExpr) isExpr() {}

func (a *AwaitExpr) Span() lexer.Span {
	return a.span
}

func (a *AwaitExpr) String() string {
	return fmt.Sprintf("await %v", a.Expr)
}

// Accept calls the VisitAwaitExpr method on the visitor
func (a *AwaitExpr) Accept(visitor Visitor) {
	visitor.VisitAwaitExpr(a)
}
