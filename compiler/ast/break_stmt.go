package ast

import (
	"biscuit/compiler/lexer"
)

// BreakStmt represents a 'break' statement.
type BreakStmt struct {
	span lexer.Span
}

func NewBreakStmt(span lexer.Span) *BreakStmt {
	return &BreakStmt{
		span: span,
	}
}

func (b *BreakStmt) isStmt() {}

func (b *BreakStmt) Span() lexer.Span {
	return b.span
}

func (b *BreakStmt) Accept(visitor Visitor) {
	visitor.VisitBreakStmt(b)
}

func (b *BreakStmt) String() string {
	return "BreakStmt()"
}
