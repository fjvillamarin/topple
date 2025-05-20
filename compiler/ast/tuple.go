package ast

import (
	"biscuit/compiler/lexer"
	"fmt"
)

// TupleExpr represents a tuple expression (items)
type TupleExpr struct {
	Elements []Expr

	span lexer.Span
}

func NewTupleExpr(elements []Expr, span lexer.Span) *TupleExpr {
	return &TupleExpr{
		Elements: elements,

		span: span,
	}
}

func (t *TupleExpr) isExpr() {}

func (t *TupleExpr) Span() lexer.Span {
	return t.span
}

func (t *TupleExpr) String() string {
	return fmt.Sprintf("(...)")
}

// Accept calls the VisitTupleExpr method on the visitor
func (t *TupleExpr) Accept(visitor Visitor) {
	visitor.VisitTupleExpr(t)
}
