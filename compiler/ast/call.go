package ast

import (
	"biscuit/compiler/lexer"
	"fmt"
)

// Call represents a function call expression (func(args))
type Call struct {
	Callee    Expr
	Arguments []Expr

	span lexer.Span
}

func NewCall(callee Expr, arguments []Expr, span lexer.Span) *Call {
	return &Call{
		Callee:    callee,
		Arguments: arguments,

		span: span,
	}
}

func (c *Call) isExpr() {}

func (c *Call) Span() lexer.Span {
	return c.span
}

func (c *Call) String() string {
	return fmt.Sprintf("%v()", c.Callee)
}

// Accept calls the VisitCall method on the visitor
func (c *Call) Accept(visitor Visitor) {
	visitor.VisitCall(c)
}
