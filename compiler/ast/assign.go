package ast

import (
	"biscuit/compiler/lexer"
	"fmt"
)

// AssignExpr represents an assignment expression (left = right)
type AssignExpr struct {
	Left  Expr
	Right Expr

	Span lexer.Span
}

func (a *AssignExpr) isExpr() {}

func (a *AssignExpr) GetSpan() lexer.Span {
	return a.Span
}

func (a *AssignExpr) String() string {
	return fmt.Sprintf("%v = %v", a.Left, a.Right)
}

// Accept calls the VisitAssignExpr method on the visitor
func (a *AssignExpr) Accept(visitor Visitor) {
	visitor.VisitAssignExpr(a)
}
