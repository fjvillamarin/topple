package ast

import (
	"fmt"
	"sylfie/compiler/lexer"
)

type With struct {
	Items   []WithItem
	IsAsync bool
	Body    []Stmt

	Span lexer.Span
}

type WithItem struct {
	Expr Expr
	As   Expr
}

func (w *With) isStmt() {}

func (w *With) GetSpan() lexer.Span {
	return w.Span
}

func (w *With) Accept(visitor Visitor) {
	visitor.VisitWith(w)
}

func (w *With) String() string {
	return fmt.Sprintf("with %s", w.Items)
}
