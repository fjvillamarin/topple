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

// indent returns the current indentation string
func (p *ASTPrinter) indent() string {
	return strings.Repeat(p.indentStr, p.indentLevel)
}

// printNodeStart prints the common start of a node representation
func (p *ASTPrinter) printNodeStart(nodeType string, node Node) {
	p.result.WriteString(fmt.Sprintf("%s%s [%s]", p.indent(), nodeType, node.Span()))
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

// VisitLiteral handles Literal nodes
func (p *ASTPrinter) VisitLiteral(node *Literal) Visitor {
	var typeStr string
	switch node.Token.Type {
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

// VisitAttribute handles Attribute nodes
func (p *ASTPrinter) VisitAttribute(node *Attribute) Visitor {
	p.printNodeStart("Attribute", node)
	p.result.WriteString(" (\n")

	p.indentLevel++
	// Visit the object expression
	if node.Object != nil {
		p.result.WriteString(fmt.Sprintf("%sobject:\n", p.indent()))
		p.indentLevel++
		node.Object.Accept(p)
		p.indentLevel--
	}

	// Display the attribute name
	p.result.WriteString(fmt.Sprintf("%sattribute: %s\n", p.indent(), node.Name.Lexeme))
	p.indentLevel--

	p.result.WriteString(fmt.Sprintf("%s)\n", p.indent()))
	return p
}

// VisitCall handles Call nodes
func (p *ASTPrinter) VisitCall(node *Call) Visitor {
	p.printNodeStart("Call", node)
	p.result.WriteString(" (\n")

	p.indentLevel++
	// Visit the function expression
	if node.Callee != nil {
		p.result.WriteString(fmt.Sprintf("%sfunction:\n", p.indent()))
		p.indentLevel++
		node.Callee.Accept(p)
		p.indentLevel--
	}

	// Visit the arguments
	if len(node.Arguments) > 0 {
		p.result.WriteString(fmt.Sprintf("%sarguments:\n", p.indent()))
		p.indentLevel++
		for i, arg := range node.Arguments {
			if arg != nil {
				p.result.WriteString(fmt.Sprintf("%sarg %d:\n", p.indent(), i))
				p.indentLevel++
				arg.Accept(p)
				p.indentLevel--
			}
		}
		p.indentLevel--
	}
	p.indentLevel--

	p.result.WriteString(fmt.Sprintf("%s)\n", p.indent()))
	return p
}

// VisitSubscript handles Subscript nodes
func (p *ASTPrinter) VisitSubscript(node *Subscript) Visitor {
	p.printNodeStart("Subscript", node)
	p.result.WriteString(" (\n")

	p.indentLevel++
	// Visit the object expression
	if node.Object != nil {
		p.result.WriteString(fmt.Sprintf("%sobject:\n", p.indent()))
		p.indentLevel++
		node.Object.Accept(p)
		p.indentLevel--
	}

	// Visit the index expression
	if node.Index != nil {
		p.result.WriteString(fmt.Sprintf("%sindex:\n", p.indent()))
		p.indentLevel++
		node.Index.Accept(p)
		p.indentLevel--
	}
	p.indentLevel--

	p.result.WriteString(fmt.Sprintf("%s)\n", p.indent()))
	return p
}

// VisitBinary handles Binary nodes
func (p *ASTPrinter) VisitBinary(node *Binary) Visitor {
	p.printNodeStart("Binary", node)
	p.result.WriteString(" (\n")

	p.indentLevel++
	// Visit the left expression
	if node.Left != nil {
		p.result.WriteString(fmt.Sprintf("%sleft:\n", p.indent()))
		p.indentLevel++
		node.Left.Accept(p)
		p.indentLevel--
	}

	// Visit the operator
	p.result.WriteString(fmt.Sprintf("%soperator: %s\n", p.indent(), node.Operator.Lexeme))

	// Visit the right expression
	if node.Right != nil {
		p.result.WriteString(fmt.Sprintf("%sright:\n", p.indent()))
		p.indentLevel++
		node.Right.Accept(p)
		p.indentLevel--
	}
	p.indentLevel--

	p.result.WriteString(fmt.Sprintf("%s)\n", p.indent()))
	return p
}

// VisitUnary handles Unary nodes
func (p *ASTPrinter) VisitUnary(node *Unary) Visitor {
	p.printNodeStart("Unary", node)
	p.result.WriteString(" (\n")

	// Visit the operator
	p.result.WriteString(fmt.Sprintf("%soperator: %s\n", p.indent(), node.Operator.Lexeme))

	p.indentLevel++
	// Visit the right expression
	if node.Right != nil {
		p.result.WriteString(fmt.Sprintf("%sright:\n", p.indent()))
		p.indentLevel++
		node.Right.Accept(p)
		p.indentLevel--
	}
	p.indentLevel--

	p.result.WriteString(fmt.Sprintf("%s)\n", p.indent()))
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
