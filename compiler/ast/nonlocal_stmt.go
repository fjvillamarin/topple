package ast

import (
	"biscuit/compiler/lexer"
	"fmt"
	"strings"
)

// NonlocalStmt represents a 'nonlocal' statement.
type NonlocalStmt struct {
	Names []*Name // The identifiers declared as nonlocal

	span lexer.Span
}

func NewNonlocalStmt(names []*Name, span lexer.Span) *NonlocalStmt {
	return &NonlocalStmt{
		Names: names,
		span:  span,
	}
}

func (n *NonlocalStmt) isStmt() {}

func (n *NonlocalStmt) Span() lexer.Span {
	return n.span
}

func (n *NonlocalStmt) Accept(visitor Visitor) {
	visitor.VisitNonlocalStmt(n)
}

func (n *NonlocalStmt) String() string {
	names := make([]string, len(n.Names))
	for i, name := range n.Names {
		names[i] = name.String()
	}
	return fmt.Sprintf("NonlocalStmt(%s)", strings.Join(names, ", "))
}
