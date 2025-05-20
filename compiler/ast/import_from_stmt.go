package ast

import (
	"biscuit/compiler/lexer"
	"fmt"
	"strings"
)

// ImportFromStmt represents a 'from ... import ...' statement.
type ImportFromStmt struct {
	DottedName *DottedName   // The module path to import from (may be nil for relative imports)
	DotCount   int           // Number of leading dots for relative imports
	Names      []*ImportName // List of imported names with optional aliases
	IsWildcard bool          // True if importing '*'

	span lexer.Span
}

func NewImportFromStmt(dottedName *DottedName, dotCount int, names []*ImportName, isWildcard bool, span lexer.Span) *ImportFromStmt {
	return &ImportFromStmt{
		DottedName: dottedName,
		DotCount:   dotCount,
		Names:      names,
		IsWildcard: isWildcard,
		span:       span,
	}
}

func (i *ImportFromStmt) isStmt() {}

func (i *ImportFromStmt) Span() lexer.Span {
	return i.span
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
