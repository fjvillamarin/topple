package nodes

import (
	"biscuit/compiler/lexer"
	"fmt"
)

// AnnotationStmt represents a variable annotation like 'x: int' or 'x: int = 5'.
type AnnotationStmt struct {
	Target   Expr // The variable or target being annotated
	Type     Expr // The type annotation expression
	Value    Expr // Optional initializer value (can be nil)
	HasValue bool // Whether an initializer value is present

	span lexer.Span
}

// NewAnnotationStmt creates a new annotation statement.
func NewAnnotationStmt(target Expr, typeExpr Expr, value Expr, hasValue bool, span lexer.Span) *AnnotationStmt {
	return &AnnotationStmt{
		Target:   target,
		Type:     typeExpr,
		Value:    value,
		HasValue: hasValue,
		span:     span,
	}
}

func (a *AnnotationStmt) isStmt() {}

func (a *AnnotationStmt) Span() lexer.Span {
	return a.span
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
