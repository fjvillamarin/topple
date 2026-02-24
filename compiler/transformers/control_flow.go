package transformers

import (
	"github.com/fjvillamarin/topple/compiler/ast"
)

// processForLoop processes a for loop in the context of an HTML context
func (vm *ViewTransformer) processForLoop(loop *ast.For) ([]ast.Stmt, error) {
	// Transform the iterable and target
	transformedIterable := vm.transformExpression(loop.Iterable)
	transformedTarget := vm.transformExpression(loop.Target)

	// Process the loop body - these statements should be processed in the current context
	// so that HTML elements are properly appended to the current children array
	var transformedBody []ast.Stmt
	for _, stmt := range loop.Body {
		processedStmts, err := vm.processViewStatement(stmt)
		if err != nil {
			return nil, err
		}
		transformedBody = append(transformedBody, processedStmts...)
	}

	// Process the else clause if it exists
	var transformedElse []ast.Stmt
	for _, stmt := range loop.Else {
		processedStmts, err := vm.processViewStatement(stmt)
		if err != nil {
			return nil, err
		}
		transformedElse = append(transformedElse, processedStmts...)
	}

	// Create the transformed for loop
	transformedLoop := &ast.For{
		Target:   transformedTarget,
		Iterable: transformedIterable,
		Body:     transformedBody,
		Else:     transformedElse,
		IsAsync:  loop.IsAsync,
		Span:     loop.Span,
	}

	return []ast.Stmt{transformedLoop}, nil
}

// processIfStatement processes an if statement in the context of an HTML context
func (vm *ViewTransformer) processIfStatement(ifStmt *ast.If) ([]ast.Stmt, error) {
	// Transform the condition
	transformedCondition := vm.transformExpression(ifStmt.Condition)

	// Process the if body - these statements should be processed in the current context
	// so that HTML elements are properly appended to the current children array
	var transformedBody []ast.Stmt
	for _, stmt := range ifStmt.Body {
		processedStmts, err := vm.processViewStatement(stmt)
		if err != nil {
			return nil, err
		}
		transformedBody = append(transformedBody, processedStmts...)
	}

	// Process the else body - these statements should be processed in the current context
	// so that HTML elements are properly appended to the current children array
	var transformedElse []ast.Stmt
	for _, stmt := range ifStmt.Else {
		processedStmts, err := vm.processViewStatement(stmt)
		if err != nil {
			return nil, err
		}
		transformedElse = append(transformedElse, processedStmts...)
	}

	// Create the transformed if statement
	transformedIf := &ast.If{
		Condition: transformedCondition,
		Body:      transformedBody,
		Else:      transformedElse,
		Span:      ifStmt.Span,
	}

	return []ast.Stmt{transformedIf}, nil
}

// processWhileLoop processes a while loop in the context of an HTML context
func (vm *ViewTransformer) processWhileLoop(loop *ast.While) ([]ast.Stmt, error) {
	// Transform the test condition
	transformedTest := vm.transformExpression(loop.Test)

	// Process the while body - these statements should be processed in the current context
	// so that HTML elements are properly appended to the current children array
	var transformedBody []ast.Stmt
	for _, stmt := range loop.Body {
		processedStmts, err := vm.processViewStatement(stmt)
		if err != nil {
			return nil, err
		}
		transformedBody = append(transformedBody, processedStmts...)
	}

	// Process the else clause if it exists
	var transformedElse []ast.Stmt
	for _, stmt := range loop.Else {
		processedStmts, err := vm.processViewStatement(stmt)
		if err != nil {
			return nil, err
		}
		transformedElse = append(transformedElse, processedStmts...)
	}

	// Create the transformed while loop
	transformedLoop := &ast.While{
		Test: transformedTest,
		Body: transformedBody,
		Else: transformedElse,
		Span: loop.Span,
	}

	return []ast.Stmt{transformedLoop}, nil
}

// processTryStatement processes a try statement in the context of an HTML context
func (vm *ViewTransformer) processTryStatement(tryStmt *ast.Try) ([]ast.Stmt, error) {
	// Process the try body - these statements should be processed in the current context
	// so that HTML elements are properly appended to the current children array
	var transformedBody []ast.Stmt
	for _, stmt := range tryStmt.Body {
		processedStmts, err := vm.processViewStatement(stmt)
		if err != nil {
			return nil, err
		}
		transformedBody = append(transformedBody, processedStmts...)
	}

	// Process except clauses
	var transformedExcepts []ast.Except
	for _, except := range tryStmt.Excepts {
		var transformedExceptBody []ast.Stmt
		for _, stmt := range except.Body {
			processedStmts, err := vm.processViewStatement(stmt)
			if err != nil {
				return nil, err
			}
			transformedExceptBody = append(transformedExceptBody, processedStmts...)
		}

		transformedExcept := ast.Except{
			Type:   vm.transformExpression(except.Type),
			Name:   except.Name, // Names don't need transformation in except clauses
			Body:   transformedExceptBody,
			IsStar: except.IsStar,
			Span:   except.Span,
		}
		transformedExcepts = append(transformedExcepts, transformedExcept)
	}

	// Process the else clause if it exists
	var transformedElse []ast.Stmt
	for _, stmt := range tryStmt.Else {
		processedStmts, err := vm.processViewStatement(stmt)
		if err != nil {
			return nil, err
		}
		transformedElse = append(transformedElse, processedStmts...)
	}

	// Process the finally clause if it exists
	var transformedFinally []ast.Stmt
	for _, stmt := range tryStmt.Finally {
		processedStmts, err := vm.processViewStatement(stmt)
		if err != nil {
			return nil, err
		}
		transformedFinally = append(transformedFinally, processedStmts...)
	}

	// Create the transformed try statement
	transformedTry := &ast.Try{
		Body:    transformedBody,
		Excepts: transformedExcepts,
		Else:    transformedElse,
		Finally: transformedFinally,
		Span:    tryStmt.Span,
	}

	return []ast.Stmt{transformedTry}, nil
}

// processMatchStatement processes a match statement in the context of an HTML context
func (vm *ViewTransformer) processMatchStatement(matchStmt *ast.MatchStmt) ([]ast.Stmt, error) {
	// Transform the subject expression
	transformedSubject := vm.transformExpression(matchStmt.Subject)

	// Process each case block
	var transformedCases []ast.CaseBlock
	for _, caseBlock := range matchStmt.Cases {
		// Process the case body - these statements should be processed in the current context
		// so that HTML elements are properly appended to the current children array
		var transformedCaseBody []ast.Stmt
		for _, stmt := range caseBlock.Body {
			processedStmts, err := vm.processViewStatement(stmt)
			if err != nil {
				return nil, err
			}
			transformedCaseBody = append(transformedCaseBody, processedStmts...)
		}

		transformedCase := ast.CaseBlock{
			Patterns: caseBlock.Patterns, // Patterns don't need transformation for view parameters
			Guard:    vm.transformExpression(caseBlock.Guard),
			Body:     transformedCaseBody,
			Span:     caseBlock.Span,
		}
		transformedCases = append(transformedCases, transformedCase)
	}

	// Create the transformed match statement
	transformedMatch := &ast.MatchStmt{
		Subject: transformedSubject,
		Cases:   transformedCases,
		Span:    matchStmt.Span,
	}

	return []ast.Stmt{transformedMatch}, nil
}

// processWithStatement processes a with statement in the context of an HTML context
func (vm *ViewTransformer) processWithStatement(withStmt *ast.With) ([]ast.Stmt, error) {
	// Transform the with items
	var transformedItems []ast.WithItem
	for _, item := range withStmt.Items {
		transformedItem := ast.WithItem{
			Expr: vm.transformExpression(item.Expr),
			As:   vm.transformExpression(item.As),
		}
		transformedItems = append(transformedItems, transformedItem)
	}

	// Process the with body - these statements should be processed in the current context
	// so that HTML elements are properly appended to the current children array
	var transformedBody []ast.Stmt
	for _, stmt := range withStmt.Body {
		processedStmts, err := vm.processViewStatement(stmt)
		if err != nil {
			return nil, err
		}
		transformedBody = append(transformedBody, processedStmts...)
	}

	// Create the transformed with statement
	transformedWith := &ast.With{
		Items:   transformedItems,
		IsAsync: withStmt.IsAsync,
		Body:    transformedBody,
		Span:    withStmt.Span,
	}

	return []ast.Stmt{transformedWith}, nil
}
