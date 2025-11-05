package ast

import (
	"fmt"
	"strings"
	"topple/compiler/lexer"
)

// ImportStmt represents an 'import' statement.
type ImportStmt struct {
	Names []*ImportName // List of imported modules with optional aliases

	Span lexer.Span
}

func (i *ImportStmt) isStmt() {}

func (i *ImportStmt) GetSpan() lexer.Span {
	return i.Span
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

	Span lexer.Span
}

func NewImportName(dottedName *DottedName, asName *Name, Span lexer.Span) *ImportName {
	return &ImportName{
		DottedName: dottedName,
		AsName:     asName,
		Span:       Span,
	}
}

func (i *ImportName) GetSpan() lexer.Span {
	return i.Span
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

	Span lexer.Span
}

func (d *DottedName) GetSpan() lexer.Span {
	return d.Span
}

func (d *DottedName) String() string {
	parts := make([]string, len(d.Names))
	for i, name := range d.Names {
		parts[i] = name.String()
	}
	return strings.Join(parts, ".")
}

func NewDottedName(names []*Name, Span lexer.Span) *DottedName {
	return &DottedName{
		Names: names,
		Span:  Span,
	}
}
