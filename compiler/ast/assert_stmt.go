package ast

import (
	"fmt"
	"topple/compiler/lexer"
)

// AssertStmt represents an 'assert' statement.
type AssertStmt struct {
	Test    Expr
	Message Expr // Optional error message

	Span lexer.Span
}

func (a *AssertStmt) isStmt() {}

func (a *AssertStmt) GetSpan() lexer.Span {
	return a.Span
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
