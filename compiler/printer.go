package compiler

import (
	"fmt"
	"strings"

	"biscuit/compiler/ast"
	"biscuit/compiler/lexer"
)

// ASTPrinter implements a visitor that prints the AST in a tree-sitter-like format.
type ASTPrinter struct {
	ast.Visitor

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
func (p *ASTPrinter) printNodeStart(nodeType string, node ast.Node) {
	p.result.WriteString(fmt.Sprintf("%s%s [%s]", p.indent(), nodeType, node.GetSpan().String()))
}

// Visit implements the visitor pattern entry point
func (p *ASTPrinter) Visit(node ast.Node) ast.Visitor {
	if node == nil {
		return nil
	}

	// Let the node's Accept method call the appropriate VisitX method
	node.Accept(p)
	return p
}

// VisitModule handles Module nodes
func (p *ASTPrinter) VisitModule(node *ast.Module) ast.Visitor {
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
func (p *ASTPrinter) VisitExprStmt(node *ast.ExprStmt) ast.Visitor {
	p.printNodeStart("ExprStmt", node)
	p.result.WriteString(" (\n")

	p.indentLevel++
	// Visit the expression inside this statement
	if node.Expr != nil {
		node.Expr.Accept(p)
	}
	p.indentLevel--

	p.result.WriteString(fmt.Sprintf("%s)\n", p.indent()))
	return p
}

// VisitName handles Name nodes
func (p *ASTPrinter) VisitName(node *ast.Name) ast.Visitor {
	p.printNodeStart("Name", node)
	p.result.WriteString(fmt.Sprintf(" (%s)\n", node.String()))
	return p
}

// VisitLiteral handles Literal nodes
func (p *ASTPrinter) VisitLiteral(node *ast.Literal) ast.Visitor {
	var typeStr string
	switch node.Token.Type {
	case lexer.String:
		typeStr = "String"
	case lexer.Number:
		typeStr = "Number"
	default:
		typeStr = "Literal"
	}

	p.printNodeStart(typeStr, node)
	p.result.WriteString(fmt.Sprintf(" (%s)\n", node.String()))
	return p
}

// VisitAttribute handles Attribute nodes
func (p *ASTPrinter) VisitAttribute(node *ast.Attribute) ast.Visitor {
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
func (p *ASTPrinter) VisitCall(node *ast.Call) ast.Visitor {
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
func (p *ASTPrinter) VisitSubscript(node *ast.Subscript) ast.Visitor {
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

	// Visit the indices
	if len(node.Indices) > 0 {
		p.result.WriteString(fmt.Sprintf("%sindices:\n", p.indent()))
		p.indentLevel++
		for i, index := range node.Indices {
			if index != nil {
				p.result.WriteString(fmt.Sprintf("%sindex %d:\n", p.indent(), i))
				p.indentLevel++
				index.Accept(p)
				p.indentLevel--
			}
		}
		p.indentLevel--
	}
	p.indentLevel--

	p.result.WriteString(fmt.Sprintf("%s)\n", p.indent()))
	return p
}

// VisitBinary handles Binary nodes
func (p *ASTPrinter) VisitBinary(node *ast.Binary) ast.Visitor {
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
func (p *ASTPrinter) VisitUnary(node *ast.Unary) ast.Visitor {
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
func (p *ASTPrinter) Print(node ast.Node) string {
	p.result.Reset()
	p.indentLevel = 0

	// Start visitor pattern
	node.Accept(p)

	return p.result.String()
}

// VisitAssignExpr handles AssignExpr nodes
func (p *ASTPrinter) VisitAssignExpr(node *ast.AssignExpr) ast.Visitor {
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
func (p *ASTPrinter) VisitStarExpr(node *ast.StarExpr) ast.Visitor {
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
func (p *ASTPrinter) VisitTernaryExpr(node *ast.TernaryExpr) ast.Visitor {
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
func (p *ASTPrinter) VisitListExpr(node *ast.ListExpr) ast.Visitor {
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
func (p *ASTPrinter) VisitTupleExpr(node *ast.TupleExpr) ast.Visitor {
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
func (p *ASTPrinter) VisitSetExpr(node *ast.SetExpr) ast.Visitor {
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
func (p *ASTPrinter) VisitYieldExpr(node *ast.YieldExpr) ast.Visitor {
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
func (p *ASTPrinter) VisitGroupExpr(node *ast.GroupExpr) ast.Visitor {
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
func (p *ASTPrinter) VisitTypeParamExpr(node *ast.TypeParamExpr) ast.Visitor {
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
func (p *ASTPrinter) VisitTypeAlias(node *ast.TypeAlias) ast.Visitor {
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
func (p *ASTPrinter) VisitReturnStmt(node *ast.ReturnStmt) ast.Visitor {
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
func (p *ASTPrinter) VisitRaiseStmt(node *ast.RaiseStmt) ast.Visitor {
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
func (p *ASTPrinter) VisitPassStmt(node *ast.PassStmt) ast.Visitor {
	p.printNodeStart("PassStmt", node)
	p.result.WriteString("\n")
	return p
}

// VisitBreakStmt handles BreakStmt nodes
func (p *ASTPrinter) VisitBreakStmt(node *ast.BreakStmt) ast.Visitor {
	p.printNodeStart("BreakStmt", node)
	p.result.WriteString("\n")
	return p
}

// VisitContinueStmt handles ContinueStmt nodes
func (p *ASTPrinter) VisitContinueStmt(node *ast.ContinueStmt) ast.Visitor {
	p.printNodeStart("ContinueStmt", node)
	p.result.WriteString("\n")
	return p
}

// VisitYieldStmt handles YieldStmt nodes
func (p *ASTPrinter) VisitYieldStmt(node *ast.YieldStmt) ast.Visitor {
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
func (p *ASTPrinter) VisitAssertStmt(node *ast.AssertStmt) ast.Visitor {
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
func (p *ASTPrinter) VisitGlobalStmt(node *ast.GlobalStmt) ast.Visitor {
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
func (p *ASTPrinter) VisitNonlocalStmt(node *ast.NonlocalStmt) ast.Visitor {
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

// VisitImportStmt handles ImportStmt nodes
func (p *ASTPrinter) VisitImportStmt(node *ast.ImportStmt) ast.Visitor {
	p.printNodeStart("ImportStmt", node)

	if len(node.Names) == 0 {
		p.result.WriteString("\n")
		return p
	}

	p.result.WriteString(" (\n")

	p.indentLevel++
	// Print imported modules
	p.result.WriteString(fmt.Sprintf("%simports:\n", p.indent()))
	p.indentLevel++
	for i, importName := range node.Names {
		p.result.WriteString(fmt.Sprintf("%s%d:\n", p.indent(), i))
		p.indentLevel++
		p.visitImportName(importName)
		p.indentLevel--
	}
	p.indentLevel--
	p.indentLevel--

	p.result.WriteString(fmt.Sprintf("%s)\n", p.indent()))
	return p
}

// VisitImportFromStmt handles ImportFromStmt nodes
func (p *ASTPrinter) VisitImportFromStmt(node *ast.ImportFromStmt) ast.Visitor {
	p.printNodeStart("ImportFromStmt", node)
	p.result.WriteString(" (\n")

	p.indentLevel++

	// Print the module path
	p.result.WriteString(fmt.Sprintf("%smodule: ", p.indent()))

	// For relative imports, print the dots
	if node.DotCount > 0 {
		p.result.WriteString(strings.Repeat(".", node.DotCount))
		if node.DottedName != nil {
			p.result.WriteString(".")
		}
	}

	// Print the dotted name if it exists
	if node.DottedName != nil {
		p.visitDottedName(node.DottedName)
	} else if node.DotCount == 0 {
		p.result.WriteString("''")
	}

	p.result.WriteString("\n")

	// Print wildcard or imported names
	if node.IsWildcard {
		p.result.WriteString(fmt.Sprintf("%simport: *\n", p.indent()))
	} else {
		p.result.WriteString(fmt.Sprintf("%simports:\n", p.indent()))
		p.indentLevel++
		for i, importName := range node.Names {
			p.result.WriteString(fmt.Sprintf("%s%d:\n", p.indent(), i))
			p.indentLevel++
			p.visitImportName(importName)
			p.indentLevel--
		}
		p.indentLevel--
	}

	p.indentLevel--

	p.result.WriteString(fmt.Sprintf("%s)\n", p.indent()))
	return p
}

// Helper method to visit an ImportName
func (p *ASTPrinter) visitImportName(node *ast.ImportName) {
	// Print the module name
	p.result.WriteString(fmt.Sprintf("%sname: ", p.indent()))
	p.visitDottedName(node.DottedName)
	p.result.WriteString("\n")

	// Print the alias if it exists
	if node.AsName != nil {
		p.result.WriteString(fmt.Sprintf("%sas: ", p.indent()))
		p.indentLevel++
		node.AsName.Accept(p)
		p.indentLevel--
	}
}

// Helper method to visit a DottedName
func (p *ASTPrinter) visitDottedName(node *ast.DottedName) {
	parts := make([]string, len(node.Names))
	for i, name := range node.Names {
		parts[i] = name.Token.Lexeme
	}
	p.result.WriteString(strings.Join(parts, "."))
}

// VisitSlice handles Slice nodes
func (p *ASTPrinter) VisitSlice(node *ast.Slice) ast.Visitor {
	p.printNodeStart("Slice", node)
	p.result.WriteString(" (\n")

	p.indentLevel++
	// Visit the start index if present
	if node.StartIndex != nil {
		p.result.WriteString(fmt.Sprintf("%sstart:\n", p.indent()))
		p.indentLevel++
		node.StartIndex.Accept(p)
		p.indentLevel--
	}

	// Visit the end index if present
	if node.EndIndex != nil {
		p.result.WriteString(fmt.Sprintf("%send:\n", p.indent()))
		p.indentLevel++
		node.EndIndex.Accept(p)
		p.indentLevel--
	}

	// Visit the step if present
	if node.Step != nil {
		p.result.WriteString(fmt.Sprintf("%sstep:\n", p.indent()))
		p.indentLevel++
		node.Step.Accept(p)
		p.indentLevel--
	}
	p.indentLevel--

	p.result.WriteString(fmt.Sprintf("%s)\n", p.indent()))
	return p
}

// VisitAwaitExpr handles AwaitExpr nodes
func (p *ASTPrinter) VisitAwaitExpr(node *ast.AwaitExpr) ast.Visitor {
	p.printNodeStart("AwaitExpr", node)
	p.result.WriteString(" (\n")

	p.indentLevel++
	// Visit the expression
	p.result.WriteString(fmt.Sprintf("%sexpr:\n", p.indent()))
	p.indentLevel++
	node.Expr.Accept(p)
	p.indentLevel--

	p.indentLevel--

	p.result.WriteString(fmt.Sprintf("%s)\n", p.indent()))
	return p
}

// VisitAssignStmt handles AssignStmt nodes
func (p *ASTPrinter) VisitAssignStmt(node *ast.AssignStmt) ast.Visitor {
	p.printNodeStart("AssignStmt", node)
	p.result.WriteString(" (\n")

	p.indentLevel++
	// Visit targets
	if len(node.Targets) > 0 {
		p.result.WriteString(fmt.Sprintf("%stargets:\n", p.indent()))
		p.indentLevel++
		for i, target := range node.Targets {
			if target != nil {
				p.result.WriteString(fmt.Sprintf("%starget %d:\n", p.indent(), i))
				p.indentLevel++
				target.Accept(p)
				p.indentLevel--
			}
		}
		p.indentLevel--
	}

	// Visit the value
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

// VisitAnnotationStmt handles AnnotationStmt nodes
func (p *ASTPrinter) VisitAnnotationStmt(node *ast.AnnotationStmt) ast.Visitor {
	p.printNodeStart("AnnotationStmt", node)
	p.result.WriteString(" (\n")

	p.indentLevel++
	// Visit the target
	if node.Target != nil {
		p.result.WriteString(fmt.Sprintf("%starget:\n", p.indent()))
		p.indentLevel++
		node.Target.Accept(p)
		p.indentLevel--
	}

	// Visit the type annotation
	if node.Type != nil {
		p.result.WriteString(fmt.Sprintf("%stype:\n", p.indent()))
		p.indentLevel++
		node.Type.Accept(p)
		p.indentLevel--
	}

	// Visit the value if it has one
	if node.HasValue && node.Value != nil {
		p.result.WriteString(fmt.Sprintf("%svalue:\n", p.indent()))
		p.indentLevel++
		node.Value.Accept(p)
		p.indentLevel--
	}
	p.indentLevel--

	p.result.WriteString(fmt.Sprintf("%s)\n", p.indent()))
	return p
}

// VisitMultiStmt handles MultiStmt nodes
func (p *ASTPrinter) VisitMultiStmt(node *ast.MultiStmt) ast.Visitor {
	for _, stmt := range node.Stmts {
		stmt.Accept(p)
	}
	return p
}

// VisitIf handles If nodes
func (p *ASTPrinter) VisitIf(node *ast.If) ast.Visitor {
	p.printNodeStart("If", node)
	p.result.WriteString(" (\n")

	p.indentLevel++
	// Visit the condition
	if node.Condition != nil {
		p.result.WriteString(fmt.Sprintf("%scondition:\n", p.indent()))
		p.indentLevel++
		node.Condition.Accept(p)
		p.indentLevel--
	}

	// Visit the body
	if node.Body != nil {
		p.result.WriteString(fmt.Sprintf("%sbody:\n", p.indent()))
		p.indentLevel++
		for _, stmt := range node.Body {
			if stmt != nil {
				stmt.Accept(p)
			}
		}
		p.indentLevel--
	}

	// Visit the else statements
	if len(node.Else) > 0 {
		p.result.WriteString(fmt.Sprintf("%selse:\n", p.indent()))
		p.indentLevel++
		for _, stmt := range node.Else {
			if stmt != nil {
				stmt.Accept(p)
			}
		}
		p.indentLevel--
	}

	p.indentLevel--

	p.result.WriteString(fmt.Sprintf("%s)\n", p.indent()))
	return p
}

// VisitWhile handles While nodes
func (p *ASTPrinter) VisitWhile(node *ast.While) ast.Visitor {
	p.printNodeStart("While", node)
	p.result.WriteString(" (\n")

	p.indentLevel++
	// Visit the condition
	if node.Condition != nil {
		p.result.WriteString(fmt.Sprintf("%scondition:\n", p.indent()))
		p.indentLevel++
		node.Condition.Accept(p)
		p.indentLevel--
	}

	// Visit the body
	if node.Body != nil {
		p.result.WriteString(fmt.Sprintf("%sbody:\n", p.indent()))
		p.indentLevel++
		for _, stmt := range node.Body {
			if stmt != nil {
				stmt.Accept(p)
			}
		}
		p.indentLevel--
	}

	// Visit the else statements
	if len(node.Else) > 0 {
		p.result.WriteString(fmt.Sprintf("%selse:\n", p.indent()))
		p.indentLevel++
		for _, stmt := range node.Else {
			if stmt != nil {
				stmt.Accept(p)
			}
		}
		p.indentLevel--
	}

	p.result.WriteString(fmt.Sprintf("%s)\n", p.indent()))
	return p
}

// VisitFor handles For nodes
func (p *ASTPrinter) VisitFor(node *ast.For) ast.Visitor {
	p.printNodeStart("For", node)
	p.result.WriteString(" (\n")

	p.indentLevel++
	// Display if this is an async for
	p.result.WriteString(fmt.Sprintf("%sisAsync: %t\n", p.indent(), node.IsAsync))

	// Visit the target
	if node.Target != nil {
		p.result.WriteString(fmt.Sprintf("%starget:\n", p.indent()))
		p.indentLevel++
		node.Target.Accept(p)
		p.indentLevel--
	}

	// Visit the iterable
	if node.Iterable != nil {
		p.result.WriteString(fmt.Sprintf("%siterable:\n", p.indent()))
		p.indentLevel++
		node.Iterable.Accept(p)
		p.indentLevel--
	}

	// Visit the body
	if node.Body != nil {
		p.result.WriteString(fmt.Sprintf("%sbody:\n", p.indent()))
		p.indentLevel++
		for _, stmt := range node.Body {
			if stmt != nil {
				stmt.Accept(p)
			}
		}
		p.indentLevel--
	}

	// Visit the else statements
	if len(node.Else) > 0 {
		p.result.WriteString(fmt.Sprintf("%selse:\n", p.indent()))
		p.indentLevel++
		for _, stmt := range node.Else {
			if stmt != nil {
				stmt.Accept(p)
			}
		}
		p.indentLevel--
	}

	p.indentLevel--

	p.result.WriteString(fmt.Sprintf("%s)\n", p.indent()))
	return p
}
