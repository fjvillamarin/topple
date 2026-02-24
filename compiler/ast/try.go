package ast

import (
	"fmt"
	"github.com/fjvillamarin/topple/compiler/lexer"
)

type Try struct {
	Body    []Stmt
	Excepts []Except
	Else    []Stmt
	Finally []Stmt

	Span lexer.Span
}

type Except struct {
	Type   Expr  // The exception type (nil for a bare except)
	Name   *Name // The name to bind the exception to (as in 'except Type as name')
	Body   []Stmt
	IsStar bool // Whether this is an 'except*' handler

	Span lexer.Span
}

func (t *Try) isStmt() {}

func (t *Try) GetSpan() lexer.Span {
	return t.Span
}

func (t *Try) Accept(visitor Visitor) {
	visitor.VisitTry(t)
}

func (t *Try) String() string {
	return fmt.Sprintf("try %s", t.Body)
}
