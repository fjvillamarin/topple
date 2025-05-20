package nodes

import (
	"biscuit/compiler/lexer"
)

// Name represents an identifier expression.
type Name struct {
	Token lexer.Token

	span lexer.Span
}

func NewName(token lexer.Token, span lexer.Span) *Name {
	return &Name{
		Token: token,

		span: span,
	}
}

func (n *Name) isExpr() {}

func (n *Name) Span() lexer.Span {
	return n.span
}

func (n *Name) String() string {
	return n.Token.Lexeme
}

// Accept calls the VisitName method on the visitor
func (n *Name) Accept(visitor Visitor) {
	visitor.VisitName(n)
}
