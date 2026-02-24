package transformers

import (
	"github.com/fjvillamarin/topple/compiler/ast"
	"github.com/fjvillamarin/topple/compiler/lexer"
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

	// Check if we need hierarchical processing FIRST before processing statements
	// This ensures we have the correct context when processing control structures
	needsHierarchy := vm.needsHierarchicalProcessing(body)

	// Handle single statement case
	if len(body) == 1 && !needsHierarchy {
		// Single simple statement (e.g., single <div> with no control structures)
		statements, err := vm.processViewStatement(body[0])
		if err != nil {
			return nil, err
		}

		// Check if it's an expression statement that contains an element creation
		if len(statements) == 1 {
			if exprStmt, ok := statements[0].(*ast.ExprStmt); ok {
				// Convert to return statement
				returnStmt := &ast.ReturnStmt{
					Value: exprStmt.Expr,
					Span:  exprStmt.Span,
				}
				return []ast.Stmt{returnStmt}, nil
			}
		}
		return statements, nil
	}

	// Multiple statements or single statement that needs hierarchical processing
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

		// Convert the last ExprStmt to a ReturnStmt
		// This handles cases where we have variable assignments followed by an HTML element
		// Example: features = [...]; <html>...</html> should return the html element
		if len(transformedBody) > 0 {
			lastIdx := len(transformedBody) - 1
			if exprStmt, ok := transformedBody[lastIdx].(*ast.ExprStmt); ok {
				transformedBody[lastIdx] = &ast.ReturnStmt{
					Value: exprStmt.Expr,
					Span:  exprStmt.Span,
				}
			}
		}
	}

	return transformedBody, nil
}

// processViewStatement processes a single statement in the view body
func (vm *ViewTransformer) processViewStatement(stmt ast.Stmt) ([]ast.Stmt, error) {
	switch s := stmt.(type) {
	case *ast.HTMLElement:
		// Process all HTML elements through processHTMLElement which handles validation
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
		// Check if this is a bare return in hierarchical processing mode
		if s.Value == nil && vm.currentContext != "" {
			// Transform bare return into return fragment(current_children)
			returnValue := &ast.Call{
				Callee: &ast.Name{
					Token: lexer.Token{Lexeme: "fragment", Type: lexer.Identifier},
					Span:  lexer.Span{},
				},
				Arguments: []*ast.Argument{{
					Value: &ast.Name{
						Token: lexer.Token{Lexeme: vm.currentContext, Type: lexer.Identifier},
						Span:  lexer.Span{},
					},
					Span: lexer.Span{},
				}},
				Span: s.Span,
			}
			return []ast.Stmt{&ast.ReturnStmt{
				Value: returnValue,
				Span:  s.Span,
			}}, nil
		}
		// Otherwise, transform normally (returns with values, or bare returns outside hierarchical mode)
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
	// Track number of HTML elements at the top level
	htmlElementCount := 0

	// Check if content has any statements that need to be wrapped in children arrays
	for _, stmt := range content {
		switch s := stmt.(type) {
		case *ast.For, *ast.If, *ast.While, *ast.Try, *ast.MatchStmt, *ast.With:
			// Control structures always need hierarchical processing
			return true
		case *ast.HTMLElement:
			htmlElementCount++
			// Check if the HTML element has complex content
			if vm.needsHierarchicalProcessing(s.Content) {
				return true
			}
		}
	}

	// Multiple top-level HTML elements need fragment wrapping
	return htmlElementCount > 1
}
