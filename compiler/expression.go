package compiler

import (
	"fmt"
)

// ExprVisitor is the interface for visitors that traverse expressions.
type ExprVisitor interface {
	VisitName(n *Name) Visitor
	VisitLiteral(l *Literal) Visitor
	VisitAttribute(a *Attribute) Visitor
	VisitCall(c *Call) Visitor
	VisitSubscript(s *Subscript) Visitor
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

// Literal represents a literal value (number, string, etc.).
type Literal struct {
	BaseNode
	Token Token
	Value any
}

func NewLiteral(token Token, value any, startPos Position, endPos Position) *Literal {
	return &Literal{
		BaseNode: BaseNode{
			StartPos: startPos,
			EndPos:   endPos,
		},
		Token: token,
		Value: value,
	}
}

func (l *Literal) isExpr() {}

func (l *Literal) String() string {
	return fmt.Sprintf("%v", l.Value)
}

// Accept calls the VisitLiteral method on the visitor
func (l *Literal) Accept(visitor Visitor) {
	visitor.VisitLiteral(l)
}

// Attribute represents an attribute access expression (obj.attr)
type Attribute struct {
	BaseNode
	Object Expr
	Name   Token
}

func NewAttribute(object Expr, name Token, startPos Position, endPos Position) *Attribute {
	return &Attribute{
		BaseNode: BaseNode{
			StartPos: startPos,
			EndPos:   endPos,
		},
		Object: object,
		Name:   name,
	}
}

func (a *Attribute) isExpr() {}

func (a *Attribute) String() string {
	return fmt.Sprintf("%v.%s", a.Object, a.Name.Lexeme)
}

// Accept calls the VisitAttribute method on the visitor
func (a *Attribute) Accept(visitor Visitor) {
	visitor.VisitAttribute(a)
}

// Call represents a function call expression (func(args))
type Call struct {
	BaseNode
	Callee    Expr
	Arguments []Expr
}

func NewCall(callee Expr, arguments []Expr, startPos Position, endPos Position) *Call {
	return &Call{
		BaseNode: BaseNode{
			StartPos: startPos,
			EndPos:   endPos,
		},
		Callee:    callee,
		Arguments: arguments,
	}
}

func (c *Call) isExpr() {}

func (c *Call) String() string {
	return fmt.Sprintf("%v()", c.Callee)
}

// Accept calls the VisitCall method on the visitor
func (c *Call) Accept(visitor Visitor) {
	visitor.VisitCall(c)
}

// Subscript represents a subscript access expression (obj[index])
type Subscript struct {
	BaseNode
	Object Expr
	Index  Expr
}

func NewSubscript(object Expr, index Expr, startPos Position, endPos Position) *Subscript {
	return &Subscript{
		BaseNode: BaseNode{
			StartPos: startPos,
			EndPos:   endPos,
		},
		Object: object,
		Index:  index,
	}
}

func (s *Subscript) isExpr() {}

func (s *Subscript) String() string {
	return fmt.Sprintf("%v[%v]", s.Object, s.Index)
}

// Accept calls the VisitSubscript method on the visitor
func (s *Subscript) Accept(visitor Visitor) {
	visitor.VisitSubscript(s)
}
