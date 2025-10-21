package ast

import (
	"topple/compiler/lexer"
)

// TupleExpr represents a tuple expression (items)
type TupleExpr struct {
	Elements []Expr

	Span lexer.Span
}

func (t *TupleExpr) isExpr() {}

func (t *TupleExpr) GetSpan() lexer.Span {
	return t.Span
}

func (t *TupleExpr) String() string {
	return "(...)"
}

// Accept calls the VisitTupleExpr method on the visitor
func (t *TupleExpr) Accept(visitor Visitor) {
	visitor.VisitTupleExpr(t)
}
