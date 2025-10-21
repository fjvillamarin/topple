package ast

import (
	"topple/compiler/lexer"
)

// Name represents an identifier expression.
type Name struct {
	Token lexer.Token

	Span lexer.Span
}

func (n *Name) isExpr() {}

func (n *Name) GetSpan() lexer.Span {
	return n.Span
}

func (n *Name) String() string {
	return n.Token.Lexeme
}

// Accept calls the VisitName method on the visitor
func (n *Name) Accept(visitor Visitor) {
	visitor.VisitName(n)
}
