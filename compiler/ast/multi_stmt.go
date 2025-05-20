package ast

import (
	"biscuit/compiler/lexer"
	"fmt"
)

// MultiStmt represents a list of statements.
// It doesn't have semantic meaning, it's just a container for a list of statements.
type MultiStmt struct {
	Stmts []Stmt

	span lexer.Span
}

func NewMultiStmt(stmts []Stmt, span lexer.Span) *MultiStmt {
	return &MultiStmt{
		Stmts: stmts,
		span:  span,
	}
}

func (m *MultiStmt) isStmt() {}

func (m *MultiStmt) Span() lexer.Span {
	return m.span
}

func (m *MultiStmt) Accept(visitor Visitor) {
	visitor.VisitMultiStmt(m)
}

func (m *MultiStmt) String() string {
	return fmt.Sprintf("MultiStmt(%d stmts)", len(m.Stmts))
}
