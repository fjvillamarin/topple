package ast

import (
	"biscuit/compiler/lexer"
	"fmt"
	"strings"
)

// GlobalStmt represents a 'global' statement.
type GlobalStmt struct {
	Names []*Name // The identifiers declared as global

	span lexer.Span
}

func NewGlobalStmt(names []*Name, span lexer.Span) *GlobalStmt {
	return &GlobalStmt{
		Names: names,
		span:  span,
	}
}

func (g *GlobalStmt) isStmt() {}

func (g *GlobalStmt) Span() lexer.Span {
	return g.span
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
