package transformers

import (
	"biscuit/compiler/ast"
	"biscuit/compiler/lexer"
	"strings"
)

type ViewTransformer struct {
	// Track if we need to add psx_runtime imports
	needsRuntimeImports bool
}

// NewViewTransformer creates a new ViewTransformer
func NewViewTransformer() *ViewTransformer {
	return &ViewTransformer{
		needsRuntimeImports: false,
	}
}

// TransformViewToClass transforms a ViewStmt into a Class that inherits from BaseView
func (vm *ViewTransformer) TransformViewToClass(viewStmt *ast.ViewStmt) (*ast.Class, error) {
	// Create the class name (same as view name)
	className := viewStmt.Name

	// Create base class reference: BaseView
	baseViewName := &ast.Name{
		Token: lexer.Token{
			Lexeme: "BaseView",
			Type:   lexer.Identifier,
		},
		Span: viewStmt.Span,
	}

	// Create __init__ method
	initMethod, err := vm.createInitMethod(viewStmt)
	if err != nil {
		return nil, err
	}

	// Create _render method
	renderMethod, err := vm.createRenderMethod(viewStmt)
	if err != nil {
		return nil, err
	}

	// Create the class body with both methods
	classBody := []ast.Stmt{initMethod, renderMethod}

	// Convert TypeParams from []*TypeParam to []TypeParam
	var typeParams []ast.TypeParam
	for _, tp := range viewStmt.TypeParams {
		typeParams = append(typeParams, *tp)
	}

	// Create base class argument
	baseViewArg := &ast.Argument{
		Value: baseViewName,
		Span:  viewStmt.Span,
	}

	// Create the class node
	classNode := &ast.Class{
		Name:       className,
		TypeParams: typeParams,
		Args:       []*ast.Argument{baseViewArg}, // inherits from BaseView
		Body:       classBody,
		Span:       viewStmt.Span,
	}

	vm.needsRuntimeImports = true
	return classNode, nil
}

// createInitMethod creates the __init__ method for the view class
func (vm *ViewTransformer) createInitMethod(viewStmt *ast.ViewStmt) (*ast.Function, error) {
	// Create __init__ method name
	initName := &ast.Name{
		Token: lexer.Token{
			Lexeme: "__init__",
			Type:   lexer.Identifier,
		},
		Span: viewStmt.Span,
	}

	// Create 'self' parameter
	selfParam := &ast.Parameter{
		Name: &ast.Name{
			Token: lexer.Token{
				Lexeme: "self",
				Type:   lexer.Identifier,
			},
			Span: viewStmt.Span,
		},
		Default:    nil,
		Annotation: nil,
		IsStar:     false,
		Span:       viewStmt.Span,
	}

	// Start with self parameter
	initParams := []*ast.Parameter{selfParam}

	// Add view parameters to __init__
	if viewStmt.Params != nil && len(viewStmt.Params.Parameters) > 0 {
		for _, param := range viewStmt.Params.Parameters {
			// Skip parameters with nil names
			if param == nil || param.Name == nil {
				continue
			}
			initParams = append(initParams, param)
		}
	}

	// Create parameter list
	paramList := &ast.ParameterList{
		Parameters:  initParams,
		SlashIndex:  -1,
		VarArgIndex: -1,
		KwArgIndex:  -1,
		Span:        viewStmt.Span,
	}

	// Create assignment statements for each view parameter
	var initBody []ast.Stmt
	if viewStmt.Params != nil && len(viewStmt.Params.Parameters) > 0 {
		for _, param := range viewStmt.Params.Parameters {
			// Skip parameters with nil names
			if param == nil || param.Name == nil {
				continue
			}

			// Create self.param_name = param_name
			selfAttr := &ast.Attribute{
				Object: &ast.Name{
					Token: lexer.Token{Lexeme: "self", Type: lexer.Identifier},
					Span:  viewStmt.Span,
				},
				Name: param.Name.Token,
				Span: viewStmt.Span,
			}

			assignment := &ast.AssignStmt{
				Targets: []ast.Expr{selfAttr},
				Value:   param.Name,
				Span:    viewStmt.Span,
			}

			initBody = append(initBody, assignment)
		}
	}

	// If no body, add pass statement
	if len(initBody) == 0 {
		passStmt := &ast.PassStmt{
			Span: viewStmt.Span,
		}
		initBody = append(initBody, passStmt)
	}

	return &ast.Function{
		Name:           initName,
		TypeParameters: nil,
		Parameters:     paramList,
		ReturnType:     nil,
		Body:           initBody,
		IsAsync:        false,
		Span:           viewStmt.Span,
	}, nil
}

// createRenderMethod creates the _render method for the view class
func (vm *ViewTransformer) createRenderMethod(viewStmt *ast.ViewStmt) (*ast.Function, error) {
	// Create _render method name
	renderName := &ast.Name{
		Token: lexer.Token{
			Lexeme: "_render",
			Type:   lexer.Identifier,
		},
		Span: viewStmt.Span,
	}

	// Create 'self' parameter
	selfParam := &ast.Parameter{
		Name: &ast.Name{
			Token: lexer.Token{
				Lexeme: "self",
				Type:   lexer.Identifier,
			},
			Span: viewStmt.Span,
		},
		Default:    nil,
		Annotation: nil,
		IsStar:     false,
		Span:       viewStmt.Span,
	}

	// Create parameter list with just self
	paramList := &ast.ParameterList{
		Parameters:  []*ast.Parameter{selfParam},
		SlashIndex:  -1,
		VarArgIndex: -1,
		KwArgIndex:  -1,
		Span:        viewStmt.Span,
	}

	// Create return type annotation: Element
	elementType := &ast.Name{
		Token: lexer.Token{
			Lexeme: "Element",
			Type:   lexer.Identifier,
		},
		Span: viewStmt.Span,
	}

	// Transform view body into _render method body
	renderBody, err := vm.transformViewBody(viewStmt.Body)
	if err != nil {
		return nil, err
	}

	return &ast.Function{
		Name:           renderName,
		TypeParameters: nil,
		Parameters:     paramList,
		ReturnType:     elementType,
		Body:           renderBody,
		IsAsync:        false,
		Span:           viewStmt.Span,
	}, nil
}

// transformViewBody transforms the view body statements into _render method statements
func (vm *ViewTransformer) transformViewBody(body []ast.Stmt) ([]ast.Stmt, error) {
	var transformedBody []ast.Stmt
	var htmlElements []ast.Expr

	// Handle empty body
	if len(body) == 0 {
		// Return empty div for empty views
		returnValue := vm.createElCall("div", &ast.Literal{
			Type:  ast.LiteralTypeString,
			Value: "",
			Span:  lexer.Span{},
		}, nil)

		returnStmt := &ast.ReturnStmt{
			Value: returnValue,
			Span:  lexer.Span{},
		}

		return []ast.Stmt{returnStmt}, nil
	}

	// Process each statement in the view body
	for _, stmt := range body {
		if stmt == nil {
			continue
		}

		switch s := stmt.(type) {
		case *ast.HTMLElement:
			// Transform HTML element to el() call
			elCall, err := vm.transformHTMLElement(s)
			if err != nil {
				return nil, err
			}
			htmlElements = append(htmlElements, elCall)
		default:
			// For non-HTML statements, keep them as-is (Python code)
			transformedBody = append(transformedBody, stmt)
		}
	}

	// Create the return statement
	var returnValue ast.Expr
	if len(htmlElements) == 1 {
		// Single element - return it directly
		returnValue = htmlElements[0]
	} else if len(htmlElements) > 1 {
		// Multiple elements - wrap in a div
		divCall := vm.createElCall("div", &ast.ListExpr{
			Elements: htmlElements,
			Span:     body[0].GetSpan(),
		}, nil)
		returnValue = divCall
	} else {
		// No HTML elements - return empty div
		returnValue = vm.createElCall("div", &ast.Literal{
			Type:  ast.LiteralTypeString,
			Value: "",
			Span:  body[0].GetSpan(),
		}, nil)
	}

	returnStmt := &ast.ReturnStmt{
		Value: returnValue,
		Span:  body[0].GetSpan(),
	}

	transformedBody = append(transformedBody, returnStmt)
	return transformedBody, nil
}

// transformHTMLElement transforms an HTMLElement into an el() call
func (vm *ViewTransformer) transformHTMLElement(element *ast.HTMLElement) (ast.Expr, error) {
	// Extract the actual tag name
	tagName := element.TagName.Lexeme

	// Transform the content
	contentExpr, err := vm.transformHTMLContent(element.Content)
	if err != nil {
		return nil, err
	}

	// TODO: Handle attributes when needed
	// For now, pass nil for attributes

	return vm.createElCall(tagName, contentExpr, nil), nil
}

// transformHTMLContent transforms HTML content (nested elements, text, etc.) into appropriate expressions
func (vm *ViewTransformer) transformHTMLContent(content []ast.Stmt) (ast.Expr, error) {
	if len(content) == 0 {
		// Empty content
		return &ast.Literal{
			Type:  ast.LiteralTypeString,
			Value: "",
			Span:  lexer.Span{},
		}, nil
	}

	if len(content) == 1 {
		// Single content item
		return vm.transformHTMLContentItem(content[0])
	}

	// Multiple content items - create a list
	var contentExprs []ast.Expr
	for _, item := range content {
		expr, err := vm.transformHTMLContentItem(item)
		if err != nil {
			return nil, err
		}
		contentExprs = append(contentExprs, expr)
	}

	return &ast.ListExpr{
		Elements: contentExprs,
		Span:     lexer.Span{},
	}, nil
}

// transformHTMLContentItem transforms a single HTML content item
func (vm *ViewTransformer) transformHTMLContentItem(item ast.Stmt) (ast.Expr, error) {
	switch content := item.(type) {
	case *ast.HTMLElement:
		// Nested HTML element - recursively transform
		return vm.transformHTMLElement(content)

	case *ast.HTMLContent:
		// HTML content with text and interpolations
		return vm.transformHTMLContentParts(content.Parts)

	default:
		// Other statements (if/for blocks, etc.) - for now, skip them
		// TODO: Handle control flow statements properly
		return &ast.Literal{
			Type:  ast.LiteralTypeString,
			Value: "", // placeholder
			Span:  lexer.Span{},
		}, nil
	}
}

// transformHTMLContentParts transforms HTML content parts (text + interpolations)
func (vm *ViewTransformer) transformHTMLContentParts(parts []ast.HTMLContentPart) (ast.Expr, error) {
	if len(parts) == 0 {
		return &ast.Literal{
			Type:  ast.LiteralTypeString,
			Value: "",
			Span:  lexer.Span{},
		}, nil
	}

	if len(parts) == 1 {
		// Single part
		switch part := parts[0].(type) {
		case *ast.HTMLText:
			// Pure text
			return &ast.Literal{
				Type:  ast.LiteralTypeString,
				Value: part.Value,
				Span:  part.Span,
			}, nil

		case *ast.HTMLInterpolation:
			// Expression interpolation - need to escape it
			escapeCall := &ast.Call{
				Callee: &ast.Name{
					Token: lexer.Token{Lexeme: "escape", Type: lexer.Identifier},
					Span:  part.Span,
				},
				Arguments: []*ast.Argument{{Value: part.Expression, Span: part.Span}},
				Span:      part.Span,
			}
			return escapeCall, nil
		}
	}

	// Multiple parts - need to build an f-string or concatenation
	// For now, concatenate text parts and handle interpolations
	var result strings.Builder
	hasInterpolation := false

	for _, part := range parts {
		switch p := part.(type) {
		case *ast.HTMLText:
			result.WriteString(p.Value)
		case *ast.HTMLInterpolation:
			hasInterpolation = true
			// For now, just put a placeholder - proper f-string handling would be more complex
			result.WriteString("{interpolation}")
		}
	}

	if !hasInterpolation {
		// Pure text, return as literal
		return &ast.Literal{
			Type:  ast.LiteralTypeString,
			Value: result.String(),
			Span:  lexer.Span{},
		}, nil
	}

	// Has interpolation - for now, return the text part only
	// TODO: Implement proper f-string or concatenation handling
	return &ast.Literal{
		Type:  ast.LiteralTypeString,
		Value: result.String(),
		Span:  lexer.Span{},
	}, nil
}

// createElCall creates an el() function call
func (vm *ViewTransformer) createElCall(tag string, content ast.Expr, attrs ast.Expr) *ast.Call {
	// Create el function reference
	elFunc := &ast.Name{
		Token: lexer.Token{
			Lexeme: "el",
			Type:   lexer.Identifier,
		},
		Span: content.GetSpan(),
	}

	// Create tag argument
	tagLiteral := &ast.Argument{
		Value: &ast.Literal{
			Type:  ast.LiteralTypeString,
			Value: tag,
			Span:  content.GetSpan(),
		},
		Span: content.GetSpan(),
	}

	contentArg := &ast.Argument{
		Value: content,
		Span:  content.GetSpan(),
	}

	args := []*ast.Argument{tagLiteral, contentArg}

	// Add attrs argument if provided
	if attrs != nil {
		attrsArg := &ast.Argument{
			Value: attrs,
			Span:  content.GetSpan(),
		}
		args = append(args, attrsArg)
	}

	return &ast.Call{
		Callee:    elFunc,
		Arguments: args,
		Span:      content.GetSpan(),
	}
}

// GetRequiredImports returns the imports needed for the transformed code
func (vm *ViewTransformer) GetRequiredImports() []*ast.ImportFromStmt {
	if !vm.needsRuntimeImports {
		return nil
	}

	// Create import from psx_runtime import BaseView, el, escape
	importNames := []*ast.ImportName{
		{
			DottedName: &ast.DottedName{
				Names: []*ast.Name{
					{
						Token: lexer.Token{Lexeme: "BaseView", Type: lexer.Identifier},
						Span:  lexer.Span{},
					},
				},
				Span: lexer.Span{},
			},
			AsName: nil,
			Span:   lexer.Span{},
		},
		{
			DottedName: &ast.DottedName{
				Names: []*ast.Name{
					{
						Token: lexer.Token{Lexeme: "el", Type: lexer.Identifier},
						Span:  lexer.Span{},
					},
				},
				Span: lexer.Span{},
			},
			AsName: nil,
			Span:   lexer.Span{},
		},
		{
			DottedName: &ast.DottedName{
				Names: []*ast.Name{
					{
						Token: lexer.Token{Lexeme: "escape", Type: lexer.Identifier},
						Span:  lexer.Span{},
					},
				},
				Span: lexer.Span{},
			},
			AsName: nil,
			Span:   lexer.Span{},
		},
		{
			DottedName: &ast.DottedName{
				Names: []*ast.Name{
					{
						Token: lexer.Token{Lexeme: "Element", Type: lexer.Identifier},
						Span:  lexer.Span{},
					},
				},
				Span: lexer.Span{},
			},
			AsName: nil,
			Span:   lexer.Span{},
		},
	}

	importStmt := &ast.ImportFromStmt{
		DottedName: &ast.DottedName{
			Names: []*ast.Name{
				{
					Token: lexer.Token{Lexeme: "runtime", Type: lexer.Identifier},
					Span:  lexer.Span{},
				},
			},
			Span: lexer.Span{},
		},
		Names:      importNames,
		IsWildcard: false,
		DotCount:   0,
		Span:       lexer.Span{},
	}

	return []*ast.ImportFromStmt{importStmt}
}
