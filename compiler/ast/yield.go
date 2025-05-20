package ast

import (
	"biscuit/compiler/lexer"
	"fmt"
)

// YieldExpr represents a yield expression (yield value)
type YieldExpr struct {
	IsFrom bool
	Value  Expr

	span lexer.Span
}

func NewYieldExpr(isFrom bool, value Expr, span lexer.Span) *YieldExpr {
	return &YieldExpr{
		IsFrom: isFrom,
		Value:  value,

		span: span,
	}
}

func (y *YieldExpr) isExpr() {}

func (y *YieldExpr) Span() lexer.Span {
	return y.span
}

func (y *YieldExpr) String() string {
	if y.IsFrom {
		return fmt.Sprintf("yield from %v", y.Value)
	}
	return fmt.Sprintf("yield %v", y.Value)
}

// Accept calls the VisitYieldExpr method on the visitor
func (y *YieldExpr) Accept(visitor Visitor) {
	visitor.VisitYieldExpr(y)
}
