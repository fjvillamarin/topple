package ast

import "strings"

// ForIfClause represents a for clause with optional if conditions in comprehensions
// This is not an AST node but a data structure used within comprehensions
type ForIfClause struct {
	IsAsync bool   // true if this is an "async for"
	Target  Expr   // The target variables (e.g., x in "for x in items")
	Iter    Expr   // The iterable expression (e.g., items in "for x in items")
	Ifs     []Expr // Zero or more if conditions
}

// String returns a string representation of the ForIfClause
func (fc *ForIfClause) String() string {
	var result strings.Builder

	if fc.IsAsync {
		result.WriteString("async ")
	}
	result.WriteString("for ")
	result.WriteString(fc.Target.String())
	result.WriteString(" in ")
	result.WriteString(fc.Iter.String())

	for _, ifCond := range fc.Ifs {
		result.WriteString(" if ")
		result.WriteString(ifCond.String())
	}

	return result.String()
}
