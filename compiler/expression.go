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
	Token Token
}

func NewName(token Token, startPos Position, endPos Position) *Name {
	return &Name{
		BaseNode: BaseNode{
			StartPos: startPos,
			EndPos:   endPos,
		},
		Token: token,
	}
}

func (n *Name) isExpr() {}

func (n *Name) String() string {
	return n.Token.Lexeme
}

// Accept calls the VisitName method on the visitor
func (n *Name) Accept(visitor Visitor) {
	visitor.VisitName(n)
}

// Constant represents a literal value (number, string, etc.).
type Constant struct {
	BaseNode
	Token Token
	Value any
}

func NewConstant(token Token, value any, startPos Position, endPos Position) *Constant {
	return &Constant{
		BaseNode: BaseNode{
			StartPos: startPos,
			EndPos:   endPos,
		},
		Token: token,
		Value: value,
	}
}

func (c *Constant) isExpr() {}

func (c *Constant) String() string {
	return fmt.Sprintf("%v", c.Value)
}

// Accept calls the VisitConstant method on the visitor
func (c *Constant) Accept(visitor Visitor) {
	visitor.VisitConstant(c)
}
