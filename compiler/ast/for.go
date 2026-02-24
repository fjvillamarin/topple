package ast

import (
	"fmt"
	"github.com/fjvillamarin/topple/compiler/lexer"
)

type For struct {
	Target   Expr
	Iterable Expr
	Body     []Stmt
	Else     []Stmt
	IsAsync  bool

	Span lexer.Span
}

func (f *For) isStmt() {}

func (f *For) GetSpan() lexer.Span {
	return f.Span
}

func (f *For) Accept(visitor Visitor) {
	visitor.VisitFor(f)
}

func (f *For) String() string {
	return fmt.Sprintf("for %s in %s", f.Target.String(), f.Iterable.String())
}
