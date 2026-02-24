package ast

import (
	"fmt"
	"github.com/fjvillamarin/topple/compiler/lexer"
)

// Binary represents a binary operation expression (left op right)
type Binary struct {
	Left     Expr
	Operator lexer.Token
	Right    Expr

	Span lexer.Span
}

func (b *Binary) isExpr() {}

func (b *Binary) GetSpan() lexer.Span {
	return b.Span
}

func (b *Binary) String() string {
	return fmt.Sprintf("%v %s %v", b.Left, b.Operator.Lexeme, b.Right)
}

// Accept calls the VisitBinary method on the visitor
func (b *Binary) Accept(visitor Visitor) {
	visitor.VisitBinary(b)
}
