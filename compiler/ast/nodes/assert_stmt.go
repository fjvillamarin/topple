package nodes

import (
	"biscuit/compiler/lexer"
	"fmt"
)

// AssertStmt represents an 'assert' statement.
type AssertStmt struct {
	Test    Expr
	Message Expr // Optional error message

	span lexer.Span
}

func NewAssertStmt(test Expr, message Expr, span lexer.Span) *AssertStmt {
	return &AssertStmt{
		Test:    test,
		Message: message,
		span:    span,
	}
}

func (a *AssertStmt) isStmt() {}

func (a *AssertStmt) Span() lexer.Span {
	return a.span
}

func (a *AssertStmt) Accept(visitor Visitor) {
	visitor.VisitAssertStmt(a)
}

func (a *AssertStmt) String() string {
	if a.Message != nil {
		return fmt.Sprintf("AssertStmt(%s, %s)", a.Test, a.Message)
	}
	return fmt.Sprintf("AssertStmt(%s)", a.Test)
}
