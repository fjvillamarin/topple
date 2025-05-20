package ast

import (
	"biscuit/compiler/lexer"
	"fmt"
)

// TypeParamExpr represents a type parameter expression
type TypeParamExpr struct {
	Name         lexer.Token
	Bound        Expr // Optional bound (: expression)
	Default      Expr // Optional default (= expression)
	IsStar       bool // Whether this is a *NAME parameter
	IsDoubleStar bool // Whether this is a **NAME parameter

	span lexer.Span
}

func NewTypeParamExpr(name lexer.Token, bound Expr, defaultValue Expr, isStar bool, isDoubleStar bool,
	span lexer.Span) *TypeParamExpr {
	return &TypeParamExpr{
		Name:         name,
		Bound:        bound,
		Default:      defaultValue,
		IsStar:       isStar,
		IsDoubleStar: isDoubleStar,

		span: span,
	}
}

func (t *TypeParamExpr) isExpr() {}

func (t *TypeParamExpr) Span() lexer.Span {
	return t.span
}

func (t *TypeParamExpr) String() string {
	prefix := ""
	if t.IsStar {
		prefix = "*"
	} else if t.IsDoubleStar {
		prefix = "**"
	}
	return fmt.Sprintf("%s%s", prefix, t.Name.Lexeme)
}

// Accept calls the VisitTypeParamExpr method on the visitor
func (t *TypeParamExpr) Accept(visitor Visitor) {
	visitor.VisitTypeParamExpr(t)
}
