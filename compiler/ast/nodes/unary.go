package nodes

import (
	"biscuit/compiler/lexer"
	"fmt"
)

// Unary represents a unary operation expression (-expr)
type Unary struct {
	Operator lexer.Token
	Right    Expr

	span lexer.Span
}

func NewUnary(operator lexer.Token, right Expr, span lexer.Span) *Unary {
	return &Unary{
		Operator: operator,
		Right:    right,

		span: span,
	}
}

func (u *Unary) isExpr() {}

func (u *Unary) Span() lexer.Span {
	return u.span
}

func (u *Unary) String() string {
	return fmt.Sprintf("%s %v", u.Operator.Lexeme, u.Right)
}

// Accept calls the VisitUnary method on the visitor
func (u *Unary) Accept(visitor Visitor) {
	visitor.VisitUnary(u)
}
