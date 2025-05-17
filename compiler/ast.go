package compiler

import "fmt"

// Node is the base interface for all AST nodes.
type Node interface {
	Accept(visitor Visitor)

	Start() Position
	End() Position
	Span() string
	String() string
}

// Expr is the interface for all expression nodes.
type Expr interface {
	Node
	isExpr()
} // only *expressions* satisfy this

// Stmt is the interface for all statement nodes.
type Stmt interface {
	Node
	isStmt()
} // only *statements* satisfy this

// Visitor is the interface for visitors that traverse the AST.
type Visitor interface {
	// Visit is the entry point for the visitor pattern
	Visit(node Node) Visitor

	// Specific visitors
	ExprVisitor
	StmtVisitor
}

// BaseNode is the base implementation of the Node interface.
type BaseNode struct {
	StartPos Position
	EndPos   Position
}

func (n *BaseNode) Start() Position {
	return n.StartPos
}

func (n *BaseNode) End() Position {
	return n.EndPos
}

func (n *BaseNode) Span() string {
	return fmt.Sprintf("%s-%s", n.StartPos, n.EndPos)
}

func (n *BaseNode) Accept(visitor Visitor) {
	visitor.Visit(n)
}

func (n *BaseNode) String() string {
	return fmt.Sprintf("Node(%s)", n.Span())
}
