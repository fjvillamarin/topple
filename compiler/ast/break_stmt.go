package ast

import (
	"biscuit/compiler/lexer"
)

// BreakStmt represents a 'break' statement.
type BreakStmt struct {
	Span lexer.Span
}

func (b *BreakStmt) isStmt() {}

func (b *BreakStmt) GetSpan() lexer.Span {
	return b.Span
}

func (b *BreakStmt) Accept(visitor Visitor) {
	visitor.VisitBreakStmt(b)
}

func (b *BreakStmt) String() string {
	return "BreakStmt()"
}
