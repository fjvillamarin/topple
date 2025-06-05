package ast

import (
	"sylfie/compiler/lexer"
)

// PassStmt represents a 'pass' statement, which is a no-op.
type PassStmt struct {
	Span lexer.Span
}

func (p *PassStmt) isStmt() {}

func (p *PassStmt) GetSpan() lexer.Span {
	return p.Span
}

func (p *PassStmt) Accept(visitor Visitor) {
	visitor.VisitPassStmt(p)
}

func (p *PassStmt) String() string {
	return "PassStmt()"
}
