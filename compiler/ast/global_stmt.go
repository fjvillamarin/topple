package ast

import (
	"biscuit/compiler/lexer"
	"fmt"
	"strings"
)

// GlobalStmt represents a 'global' statement.
type GlobalStmt struct {
	Names []*Name // The identifiers declared as global

	Span lexer.Span
}

func (g *GlobalStmt) isStmt() {}

func (g *GlobalStmt) GetSpan() lexer.Span {
	return g.Span
}

func (g *GlobalStmt) Accept(visitor Visitor) {
	visitor.VisitGlobalStmt(g)
}

func (g *GlobalStmt) String() string {
	names := make([]string, len(g.Names))
	for i, name := range g.Names {
		names[i] = name.String()
	}
	return fmt.Sprintf("GlobalStmt(%s)", strings.Join(names, ", "))
}
