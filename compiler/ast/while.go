package ast

import (
	"biscuit/compiler/lexer"
	"fmt"
)

type While struct {
	Condition Expr
	Body      []Stmt
	Else      []Stmt

	Span lexer.Span
}

func (w *While) isStmt() {}

func (w *While) GetSpan() lexer.Span {
	return w.Span
}

func (w *While) Accept(visitor Visitor) {
	visitor.VisitWhile(w)
}

func (w *While) String() string {
	return fmt.Sprintf("while %s", w.Condition.String())
}
