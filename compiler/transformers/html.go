package transformers

import (
	"sylfie/compiler/ast"
	"sylfie/compiler/lexer"
)

// processHTMLElement processes an HTMLElement and returns the transformed statements
func (vm *ViewTransformer) processHTMLElement(element *ast.HTMLElement) ([]ast.Stmt, error) {
	var statements []ast.Stmt

	// Check if this element is actually a view composition
	if viewStmt, isView := vm.isViewElement(element); isView {
		// Validate slot usage before processing
		if err := vm.validateSlotUsage(element); err != nil {
			return nil, err
		}

		// Create the view call with slot processing
		transformedView, err := vm.transformViewCallWithSlots(viewStmt, element)
		if err != nil {
			return nil, err
		}

		// Append to current context if we have one
		if vm.currentContext != "" {
			appendStmt := vm.createAppendStatement(vm.currentContext, transformedView)
			statements = append(statements, appendStmt)
		} else {
			// If no current context, add as expression statement
			statements = append(statements, &ast.ExprStmt{
				Expr: transformedView,
				Span: element.Span,
			})
		}

		return statements, nil
	}

	// Regular HTML element processing...
	// Transform the element into an el() call
	transformedElement, err := vm.transformHTMLElement(element)
	if err != nil {
		return nil, err
	}

	// Append to current context if we have one
	if vm.currentContext != "" {
		appendStmt := vm.createAppendStatement(vm.currentContext, transformedElement)
		return []ast.Stmt{appendStmt}, nil
	}

	// If no current context, return as expression statement
	return []ast.Stmt{&ast.ExprStmt{
		Expr: transformedElement,
		Span: element.Span,
	}}, nil
}

// createAppendStatement creates a statement that appends an element to a children array
func (vm *ViewTransformer) createAppendStatement(arrayName string, element ast.Expr) ast.Stmt {
	// Create: arrayName.append(element)
	appendCall := &ast.Call{
		Callee: &ast.Attribute{
			Object: &ast.Name{
				Token: lexer.Token{Lexeme: arrayName, Type: lexer.Identifier},
				Span:  lexer.Span{},
			},
			Name: lexer.Token{Lexeme: "append", Type: lexer.Identifier},
			Span: lexer.Span{},
		},
		Arguments: []*ast.Argument{{
			Value: element,
			Span:  lexer.Span{},
		}},
		Span: lexer.Span{},
	}

	return &ast.ExprStmt{
		Expr: appendCall,
		Span: lexer.Span{},
	}
}

// transformHTMLElement transforms an HTMLElement into an el() call
func (vm *ViewTransformer) transformHTMLElement(element *ast.HTMLElement) (ast.Expr, error) {
	// Check if this element is actually a view composition
	if viewStmt, isView := vm.isViewElement(element); isView {
		// This is a view composition - create a view instantiation call
		return vm.transformViewCall(viewStmt, element.Attributes), nil
	}

	// Regular HTML element processing...
	// Extract the actual tag name
	tagName := element.TagName.Lexeme

	// Transform attributes
	var attrsExpr ast.Expr
	if len(element.Attributes) > 0 {
		transformedAttrs, err := vm.transformHTMLAttributes(element.Attributes)
		if err != nil {
			return nil, err
		}
		attrsExpr = transformedAttrs
	}

	// Transform the content
	contentExpr, err := vm.transformHTMLContent(element.Content)
	if err != nil {
		return nil, err
	}

	return vm.createElCall(tagName, contentExpr, attrsExpr), nil
}

// transformHTMLAttributes transforms HTML attributes into a Python dictionary expression
func (vm *ViewTransformer) transformHTMLAttributes(attributes []ast.HTMLAttribute) (ast.Expr, error) {
	if len(attributes) == 0 {
		return nil, nil
	}

	var dictPairs []ast.DictPair

	for _, attr := range attributes {
		// Create the key (attribute name)
		keyExpr := &ast.Literal{
			Type:  ast.LiteralTypeString,
			Value: attr.Name.Lexeme,
			Span:  lexer.Span{Start: attr.Name.Start(), End: attr.Name.End()},
		}

		var valueExpr ast.Expr

		if attr.Value == nil {
			// Boolean attribute (no value) - use True
			valueExpr = &ast.Literal{
				Type:  ast.LiteralTypeBool,
				Value: true,
				Span:  attr.Span,
			}
		} else {
			// Transform the attribute value, applying view parameter transformation
			transformedValue := vm.transformExpression(attr.Value)

			// Check if this is a static string literal - no need to escape
			if literal, ok := attr.Value.(*ast.Literal); ok && literal.Type == ast.LiteralTypeString {
				valueExpr = transformedValue
			} else {
				// Dynamic expression - wrap with escape() for security
				valueExpr = &ast.Call{
					Callee: &ast.Name{
						Token: lexer.Token{Lexeme: "escape", Type: lexer.Identifier},
						Span:  attr.Span,
					},
					Arguments: []*ast.Argument{{Value: transformedValue, Span: attr.Span}},
					Span:      attr.Span,
				}
			}
		}

		// Create the key-value pair
		pair := &ast.KeyValuePair{
			Key:   keyExpr,
			Value: valueExpr,
			Span:  attr.Span,
		}

		dictPairs = append(dictPairs, pair)
	}

	// Create the dictionary expression
	return &ast.DictExpr{
		Pairs: dictPairs,
		Span:  lexer.Span{},
	}, nil
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
		// Check if this is a slot element
		if content.TagName.Lexeme == "slot" {
			return vm.transformSlotElementToExpression(content)
		}
		// Nested HTML element - recursively transform
		return vm.transformHTMLElement(content)

	case *ast.HTMLContent:
		// HTML content with text and interpolations
		return vm.transformHTMLContentParts(content.Parts)

	case *ast.ExprStmt:
		// Expression statement - only treat string-like expressions as content
		if vm.isStringLikeExpression(content.Expr) {
			transformedExpr := vm.transformExpression(content.Expr)
			return &ast.Call{
				Callee: &ast.Name{
					Token: lexer.Token{Lexeme: "escape", Type: lexer.Identifier},
					Span:  content.Span,
				},
				Arguments: []*ast.Argument{{
					Value: transformedExpr,
					Span:  content.Span,
				}},
				Span: content.Span,
			}, nil
		} else {
			// Non-string expression - return empty string (ignore it)
			return &ast.Literal{
				Type:  ast.LiteralTypeString,
				Value: "",
				Span:  content.Span,
			}, nil
		}

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
			// Expression interpolation - transform the expression for view parameters
			transformedExpr := vm.transformExpression(part.Expression)
			escapeCall := &ast.Call{
				Callee: &ast.Name{
					Token: lexer.Token{Lexeme: "escape", Type: lexer.Identifier},
					Span:  part.Span,
				},
				Arguments: []*ast.Argument{{Value: transformedExpr, Span: part.Span}},
				Span:      part.Span,
			}
			return escapeCall, nil
		}
	}

	// Multiple parts - build an f-string expression
	var fStringParts []ast.FStringPart

	for _, part := range parts {
		switch p := part.(type) {
		case *ast.HTMLText:
			// Add text as an f-string middle part
			fStringParts = append(fStringParts, &ast.FStringMiddle{
				Value: p.Value,
				Span:  p.Span,
			})

		case *ast.HTMLInterpolation:
			// Transform the expression for view parameters and add as replacement field
			transformedExpr := vm.transformExpression(p.Expression)
			escapeCall := &ast.Call{
				Callee: &ast.Name{
					Token: lexer.Token{Lexeme: "escape", Type: lexer.Identifier},
					Span:  p.Span,
				},
				Arguments: []*ast.Argument{{Value: transformedExpr, Span: p.Span}},
				Span:      p.Span,
			}

			replacementField := &ast.FStringReplacementField{
				Expression: escapeCall,
				Equal:      false,
				Conversion: nil,
				FormatSpec: nil,
				Span:       p.Span,
			}
			fStringParts = append(fStringParts, replacementField)
		}
	}

	// Build the f-string
	return &ast.FString{
		Parts: fStringParts,
		Span:  lexer.Span{},
	}, nil
}

// processHTMLContent processes HTMLContent and returns the transformed statements
func (vm *ViewTransformer) processHTMLContent(content *ast.HTMLContent) ([]ast.Stmt, error) {
	// Transform HTML content parts (text + interpolations)
	contentExpr, err := vm.transformHTMLContentParts(content.Parts)
	if err != nil {
		return nil, err
	}

	// Append to current context if we have one
	if vm.currentContext != "" {
		appendStmt := vm.createAppendStatement(vm.currentContext, contentExpr)
		return []ast.Stmt{appendStmt}, nil
	}

	// If no current context, return as expression statement
	return []ast.Stmt{&ast.ExprStmt{
		Expr: contentExpr,
		Span: content.Span,
	}}, nil
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