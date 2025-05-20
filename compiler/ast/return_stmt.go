package ast

import (
	"biscuit/compiler/lexer"
	"fmt"
)

// ReturnStmt represents a 'return' statement.
type ReturnStmt struct {
	Value Expr

	span lexer.Span
}

func NewReturnStmt(value Expr, span lexer.Span) *ReturnStmt {
	return &ReturnStmt{
		Value: value,

		span: span,
	}
}

func (r *ReturnStmt) isStmt() {}

func (r *ReturnStmt) Span() lexer.Span {
	return r.span
}

func (r *ReturnStmt) Accept(visitor Visitor) {
	visitor.VisitReturnStmt(r)
}

func (r *ReturnStmt) String() string {
	return fmt.Sprintf("ReturnStmt(%s)", r.Value)
}
