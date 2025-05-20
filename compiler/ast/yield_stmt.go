package ast

import (
	"biscuit/compiler/lexer"
	"fmt"
)

// YieldStmt represents a 'yield' statement.
type YieldStmt struct {
	Value Expr

	span lexer.Span
}

func NewYieldStmt(value Expr, span lexer.Span) *YieldStmt {
	return &YieldStmt{
		Value: value,
		span:  span,
	}
}

func (y *YieldStmt) isStmt() {}

func (y *YieldStmt) Span() lexer.Span {
	return y.span
}

func (y *YieldStmt) Accept(visitor Visitor) {
	visitor.VisitYieldStmt(y)
}

func (y *YieldStmt) String() string {
	return fmt.Sprintf("YieldStmt(%s)", y.Value)
}
