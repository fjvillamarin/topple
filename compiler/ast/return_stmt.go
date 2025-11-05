package ast

import (
	"fmt"
	"topple/compiler/lexer"
)

// ReturnStmt represents a 'return' statement.
type ReturnStmt struct {
	Value Expr

	Span lexer.Span
}

func (r *ReturnStmt) isStmt() {}

func (r *ReturnStmt) GetSpan() lexer.Span {
	return r.Span
}

func (r *ReturnStmt) Accept(visitor Visitor) {
	visitor.VisitReturnStmt(r)
}

func (r *ReturnStmt) String() string {
	return fmt.Sprintf("ReturnStmt(%s)", r.Value)
}
