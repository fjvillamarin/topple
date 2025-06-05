package ast

import (
	"fmt"
	"sylfie/compiler/lexer"
)

// TypeParam represents a type parameter expression
type TypeParam struct {
	Name         lexer.Token
	Bound        Expr // Optional bound (: expression)
	Default      Expr // Optional default (= expression)
	IsStar       bool // Whether this is a *NAME parameter
	IsDoubleStar bool // Whether this is a **NAME parameter

	Span lexer.Span
}

func (t *TypeParam) isExpr() {}

func (t *TypeParam) GetSpan() lexer.Span {
	return t.Span
}

func (t *TypeParam) String() string {
	prefix := ""
	if t.IsStar {
		prefix = "*"
	} else if t.IsDoubleStar {
		prefix = "**"
	}
	return fmt.Sprintf("%s%s", prefix, t.Name.Lexeme)
}

// Accept calls the VisitTypeParamExpr method on the visitor
func (t *TypeParam) Accept(visitor Visitor) {
	visitor.VisitTypeParamExpr(t)
}
