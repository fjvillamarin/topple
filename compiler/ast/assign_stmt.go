package ast

import (
	"fmt"
	"github.com/fjvillamarin/topple/compiler/lexer"
	"strings"
)

// AssignStmt represents an assignment statement like 'x = y' or 'a, b = c, d'.
type AssignStmt struct {
	Targets []Expr // Left-hand side targets (can be multiple for unpacking)
	Value   Expr   // Right-hand side expression

	Span lexer.Span
}

func (a *AssignStmt) isStmt() {}

func (a *AssignStmt) GetSpan() lexer.Span {
	return a.Span
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
