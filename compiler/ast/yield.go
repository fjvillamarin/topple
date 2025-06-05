package ast

import (
	"sylfie/compiler/lexer"
	"fmt"
)

// YieldExpr represents a yield expression (yield value)
type YieldExpr struct {
	IsFrom bool
	Value  Expr

	Span lexer.Span
}

func (y *YieldExpr) isExpr() {}

func (y *YieldExpr) GetSpan() lexer.Span {
	return y.Span
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
