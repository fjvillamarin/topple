package ast

import (
	"biscuit/compiler/lexer"
	"fmt"
)

// ExprStmt represents an expression used as a statement.
type ExprStmt struct {
	Value Expr

	span lexer.Span
}

func NewExprStmt(value Expr, span lexer.Span) *ExprStmt {
	return &ExprStmt{
		Value: value,

		span: span,
	}
}

func (e *ExprStmt) isStmt() {}

func (e *ExprStmt) Span() lexer.Span {
	return e.span
}

// Accept calls the VisitExprStmt method on the visitor
func (e *ExprStmt) Accept(visitor Visitor) {
	visitor.VisitExprStmt(e)
}

func (e *ExprStmt) String() string {
	return fmt.Sprintf("ExprStmt(%s)", e.Value)
}
