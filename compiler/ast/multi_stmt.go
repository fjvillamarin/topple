package ast

import (
	"fmt"
	"github.com/fjvillamarin/topple/compiler/lexer"
)

// MultiStmt represents a list of statements.
// It doesn't have semantic meaning, it's just a container for a list of statements.
type MultiStmt struct {
	Stmts []Stmt

	Span lexer.Span
}

func (m *MultiStmt) isStmt() {}

func (m *MultiStmt) GetSpan() lexer.Span {
	return m.Span
}

func (m *MultiStmt) Accept(visitor Visitor) {
	visitor.VisitMultiStmt(m)
}

func (m *MultiStmt) String() string {
	return fmt.Sprintf("MultiStmt(%d stmts)", len(m.Stmts))
}
