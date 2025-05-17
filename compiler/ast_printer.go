package compiler

import (
	"fmt"
	"strings"
)

// ASTPrinter implements a visitor that prints the AST in a tree-sitter-like format.
type ASTPrinter struct {
	result      strings.Builder
	indentLevel int
	indentStr   string
}

// NewASTPrinter creates a new ASTPrinter with the specified indent string.
func NewASTPrinter(indentStr string) *ASTPrinter {
	return &ASTPrinter{
		indentStr: indentStr,
	}
}

// formatSpan formats the position span for a node
func formatSpan(node Node) string {
	return fmt.Sprintf("%s-%s", node.Start(), node.End())
}

// indent returns the current indentation string
func (p *ASTPrinter) indent() string {
	return strings.Repeat(p.indentStr, p.indentLevel)
}

// printNodeStart prints the common start of a node representation
func (p *ASTPrinter) printNodeStart(nodeType string, node Node) {
	p.result.WriteString(fmt.Sprintf("%s%s [%s]", p.indent(), nodeType, formatSpan(node)))
}

// Visit implements the visitor pattern entry point
func (p *ASTPrinter) Visit(node Node) Visitor {
	if node == nil {
		return nil
	}

	// Let the node's Accept method call the appropriate VisitX method
	node.Accept(p)
	return p
}

// VisitModule handles Module nodes
func (p *ASTPrinter) VisitModule(node *Module) Visitor {
	p.printNodeStart("Module", node)
	p.result.WriteString(" (\n")

	p.indentLevel++
	// Visit all statements in the module body
	for _, stmt := range node.Body {
		if stmt != nil {
			stmt.Accept(p)
		}
	}
	p.indentLevel--

	p.result.WriteString(fmt.Sprintf("%s)\n", p.indent()))
	return p
}

// VisitExprStmt handles ExprStmt nodes
func (p *ASTPrinter) VisitExprStmt(node *ExprStmt) Visitor {
	p.printNodeStart("ExprStmt", node)
	p.result.WriteString(" (\n")

	p.indentLevel++
	// Visit the expression inside this statement
	if node.Value != nil {
		node.Value.Accept(p)
	}
	p.indentLevel--

	p.result.WriteString(fmt.Sprintf("%s)\n", p.indent()))
	return p
}

// VisitName handles Name nodes
func (p *ASTPrinter) VisitName(node *Name) Visitor {
	p.printNodeStart("Name", node)
	p.result.WriteString(fmt.Sprintf(" (%s)\n", node.String()))
	return p
}

// VisitConstant handles Constant nodes
func (p *ASTPrinter) VisitConstant(node *Constant) Visitor {
	var typeStr string
	switch node.Tok.Type {
	case String:
		typeStr = "String"
	case Number:
		typeStr = "Number"
	default:
		typeStr = "Literal"
	}

	p.printNodeStart(typeStr, node)
	p.result.WriteString(fmt.Sprintf(" (%s)\n", node.String()))
	return p
}

// Print visits the AST starting from the given node and returns the string representation.
func (p *ASTPrinter) Print(node Node) string {
	p.result.Reset()
	p.indentLevel = 0

	// Start visitor pattern
	node.Accept(p)

	return p.result.String()
}
