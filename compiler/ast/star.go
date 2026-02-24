package ast

import (
	"fmt"
	"github.com/fjvillamarin/topple/compiler/lexer"
)

// StarExpr represents a star expression (*expr)
type StarExpr struct {
	Expr Expr

	Span lexer.Span
}

func (s *StarExpr) isExpr() {}

func (s *StarExpr) GetSpan() lexer.Span {
	return s.Span
}

func (s *StarExpr) String() string {
	return fmt.Sprintf("*%v", s.Expr)
}

// Accept calls the VisitStarExpr method on the visitor
func (s *StarExpr) Accept(visitor Visitor) {
	visitor.VisitStarExpr(s)
}
