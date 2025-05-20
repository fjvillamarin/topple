package ast

import (
	"biscuit/compiler/lexer"
)

// PassStmt represents a 'pass' statement, which is a no-op.
type PassStmt struct {
	span lexer.Span
}

func NewPassStmt(span lexer.Span) *PassStmt {
	return &PassStmt{
		span: span,
	}
}

func (p *PassStmt) isStmt() {}

func (p *PassStmt) Span() lexer.Span {
	return p.span
}

func (p *PassStmt) Accept(visitor Visitor) {
	visitor.VisitPassStmt(p)
}

func (p *PassStmt) String() string {
	return "PassStmt()"
}
