package parser

import (
	"topple/compiler/ast"
	"topple/compiler/lexer"
)

// tPrimary parses a primary expression that must be followed by an accessor.
// This is used for parsing chained targets in assignment contexts.
// Implementation of the grammar rule:
//
// t_primary:
//
//	| t_primary '.' NAME &t_lookahead
//	| t_primary '[' slices ']' &t_lookahead
//	| t_primary genexp &t_lookahead
//	| t_primary '(' [arguments] ')' &t_lookahead
//	| atom &t_lookahead
func (p *Parser) tPrimary() (ast.Expr, error) {
	// Start with the base case: atom &t_lookahead
	expr, err := p.atom()
	if err != nil {
		return nil, err
	}

	// Check if there's a lookahead token - required for all t_primary rules
	// This implements the &t_lookahead part of the grammar
	// The grammar's &t_lookahead means "atom must have at least one accessor",
	// not "check lookahead after every accessor"
	if !p.tLookahead() {
		return nil, p.error(p.peek(), "expected accessor token ('.', '[', or '(')")
	}

	// Consume ALL accessors greedily, like primary() does
	// This fixes the bug where internal lookahead checks caused premature loop exit
	for p.tLookahead() {
		if p.match(lexer.Dot) {
			// Rule: t_primary '.' NAME
			name, err := p.consume(lexer.Identifier, "expected identifier after '.'")
			if err != nil {
				return nil, err
			}

			expr = &ast.Attribute{
				Object: expr,
				Name:   name,
				Span:   lexer.Span{Start: expr.GetSpan().Start, End: name.End()},
			}
		} else if p.match(lexer.LeftParen) {
			// Rule: t_primary '(' [arguments] ')'
			expr, err = p.finishCall(expr)
			if err != nil {
				return nil, err
			}
		} else if p.match(lexer.LeftBracket) {
			// Rule: t_primary '[' slices ']'
			indices, err := p.slices()
			if err != nil {
				return nil, err
			}

			right, err := p.consume(lexer.RightBracket, "expected ']' after index")
			if err != nil {
				return nil, err
			}

			expr = &ast.Subscript{
				Object:  expr,
				Indices: indices,
				Span:    lexer.Span{Start: expr.GetSpan().Start, End: right.End()},
			}
		} else {
			// Not an accessor token, stop
			break
		}
	}

	return expr, nil
}

// tLookahead checks if the current token is one of the t_lookahead tokens: '(', '[', '.'
func (p *Parser) tLookahead() bool {
	if p.isAtEnd() {
		return false
	}
	tokenType := p.peek().Type
	return tokenType == lexer.LeftParen || tokenType == lexer.LeftBracket || tokenType == lexer.Dot
}

// singleSubscriptAttributeTarget parses a single target with attribute or subscription
// as per the grammar:
//
// single_subscript_attribute_target:
//
//	| t_primary '.' NAME !t_lookahead
//	| t_primary '[' slices ']' !t_lookahead
func (p *Parser) singleSubscriptAttributeTarget() (ast.Expr, error) {
	// Parse the t_primary expression (which now greedily consumes all accessors)
	expr, err := p.tPrimary()
	if err != nil {
		return nil, err
	}

	// With the fixed greedy tPrimary(), it already consumed all accessors
	// Just verify it's an Attribute or Subscript (grammar requirement)
	switch expr.(type) {
	case *ast.Attribute, *ast.Subscript:
		// Check negative lookahead - must NOT be followed by another accessor
		if p.tLookahead() {
			return nil, p.error(p.peek(), "unexpected accessor after target")
		}
		return expr, nil
	case *ast.Call:
		// Function calls cannot be assignment targets
		return nil, p.error(p.peek(), "cannot assign to function call")
	default:
		// tPrimary() returned something other than Attribute/Subscript
		// This shouldn't happen with correct grammar, but handle gracefully
		return nil, p.error(p.peek(), "expected attribute or subscript target")
	}
}

// singleTarget parses a single target as per the grammar:
//
// single_target:
//
//	| single_subscript_attribute_target
//	| NAME
//	| '(' single_target ')'
func (p *Parser) singleTarget() (ast.Expr, error) {
	if p.check(lexer.Identifier) {
		// Save position in case we need to backtrack
		startPos := p.Current
		name := p.advance()

		// Check if this is a simple subscript like arr[0]
		if p.match(lexer.LeftBracket) {
			// Handle simple subscript: name[index]
			indices, err := p.slices()
			if err != nil {
				return nil, err
			}

			right, err := p.consume(lexer.RightBracket, "expected ']' after index")
			if err != nil {
				return nil, err
			}

			// Check if there's more chaining
			if !p.tLookahead() {
				// Simple subscript, return it
				return &ast.Subscript{
					Object: &ast.Name{
						Token: name,
						Span:  lexer.Span{Start: name.Start(), End: name.End()},
					},
					Indices: indices,
					Span:    lexer.Span{Start: name.Start(), End: right.End()},
				}, nil
			}

			// There's more chaining, restore position and use the complex path
			p.Current = startPos
			return p.singleSubscriptAttributeTarget()
		}

		// Check if it might be a single_subscript_attribute_target
		// by seeing if there's a lookahead accessor after the identifier
		if p.check(lexer.Dot) || p.check(lexer.LeftParen) {
			// Restore position and use the complex path
			p.Current = startPos
			return p.singleSubscriptAttributeTarget()
		}

		// Just a NAME
		return &ast.Name{
			Token: name,
			Span:  lexer.Span{Start: name.Start(), End: name.End()},
		}, nil
	} else if p.match(lexer.LeftParen) {
		// Handle parenthesized form: '(' single_target ')'
		target, err := p.singleTarget()
		if err != nil {
			return nil, err
		}

		_, err = p.consume(lexer.RightParen, "expected ')' after target")
		if err != nil {
			return nil, err
		}

		return &ast.GroupExpr{
			Expression: target,
			Span:       lexer.Span{Start: p.previous().Start(), End: p.previous().End()},
		}, nil
	}

	// Try to parse as single_subscript_attribute_target
	return p.singleSubscriptAttributeTarget()
}

// starTarget parses a star target as per the grammar:
//
// star_target:
//
//	| '*' (!'*' star_target)
//	| target_with_star_atom
func (p *Parser) starTarget() (ast.Expr, error) {
	if p.match(lexer.Star) {
		// Handle starred expression
		star := p.previous()

		// Check not followed by another star
		if p.check(lexer.Star) {
			return nil, p.error(p.peek(), "cannot use ** in target expressions")
		}

		// Parse the inner star_target
		expr, err := p.starTarget()
		if err != nil {
			return nil, err
		}

		return &ast.StarExpr{
			Expr: expr,

			Span: lexer.Span{Start: star.Start(), End: expr.GetSpan().End},
		}, nil
	}

	// Not a starred expression, parse as target_with_star_atom
	return p.targetWithStarAtom()
}

// starTargets parses star targets as per the grammar:
//
// star_targets:
//
//	| star_target !','
//	| star_target (',' star_target)* [',']
func (p *Parser) starTargets() ([]ast.Expr, error) {
	// Parse the first star_target
	target, err := p.starTarget()
	if err != nil {
		return nil, err
	}

	targets := []ast.Expr{target}

	// If there's no comma, return the single target
	if !p.match(lexer.Comma) {
		return targets, nil
	}

	// We've consumed a comma
	// Check if we're at the end (trailing comma case)
	if p.check(lexer.RightParen) || p.check(lexer.RightBracket) ||
		p.check(lexer.Colon) || p.check(lexer.Equal) || p.check(lexer.Newline) ||
		p.check(lexer.Semicolon) || p.isAtEnd() {
		// Just a trailing comma, we're done
		return targets, nil
	}

	// Keep parsing targets and commas
	for {
		target, err = p.starTarget()
		if err != nil {
			return nil, err
		}
		targets = append(targets, target)

		// If no more commas, we're done
		if !p.match(lexer.Comma) {
			break
		}

		// If we've found a trailing comma, we're done
		if p.check(lexer.RightParen) || p.check(lexer.RightBracket) ||
			p.check(lexer.Colon) || p.check(lexer.Equal) || p.check(lexer.Newline) ||
			p.check(lexer.Semicolon) || p.isAtEnd() {
			break
		}
	}

	// Validate that only one starred expression exists at this level
	if err := p.validateStarredTargets(targets); err != nil {
		return nil, err
	}

	return targets, nil
}

// parseStarTargetSequence parses a sequence of star targets for either list or tuple contexts
// isTuple indicates whether tuple rules should be enforced (requiring comma for single element)
// closingToken specifies the token that would terminate the sequence (RightParen or RightBracket)
func (p *Parser) parseStarTargetSequence(isTuple bool, closingToken lexer.TokenType) ([]ast.Expr, error) {
	// Parse the first star_target
	target, err := p.starTarget()
	if err != nil {
		return nil, err
	}

	elements := []ast.Expr{target}

	// For tuples, a comma is required
	if isTuple && !p.match(lexer.Comma) {
		return nil, p.error(p.peek(), "expected ',' after target in tuple")
	} else if !isTuple {
		// For lists, comma is optional
		if !p.match(lexer.Comma) {
			return elements, nil
		}
	}

	// Check if it's just a single-element tuple (star_target ',')
	if isTuple && p.check(closingToken) {
		return elements, nil
	}

	// If we don't have another target after comma for a list, we're done
	if !isTuple && p.check(closingToken) {
		return elements, nil
	}

	// Parse additional targets if there are any
	if !p.check(closingToken) {
		// For tuples, we need at least a second element if not single-element
		if isTuple {
			target, err = p.starTarget()
			if err != nil {
				return nil, err
			}
			elements = append(elements, target)

			// Parse more elements with commas
			for p.match(lexer.Comma) {
				// Check for trailing comma
				if p.check(closingToken) {
					break
				}

				target, err = p.starTarget()
				if err != nil {
					return nil, err
				}
				elements = append(elements, target)
			}
		} else {
			// For lists, parse additional elements while there are commas
			for {
				target, err = p.starTarget()
				if err != nil {
					return nil, err
				}
				elements = append(elements, target)

				if !p.match(lexer.Comma) {
					break
				}

				// Check for trailing comma
				if p.check(closingToken) {
					break
				}
			}
		}
	}

	return elements, nil
}

// starAtom parses a star atom as per the grammar:
//
// star_atom:
//
//	| NAME
//	| '(' target_with_star_atom ')'
//	| '(' [star_targets_tuple_seq] ')'
//	| '[' [star_targets_list_seq] ']'
func (p *Parser) starAtom() (ast.Expr, error) {
	startPos := p.peek().Start()

	if p.match(lexer.Identifier) {
		// Handle simple NAME case
		name := p.previous()
		return &ast.Name{
			Token: name,

			Span: lexer.Span{Start: startPos, End: name.End()},
		}, nil
	} else if p.match(lexer.LeftParen) {
		// Handle parenthesized forms
		if p.match(lexer.RightParen) {
			// Empty tuple
			return &ast.TupleExpr{
				Elements: []ast.Expr{},

				Span: lexer.Span{Start: startPos, End: p.previous().End()},
			}, nil
		}

		// Try to parse as target_with_star_atom first
		nextPos := p.Current
		target, err := p.targetWithStarAtom()

		if err == nil {
			// Successfully parsed as target_with_star_atom
			// Consume the closing parenthesis
			_, err = p.consume(lexer.RightParen, "expected ')' after target")
			if err != nil {
				goto tryStartTargetSequence
			}
			return &ast.GroupExpr{
				Expression: target,

				Span: lexer.Span{Start: startPos, End: p.previous().End()},
			}, nil
		}

	tryStartTargetSequence:
		// Restore position and try as star_targets_tuple_seq
		p.Current = nextPos

		elements, err := p.parseStarTargetSequence(true, lexer.RightParen)
		if err != nil {
			return nil, err
		}

		// Consume the closing parenthesis
		_, err = p.consume(lexer.RightParen, "expected ')' after tuple targets")
		if err != nil {
			return nil, err
		}

		return &ast.TupleExpr{
			Elements: elements,

			Span: lexer.Span{Start: startPos, End: p.previous().End()},
		}, nil
	} else if p.match(lexer.LeftBracket) {
		// Handle list form
		if p.match(lexer.RightBracket) {
			// Empty list
			return &ast.ListExpr{
				Elements: []ast.Expr{},

				Span: lexer.Span{Start: startPos, End: p.previous().End()},
			}, nil
		}

		elements, err := p.parseStarTargetSequence(false, lexer.RightBracket)
		if err != nil {
			return nil, err
		}

		// Consume the closing bracket
		_, err = p.consume(lexer.RightBracket, "expected ']' after list targets")
		if err != nil {
			return nil, err
		}

		return &ast.ListExpr{
			Elements: elements,

			Span: lexer.Span{Start: startPos, End: p.previous().End()},
		}, nil
	}

	return nil, p.error(p.peek(), "expected NAME, '(' or '[' in star atom")
}

// targetWithStarAtom parses a target with star atom as per the grammar:
//
// target_with_star_atom:
//
//	| t_primary '.' NAME !t_lookahead
//	| t_primary '[' slices ']' !t_lookahead
//	| star_atom
func (p *Parser) targetWithStarAtom() (ast.Expr, error) {
	// Try to parse as t_primary if the next token could start one
	// With the fixed greedy tPrimary(), it will consume all chained accessors
	if p.check(lexer.Identifier) || p.check(lexer.LeftParen) || p.check(lexer.LeftBracket) ||
		p.check(lexer.False) || p.check(lexer.True) || p.check(lexer.None) ||
		p.check(lexer.Number) || p.check(lexer.String) || p.check(lexer.Ellipsis) {

		startPos := p.Current

		// Try to parse a t_primary (which now greedily consumes all accessors)
		primary, err := p.tPrimary()
		if err != nil {
			// Failed to parse as t_primary, restore and try star_atom
			p.Current = startPos
			goto tryStarAtom
		}

		// tPrimary() already consumed all accessors, just verify no more follow
		// (implements the !t_lookahead in the grammar)
		if p.tLookahead() {
			return nil, p.error(p.peek(), "unexpected accessor after target")
		}

		// Check that tPrimary() returned a valid target type
		// The grammar requires that the chain ends with .NAME or [slices], not ()
		switch primary.(type) {
		case *ast.Attribute, *ast.Subscript:
			// Valid: chain ended with .NAME or [slices]
			return primary, nil
		case *ast.Call:
			// Invalid: chain ended with () - function calls cannot be assignment targets
			return nil, p.error(p.peek(), "cannot assign to function call")
		default:
			// Shouldn't happen with greedy tPrimary(), but handle gracefully
			p.Current = startPos
			goto tryStarAtom
		}
	}

tryStarAtom:
	// If we couldn't parse as t_primary, try as star_atom
	return p.starAtom()
}

// validateStarredTargets checks that only one starred expression exists at the current nesting level
func (p *Parser) validateStarredTargets(targets []ast.Expr) error {
	starCount := 0
	for _, target := range targets {
		if _, isStar := target.(*ast.StarExpr); isStar {
			starCount++
			if starCount > 1 {
				return p.error(p.peek(), "multiple starred expressions in assignment")
			}
		}
	}
	return nil
}
