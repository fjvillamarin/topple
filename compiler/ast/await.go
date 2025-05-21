package ast

import (
	"biscuit/compiler/lexer"
	"fmt"
)

// AwaitExpr represents an await expression (await expr)
type AwaitExpr struct {
	Expr Expr

	Span lexer.Span
}

func (a *AwaitExpr) isExpr() {}

func (a *AwaitExpr) GetSpan() lexer.Span {
	return a.Span
}

func (a *AwaitExpr) String() string {
	return fmt.Sprintf("await %v", a.Expr)
}

// Accept calls the VisitAwaitExpr method on the visitor
func (a *AwaitExpr) Accept(visitor Visitor) {
	visitor.VisitAwaitExpr(a)
}
