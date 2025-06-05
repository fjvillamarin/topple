package ast

import (
	"sylfie/compiler/lexer"
	"fmt"
)

// AnnotationStmt represents a variable annotation like 'x: int' or 'x: int = 5'.
type AnnotationStmt struct {
	Target   Expr // The variable or target being annotated
	Type     Expr // The type annotation expression
	Value    Expr // Optional initializer value (can be nil)
	HasValue bool // Whether an initializer value is present

	Span lexer.Span
}

func (a *AnnotationStmt) isStmt() {}

func (a *AnnotationStmt) GetSpan() lexer.Span {
	return a.Span
}

func (a *AnnotationStmt) Accept(visitor Visitor) {
	visitor.VisitAnnotationStmt(a)
}

func (a *AnnotationStmt) String() string {
	if a.HasValue {
		return fmt.Sprintf("AnnotationStmt(%s: %s = %s)", a.Target, a.Type, a.Value)
	}
	return fmt.Sprintf("AnnotationStmt(%s: %s)", a.Target, a.Type)
}
