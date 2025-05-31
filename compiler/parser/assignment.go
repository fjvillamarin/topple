package parser

import (
	"biscuit/compiler/ast"
	"biscuit/compiler/lexer"
	"fmt"
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

		endPos := typeExpr.GetSpan().End
		if valueExpr != nil {
			endPos = valueExpr.GetSpan().End
		}

		// Create a variable annotation statement
		nameExpr := &ast.Name{
			Token: name,

			Span: name.Span,
		}
		return &ast.AnnotationStmt{
			Target:   nameExpr,
			Type:     typeExpr,
			Value:    valueExpr,
			HasValue: hasValue,

			Span: lexer.Span{
				Start: startPos,
				End:   endPos,
			},
		}, nil
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

		endPos := typeExpr.GetSpan().End
		if valueExpr != nil {
			endPos = valueExpr.GetSpan().End
		}

		// Create annotation statement
		return &ast.AnnotationStmt{
			Target:   target,
			Type:     typeExpr,
			Value:    valueExpr,
			HasValue: hasValue,

			Span: lexer.Span{
				Start: startPos,
				End:   endPos,
			},
		}, nil
	}

trySingleSubscriptAttributeTarget:
	// If any part fails, restore position and try next alternative
	p.Current = originalPos

	// Try single_subscript_attribute_target ':' ...
	target, err := p.singleSubscriptAttributeTarget()
	if err == nil {
		_, err = p.consume(lexer.Colon, "expected ':' after target")
		if err != nil {
			goto tryStarTargets
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

		endPos := typeExpr.GetSpan().End
		if valueExpr != nil {
			endPos = valueExpr.GetSpan().End
		}

		// Create annotation statement
		return &ast.AnnotationStmt{
			Target:   target,
			Type:     typeExpr,
			Value:    valueExpr,
			HasValue: hasValue,

			Span: lexer.Span{
				Start: startPos,
				End:   endPos,
			},
		}, nil
	}

tryStarTargets:
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

			// Parse additional star_targets '=' pairs
			for p.match(lexer.Equal) {
				// Save position before trying to parse more targets
				posBeforeTargets := p.Current

				// Try to parse more targets
				moreTargets, err := p.starTargets()
				if err != nil {
					// If we can't parse targets, this isn't a chain assignment
					// Restore position to the equal sign we just consumed
					p.Current = posBeforeTargets - 1
					break
				}

				// Check if there's another '=' after these targets
				if !p.check(lexer.Equal) {
					// No '=' after these targets, so they're part of the RHS expression
					// Restore position to before we parsed these "targets"
					p.Current = posBeforeTargets - 1
					break
				}

				// We have valid targets followed by '=', add them to the chain
				targetChain = append(targetChain, moreTargets)
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
				stmts = append(stmts, &ast.AssignStmt{
					Targets: targetChain[i],
					Value:   rhs,

					Span: lexer.Span{
						Start: startPos,
						End:   rhs.GetSpan().End,
					},
				})
			}
			return &ast.MultiStmt{
				Stmts: stmts,
				Span:  lexer.Span{Start: startPos, End: rhs.GetSpan().End},
			}, nil
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
	if op, err := p.augassign(); err == nil {
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

		// Augmented assignment is syntactic sugar for:
		// target = target op value
		rhs := &ast.Binary{
			Left:     singleTarget, // We are assigning the same node, this may cause issues in the future - maybe we should clone the node?
			Operator: augassignToOperator(op),
			Right:    value,

			Span: lexer.Span{
				Start: singleTarget.GetSpan().Start,
				End:   value.GetSpan().End,
			},
		}
		return &ast.AssignStmt{
			Targets: []ast.Expr{singleTarget},
			Value:   rhs,

			Span: lexer.Span{Start: singleTarget.GetSpan().Start, End: rhs.GetSpan().End},
		}, nil
	}

	// If we get here, none of the assignment forms matched
	return nil, p.error(p.peek(), "invalid assignment")
}

// annotatedRhs parses the right-hand side of an annotated assignment:
// annotated_rhs: yield_expr | star_expressions
func (p *Parser) annotatedRhs() (ast.Expr, error) {
	fmt.Printf("[DEBUG] annotatedRhs() entry, current token: %d %s at pos %d\n", p.Current, p.peek().Type, p.Current)

	if p.check(lexer.Yield) {
		result, err := p.yieldExpression()
		fmt.Printf("[DEBUG] annotatedRhs() yield exit, pos %d\n", p.Current)
		return result, err
	}

	fmt.Printf("[DEBUG] annotatedRhs() calling starExpressions(), pos %d\n", p.Current)
	result, err := p.starExpressions()
	fmt.Printf("[DEBUG] annotatedRhs() starExpressions() returned, pos %d\n", p.Current)
	return result, err
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

// augassignToOperator converts an augmented assignment operator to its corresponding binary operator
func augassignToOperator(augassign lexer.Token) lexer.Token {
	switch augassign.Type {
	case lexer.PlusEqual:
		return lexer.Token{
			Type:    lexer.Plus,
			Lexeme:  "+",
			Literal: "+",
			Span:    augassign.Span,
		}
	case lexer.MinusEqual:
		return lexer.Token{
			Type:    lexer.Minus,
			Lexeme:  "-",
			Literal: "-",
			Span:    augassign.Span,
		}
	case lexer.StarEqual:
		return lexer.Token{
			Type:    lexer.Star,
			Lexeme:  "*",
			Literal: "*",
			Span:    augassign.Span,
		}
	case lexer.AtEqual:
		return lexer.Token{
			Type:    lexer.At,
			Lexeme:  "@",
			Literal: "@",
			Span:    augassign.Span,
		}
	case lexer.SlashEqual:
		return lexer.Token{
			Type:    lexer.Slash,
			Lexeme:  "/",
			Literal: "/",
			Span:    augassign.Span,
		}
	case lexer.PercentEqual:
		return lexer.Token{
			Type:    lexer.Percent,
			Lexeme:  "%",
			Literal: "%",
			Span:    augassign.Span,
		}
	case lexer.AmpEqual:
		return lexer.Token{
			Type:    lexer.Ampersand,
			Lexeme:  "&",
			Literal: "&",
			Span:    augassign.Span,
		}
	case lexer.PipeEqual:
		return lexer.Token{
			Type:    lexer.Pipe,
			Lexeme:  "|",
			Literal: "|",
			Span:    augassign.Span,
		}
	case lexer.CaretEqual:
		return lexer.Token{
			Type:    lexer.Caret,
			Lexeme:  "^",
			Literal: "^",
			Span:    augassign.Span,
		}
	case lexer.LessLessEqual:
		return lexer.Token{
			Type:    lexer.LessLess,
			Lexeme:  "<<",
			Literal: "<<",
			Span:    augassign.Span,
		}
	case lexer.GreaterGreaterEqual:
		return lexer.Token{
			Type:    lexer.GreaterGreater,
			Lexeme:  ">>",
			Literal: ">>",
			Span:    augassign.Span,
		}
	case lexer.StarStarEqual:
		return lexer.Token{
			Type:    lexer.StarStar,
			Lexeme:  "**",
			Literal: "**",
			Span:    augassign.Span,
		}
	case lexer.SlashSlashEqual:
		return lexer.Token{
			Type:    lexer.SlashSlash,
			Lexeme:  "//",
			Literal: "//",
			Span:    augassign.Span,
		}
	}
	return lexer.Token{}
}
