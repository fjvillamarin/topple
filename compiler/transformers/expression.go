package transformers

import (
	"github.com/fjvillamarin/topple/compiler/ast"
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

	case *ast.ListComp:
		// Transform the element expression and the iterable/conditions in clauses
		// Note: Target variables in ForIfClause create new bindings, don't transform them
		transformedClauses := make([]ast.ForIfClause, len(e.Clauses))
		for i, clause := range e.Clauses {
			transformedIfs := make([]ast.Expr, len(clause.Ifs))
			for j, ifExpr := range clause.Ifs {
				transformedIfs[j] = vm.transformExpression(ifExpr)
			}
			transformedClauses[i] = ast.ForIfClause{
				IsAsync: clause.IsAsync,
				Target:  clause.Target, // Keep target as-is (creates new binding)
				Iter:    vm.transformExpression(clause.Iter),
				Ifs:     transformedIfs,
			}
		}
		return &ast.ListComp{
			Element: vm.transformExpression(e.Element),
			Clauses: transformedClauses,
			Span:    e.Span,
		}

	case *ast.SetComp:
		// Same pattern as ListComp
		transformedClauses := make([]ast.ForIfClause, len(e.Clauses))
		for i, clause := range e.Clauses {
			transformedIfs := make([]ast.Expr, len(clause.Ifs))
			for j, ifExpr := range clause.Ifs {
				transformedIfs[j] = vm.transformExpression(ifExpr)
			}
			transformedClauses[i] = ast.ForIfClause{
				IsAsync: clause.IsAsync,
				Target:  clause.Target, // Keep target as-is (creates new binding)
				Iter:    vm.transformExpression(clause.Iter),
				Ifs:     transformedIfs,
			}
		}
		return &ast.SetComp{
			Element: vm.transformExpression(e.Element),
			Clauses: transformedClauses,
			Span:    e.Span,
		}

	case *ast.DictComp:
		// Transform key, value, and clauses
		transformedClauses := make([]ast.ForIfClause, len(e.Clauses))
		for i, clause := range e.Clauses {
			transformedIfs := make([]ast.Expr, len(clause.Ifs))
			for j, ifExpr := range clause.Ifs {
				transformedIfs[j] = vm.transformExpression(ifExpr)
			}
			transformedClauses[i] = ast.ForIfClause{
				IsAsync: clause.IsAsync,
				Target:  clause.Target, // Keep target as-is (creates new binding)
				Iter:    vm.transformExpression(clause.Iter),
				Ifs:     transformedIfs,
			}
		}
		return &ast.DictComp{
			Key:     vm.transformExpression(e.Key),
			Value:   vm.transformExpression(e.Value),
			Clauses: transformedClauses,
			Span:    e.Span,
		}

	case *ast.GenExpr:
		// Same pattern as ListComp
		transformedClauses := make([]ast.ForIfClause, len(e.Clauses))
		for i, clause := range e.Clauses {
			transformedIfs := make([]ast.Expr, len(clause.Ifs))
			for j, ifExpr := range clause.Ifs {
				transformedIfs[j] = vm.transformExpression(ifExpr)
			}
			transformedClauses[i] = ast.ForIfClause{
				IsAsync: clause.IsAsync,
				Target:  clause.Target, // Keep target as-is (creates new binding)
				Iter:    vm.transformExpression(clause.Iter),
				Ifs:     transformedIfs,
			}
		}
		return &ast.GenExpr{
			Element: vm.transformExpression(e.Element),
			Clauses: transformedClauses,
			Span:    e.Span,
		}

	case *ast.AwaitExpr:
		return &ast.AwaitExpr{
			Expr: vm.transformExpression(e.Expr),
			Span: e.Span,
		}

	case *ast.YieldExpr:
		return &ast.YieldExpr{
			IsFrom: e.IsFrom,
			Value:  vm.transformExpression(e.Value),
			Span:   e.Span,
		}

	case *ast.FString:
		// Transform expressions in f-string replacement fields
		transformedParts := make([]ast.FStringPart, len(e.Parts))
		for i, part := range e.Parts {
			switch p := part.(type) {
			case *ast.FStringReplacementField:
				transformedParts[i] = &ast.FStringReplacementField{
					Expression: vm.transformExpression(p.Expression),
					Equal:      p.Equal,
					Conversion: p.Conversion,
					FormatSpec: p.FormatSpec,
					Span:       p.Span,
				}
			default:
				// FStringMiddle and others don't need transformation
				transformedParts[i] = part
			}
		}
		return &ast.FString{
			Parts: transformedParts,
			Span:  e.Span,
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

	case *ast.Try:
		// Transform try body
		transformedBody := make([]ast.Stmt, len(s.Body))
		for i, stmt := range s.Body {
			transformedBody[i] = vm.transformStatement(stmt)
		}
		// Transform except handlers
		transformedExcepts := make([]ast.Except, len(s.Excepts))
		for i, except := range s.Excepts {
			transformedExceptBody := make([]ast.Stmt, len(except.Body))
			for j, stmt := range except.Body {
				transformedExceptBody[j] = vm.transformStatement(stmt)
			}
			transformedExcepts[i] = ast.Except{
				Type:   vm.transformExpression(except.Type),
				Name:   except.Name, // Keep name as-is (creates new binding)
				Body:   transformedExceptBody,
				IsStar: except.IsStar,
				Span:   except.Span,
			}
		}
		// Transform else clause
		transformedElse := make([]ast.Stmt, len(s.Else))
		for i, stmt := range s.Else {
			transformedElse[i] = vm.transformStatement(stmt)
		}
		// Transform finally clause
		transformedFinally := make([]ast.Stmt, len(s.Finally))
		for i, stmt := range s.Finally {
			transformedFinally[i] = vm.transformStatement(stmt)
		}
		return &ast.Try{
			Body:    transformedBody,
			Excepts: transformedExcepts,
			Else:    transformedElse,
			Finally: transformedFinally,
			Span:    s.Span,
		}

	case *ast.While:
		transformedBody := make([]ast.Stmt, len(s.Body))
		for i, stmt := range s.Body {
			transformedBody[i] = vm.transformStatement(stmt)
		}
		transformedElse := make([]ast.Stmt, len(s.Else))
		for i, stmt := range s.Else {
			transformedElse[i] = vm.transformStatement(stmt)
		}
		return &ast.While{
			Test: vm.transformExpression(s.Test),
			Body: transformedBody,
			Else: transformedElse,
			Span: s.Span,
		}

	case *ast.With:
		// Transform with items
		transformedItems := make([]ast.WithItem, len(s.Items))
		for i, item := range s.Items {
			transformedItems[i] = ast.WithItem{
				Expr: vm.transformExpression(item.Expr),
				As:   item.As, // Keep as-is (creates new binding)
			}
		}
		// Transform body
		transformedBody := make([]ast.Stmt, len(s.Body))
		for i, stmt := range s.Body {
			transformedBody[i] = vm.transformStatement(stmt)
		}
		return &ast.With{
			Items:   transformedItems,
			IsAsync: s.IsAsync,
			Body:    transformedBody,
			Span:    s.Span,
		}

	case *ast.MatchStmt:
		// Transform match cases
		transformedCases := make([]ast.CaseBlock, len(s.Cases))
		for i, caseBlock := range s.Cases {
			transformedBody := make([]ast.Stmt, len(caseBlock.Body))
			for j, stmt := range caseBlock.Body {
				transformedBody[j] = vm.transformStatement(stmt)
			}
			transformedCases[i] = ast.CaseBlock{
				Patterns: caseBlock.Patterns, // Patterns may create bindings, keep as-is for now
				Guard:    vm.transformExpression(caseBlock.Guard),
				Body:     transformedBody,
				Span:     caseBlock.Span,
			}
		}
		return &ast.MatchStmt{
			Subject: vm.transformExpression(s.Subject),
			Cases:   transformedCases,
			Span:    s.Span,
		}

	case *ast.MultiStmt:
		// MultiStmt should have been unwrapped in the parser
		panic("MultiStmt should not reach transformer - it should be unwrapped in the parser")

	// For other statements, return as-is for now
	default:
		return s
	}
}
