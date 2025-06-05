package ast

import (
	"fmt"
	"sylfie/compiler/lexer"
)

// Unary represents a unary operation expression (-expr)
type Unary struct {
	Operator lexer.Token
	Right    Expr

	Span lexer.Span
}

func (u *Unary) isExpr() {}

func (u *Unary) GetSpan() lexer.Span {
	return u.Span
}

func (u *Unary) String() string {
	return fmt.Sprintf("%s %v", u.Operator.Lexeme, u.Right)
}

// Accept calls the VisitUnary method on the visitor
func (u *Unary) Accept(visitor Visitor) {
	visitor.VisitUnary(u)
}
