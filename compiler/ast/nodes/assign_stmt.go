package nodes

import (
	"biscuit/compiler/lexer"
	"fmt"
	"strings"
)

// AssignStmt represents an assignment statement like 'x = y' or 'a, b = c, d'.
type AssignStmt struct {
	Targets []Expr // Left-hand side targets (can be multiple for unpacking)
	Value   Expr   // Right-hand side expression

	span lexer.Span
}

// NewAssignStmt creates a new assignment statement.
func NewAssignStmt(targets []Expr, value Expr, span lexer.Span) *AssignStmt {
	return &AssignStmt{
		Targets: targets,
		Value:   value,
		span:    span,
	}
}

func (a *AssignStmt) isStmt() {}

func (a *AssignStmt) Span() lexer.Span {
	return a.span
}

func (a *AssignStmt) Accept(visitor Visitor) {
	visitor.VisitAssignStmt(a)
}

func (a *AssignStmt) String() string {
	var targetStrs []string
	for _, target := range a.Targets {
		targetStrs = append(targetStrs, target.String())
	}
	return fmt.Sprintf("AssignStmt(%s = %s)", strings.Join(targetStrs, ", "), a.Value)
}
