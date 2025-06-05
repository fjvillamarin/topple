package ast

import (
	"sylfie/compiler/lexer"
	"fmt"
)

type If struct {
	Condition Expr
	Body      []Stmt
	Else      []Stmt

	Span lexer.Span
}

func (i *If) isStmt() {}

func (i *If) GetSpan() lexer.Span {
	return i.Span
}

func (i *If) Accept(visitor Visitor) {
	visitor.VisitIf(i)
}

func (i *If) String() string {
	return fmt.Sprintf("if %s", i.Condition.String())
}
