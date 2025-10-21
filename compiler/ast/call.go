package ast

import (
	"fmt"
	"topple/compiler/lexer"
)

// Call represents a function call expression (func(args))
type Call struct {
	Callee    Expr
	Arguments []*Argument

	Span lexer.Span
}

func (c *Call) isExpr() {}

func (c *Call) GetSpan() lexer.Span {
	return c.Span
}

func (c *Call) String() string {
	return fmt.Sprintf("%v()", c.Callee)
}

// Accept calls the VisitCall method on the visitor
func (c *Call) Accept(visitor Visitor) {
	visitor.VisitCall(c)
}
