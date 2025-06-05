package transformers

import (
	"biscuit/compiler/ast"
)

// transformExpression recursively transforms expressions, converting view parameters to self attributes
func (vm *ViewTransformer) transformExpression(expr ast.Expr) ast.Expr {
	switch e := expr.(type) {
	case *ast.Name:
		// Check if this is a view parameter and transform to self.param
		if vm.isViewParameter(e) {
			return vm.transformNameToSelfAttribute(e)
		}
		return e

	case *ast.Binary:
		return &ast.Binary{
			Left:     vm.transformExpression(e.Left),
			Operator: e.Operator,
			Right:    vm.transformExpression(e.Right),
			Span:     e.Span,
		}

	case *ast.Unary:
		return &ast.Unary{
			Operator: e.Operator,
			Right:    vm.transformExpression(e.Right),
			Span:     e.Span,
		}

	case *ast.Call:
		transformedCallee := vm.transformExpression(e.Callee)
		transformedArgs := make([]*ast.Argument, len(e.Arguments))
		for i, arg := range e.Arguments {
			transformedArgs[i] = &ast.Argument{
				Value: vm.transformExpression(arg.Value),
				Name:  arg.Name,
				Span:  arg.Span,
			}
		}
		return &ast.Call{
			Callee:    transformedCallee,
			Arguments: transformedArgs,
			Span:      e.Span,
		}

	case *ast.Attribute:
		return &ast.Attribute{
			Object: vm.transformExpression(e.Object),
			Name:   e.Name,
			Span:   e.Span,
		}

	case *ast.Subscript:
		transformedIndices := make([]ast.Expr, len(e.Indices))
		for i, index := range e.Indices {
			transformedIndices[i] = vm.transformExpression(index)
		}
		return &ast.Subscript{
			Object:  vm.transformExpression(e.Object),
			Indices: transformedIndices,
			Span:    e.Span,
		}

	case *ast.ListExpr:
		transformedElements := make([]ast.Expr, len(e.Elements))
		for i, elem := range e.Elements {
			transformedElements[i] = vm.transformExpression(elem)
		}
		return &ast.ListExpr{
			Elements: transformedElements,
			Span:     e.Span,
		}

	case *ast.TupleExpr:
		transformedElements := make([]ast.Expr, len(e.Elements))
		for i, elem := range e.Elements {
			transformedElements[i] = vm.transformExpression(elem)
		}
		return &ast.TupleExpr{
			Elements: transformedElements,
			Span:     e.Span,
		}

	case *ast.DictExpr:
		transformedPairs := make([]ast.DictPair, len(e.Pairs))
		for i, pair := range e.Pairs {
			switch p := pair.(type) {
			case *ast.KeyValuePair:
				transformedPairs[i] = &ast.KeyValuePair{
					Key:   vm.transformExpression(p.Key),
					Value: vm.transformExpression(p.Value),
					Span:  p.Span,
				}
			default:
				transformedPairs[i] = pair
			}
		}
		return &ast.DictExpr{
			Pairs: transformedPairs,
			Span:  e.Span,
		}

	case *ast.SetExpr:
		transformedElements := make([]ast.Expr, len(e.Elements))
		for i, elem := range e.Elements {
			transformedElements[i] = vm.transformExpression(elem)
		}
		return &ast.SetExpr{
			Elements: transformedElements,
			Span:     e.Span,
		}

	case *ast.TernaryExpr:
		return &ast.TernaryExpr{
			Condition: vm.transformExpression(e.Condition),
			TrueExpr:  vm.transformExpression(e.TrueExpr),
			FalseExpr: vm.transformExpression(e.FalseExpr),
			Span:      e.Span,
		}

	case *ast.Lambda:
		// Note: Lambda parameters are not transformed as they create new local scope
		return &ast.Lambda{
			Parameters: e.Parameters,
			Body:       vm.transformExpression(e.Body),
			Span:       e.Span,
		}

	case *ast.Slice:
		var startIndex, endIndex, step ast.Expr
		if e.StartIndex != nil {
			startIndex = vm.transformExpression(e.StartIndex)
		}
		if e.EndIndex != nil {
			endIndex = vm.transformExpression(e.EndIndex)
		}
		if e.Step != nil {
			step = vm.transformExpression(e.Step)
		}
		return &ast.Slice{
			StartIndex: startIndex,
			EndIndex:   endIndex,
			Step:       step,
			Span:       e.Span,
		}

	// Expressions that don't need transformation
	case *ast.Literal:
		return e

	default:
		// For any other expression types, return as-is
		return e
	}
}

// transformStatement recursively transforms statements
func (vm *ViewTransformer) transformStatement(stmt ast.Stmt) ast.Stmt {
	switch s := stmt.(type) {
	case *ast.ExprStmt:
		return &ast.ExprStmt{
			Expr: vm.transformExpression(s.Expr),
			Span: s.Span,
		}

	case *ast.AssignStmt:
		// Transform the value but not the targets (assignments create new bindings)
		transformedTargets := make([]ast.Expr, len(s.Targets))
		for i, target := range s.Targets {
			transformedTargets[i] = vm.transformExpression(target)
		}
		return &ast.AssignStmt{
			Targets: transformedTargets,
			Value:   vm.transformExpression(s.Value),
			Span:    s.Span,
		}

	case *ast.ReturnStmt:
		return &ast.ReturnStmt{
			Value: vm.transformExpression(s.Value),
			Span:  s.Span,
		}

	case *ast.If:
		transformedBody := make([]ast.Stmt, len(s.Body))
		for i, stmt := range s.Body {
			transformedBody[i] = vm.transformStatement(stmt)
		}
		transformedElse := make([]ast.Stmt, len(s.Else))
		for i, stmt := range s.Else {
			transformedElse[i] = vm.transformStatement(stmt)
		}
		return &ast.If{
			Condition: vm.transformExpression(s.Condition),
			Body:      transformedBody,
			Else:      transformedElse,
			Span:      s.Span,
		}

	case *ast.For:
		transformedBody := make([]ast.Stmt, len(s.Body))
		for i, stmt := range s.Body {
			transformedBody[i] = vm.transformStatement(stmt)
		}
		transformedElse := make([]ast.Stmt, len(s.Else))
		for i, stmt := range s.Else {
			transformedElse[i] = vm.transformStatement(stmt)
		}
		return &ast.For{
			Target:   vm.transformExpression(s.Target),
			Iterable: vm.transformExpression(s.Iterable),
			Body:     transformedBody,
			Else:     transformedElse,
			IsAsync:  s.IsAsync,
			Span:     s.Span,
		}

	case *ast.Function:
		// Transform function body
		transformedBody := make([]ast.Stmt, len(s.Body))
		for i, stmt := range s.Body {
			transformedBody[i] = vm.transformStatement(stmt)
		}
		// Note: Parameters are not transformed as they create new local scope
		return &ast.Function{
			Name:           s.Name,
			TypeParameters: s.TypeParameters,
			Parameters:     s.Parameters, // Keep parameters as-is
			ReturnType:     vm.transformExpression(s.ReturnType),
			Body:           transformedBody,
			IsAsync:        s.IsAsync,
			Span:           s.Span,
		}

	case *ast.MultiStmt:
		transformedStmts := make([]ast.Stmt, len(s.Stmts))
		for i, stmt := range s.Stmts {
			transformedStmts[i] = vm.transformStatement(stmt)
		}
		return &ast.MultiStmt{
			Stmts: transformedStmts,
			Span:  s.Span,
		}

	// For other statements, return as-is for now
	default:
		return s
	}
}