package ast

import (
	"biscuit/compiler/lexer"
	"fmt"
)

// Subscript represents a subscript access expression (obj[index] or obj[start:end:step])
type Subscript struct {
	Object  Expr
	Indices []Expr // Multiple indices or slices

	span lexer.Span
}

func NewSubscript(object Expr, indices []Expr, span lexer.Span) *Subscript {
	return &Subscript{
		Object:  object,
		Indices: indices,

		span: span,
	}
}

func (s *Subscript) isExpr() {}

func (s *Subscript) Span() lexer.Span {
	return s.span
}

func (s *Subscript) String() string {
	return fmt.Sprintf("%v[...]", s.Object)
}

// Accept calls the VisitSubscript method on the visitor
func (s *Subscript) Accept(visitor Visitor) {
	visitor.VisitSubscript(s)
}
