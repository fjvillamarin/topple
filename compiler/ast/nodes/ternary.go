package nodes

import (
	"biscuit/compiler/lexer"
	"fmt"
)

// TernaryExpr represents a ternary expression (condition ? trueExpr : falseExpr)
type TernaryExpr struct {
	Condition Expr
	TrueExpr  Expr
	FalseExpr Expr

	span lexer.Span
}

func NewTernaryExpr(condition Expr, trueExpr Expr, falseExpr Expr, span lexer.Span) *TernaryExpr {
	return &TernaryExpr{
		Condition: condition,
		TrueExpr:  trueExpr,
		FalseExpr: falseExpr,

		span: span,
	}
}

func (t *TernaryExpr) isExpr() {}

func (t *TernaryExpr) Span() lexer.Span {
	return t.span
}

func (t *TernaryExpr) String() string {
	return fmt.Sprintf("%v ? %v : %v", t.Condition, t.TrueExpr, t.FalseExpr)
}

// Accept calls the VisitTernaryExpr method on the visitor
func (t *TernaryExpr) Accept(visitor Visitor) {
	visitor.VisitTernaryExpr(t)
}
