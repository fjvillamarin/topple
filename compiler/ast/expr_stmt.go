package ast

import (
	"fmt"
	"sylfie/compiler/lexer"
)

// ExprStmt represents an expression used as a statement.
type ExprStmt struct {
	Expr Expr

	Span lexer.Span
}

func (e *ExprStmt) isStmt() {}

func (e *ExprStmt) GetSpan() lexer.Span {
	return e.Span
}

// Accept calls the VisitExprStmt method on the visitor
func (e *ExprStmt) Accept(visitor Visitor) {
	visitor.VisitExprStmt(e)
}

func (e *ExprStmt) String() string {
	return fmt.Sprintf("ExprStmt(%s)", e.Expr)
}
