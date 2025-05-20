package ast

import (
	"biscuit/compiler/lexer"
)

// ContinueStmt represents a 'continue' statement.
type ContinueStmt struct {
	span lexer.Span
}

func NewContinueStmt(span lexer.Span) *ContinueStmt {
	return &ContinueStmt{
		span: span,
	}
}

func (c *ContinueStmt) isStmt() {}

func (c *ContinueStmt) Span() lexer.Span {
	return c.span
}

func (c *ContinueStmt) Accept(visitor Visitor) {
	visitor.VisitContinueStmt(c)
}

func (c *ContinueStmt) String() string {
	return "ContinueStmt()"
}
