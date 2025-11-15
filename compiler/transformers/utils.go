package transformers

import (
	"topple/compiler/ast"
	"topple/compiler/lexer"
)

// isViewParameter checks if a name is a view parameter using the resolution table
func (vm *ViewTransformer) isViewParameter(name *ast.Name) bool {
	if vm.resolutionTable == nil {
		return false
	}

	// PRIMARY: Try binding-based lookup for scope-aware resolution
	// This correctly handles parameters with the same name in different views
	if binding, exists := vm.resolutionTable.NameToBinding[name]; exists {
		// Check if this specific binding is a view parameter
		return binding.Variable.IsViewParameter
	}

	// FALLBACK 1: Try pointer-based Variables map for original nodes
	if variable, exists := vm.resolutionTable.Variables[name]; exists {
		return variable.IsViewParameter
	}

	// FALLBACK 2: Try ViewParameters map by name (least reliable)
	// Only use this as last resort for newly created nodes not in binding maps
	varName := name.Token.Lexeme
	if variable, exists := vm.resolutionTable.ViewParameters[varName]; exists {
		return variable.IsViewParameter
	}

	return false
}

// transformNameToSelfAttribute transforms a view parameter name to self.param
func (vm *ViewTransformer) transformNameToSelfAttribute(name *ast.Name) *ast.Attribute {
	return &ast.Attribute{
		Object: &ast.Name{
			Token: lexer.Token{
				Lexeme: "self",
				Type:   lexer.Identifier,
			},
			Span: name.Span,
		},
		Name: name.Token,
		Span: name.Span,
	}
}

// isStringLikeExpression checks if an expression is string-like (string literal, f-string, or concatenation)
func (vm *ViewTransformer) isStringLikeExpression(expr ast.Expr) bool {
	switch e := expr.(type) {
	case *ast.Literal:
		return e.Type == ast.LiteralTypeString
	case *ast.FString:
		return true
	case *ast.Binary:
		// Check if it's string concatenation (both operands are string-like)
		if e.Operator.Type == lexer.Plus {
			return vm.isStringLikeExpression(e.Left) && vm.isStringLikeExpression(e.Right)
		}
	case *ast.Call:
		// Check for str() calls or other string-returning functions
		if name, ok := e.Callee.(*ast.Name); ok {
			return name.Token.Lexeme == "str"
		}
	}
	return false
}
