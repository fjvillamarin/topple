package ast

import (
	"topple/compiler/lexer"
)

// ContinueStmt represents a 'continue' statement.
type ContinueStmt struct {
	Span lexer.Span
}

func (c *ContinueStmt) isStmt() {}

func (c *ContinueStmt) GetSpan() lexer.Span {
	return c.Span
}

func (c *ContinueStmt) Accept(visitor Visitor) {
	visitor.VisitContinueStmt(c)
}

func (c *ContinueStmt) String() string {
	return "ContinueStmt()"
}
