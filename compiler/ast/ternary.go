package ast

import (
	"fmt"
	"sylfie/compiler/lexer"
)

// TernaryExpr represents a ternary expression (condition ? trueExpr : falseExpr)
type TernaryExpr struct {
	Condition Expr
	TrueExpr  Expr
	FalseExpr Expr

	Span lexer.Span
}

func (t *TernaryExpr) isExpr() {}

func (t *TernaryExpr) GetSpan() lexer.Span {
	return t.Span
}

func (t *TernaryExpr) String() string {
	return fmt.Sprintf("%v ? %v : %v", t.Condition, t.TrueExpr, t.FalseExpr)
}

// Accept calls the VisitTernaryExpr method on the visitor
func (t *TernaryExpr) Accept(visitor Visitor) {
	visitor.VisitTernaryExpr(t)
}
