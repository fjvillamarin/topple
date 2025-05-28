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

// VisitFString handles FString nodes
func (p *ASTPrinter) VisitFString(node *ast.FString) ast.Visitor {
	p.printNodeStart("FString", node)
	p.result.WriteString(" (\n")

	p.indentLevel++
	if len(node.Parts) > 0 {
		p.result.WriteString(fmt.Sprintf("%sparts:\n", p.indent()))
		p.indentLevel++
		for i, part := range node.Parts {
			p.result.WriteString(fmt.Sprintf("%spart %d:\n", p.indent(), i))
			p.indentLevel++
			part.Accept(p)
			p.indentLevel--
		}
		p.indentLevel--
	}
	p.indentLevel--

	p.result.WriteString(fmt.Sprintf("%s)\n", p.indent()))
	return p
}

// VisitFStringMiddle handles FStringMiddle nodes
func (p *ASTPrinter) VisitFStringMiddle(node *ast.FStringMiddle) ast.Visitor {
	p.printNodeStart("FStringMiddle", node)
	p.result.WriteString(fmt.Sprintf(" (%s)\n", node.Value))
	return p
}

// VisitFStringReplacementField handles FStringReplacementField nodes
func (p *ASTPrinter) VisitFStringReplacementField(node *ast.FStringReplacementField) ast.Visitor {
	p.printNodeStart("FStringReplacementField", node)
	p.result.WriteString(" (\n")

	p.indentLevel++
	if node.Expression != nil {
		p.result.WriteString(fmt.Sprintf("%sexpression:\n", p.indent()))
		p.indentLevel++
		node.Expression.Accept(p)
		p.indentLevel--
	}

	if node.Equal {
		p.result.WriteString(fmt.Sprintf("%sequal: true\n", p.indent()))
	}

	if node.Conversion != nil {
		p.result.WriteString(fmt.Sprintf("%sconversion:\n", p.indent()))
		p.indentLevel++
		node.Conversion.Accept(p)
		p.indentLevel--
	}

	if node.FormatSpec != nil {
		p.result.WriteString(fmt.Sprintf("%sformatSpec:\n", p.indent()))
		p.indentLevel++
		node.FormatSpec.Accept(p)
		p.indentLevel--
	}
	p.indentLevel--

	p.result.WriteString(fmt.Sprintf("%s)\n", p.indent()))
	return p
}

// VisitFStringConversion handles FStringConversion nodes
func (p *ASTPrinter) VisitFStringConversion(node *ast.FStringConversion) ast.Visitor {
	p.printNodeStart("FStringConversion", node)
	p.result.WriteString(fmt.Sprintf(" (%s)\n", node.Type))
	return p
}

// VisitFStringFormatSpec handles FStringFormatSpec nodes
func (p *ASTPrinter) VisitFStringFormatSpec(node *ast.FStringFormatSpec) ast.Visitor {
	p.printNodeStart("FStringFormatSpec", node)
	p.result.WriteString(" (\n")

	p.indentLevel++
	if len(node.Spec) > 0 {
		p.result.WriteString(fmt.Sprintf("%sspec:\n", p.indent()))
		p.indentLevel++
		for i, part := range node.Spec {
			p.result.WriteString(fmt.Sprintf("%spart %d:\n", p.indent(), i))
			p.indentLevel++
			part.Accept(p)
			p.indentLevel--
		}
		p.indentLevel--
	}
	p.indentLevel--

	p.result.WriteString(fmt.Sprintf("%s)\n", p.indent()))
	return p
}

// VisitFStringFormatMiddle handles FStringFormatMiddle nodes
func (p *ASTPrinter) VisitFStringFormatMiddle(node *ast.FStringFormatMiddle) ast.Visitor {
	p.printNodeStart("FStringFormatMiddle", node)
	p.result.WriteString(fmt.Sprintf(" (%s)\n", node.Value))
	return p
}

// VisitFStringFormatReplacementField handles FStringFormatReplacementField nodes
func (p *ASTPrinter) VisitFStringFormatReplacementField(node *ast.FStringFormatReplacementField) ast.Visitor {
	p.printNodeStart("FStringFormatReplacementField", node)
	p.result.WriteString(" (\n")

	p.indentLevel++
	// Visit the expression
	if node.Expression != nil {
		p.result.WriteString(fmt.Sprintf("%sexpression:\n", p.indent()))
		p.indentLevel++
		node.Expression.Accept(p)
		p.indentLevel--
	}

	// Display debugging equals if present
	if node.Equal {
		p.result.WriteString(fmt.Sprintf("%sequal: true\n", p.indent()))
	}

	// Visit the conversion if present
	if node.Conversion != nil {
		p.result.WriteString(fmt.Sprintf("%sconversion:\n", p.indent()))
		p.indentLevel++
		node.Conversion.Accept(p)
		p.indentLevel--
	}

	// Visit the format spec if present
	if node.FormatSpec != nil {
		p.result.WriteString(fmt.Sprintf("%sformat_spec:\n", p.indent()))
		p.indentLevel++
		node.FormatSpec.Accept(p)
		p.indentLevel--
	}
	p.indentLevel--

	p.result.WriteString(fmt.Sprintf("%s)\n", p.indent()))
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

	if len(node.Elements) > 0 {
		p.result.WriteString(fmt.Sprintf("%selements:\n", p.indent()))
		p.indentLevel++
		for _, element := range node.Elements {
			element.Accept(p)
		}
		p.indentLevel--
	}

	p.indentLevel--

	p.result.WriteString(fmt.Sprintf("%s)\n", p.indent()))
	return p
}

// VisitDictExpr handles DictExpr nodes
func (p *ASTPrinter) VisitDictExpr(node *ast.DictExpr) ast.Visitor {
	p.printNodeStart("DictExpr", node)
	p.result.WriteString(" (\n")

	p.indentLevel++

	if len(node.Pairs) > 0 {
		p.result.WriteString(fmt.Sprintf("%spairs:\n", p.indent()))
		p.indentLevel++
		for i, pair := range node.Pairs {
			switch kvp := pair.(type) {
			case *ast.KeyValuePair:
				p.result.WriteString(fmt.Sprintf("%skvpair_%d:\n", p.indent(), i))
				p.indentLevel++
				p.result.WriteString(fmt.Sprintf("%skey:\n", p.indent()))
				p.indentLevel++
				kvp.Key.Accept(p)
				p.indentLevel--
				p.result.WriteString(fmt.Sprintf("%svalue:\n", p.indent()))
				p.indentLevel++
				kvp.Value.Accept(p)
				p.indentLevel--
				p.indentLevel--
			case *ast.DoubleStarredPair:
				p.result.WriteString(fmt.Sprintf("%sstarred_%d:\n", p.indent(), i))
				p.indentLevel++
				kvp.Expr.Accept(p)
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
func (p *ASTPrinter) VisitTypeParamExpr(node *ast.TypeParam) ast.Visitor {
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

	// Display start index if present
	if node.StartIndex != nil {
		p.result.WriteString(fmt.Sprintf("%sstart:\n", p.indent()))
		p.indentLevel++
		node.StartIndex.Accept(p)
		p.indentLevel--
	}

	// Display end index if present
	if node.EndIndex != nil {
		p.result.WriteString(fmt.Sprintf("%send:\n", p.indent()))
		p.indentLevel++
		node.EndIndex.Accept(p)
		p.indentLevel--
	}

	// Display step if present
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
	if node.Test != nil {
		p.result.WriteString(fmt.Sprintf("%scondition:\n", p.indent()))
		p.indentLevel++
		node.Test.Accept(p)
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

// VisitWith handles With nodes
func (p *ASTPrinter) VisitWith(node *ast.With) ast.Visitor {
	p.printNodeStart("With", node)
	p.result.WriteString(" (\n")

	p.indentLevel++
	// Display if this is an async with
	p.result.WriteString(fmt.Sprintf("%sisAsync: %t\n", p.indent(), node.IsAsync))

	// Print the with items
	if len(node.Items) > 0 {
		p.result.WriteString(fmt.Sprintf("%sitems:\n", p.indent()))
		p.indentLevel++
		for i, item := range node.Items {
			p.result.WriteString(fmt.Sprintf("%sitem %d:\n", p.indent(), i))
			p.indentLevel++

			// Print the expression
			p.result.WriteString(fmt.Sprintf("%sexpr:\n", p.indent()))
			p.indentLevel++
			item.Expr.Accept(p)
			p.indentLevel--

			// Print the 'as' target if it exists
			if item.As != nil {
				p.result.WriteString(fmt.Sprintf("%sas:\n", p.indent()))
				p.indentLevel++
				item.As.Accept(p)
				p.indentLevel--
			}

			p.indentLevel--
		}
		p.indentLevel--
	}

	// Visit the body
	if len(node.Body) > 0 {
		p.result.WriteString(fmt.Sprintf("%sbody:\n", p.indent()))
		p.indentLevel++
		for _, stmt := range node.Body {
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

// VisitTry handles Try nodes
func (p *ASTPrinter) VisitTry(node *ast.Try) ast.Visitor {
	p.printNodeStart("Try", node)
	p.result.WriteString(" (\n")

	p.indentLevel++

	// Visit the try body
	if len(node.Body) > 0 {
		p.result.WriteString(fmt.Sprintf("%sbody:\n", p.indent()))
		p.indentLevel++
		for _, stmt := range node.Body {
			if stmt != nil {
				stmt.Accept(p)
			}
		}
		p.indentLevel--
	}

	// Visit the except blocks
	if len(node.Excepts) > 0 {
		p.result.WriteString(fmt.Sprintf("%sexcept blocks:\n", p.indent()))
		p.indentLevel++
		for i, except := range node.Excepts {
			p.result.WriteString(fmt.Sprintf("%sexcept %d:\n", p.indent(), i))
			p.indentLevel++

			// Print if this is an except* block
			p.result.WriteString(fmt.Sprintf("%sisStar: %t\n", p.indent(), except.IsStar))

			// Print the exception type if present
			if except.Type != nil {
				p.result.WriteString(fmt.Sprintf("%stype:\n", p.indent()))
				p.indentLevel++
				except.Type.Accept(p)
				p.indentLevel--
			}

			// Print the name if present
			if except.Name != nil {
				p.result.WriteString(fmt.Sprintf("%sname:\n", p.indent()))
				p.indentLevel++
				except.Name.Accept(p)
				p.indentLevel--
			}

			// Print the except body
			if len(except.Body) > 0 {
				p.result.WriteString(fmt.Sprintf("%sbody:\n", p.indent()))
				p.indentLevel++
				for _, stmt := range except.Body {
					if stmt != nil {
						stmt.Accept(p)
					}
				}
				p.indentLevel--
			}

			p.indentLevel--
		}
		p.indentLevel--
	}

	// Visit the else block if present
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

	// Visit the finally block if present
	if len(node.Finally) > 0 {
		p.result.WriteString(fmt.Sprintf("%sfinally:\n", p.indent()))
		p.indentLevel++
		for _, stmt := range node.Finally {
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

// VisitArgument handles Argument nodes
func (p *ASTPrinter) VisitArgument(node *ast.Argument) ast.Visitor {
	p.printNodeStart("Argument", node)
	p.result.WriteString(" (\n")

	p.indentLevel++

	// Print if this is a starred or double-starred argument
	if node.IsStar {
		p.result.WriteString(fmt.Sprintf("%sisStar: true\n", p.indent()))
	}
	if node.IsDoubleStar {
		p.result.WriteString(fmt.Sprintf("%sisDoubleStar: true\n", p.indent()))
	}

	// Print the keyword name if present
	if node.Name != nil {
		p.result.WriteString(fmt.Sprintf("%sname:\n", p.indent()))
		p.indentLevel++
		node.Name.Accept(p)
		p.indentLevel--
	}

	// Print the value
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

// VisitDecorator handles Decorator nodes
func (p *ASTPrinter) VisitDecorator(node *ast.Decorator) ast.Visitor {
	p.printNodeStart("Decorator", node)
	p.result.WriteString(" (\n")

	p.indentLevel++

	// Print the decorator expression
	if node.Expr != nil {
		p.result.WriteString(fmt.Sprintf("%sexpr:\n", p.indent()))
		p.indentLevel++
		node.Expr.Accept(p)
		p.indentLevel--
	}

	// Print the decorated statement
	if node.Stmt != nil {
		p.result.WriteString(fmt.Sprintf("%sdecorated:\n", p.indent()))
		p.indentLevel++
		node.Stmt.Accept(p)
		p.indentLevel--
	}

	p.indentLevel--

	p.result.WriteString(fmt.Sprintf("%s)\n", p.indent()))
	return p
}

// VisitClass handles Class nodes
func (p *ASTPrinter) VisitClass(node *ast.Class) ast.Visitor {
	p.printNodeStart("Class", node)
	p.result.WriteString(" (\n")

	p.indentLevel++

	// Print the class name
	p.result.WriteString(fmt.Sprintf("%sname:\n", p.indent()))
	p.indentLevel++
	node.Name.Accept(p)
	p.indentLevel--

	// Print type parameters if any
	if len(node.TypeParams) > 0 {
		p.result.WriteString(fmt.Sprintf("%stype parameters:\n", p.indent()))
		p.indentLevel++
		for i, param := range node.TypeParams {
			p.result.WriteString(fmt.Sprintf("%sparam %d: %s\n", p.indent(), i, param.String()))
			p.indentLevel++
			// TypeParam.Name is a lexer.Token, not a *Name node
			p.result.WriteString(fmt.Sprintf("%sname: %s\n", p.indent(), param.Name.Lexeme))

			if param.Bound != nil {
				p.result.WriteString(fmt.Sprintf("%sbound:\n", p.indent()))
				p.indentLevel++
				param.Bound.Accept(p)
				p.indentLevel--
			}
			if param.Default != nil {
				p.result.WriteString(fmt.Sprintf("%sdefault:\n", p.indent()))
				p.indentLevel++
				param.Default.Accept(p)
				p.indentLevel--
			}
			p.indentLevel--
		}
		p.indentLevel--
	}

	// Print constructor arguments (base classes) if any
	if len(node.Args) > 0 {
		p.result.WriteString(fmt.Sprintf("%sarguments:\n", p.indent()))
		p.indentLevel++
		for i, arg := range node.Args {
			p.result.WriteString(fmt.Sprintf("%sarg %d:\n", p.indent(), i))
			p.indentLevel++
			argCopy := arg // Create a copy of the argument
			p.VisitArgument(&argCopy)
			p.indentLevel--
		}
		p.indentLevel--
	}

	// Print the class body
	if len(node.Body) > 0 {
		p.result.WriteString(fmt.Sprintf("%sbody:\n", p.indent()))
		p.indentLevel++
		for _, stmt := range node.Body {
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

// VisitFunction handles Function nodes
func (p *ASTPrinter) VisitFunction(node *ast.Function) ast.Visitor {
	p.printNodeStart("Function", node)
	p.result.WriteString(" (\n")

	p.indentLevel++

	// Print if async
	p.result.WriteString(fmt.Sprintf("%sisAsync: %t\n", p.indent(), node.IsAsync))

	// Print the function name
	p.result.WriteString(fmt.Sprintf("%sname:\n", p.indent()))
	p.indentLevel++
	node.Name.Accept(p)
	p.indentLevel--

	// Print type parameters if any
	if len(node.TypeParameters) > 0 {
		p.result.WriteString(fmt.Sprintf("%stype parameters:\n", p.indent()))
		p.indentLevel++
		for i, param := range node.TypeParameters {
			p.result.WriteString(fmt.Sprintf("%sparam %d: %s\n", p.indent(), i, param.String()))
			p.indentLevel++
			// TypeParam.Name is a lexer.Token, not a *Name node
			p.result.WriteString(fmt.Sprintf("%sname: %s\n", p.indent(), param.Name.Lexeme))

			if param.Bound != nil {
				p.result.WriteString(fmt.Sprintf("%sbound:\n", p.indent()))
				p.indentLevel++
				param.Bound.Accept(p)
				p.indentLevel--
			}
			if param.Default != nil {
				p.result.WriteString(fmt.Sprintf("%sdefault:\n", p.indent()))
				p.indentLevel++
				param.Default.Accept(p)
				p.indentLevel--
			}
			p.indentLevel--
		}
		p.indentLevel--
	}

	// Print parameter list
	p.result.WriteString(fmt.Sprintf("%sparameters:\n", p.indent()))
	p.indentLevel++
	if node.Parameters != nil {
		for i, param := range node.Parameters.Parameters {
			p.result.WriteString(fmt.Sprintf("%sparam %d:\n", p.indent(), i))
			p.indentLevel++

			// Print parameter attributes
			if param.IsStar {
				p.result.WriteString(fmt.Sprintf("%sisStar: true\n", p.indent()))
			}
			if param.IsDoubleStar {
				p.result.WriteString(fmt.Sprintf("%sisDoubleStar: true\n", p.indent()))
			}
			if param.IsSlash {
				p.result.WriteString(fmt.Sprintf("%sisSlash: true\n", p.indent()))
			}

			// Print the parameter name
			if param.Name != nil {
				p.result.WriteString(fmt.Sprintf("%sname:\n", p.indent()))
				p.indentLevel++
				param.Name.Accept(p)
				p.indentLevel--
			}

			// Print the parameter annotation if present
			if param.Annotation != nil {
				p.result.WriteString(fmt.Sprintf("%sannotation:\n", p.indent()))
				p.indentLevel++
				param.Annotation.Accept(p)
				p.indentLevel--
			}

			// Print the parameter default value if present
			if param.Default != nil {
				p.result.WriteString(fmt.Sprintf("%sdefault:\n", p.indent()))
				p.indentLevel++
				param.Default.Accept(p)
				p.indentLevel--
			}

			p.indentLevel--
		}
	}
	p.indentLevel--

	// Print return type if present
	if node.ReturnType != nil {
		p.result.WriteString(fmt.Sprintf("%sreturn type:\n", p.indent()))
		p.indentLevel++
		node.ReturnType.Accept(p)
		p.indentLevel--
	}

	// Print the function body
	if len(node.Body) > 0 {
		p.result.WriteString(fmt.Sprintf("%sbody:\n", p.indent()))
		p.indentLevel++
		for _, stmt := range node.Body {
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

// VisitLambda handles Lambda nodes
func (p *ASTPrinter) VisitLambda(node *ast.Lambda) ast.Visitor {
	p.printNodeStart("Lambda", node)
	p.result.WriteString(" (\n")

	p.indentLevel++

	// Display parameters if present
	if node.Parameters != nil {
		p.result.WriteString(fmt.Sprintf("%sparameters:\n", p.indent()))
		p.indentLevel++
		// Note: ParameterList doesn't implement the visitor pattern yet,
		// so we'll just print a string representation for now
		p.result.WriteString(fmt.Sprintf("%s%s\n", p.indent(), node.Parameters.String()))
		p.indentLevel--
	}

	// Display the body expression
	if node.Body != nil {
		p.result.WriteString(fmt.Sprintf("%sbody:\n", p.indent()))
		p.indentLevel++
		node.Body.Accept(p)
		p.indentLevel--
	}

	p.indentLevel--

	p.result.WriteString(fmt.Sprintf("%s)\n", p.indent()))
	return p
}

// VisitListComp handles ListComp nodes
func (p *ASTPrinter) VisitListComp(node *ast.ListComp) ast.Visitor {
	p.printNodeStart("ListComp", node)
	p.result.WriteString(" (\n")

	p.indentLevel++

	// Display the element expression
	p.result.WriteString(fmt.Sprintf("%selement:\n", p.indent()))
	p.indentLevel++
	node.Element.Accept(p)
	p.indentLevel--

	// Display the for/if clauses
	if len(node.Clauses) > 0 {
		p.result.WriteString(fmt.Sprintf("%sclauses:\n", p.indent()))
		p.indentLevel++
		for i, clause := range node.Clauses {
			p.result.WriteString(fmt.Sprintf("%sclause_%d:\n", p.indent(), i))
			p.indentLevel++

			p.result.WriteString(fmt.Sprintf("%sisAsync: %t\n", p.indent(), clause.IsAsync))

			p.result.WriteString(fmt.Sprintf("%starget:\n", p.indent()))
			p.indentLevel++
			clause.Target.Accept(p)
			p.indentLevel--

			p.result.WriteString(fmt.Sprintf("%siter:\n", p.indent()))
			p.indentLevel++
			clause.Iter.Accept(p)
			p.indentLevel--

			if len(clause.Ifs) > 0 {
				p.result.WriteString(fmt.Sprintf("%sifs:\n", p.indent()))
				p.indentLevel++
				for j, ifCond := range clause.Ifs {
					p.result.WriteString(fmt.Sprintf("%sif_%d:\n", p.indent(), j))
					p.indentLevel++
					ifCond.Accept(p)
					p.indentLevel--
				}
				p.indentLevel--
			}

			p.indentLevel--
		}
		p.indentLevel--
	}

	p.indentLevel--

	p.result.WriteString(fmt.Sprintf("%s)\n", p.indent()))
	return p
}

// VisitSetComp handles SetComp nodes
func (p *ASTPrinter) VisitSetComp(node *ast.SetComp) ast.Visitor {
	p.printNodeStart("SetComp", node)
	p.result.WriteString(" (\n")

	p.indentLevel++

	// Display the element expression
	p.result.WriteString(fmt.Sprintf("%selement:\n", p.indent()))
	p.indentLevel++
	node.Element.Accept(p)
	p.indentLevel--

	// Display the for/if clauses
	if len(node.Clauses) > 0 {
		p.result.WriteString(fmt.Sprintf("%sclauses:\n", p.indent()))
		p.indentLevel++
		for i, clause := range node.Clauses {
			p.result.WriteString(fmt.Sprintf("%sclause_%d:\n", p.indent(), i))
			p.indentLevel++

			p.result.WriteString(fmt.Sprintf("%sisAsync: %t\n", p.indent(), clause.IsAsync))

			p.result.WriteString(fmt.Sprintf("%starget:\n", p.indent()))
			p.indentLevel++
			clause.Target.Accept(p)
			p.indentLevel--

			p.result.WriteString(fmt.Sprintf("%siter:\n", p.indent()))
			p.indentLevel++
			clause.Iter.Accept(p)
			p.indentLevel--

			if len(clause.Ifs) > 0 {
				p.result.WriteString(fmt.Sprintf("%sifs:\n", p.indent()))
				p.indentLevel++
				for j, ifCond := range clause.Ifs {
					p.result.WriteString(fmt.Sprintf("%sif_%d:\n", p.indent(), j))
					p.indentLevel++
					ifCond.Accept(p)
					p.indentLevel--
				}
				p.indentLevel--
			}

			p.indentLevel--
		}
		p.indentLevel--
	}

	p.indentLevel--

	p.result.WriteString(fmt.Sprintf("%s)\n", p.indent()))
	return p
}

// VisitDictComp handles DictComp nodes
func (p *ASTPrinter) VisitDictComp(node *ast.DictComp) ast.Visitor {
	p.printNodeStart("DictComp", node)
	p.result.WriteString(" (\n")

	p.indentLevel++

	// Display the key expression
	p.result.WriteString(fmt.Sprintf("%skey:\n", p.indent()))
	p.indentLevel++
	node.Key.Accept(p)
	p.indentLevel--

	// Display the value expression
	p.result.WriteString(fmt.Sprintf("%svalue:\n", p.indent()))
	p.indentLevel++
	node.Value.Accept(p)
	p.indentLevel--

	// Display the for/if clauses
	if len(node.Clauses) > 0 {
		p.result.WriteString(fmt.Sprintf("%sclauses:\n", p.indent()))
		p.indentLevel++
		for i, clause := range node.Clauses {
			p.result.WriteString(fmt.Sprintf("%sclause_%d:\n", p.indent(), i))
			p.indentLevel++

			p.result.WriteString(fmt.Sprintf("%sisAsync: %t\n", p.indent(), clause.IsAsync))

			p.result.WriteString(fmt.Sprintf("%starget:\n", p.indent()))
			p.indentLevel++
			clause.Target.Accept(p)
			p.indentLevel--

			p.result.WriteString(fmt.Sprintf("%siter:\n", p.indent()))
			p.indentLevel++
			clause.Iter.Accept(p)
			p.indentLevel--

			if len(clause.Ifs) > 0 {
				p.result.WriteString(fmt.Sprintf("%sifs:\n", p.indent()))
				p.indentLevel++
				for j, ifCond := range clause.Ifs {
					p.result.WriteString(fmt.Sprintf("%sif_%d:\n", p.indent(), j))
					p.indentLevel++
					ifCond.Accept(p)
					p.indentLevel--
				}
				p.indentLevel--
			}

			p.indentLevel--
		}
		p.indentLevel--
	}

	p.indentLevel--

	p.result.WriteString(fmt.Sprintf("%s)\n", p.indent()))
	return p
}

// VisitGenExpr handles GenExpr nodes
func (p *ASTPrinter) VisitGenExpr(node *ast.GenExpr) ast.Visitor {
	p.printNodeStart("GenExpr", node)
	p.result.WriteString(" (\n")

	p.indentLevel++

	// Display the element expression
	p.result.WriteString(fmt.Sprintf("%selement:\n", p.indent()))
	p.indentLevel++
	node.Element.Accept(p)
	p.indentLevel--

	// Display the for/if clauses
	if len(node.Clauses) > 0 {
		p.result.WriteString(fmt.Sprintf("%sclauses:\n", p.indent()))
		p.indentLevel++
		for i, clause := range node.Clauses {
			p.result.WriteString(fmt.Sprintf("%sclause_%d:\n", p.indent(), i))
			p.indentLevel++

			p.result.WriteString(fmt.Sprintf("%sisAsync: %t\n", p.indent(), clause.IsAsync))

			p.result.WriteString(fmt.Sprintf("%starget:\n", p.indent()))
			p.indentLevel++
			clause.Target.Accept(p)
			p.indentLevel--

			p.result.WriteString(fmt.Sprintf("%siter:\n", p.indent()))
			p.indentLevel++
			clause.Iter.Accept(p)
			p.indentLevel--

			if len(clause.Ifs) > 0 {
				p.result.WriteString(fmt.Sprintf("%sifs:\n", p.indent()))
				p.indentLevel++
				for j, ifCond := range clause.Ifs {
					p.result.WriteString(fmt.Sprintf("%sif_%d:\n", p.indent(), j))
					p.indentLevel++
					ifCond.Accept(p)
					p.indentLevel--
				}
				p.indentLevel--
			}

			p.indentLevel--
		}
		p.indentLevel--
	}

	p.indentLevel--

	p.result.WriteString(fmt.Sprintf("%s)\n", p.indent()))
	return p
}

// VisitMatch handles MatchStmt nodes
func (p *ASTPrinter) VisitMatch(node *ast.MatchStmt) ast.Visitor {
	p.printNodeStart("MatchStmt", node)
	p.result.WriteString(" (\n")

	p.indentLevel++

	// Print the subject expression
	p.result.WriteString(fmt.Sprintf("%ssubject:\n", p.indent()))
	p.indentLevel++
	node.Subject.Accept(p)
	p.indentLevel--

	// Print the case blocks
	if len(node.Cases) > 0 {
		p.result.WriteString(fmt.Sprintf("%scases:\n", p.indent()))
		p.indentLevel++
		for i, caseBlock := range node.Cases {
			p.result.WriteString(fmt.Sprintf("%scase_%d:\n", p.indent(), i))
			p.indentLevel++

			// Print patterns
			if len(caseBlock.Patterns) > 0 {
				p.result.WriteString(fmt.Sprintf("%spatterns:\n", p.indent()))
				p.indentLevel++
				for j, pattern := range caseBlock.Patterns {
					p.result.WriteString(fmt.Sprintf("%spattern_%d:\n", p.indent(), j))
					p.indentLevel++
					pattern.Accept(p)
					p.indentLevel--
				}
				p.indentLevel--
			}

			// Print guard if present
			if caseBlock.Guard != nil {
				p.result.WriteString(fmt.Sprintf("%sguard:\n", p.indent()))
				p.indentLevel++
				caseBlock.Guard.Accept(p)
				p.indentLevel--
			}

			// Print body
			if len(caseBlock.Body) > 0 {
				p.result.WriteString(fmt.Sprintf("%sbody:\n", p.indent()))
				p.indentLevel++
				for _, stmt := range caseBlock.Body {
					if stmt != nil {
						stmt.Accept(p)
					}
				}
				p.indentLevel--
			}

			p.indentLevel--
		}
		p.indentLevel--
	}

	p.indentLevel--

	p.result.WriteString(fmt.Sprintf("%s)\n", p.indent()))
	return p
}

// VisitLiteralPattern handles LiteralPattern nodes
func (p *ASTPrinter) VisitLiteralPattern(node *ast.LiteralPattern) ast.Visitor {
	p.printNodeStart("LiteralPattern", node)
	p.result.WriteString(" (\n")

	p.indentLevel++
	p.result.WriteString(fmt.Sprintf("%svalue:\n", p.indent()))
	p.indentLevel++
	node.Value.Accept(p)
	p.indentLevel--
	p.indentLevel--

	p.result.WriteString(fmt.Sprintf("%s)\n", p.indent()))
	return p
}

// VisitCapturePattern handles CapturePattern nodes
func (p *ASTPrinter) VisitCapturePattern(node *ast.CapturePattern) ast.Visitor {
	p.printNodeStart("CapturePattern", node)
	p.result.WriteString(" (\n")

	p.indentLevel++
	p.result.WriteString(fmt.Sprintf("%sname:\n", p.indent()))
	p.indentLevel++
	node.Name.Accept(p)
	p.indentLevel--
	p.indentLevel--

	p.result.WriteString(fmt.Sprintf("%s)\n", p.indent()))
	return p
}

// VisitWildcardPattern handles WildcardPattern nodes
func (p *ASTPrinter) VisitWildcardPattern(node *ast.WildcardPattern) ast.Visitor {
	p.printNodeStart("WildcardPattern", node)
	p.result.WriteString(" (_)\n")
	return p
}

// VisitValuePattern handles ValuePattern nodes
func (p *ASTPrinter) VisitValuePattern(node *ast.ValuePattern) ast.Visitor {
	p.printNodeStart("ValuePattern", node)
	p.result.WriteString(" (\n")

	p.indentLevel++
	p.result.WriteString(fmt.Sprintf("%svalue:\n", p.indent()))
	p.indentLevel++
	node.Value.Accept(p)
	p.indentLevel--
	p.indentLevel--

	p.result.WriteString(fmt.Sprintf("%s)\n", p.indent()))
	return p
}

// VisitGroupPattern handles GroupPattern nodes
func (p *ASTPrinter) VisitGroupPattern(node *ast.GroupPattern) ast.Visitor {
	p.printNodeStart("GroupPattern", node)
	p.result.WriteString(" (\n")

	p.indentLevel++
	p.result.WriteString(fmt.Sprintf("%spattern:\n", p.indent()))
	p.indentLevel++
	node.Pattern.Accept(p)
	p.indentLevel--
	p.indentLevel--

	p.result.WriteString(fmt.Sprintf("%s)\n", p.indent()))
	return p
}

// VisitSequencePattern handles SequencePattern nodes
func (p *ASTPrinter) VisitSequencePattern(node *ast.SequencePattern) ast.Visitor {
	p.printNodeStart("SequencePattern", node)
	p.result.WriteString(" (\n")

	p.indentLevel++
	p.result.WriteString(fmt.Sprintf("%sisTuple: %t\n", p.indent(), node.IsTuple))

	if len(node.Patterns) > 0 {
		p.result.WriteString(fmt.Sprintf("%spatterns:\n", p.indent()))
		p.indentLevel++
		for i, pattern := range node.Patterns {
			p.result.WriteString(fmt.Sprintf("%spattern_%d:\n", p.indent(), i))
			p.indentLevel++
			pattern.Accept(p)
			p.indentLevel--
		}
		p.indentLevel--
	}

	p.indentLevel--

	p.result.WriteString(fmt.Sprintf("%s)\n", p.indent()))
	return p
}

// VisitStarPattern handles StarPattern nodes
func (p *ASTPrinter) VisitStarPattern(node *ast.StarPattern) ast.Visitor {
	p.printNodeStart("StarPattern", node)
	p.result.WriteString(" (\n")

	p.indentLevel++
	p.result.WriteString(fmt.Sprintf("%spattern:\n", p.indent()))
	p.indentLevel++
	node.Pattern.Accept(p)
	p.indentLevel--
	p.indentLevel--

	p.result.WriteString(fmt.Sprintf("%s)\n", p.indent()))
	return p
}

// VisitMappingPattern handles MappingPattern nodes
func (p *ASTPrinter) VisitMappingPattern(node *ast.MappingPattern) ast.Visitor {
	p.printNodeStart("MappingPattern", node)
	p.result.WriteString(" (\n")

	p.indentLevel++

	if len(node.Pairs) > 0 {
		p.result.WriteString(fmt.Sprintf("%spairs:\n", p.indent()))
		p.indentLevel++
		for i, pair := range node.Pairs {
			p.result.WriteString(fmt.Sprintf("%spair_%d:\n", p.indent(), i))
			p.indentLevel++
			p.result.WriteString(fmt.Sprintf("%skey:\n", p.indent()))
			p.indentLevel++
			pair.Key.Accept(p)
			p.indentLevel--
			p.result.WriteString(fmt.Sprintf("%spattern:\n", p.indent()))
			p.indentLevel++
			pair.Pattern.Accept(p)
			p.indentLevel--
			p.indentLevel--
		}
		p.indentLevel--
	}

	if node.HasRest {
		p.result.WriteString(fmt.Sprintf("%srest:\n", p.indent()))
		p.indentLevel++
		node.DoubleStar.Accept(p)
		p.indentLevel--
	}

	p.indentLevel--

	p.result.WriteString(fmt.Sprintf("%s)\n", p.indent()))
	return p
}

// VisitClassPattern handles ClassPattern nodes
func (p *ASTPrinter) VisitClassPattern(node *ast.ClassPattern) ast.Visitor {
	p.printNodeStart("ClassPattern", node)
	p.result.WriteString(" (\n")

	p.indentLevel++

	p.result.WriteString(fmt.Sprintf("%sclass:\n", p.indent()))
	p.indentLevel++
	node.Class.Accept(p)
	p.indentLevel--

	if len(node.Patterns) > 0 {
		p.result.WriteString(fmt.Sprintf("%spatterns:\n", p.indent()))
		p.indentLevel++
		for i, pattern := range node.Patterns {
			p.result.WriteString(fmt.Sprintf("%spattern_%d:\n", p.indent(), i))
			p.indentLevel++
			pattern.Accept(p)
			p.indentLevel--
		}
		p.indentLevel--
	}

	if len(node.KwdPatterns) > 0 {
		p.result.WriteString(fmt.Sprintf("%skeyword_patterns:\n", p.indent()))
		p.indentLevel++
		for i, kwdPattern := range node.KwdPatterns {
			p.result.WriteString(fmt.Sprintf("%skwd_%d:\n", p.indent(), i))
			p.indentLevel++
			p.result.WriteString(fmt.Sprintf("%sname:\n", p.indent()))
			p.indentLevel++
			kwdPattern.Name.Accept(p)
			p.indentLevel--
			p.result.WriteString(fmt.Sprintf("%spattern:\n", p.indent()))
			p.indentLevel++
			kwdPattern.Pattern.Accept(p)
			p.indentLevel--
			p.indentLevel--
		}
		p.indentLevel--
	}

	p.indentLevel--

	p.result.WriteString(fmt.Sprintf("%s)\n", p.indent()))
	return p
}

// VisitAsPattern handles AsPattern nodes
func (p *ASTPrinter) VisitAsPattern(node *ast.AsPattern) ast.Visitor {
	p.printNodeStart("AsPattern", node)
	p.result.WriteString(" (\n")

	p.indentLevel++

	p.result.WriteString(fmt.Sprintf("%spattern:\n", p.indent()))
	p.indentLevel++
	node.Pattern.Accept(p)
	p.indentLevel--

	p.result.WriteString(fmt.Sprintf("%starget:\n", p.indent()))
	p.indentLevel++
	node.Target.Accept(p)
	p.indentLevel--

	p.indentLevel--

	p.result.WriteString(fmt.Sprintf("%s)\n", p.indent()))
	return p
}

// VisitOrPattern handles OrPattern nodes
func (p *ASTPrinter) VisitOrPattern(node *ast.OrPattern) ast.Visitor {
	p.printNodeStart("OrPattern", node)
	p.result.WriteString(" (\n")

	p.indentLevel++

	if len(node.Patterns) > 0 {
		p.result.WriteString(fmt.Sprintf("%salternatives:\n", p.indent()))
		p.indentLevel++
		for i, pattern := range node.Patterns {
			p.result.WriteString(fmt.Sprintf("%salt_%d:\n", p.indent(), i))
			p.indentLevel++
			pattern.Accept(p)
			p.indentLevel--
		}
		p.indentLevel--
	}

	p.indentLevel--

	p.result.WriteString(fmt.Sprintf("%s)\n", p.indent()))
	return p
}
