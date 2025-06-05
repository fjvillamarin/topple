package ast

import (
	"fmt"
	"strings"
	"sylfie/compiler/lexer"
)

// ImportFromStmt represents a 'from ... import ...' statement.
type ImportFromStmt struct {
	DottedName *DottedName   // The module path to import from (may be nil for relative imports)
	DotCount   int           // Number of leading dots for relative imports
	Names      []*ImportName // List of imported names with optional aliases
	IsWildcard bool          // True if importing '*'

	Span lexer.Span
}

func (i *ImportFromStmt) isStmt() {}

func (i *ImportFromStmt) GetSpan() lexer.Span {
	return i.Span
}

func (i *ImportFromStmt) Accept(visitor Visitor) {
	visitor.VisitImportFromStmt(i)
}

func (i *ImportFromStmt) String() string {
	var module string
	if i.DottedName != nil {
		module = i.DottedName.String()
	} else {
		module = strings.Repeat(".", i.DotCount)
	}

	if i.IsWildcard {
		return fmt.Sprintf("ImportFromStmt(%s, *)", module)
	}

	names := make([]string, len(i.Names))
	for j, name := range i.Names {
		names[j] = name.String()
	}
	return fmt.Sprintf("ImportFromStmt(%s, [%s])", module, strings.Join(names, ", "))
}
