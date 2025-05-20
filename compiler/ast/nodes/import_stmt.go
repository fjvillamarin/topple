package nodes

import (
	"biscuit/compiler/lexer"
	"fmt"
	"strings"
)

// ImportStmt represents an 'import' statement.
type ImportStmt struct {
	Names []*ImportName // List of imported modules with optional aliases

	span lexer.Span
}

func NewImportStmt(names []*ImportName, span lexer.Span) *ImportStmt {
	return &ImportStmt{
		Names: names,
		span:  span,
	}
}

func (i *ImportStmt) isStmt() {}

func (i *ImportStmt) Span() lexer.Span {
	return i.span
}

func (i *ImportStmt) Accept(visitor Visitor) {
	visitor.VisitImportStmt(i)
}

func (i *ImportStmt) String() string {
	names := make([]string, len(i.Names))
	for j, name := range i.Names {
		names[j] = name.String()
	}
	return fmt.Sprintf("ImportStmt(%s)", strings.Join(names, ", "))
}

// ImportName represents a single module import with optional alias.
type ImportName struct {
	DottedName *DottedName // The module path
	AsName     *Name       // Optional alias name

	span lexer.Span
}

func NewImportName(dottedName *DottedName, asName *Name, span lexer.Span) *ImportName {
	return &ImportName{
		DottedName: dottedName,
		AsName:     asName,
		span:       span,
	}
}

func (i *ImportName) Span() lexer.Span {
	return i.span
}

func (i *ImportName) String() string {
	if i.AsName != nil {
		return fmt.Sprintf("%s as %s", i.DottedName, i.AsName)
	}
	return i.DottedName.String()
}

// DottedName represents a dotted module path.
type DottedName struct {
	Names []*Name // Parts of the dotted path

	span lexer.Span
}

func NewDottedName(names []*Name, span lexer.Span) *DottedName {
	return &DottedName{
		Names: names,
		span:  span,
	}
}

func (d *DottedName) Span() lexer.Span {
	return d.span
}

func (d *DottedName) String() string {
	parts := make([]string, len(d.Names))
	for i, name := range d.Names {
		parts[i] = name.String()
	}
	return strings.Join(parts, ".")
}
