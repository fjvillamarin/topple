package nodes

import (
	"biscuit/compiler/lexer"
	"fmt"
)

// AugAssignStmt represents an augmented assignment statement like 'x += y'.
type AugAssignStmt struct {
	Target   Expr        // Left-hand side target
	Operator lexer.Token // The augmented assignment operator
	Value    Expr        // Right-hand side expression

	span lexer.Span
}

// NewAugAssignStmt creates a new augmented assignment statement.
func NewAugAssignStmt(target Expr, operator lexer.Token, value Expr, span lexer.Span) *AugAssignStmt {
	return &AugAssignStmt{
		Target:   target,
		Operator: operator,
		Value:    value,
		span:     span,
	}
}

func (a *AugAssignStmt) isStmt() {}

func (a *AugAssignStmt) Span() lexer.Span {
	return a.span
}

func (a *AugAssignStmt) Accept(visitor Visitor) {
	visitor.VisitAugAssignStmt(a)
}

func (a *AugAssignStmt) String() string {
	return fmt.Sprintf("AugAssignStmt(%s %s %s)", a.Target, a.Operator.Lexeme, a.Value)
}
