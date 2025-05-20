package nodes

import (
	"biscuit/compiler/lexer"
	"fmt"
)

// Binary represents a binary operation expression (left op right)
type Binary struct {
	Left     Expr
	Operator lexer.Token
	Right    Expr

	span lexer.Span
}

func NewBinary(left Expr, operator lexer.Token, right Expr, span lexer.Span) *Binary {
	return &Binary{
		Left:     left,
		Operator: operator,
		Right:    right,

		span: span,
	}
}

func (b *Binary) isExpr() {}

func (b *Binary) Span() lexer.Span {
	return b.span
}

func (b *Binary) String() string {
	return fmt.Sprintf("%v %s %v", b.Left, b.Operator.Lexeme, b.Right)
}

// Accept calls the VisitBinary method on the visitor
func (b *Binary) Accept(visitor Visitor) {
	visitor.VisitBinary(b)
}
