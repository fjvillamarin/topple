package transformers

import (
	"github.com/fjvillamarin/topple/compiler/ast"
	"github.com/fjvillamarin/topple/compiler/lexer"
)

// isViewElement checks if an HTML element represents a view component
func (vm *ViewTransformer) isViewElement(element *ast.HTMLElement) (*ast.ViewStmt, bool) {
	if vm.resolutionTable == nil {
		return nil, false
	}

	if viewStmt, exists := vm.resolutionTable.ViewElements[element]; exists {
		return viewStmt, true
	}

	return nil, false
}

// transformViewCall creates a view instantiation call from an HTML element and its attributes,
// now with support for slot content
func (vm *ViewTransformer) transformViewCall(viewStmt *ast.ViewStmt, attributes []ast.HTMLAttribute) *ast.Call {
	// Create the view class name reference
	viewName := &ast.Name{
		Token: lexer.Token{
			Lexeme: viewStmt.Name.Token.Lexeme,
			Type:   lexer.Identifier,
		},
		Span: viewStmt.Span,
	}

	// Transform attributes into keyword constructor arguments
	var args []*ast.Argument

	// Create a map of valid parameter names for validation
	validParams := make(map[string]bool)
	if viewStmt.Params != nil {
		for _, param := range viewStmt.Params.Parameters {
			if param != nil && param.Name != nil {
				validParams[param.Name.Token.Lexeme] = true
			}
		}
	}

	// Process attributes into keyword arguments
	for _, attr := range attributes {
		// Only include attributes that match view parameters
		if _, isValid := validParams[attr.Name.Lexeme]; isValid {
			var value ast.Expr
			if attr.Value != nil {
				value = vm.transformExpression(attr.Value)
			} else {
				// Boolean attribute (no value) - use True
				value = &ast.Literal{
					Type:  ast.LiteralTypeBool,
					Value: true,
					Span:  attr.Span,
				}
			}

			arg := &ast.Argument{
				Name: &ast.Name{
					Token: lexer.Token{
						Lexeme: attr.Name.Lexeme,
						Type:   lexer.Identifier,
					},
					Span: attr.Span,
				},
				Value: value,
				Span:  attr.Span,
			}
			args = append(args, arg)
		}
	}

	return &ast.Call{
		Callee:    viewName,
		Arguments: args,
		Span:      viewStmt.Span,
	}
}
