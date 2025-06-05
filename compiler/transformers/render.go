package transformers

import (
	"sylfie/compiler/ast"
	"sylfie/compiler/lexer"
)

// transformViewBody transforms the view body statements into _render method statements
func (vm *ViewTransformer) transformViewBody(body []ast.Stmt) ([]ast.Stmt, error) {
	var transformedBody []ast.Stmt

	// Handle empty body
	if len(body) == 0 {
		// Return empty fragment for empty views
		returnValue := &ast.Call{
			Callee: &ast.Name{
				Token: lexer.Token{Lexeme: "fragment", Type: lexer.Identifier},
				Span:  lexer.Span{},
			},
			Arguments: []*ast.Argument{{
				Value: &ast.ListExpr{Elements: []ast.Expr{}, Span: lexer.Span{}},
				Span:  lexer.Span{},
			}},
			Span: lexer.Span{},
		}
		returnStmt := &ast.ReturnStmt{
			Value: returnValue,
			Span:  lexer.Span{},
		}
		return []ast.Stmt{returnStmt}, nil
	}

	// Transform each statement in the body
	statements, err := vm.processViewStatement(body[0])
	if err != nil {
		return nil, err
	}

	// Check if we got exactly one HTML element and no explicit returns
	if len(body) == 1 && len(statements) == 1 {
		// Check if it's an expression statement that contains an element creation
		if exprStmt, ok := statements[0].(*ast.ExprStmt); ok {
			// Convert to return statement
			returnStmt := &ast.ReturnStmt{
				Value: exprStmt.Expr,
				Span:  exprStmt.Span,
			}
			transformedBody = append(transformedBody, returnStmt)
		} else {
			transformedBody = append(transformedBody, statements...)
		}
	} else {
		// Multiple statements - need hierarchical processing
		// First, check if we need hierarchical processing
		needsHierarchy := vm.needsHierarchicalProcessing(body)

		if needsHierarchy {
			// Push a new context for the root elements
			rootContext := vm.pushContext("root")

			// Create the array for root elements
			createArray := &ast.AssignStmt{
				Targets: []ast.Expr{
					&ast.Name{
						Token: lexer.Token{Lexeme: rootContext, Type: lexer.Identifier},
						Span:  lexer.Span{},
					},
				},
				Value: &ast.ListExpr{
					Elements: []ast.Expr{},
					Span:     lexer.Span{},
				},
				Span: lexer.Span{},
			}
			transformedBody = append(transformedBody, createArray)

			// Process all statements
			for _, stmt := range body {
				stmts, err := vm.processViewStatement(stmt)
				if err != nil {
					vm.popContext()
					return nil, err
				}
				transformedBody = append(transformedBody, stmts...)
			}

			// Pop context
			vm.popContext()

			// Return the fragment with the root elements
			returnValue := &ast.Call{
				Callee: &ast.Name{
					Token: lexer.Token{Lexeme: "fragment", Type: lexer.Identifier},
					Span:  lexer.Span{},
				},
				Arguments: []*ast.Argument{{
					Value: &ast.Name{
						Token: lexer.Token{Lexeme: rootContext, Type: lexer.Identifier},
						Span:  lexer.Span{},
					},
					Span: lexer.Span{},
				}},
				Span: lexer.Span{},
			}
			returnStmt := &ast.ReturnStmt{
				Value: returnValue,
				Span:  lexer.Span{},
			}
			transformedBody = append(transformedBody, returnStmt)
		} else {
			// Simple case - just transform all statements
			for _, stmt := range body {
				stmts, err := vm.processViewStatement(stmt)
				if err != nil {
					return nil, err
				}
				transformedBody = append(transformedBody, stmts...)
			}
		}
	}

	return transformedBody, nil
}

// processViewStatement processes a single statement in the view body
func (vm *ViewTransformer) processViewStatement(stmt ast.Stmt) ([]ast.Stmt, error) {
	switch s := stmt.(type) {
	case *ast.HTMLElement:
		// Check if this is a view component
		if viewStmt, isView := vm.isViewElement(s); isView {
			// Transform view component with slots
			transformedCall, err := vm.transformViewCallWithSlots(viewStmt, s)
			if err != nil {
				return nil, err
			}

			// If we're in a context, append to it
			if vm.currentContext != "" {
				appendStmt := vm.createAppendStatement(vm.currentContext, transformedCall)
				return []ast.Stmt{appendStmt}, nil
			}

			// Otherwise return as expression statement
			return []ast.Stmt{&ast.ExprStmt{Expr: transformedCall, Span: s.Span}}, nil
		}
		// Regular HTML element
		return vm.processHTMLElement(s)
	case *ast.For:
		return vm.processForLoop(s)
	case *ast.If:
		return vm.processIfStatement(s)
	case *ast.While:
		return vm.processWhileLoop(s)
	case *ast.Try:
		return vm.processTryStatement(s)
	case *ast.MatchStmt:
		return vm.processMatchStatement(s)
	case *ast.With:
		return vm.processWithStatement(s)
	case *ast.HTMLContent:
		// Process HTML content
		return vm.processHTMLContent(s)
	case *ast.ExprStmt:
		// Transform the expression normally
		transformed := vm.transformStatement(s)
		return []ast.Stmt{transformed}, nil
	case *ast.AssignStmt:
		// Transform assignments (preserving the assignment)
		transformed := vm.transformStatement(s)
		return []ast.Stmt{transformed}, nil
	case *ast.ReturnStmt:
		// Transform return statements
		transformed := vm.transformStatement(s)
		return []ast.Stmt{transformed}, nil
	default:
		// For other statements, transform them normally
		transformed := vm.transformStatement(stmt)
		return []ast.Stmt{transformed}, nil
	}
}

// needsHierarchicalProcessing checks if content needs hierarchical processing
func (vm *ViewTransformer) needsHierarchicalProcessing(content []ast.Stmt) bool {
	// Check if content has any statements that need to be wrapped in children arrays
	for _, stmt := range content {
		switch s := stmt.(type) {
		case *ast.For, *ast.If, *ast.While, *ast.Try, *ast.MatchStmt, *ast.With:
			return true
		case *ast.HTMLElement:
			// Check if the HTML element has complex content
			if vm.needsHierarchicalProcessing(s.Content) {
				return true
			}
		}
	}
	return false
}
