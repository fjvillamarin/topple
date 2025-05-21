package ast

import (
	"biscuit/compiler/lexer"
	"fmt"
)

// YieldStmt represents a 'yield' statement.
type YieldStmt struct {
	Value Expr

	Span lexer.Span
}

func (y *YieldStmt) isStmt() {}

func (y *YieldStmt) GetSpan() lexer.Span {
	return y.Span
}

func (y *YieldStmt) Accept(visitor Visitor) {
	visitor.VisitYieldStmt(y)
}

func (y *YieldStmt) String() string {
	return fmt.Sprintf("YieldStmt(%s)", y.Value)
}
