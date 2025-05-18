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

// VisitAssignExpr handles AssignExpr nodes
func (p *ASTPrinter) VisitAssignExpr(node *AssignExpr) Visitor {
	p.printNodeStart("AssignExpr", node)
	p.result.WriteString(" (\n")

	p.indentLevel++
	// Visit the left expression
	if node.Left != nil {
		p.result.WriteString(fmt.Sprintf("%sleft:\n", p.indent()))
		p.indentLevel++
		node.Left.Accept(p)
		p.indentLevel--
	}

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

// VisitStarExpr handles StarExpr nodes
func (p *ASTPrinter) VisitStarExpr(node *StarExpr) Visitor {
	p.printNodeStart("StarExpr", node)
	p.result.WriteString(" (\n")

	p.indentLevel++
	// Visit the expression
	if node.Expr != nil {
		p.result.WriteString(fmt.Sprintf("%sexpr:\n", p.indent()))
		p.indentLevel++
		node.Expr.Accept(p)
		p.indentLevel--
	}
	p.indentLevel--

	p.result.WriteString(fmt.Sprintf("%s)\n", p.indent()))
	return p
}

// VisitTernaryExpr handles TernaryExpr nodes
func (p *ASTPrinter) VisitTernaryExpr(node *TernaryExpr) Visitor {
	p.printNodeStart("TernaryExpr", node)
	p.result.WriteString(" (\n")

	p.indentLevel++
	// Visit the condition expression
	if node.Condition != nil {
		p.result.WriteString(fmt.Sprintf("%scondition:\n", p.indent()))
		p.indentLevel++
		node.Condition.Accept(p)
		p.indentLevel--
	}

	// Visit the true expression
	if node.TrueExpr != nil {
		p.result.WriteString(fmt.Sprintf("%strue:\n", p.indent()))
		p.indentLevel++
		node.TrueExpr.Accept(p)
		p.indentLevel--
	}

	// Visit the false expression
	if node.FalseExpr != nil {
		p.result.WriteString(fmt.Sprintf("%sfalse:\n", p.indent()))
		p.indentLevel++
		node.FalseExpr.Accept(p)
		p.indentLevel--
	}
	p.indentLevel--

	p.result.WriteString(fmt.Sprintf("%s)\n", p.indent()))
	return p
}

// VisitListExpr handles ListExpr nodes
func (p *ASTPrinter) VisitListExpr(node *ListExpr) Visitor {
	p.printNodeStart("ListExpr", node)
	p.result.WriteString(" (\n")

	p.indentLevel++
	// Visit all elements in the list
	if len(node.Elements) > 0 {
		p.result.WriteString(fmt.Sprintf("%selements:\n", p.indent()))
		p.indentLevel++
		for i, elem := range node.Elements {
			if elem != nil {
				p.result.WriteString(fmt.Sprintf("%sitem %d:\n", p.indent(), i))
				p.indentLevel++
				elem.Accept(p)
				p.indentLevel--
			}
		}
		p.indentLevel--
	}
	p.indentLevel--

	p.result.WriteString(fmt.Sprintf("%s)\n", p.indent()))
	return p
}

// VisitTupleExpr handles TupleExpr nodes
func (p *ASTPrinter) VisitTupleExpr(node *TupleExpr) Visitor {
	p.printNodeStart("TupleExpr", node)
	p.result.WriteString(" (\n")

	p.indentLevel++
	// Visit all elements in the tuple
	if len(node.Elements) > 0 {
		p.result.WriteString(fmt.Sprintf("%selements:\n", p.indent()))
		p.indentLevel++
		for i, elem := range node.Elements {
			if elem != nil {
				p.result.WriteString(fmt.Sprintf("%sitem %d:\n", p.indent(), i))
				p.indentLevel++
				elem.Accept(p)
				p.indentLevel--
			}
		}
		p.indentLevel--
	}
	p.indentLevel--

	p.result.WriteString(fmt.Sprintf("%s)\n", p.indent()))
	return p
}

// VisitSetExpr handles SetExpr nodes
func (p *ASTPrinter) VisitSetExpr(node *SetExpr) Visitor {
	p.printNodeStart("SetExpr", node)
	p.result.WriteString(" (\n")

	p.indentLevel++
	// Visit all elements in the set
	if len(node.Elements) > 0 {
		p.result.WriteString(fmt.Sprintf("%selements:\n", p.indent()))
		p.indentLevel++
		for i, elem := range node.Elements {
			if elem != nil {
				p.result.WriteString(fmt.Sprintf("%sitem %d:\n", p.indent(), i))
				p.indentLevel++
				elem.Accept(p)
				p.indentLevel--
			}
		}
		p.indentLevel--
	}
	p.indentLevel--

	p.result.WriteString(fmt.Sprintf("%s)\n", p.indent()))
	return p
}

// VisitYieldExpr handles YieldExpr nodes
func (p *ASTPrinter) VisitYieldExpr(node *YieldExpr) Visitor {
	p.printNodeStart("YieldExpr", node)
	p.result.WriteString(" (\n")

	p.indentLevel++
	// Display whether this is a "yield from" expression
	p.result.WriteString(fmt.Sprintf("%sisFrom: %t\n", p.indent(), node.IsFrom))

	// Visit the yield value if present
	if node.Value != nil {
		p.result.WriteString(fmt.Sprintf("%svalue:\n", p.indent()))
		p.indentLevel++
		node.Value.Accept(p)
		p.indentLevel--
	}
	p.indentLevel--

	p.result.WriteString(fmt.Sprintf("%s)\n", p.indent()))
	return p
}

// VisitGroupExpr handles GroupExpr nodes
func (p *ASTPrinter) VisitGroupExpr(node *GroupExpr) Visitor {
	p.printNodeStart("GroupExpr", node)
	p.result.WriteString(" (\n")

	p.indentLevel++
	// Visit the inner expression
	if node.Expression != nil {
		p.result.WriteString(fmt.Sprintf("%sexpression:\n", p.indent()))
		p.indentLevel++
		node.Expression.Accept(p)
		p.indentLevel--
	}
	p.indentLevel--

	p.result.WriteString(fmt.Sprintf("%s)\n", p.indent()))
	return p
}

// VisitTypeParamExpr handles TypeParamExpr nodes
func (p *ASTPrinter) VisitTypeParamExpr(node *TypeParamExpr) Visitor {
	p.printNodeStart("TypeParamExpr", node)

	// Format the type parameter
	var paramStr string
	if node.IsStar {
		paramStr = "*"
	} else if node.IsDoubleStar {
		paramStr = "**"
	}
	paramStr += node.Name.Lexeme

	p.result.WriteString(fmt.Sprintf(" (%s)", paramStr))

	if node.Bound != nil || node.Default != nil {
		p.result.WriteString(" (\n")
	} else {
		p.result.WriteString("\n")
	}

	p.indentLevel++

	// Display parameter bound if present
	if node.Bound != nil {
		p.result.WriteString(fmt.Sprintf("%sbound:\n", p.indent()))
		p.indentLevel++
		node.Bound.Accept(p)
		p.indentLevel--
	}

	// Display parameter default if present
	if node.Default != nil {
		p.result.WriteString(fmt.Sprintf("%sdefault:\n", p.indent()))
		p.indentLevel++
		node.Default.Accept(p)
		p.indentLevel--
	}

	p.indentLevel--

	if node.Bound != nil || node.Default != nil {
		p.result.WriteString(fmt.Sprintf("%s)\n", p.indent()))
	}
	return p
}

// VisitTypeAlias handles TypeAlias nodes
func (p *ASTPrinter) VisitTypeAlias(node *TypeAlias) Visitor {
	p.printNodeStart("TypeAlias", node)
	p.result.WriteString(" (\n")

	p.indentLevel++
	// Display the type name
	p.result.WriteString(fmt.Sprintf("%sname: %s\n", p.indent(), node.Name.Lexeme))

	// Display type parameters if any
	if len(node.Params) > 0 {
		p.result.WriteString(fmt.Sprintf("%sparameters:\n", p.indent()))
		p.indentLevel++
		for i, param := range node.Params {
			p.result.WriteString(fmt.Sprintf("%sparam %d:\n", p.indent(), i))
			p.indentLevel++
			param.Accept(p)
			p.indentLevel--
		}
		p.indentLevel--
	}

	// Visit the value expression
	if node.Value != nil {
		p.result.WriteString(fmt.Sprintf("%svalue:\n", p.indent()))
		p.indentLevel++
		node.Value.Accept(p)
		p.indentLevel--
	}
	p.indentLevel--

	p.result.WriteString(fmt.Sprintf("%s)\n", p.indent()))
	return p
}

// VisitReturnStmt handles ReturnStmt nodes
func (p *ASTPrinter) VisitReturnStmt(node *ReturnStmt) Visitor {
	p.printNodeStart("ReturnStmt", node)

	if node.Value == nil {
		p.result.WriteString("\n")
		return p
	}

	p.result.WriteString(" (\n")

	p.indentLevel++
	// Visit the return expression
	if node.Value != nil {
		p.result.WriteString(fmt.Sprintf("%svalue:\n", p.indent()))
		p.indentLevel++
		node.Value.Accept(p)
		p.indentLevel--
	}
	p.indentLevel--

	p.result.WriteString(fmt.Sprintf("%s)\n", p.indent()))
	return p
}

// VisitRaiseStmt handles RaiseStmt nodes
func (p *ASTPrinter) VisitRaiseStmt(node *RaiseStmt) Visitor {
	p.printNodeStart("RaiseStmt", node)

	if !node.HasException {
		p.result.WriteString("\n")
		return p
	}

	p.result.WriteString(" (\n")

	p.indentLevel++
	// Visit the exception expression
	if node.Exception != nil {
		p.result.WriteString(fmt.Sprintf("%sexception:\n", p.indent()))
		p.indentLevel++
		node.Exception.Accept(p)
		p.indentLevel--
	}

	// Visit the from expression if it exists
	if node.HasFrom && node.FromExpr != nil {
		p.result.WriteString(fmt.Sprintf("%sfrom:\n", p.indent()))
		p.indentLevel++
		node.FromExpr.Accept(p)
		p.indentLevel--
	}
	p.indentLevel--

	p.result.WriteString(fmt.Sprintf("%s)\n", p.indent()))
	return p
}

// VisitPassStmt handles PassStmt nodes
func (p *ASTPrinter) VisitPassStmt(node *PassStmt) Visitor {
	p.printNodeStart("PassStmt", node)
	p.result.WriteString("\n")
	return p
}

// VisitBreakStmt handles BreakStmt nodes
func (p *ASTPrinter) VisitBreakStmt(node *BreakStmt) Visitor {
	p.printNodeStart("BreakStmt", node)
	p.result.WriteString("\n")
	return p
}

// VisitContinueStmt handles ContinueStmt nodes
func (p *ASTPrinter) VisitContinueStmt(node *ContinueStmt) Visitor {
	p.printNodeStart("ContinueStmt", node)
	p.result.WriteString("\n")
	return p
}

// VisitYieldStmt handles YieldStmt nodes
func (p *ASTPrinter) VisitYieldStmt(node *YieldStmt) Visitor {
	p.printNodeStart("YieldStmt", node)
	p.result.WriteString(" (\n")

	p.indentLevel++
	// Visit the yield expression
	if node.Value != nil {
		p.result.WriteString(fmt.Sprintf("%svalue:\n", p.indent()))
		p.indentLevel++
		node.Value.Accept(p)
		p.indentLevel--
	}
	p.indentLevel--

	p.result.WriteString(fmt.Sprintf("%s)\n", p.indent()))
	return p
}

// VisitAssertStmt handles AssertStmt nodes
func (p *ASTPrinter) VisitAssertStmt(node *AssertStmt) Visitor {
	p.printNodeStart("AssertStmt", node)
	p.result.WriteString(" (\n")

	p.indentLevel++
	// Visit the test expression
	p.result.WriteString(fmt.Sprintf("%stest:\n", p.indent()))
	p.indentLevel++
	node.Test.Accept(p)
	p.indentLevel--

	// Visit the message expression if present
	if node.Message != nil {
		p.result.WriteString(fmt.Sprintf("%smessage:\n", p.indent()))
		p.indentLevel++
		node.Message.Accept(p)
		p.indentLevel--
	}
	p.indentLevel--

	p.result.WriteString(fmt.Sprintf("%s)\n", p.indent()))
	return p
}

// VisitGlobalStmt handles GlobalStmt nodes
func (p *ASTPrinter) VisitGlobalStmt(node *GlobalStmt) Visitor {
	p.printNodeStart("GlobalStmt", node)

	if len(node.Names) == 0 {
		p.result.WriteString("\n")
		return p
	}

	p.result.WriteString(" (\n")

	p.indentLevel++
	// Print the names
	p.result.WriteString(fmt.Sprintf("%snames:\n", p.indent()))
	p.indentLevel++
	for i, name := range node.Names {
		p.result.WriteString(fmt.Sprintf("%s%d:\n", p.indent(), i))
		p.indentLevel++
		name.Accept(p)
		p.indentLevel--
	}
	p.indentLevel--
	p.indentLevel--

	p.result.WriteString(fmt.Sprintf("%s)\n", p.indent()))
	return p
}

// VisitNonlocalStmt handles NonlocalStmt nodes
func (p *ASTPrinter) VisitNonlocalStmt(node *NonlocalStmt) Visitor {
	p.printNodeStart("NonlocalStmt", node)

	if len(node.Names) == 0 {
		p.result.WriteString("\n")
		return p
	}

	p.result.WriteString(" (\n")

	p.indentLevel++
	// Print the names
	p.result.WriteString(fmt.Sprintf("%snames:\n", p.indent()))
	p.indentLevel++
	for i, name := range node.Names {
		p.result.WriteString(fmt.Sprintf("%s%d:\n", p.indent(), i))
		p.indentLevel++
		name.Accept(p)
		p.indentLevel--
	}
	p.indentLevel--
	p.indentLevel--

	p.result.WriteString(fmt.Sprintf("%s)\n", p.indent()))
	return p
}
