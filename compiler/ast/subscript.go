package ast

import (
	"biscuit/compiler/lexer"
	"fmt"
)

// Subscript represents a subscript access expression (obj[index] or obj[start:end:step])
type Subscript struct {
	Object  Expr
	Indices []Expr // Multiple indices or slices

	Span lexer.Span
}

func (s *Subscript) isExpr() {}

func (s *Subscript) GetSpan() lexer.Span {
	return s.Span
}

func (s *Subscript) String() string {
	return fmt.Sprintf("%v[...]", s.Object)
}

// Accept calls the VisitSubscript method on the visitor
func (s *Subscript) Accept(visitor Visitor) {
	visitor.VisitSubscript(s)
}
