package nodes

import (
	"biscuit/compiler/lexer"
	"fmt"
)

// AssignExpr represents an assignment expression (left = right)
type AssignExpr struct {
	Left  Expr
	Right Expr

	span lexer.Span
}

func NewAssignExpr(left Expr, right Expr, span lexer.Span) *AssignExpr {
	return &AssignExpr{
		Left:  left,
		Right: right,

		span: span,
	}
}

func (a *AssignExpr) isExpr() {}

func (a *AssignExpr) Span() lexer.Span {
	return a.span
}

func (a *AssignExpr) String() string {
	return fmt.Sprintf("%v = %v", a.Left, a.Right)
}

// Accept calls the VisitAssignExpr method on the visitor
func (a *AssignExpr) Accept(visitor Visitor) {
	visitor.VisitAssignExpr(a)
}
