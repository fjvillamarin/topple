package compiler

import (
	"fmt"
)

// ExprVisitor is the interface for visitors that traverse expressions.
type ExprVisitor interface {
	VisitName(n *Name) Visitor
	VisitConstant(c *Constant) Visitor
}

// Name represents an identifier expression.
type Name struct {
	BaseNode
	Tok Token
}

func (n *Name) isExpr() {}

func (n *Name) String() string {
	return n.Tok.Lexeme
}

// Accept calls the VisitName method on the visitor
func (n *Name) Accept(visitor Visitor) {
	visitor.VisitName(n)
}

// Constant represents a literal value (number, string, etc.).
type Constant struct {
	BaseNode
	Tok   Token
	Value interface{}
}

func (c *Constant) isExpr() {}

func (c *Constant) String() string {
	return fmt.Sprintf("%v", c.Value)
}

// Accept calls the VisitConstant method on the visitor
func (c *Constant) Accept(visitor Visitor) {
	visitor.VisitConstant(c)
}
