package ast

import (
	"fmt"
	"sylfie/compiler/lexer"
)

// Lambda represents a lambda expression: lambda x, y=10: x + y
type Lambda struct {
	Parameters *ParameterList // Function parameters (can be nil for no parameters)
	Body       Expr           // The body expression

	Span lexer.Span
}

func (l *Lambda) isExpr() {}

func (l *Lambda) GetSpan() lexer.Span {
	return l.Span
}

func (l *Lambda) String() string {
	if l.Parameters != nil {
		return fmt.Sprintf("lambda %s: %s", l.Parameters.String(), l.Body.String())
	}
	return fmt.Sprintf("lambda: %s", l.Body.String())
}

// Accept calls the VisitLambda method on the visitor
func (l *Lambda) Accept(visitor Visitor) {
	visitor.VisitLambda(l)
}
