package parser

import (
	"biscuit/compiler/ast"
	"biscuit/compiler/lexer"
)

// assignment parses an assignment statement.
func (p *Parser) assignment() (ast.Stmt, error) {
	startPos := p.peek().Start()
	originalPos := p.Current

	// Try form 1: NAME ':' expression ['=' annotated_rhs]
	if p.check(lexer.Identifier) && p.checkNext(lexer.Colon) {
		name := p.advance()                                         // Consume the NAME
		_, err := p.consume(lexer.Colon, "expected ':' after name") // Consume the ':'
		if err != nil {
			return nil, err
		}

		// Parse type expression
		typeExpr, err := p.expression()
		if err != nil {
			return nil, err
		}

		// Check for optional assignment
		var valueExpr ast.Expr = nil
		hasValue := false
		if p.match(lexer.Equal) {
			hasValue = true
			valueExpr, err = p.annotatedRhs()
			if err != nil {
				return nil, err
			}
		}

		endPos := typeExpr.Span().End
		if valueExpr != nil {
			endPos = valueExpr.Span().End
		}

		// Create a variable annotation statement
		nameExpr := ast.NewName(name, lexer.Span{Start: name.Start(), End: name.End()})
		return ast.NewAnnotationStmt(nameExpr, typeExpr, valueExpr, hasValue, lexer.Span{Start: startPos, End: endPos}), nil
	}

	// ('(' single_target ')' | single_subscript_attribute_target) ':' expression ['=' annotated_rhs]

	// First try '(' single_target ')'
	if p.match(lexer.LeftParen) {
		target, err := p.singleTarget()
		if err != nil {
			goto trySingleSubscriptAttributeTarget
		}

		// Successfully parsed single_target, now expect closing paren
		_, err = p.consume(lexer.RightParen, "expected ')' after target")
		if err != nil {
			goto trySingleSubscriptAttributeTarget
		}

		// Now expect colon
		_, err = p.consume(lexer.Colon, "expected ':' after target")
		if err != nil {
			goto trySingleSubscriptAttributeTarget
		}

		// Parse type expression
		typeExpr, err := p.expression()
		if err != nil {
			return nil, err
		}

		// Check for optional assignment
		var valueExpr ast.Expr = nil
		hasValue := false
		if p.match(lexer.Equal) {
			hasValue = true
			valueExpr, err = p.annotatedRhs()
			if err != nil {
				return nil, err
			}
		}

		endPos := typeExpr.Span().End
		if valueExpr != nil {
			endPos = valueExpr.Span().End
		}

		// Create annotation statement
		return ast.NewAnnotationStmt(target, typeExpr, valueExpr, hasValue, lexer.Span{Start: startPos, End: endPos}), nil
	}

trySingleSubscriptAttributeTarget:
	// If any part fails, restore position and try next alternative
	p.Current = originalPos

	// Try single_subscript_attribute_target ':' ...
	target, err := p.singleSubscriptAttributeTarget()
	if err == nil {
		_, err = p.consume(lexer.Colon, "expected ':' after target")
		if err != nil {
			return nil, err
		}

		// Parse type expression
		typeExpr, err := p.expression()
		if err == nil {
			return nil, err
		}

		// Check for optional assignment
		var valueExpr ast.Expr = nil
		hasValue := false
		if p.match(lexer.Equal) {
			hasValue = true
			valueExpr, err = p.annotatedRhs()
			if err != nil {
				return nil, err
			}
		}

		endPos := typeExpr.Span().End
		if valueExpr != nil {
			endPos = valueExpr.Span().End
		}

		// Create annotation statement
		return ast.NewAnnotationStmt(target, typeExpr, valueExpr, hasValue, lexer.Span{Start: startPos, End: endPos}), nil
	}

	// Restore position and try form 3: (star_targets '=' )+ (yield_expr | star_expressions) !'=' [TYPE_COMMENT]
	p.Current = originalPos

	// Parse the first star_targets
	targets, err := p.starTargets()
	if err == nil {
		// We have valid targets, now check for '='
		if p.check(lexer.Equal) {
			// Start building the chain of targets
			var targetChain [][]ast.Expr
			targetChain = append(targetChain, targets)

			lastPos := p.Current
			// Parse additional star_targets '=' pairs
			for p.match(lexer.Equal) {
				moreTargets, err := p.starTargets()
				if err != nil {
					// We've probably consumed the right-hand side expression
					// so we need to restore the position, which should be the last equal sign
					p.Current = lastPos
					break
				}
				targetChain = append(targetChain, moreTargets)
				lastPos = p.Current
			}

			_, err = p.consume(lexer.Equal, "expected '=' after targets")
			if err != nil {
				return nil, err
			}

			// Parse the right-hand side expression
			var rhs ast.Expr
			if p.check(lexer.Yield) {
				rhs, err = p.yieldExpression()
			} else {
				rhs, err = p.starExpressions()
			}
			if err != nil {
				return nil, err
			}

			// Make sure '=' doesn't follow (used in the grammar to disambiguate)
			if p.check(lexer.Equal) {
				return nil, p.error(p.peek(), "unexpected '=' in assignment")
			}

			// For chain assignments (a = b = c = 1), we create multiple AssignStmt nodes
			// The last one gets the right-hand side expression, and then assign left to right
			// We iterate over the targetChain, and create an AssignStmt for each target
			// TODO: we should assign the RHS expression to a temp variable, and then assign the temp variable to the targets
			var stmts []ast.Stmt
			for i := 0; i < len(targetChain); i++ {
				stmts = append(stmts, ast.NewAssignStmt(targetChain[i], rhs, lexer.Span{Start: startPos, End: rhs.Span().End}))
			}
			return ast.NewMultiStmt(stmts, lexer.Span{Start: startPos, End: rhs.Span().End}), nil
		}
	}

	// Restore position and try form 4: single_target augassign ~ (yield_expr | star_expressions)
	p.Current = originalPos

	// Parse single_target
	singleTarget, err := p.singleTarget()
	if err != nil {
		return nil, err
	}

	// Try to parse augassign
	if p.match(lexer.PlusEqual, lexer.MinusEqual, lexer.StarEqual, lexer.AtEqual, lexer.SlashEqual, lexer.PercentEqual,
		lexer.AmpEqual, lexer.PipeEqual, lexer.CaretEqual, lexer.LessLessEqual, lexer.GreaterGreaterEqual,
		lexer.StarStarEqual, lexer.SlashSlashEqual) {

		op := p.previous()

		// Parse the right-hand side expression
		var value ast.Expr
		if p.check(lexer.Yield) {
			value, err = p.yieldExpression()
		} else {
			value, err = p.starExpressions()
		}
		if err != nil {
			return nil, err
		}

		// Create and return an AugAssignStmt node
		return ast.NewAugAssignStmt(singleTarget, op, value, lexer.Span{Start: startPos, End: value.Span().End}), nil
	}

	// If we get here, none of the assignment forms matched
	return nil, p.error(p.peek(), "invalid assignment")
}

// annotatedRhs parses the right-hand side of an annotated assignment:
// annotated_rhs: yield_expr | star_expressions
func (p *Parser) annotatedRhs() (ast.Expr, error) {
	if p.check(lexer.Yield) {
		return p.yieldExpression()
	}
	return p.starExpressions()
}

// augassign parses an augmented assignment operator:
// augassign:
//
//	| '+=' | '-=' | '*=' | '@=' | '/=' | '%=' | '&=' | '|=' | '^=' | '<<=' | '>>=' | '**=' | '//='
func (p *Parser) augassign() (lexer.Token, error) {
	if p.match(lexer.PlusEqual, lexer.MinusEqual, lexer.StarEqual, lexer.AtEqual, lexer.SlashEqual, lexer.PercentEqual,
		lexer.AmpEqual, lexer.PipeEqual, lexer.CaretEqual, lexer.LessLessEqual, lexer.GreaterGreaterEqual,
		lexer.StarStarEqual, lexer.SlashSlashEqual) {
		return p.previous(), nil
	}
	return lexer.Token{}, p.error(p.peek(), "expected augmented assignment operator")
}
