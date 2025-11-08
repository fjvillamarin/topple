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

	// PRIMARY: Try binding-based lookup (works for both original and new nodes)
	// This is the KEY FIX - lookup by name string instead of pointer identity
	varName := name.Token.Lexeme

	// Check the ViewParameters map first (most direct approach)
	if variable, exists := vm.resolutionTable.ViewParameters[varName]; exists {
		// Verify it's actually a view parameter
		return variable.IsViewParameter
	}

	// FALLBACK: Try pointer-based lookup for original nodes (backward compatibility)
	if variable, exists := vm.resolutionTable.Variables[name]; exists {
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
