package parser

import (
	"biscuit/compiler/ast"
	"biscuit/compiler/lexer"
)

func (p *Parser) delStatement() (ast.Stmt, error) {
	// Consume the 'del' keyword
	delToken, err := p.consume(lexer.Del, "expected 'del'")
	if err != nil {
		return nil, err
	}

	// Parse the del_targets
	targets, err := p.delTargets()
	if err != nil {
		return nil, err
	}

	// TODO: Create and return a DelStmt node
	// For now, return a placeholder
	return &ast.ExprStmt{
		Expr: targets,

		Span: lexer.Span{Start: delToken.Start(), End: targets.GetSpan().End},
	}, nil
}

// delTarget parses a target for the del statement as per the grammar:
// del_target:
//
//	| t_primary '.' NAME !t_lookahead
//	| t_primary '[' slices ']' !t_lookahead
//	| del_t_atom
func (p *Parser) delTarget() (ast.Expr, error) {
	// Try to parse as t_primary if the next token could start a t_primary
	if p.check(lexer.Identifier) || p.check(lexer.LeftParen) || p.check(lexer.LeftBracket) ||
		p.check(lexer.False) || p.check(lexer.True) || p.check(lexer.None) ||
		p.check(lexer.Number) || p.check(lexer.String) || p.check(lexer.Ellipsis) {

		// First, save the current position
		startPos := p.Current

		// Try to parse a t_primary followed by '.'
		primary, err := p.tPrimary()
		if err == nil && p.match(lexer.Dot) {
			// Handle attribute access: t_primary.NAME
			name, err := p.consume(lexer.Identifier, "expected identifier after '.'")
			if err != nil {
				return nil, err
			}
			result := &ast.Attribute{
				Object: primary,
				Name:   name,

				Span: lexer.Span{Start: primary.GetSpan().Start, End: name.End()},
			}

			// Check negative lookahead - must NOT be followed by another accessor
			if p.tLookahead() {
				return nil, p.error(p.peek(), "unexpected accessor after attribute in del target")
			}

			return result, nil
		}

		// Restore position and try t_primary followed by '['
		p.Current = startPos
		primary, err = p.tPrimary()
		if err == nil && p.match(lexer.LeftBracket) {
			// Handle subscript access: t_primary[slices]
			indices, err := p.slices()
			if err != nil {
				return nil, err
			}

			right, err := p.consume(lexer.RightBracket, "expected ']' after index")
			if err != nil {
				return nil, err
			}
			result := &ast.Subscript{
				Object:  primary,
				Indices: indices,

				Span: lexer.Span{Start: primary.GetSpan().Start, End: right.End()},
			}

			// Check negative lookahead - must NOT be followed by another accessor
			if p.tLookahead() {
				return nil, p.error(p.peek(), "unexpected accessor after subscript in del target")
			}

			return result, nil
		}

		// Reset position if we couldn't match t_primary with an accessor
		p.Current = startPos
	}

	// If we couldn't parse as t_primary with an accessor, try as del_t_atom
	return p.delTAtom()
}

// delTAtom parses a del_t_atom as per the grammar:
// del_t_atom:
//
//	| NAME
//	| '(' del_target ')'
//	| '(' [del_targets] ')'
//	| '[' [del_targets] ']'
func (p *Parser) delTAtom() (ast.Expr, error) {
	startPos := p.peek().Start()

	if p.check(lexer.Identifier) {
		// Handle simple NAME case
		name := p.advance()
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

		// Try to parse as single del_target first
		nextPos := p.Current
		target, err := p.delTarget()

		if err == nil {
			// Check if there's a comma after, which means it's a tuple
			if p.match(lexer.Comma) {
				// Start a tuple with the first target
				elements := []ast.Expr{target}

				// Check for empty rest of tuple
				if p.match(lexer.RightParen) {
					return &ast.TupleExpr{
						Elements: elements,

						Span: lexer.Span{Start: startPos, End: p.previous().End()},
					}, nil
				}

				// Parse rest of del_targets
				for !p.check(lexer.RightParen) {
					target, err = p.delTarget()
					if err != nil {
						return nil, err
					}
					elements = append(elements, target)

					if !p.match(lexer.Comma) {
						break
					}
				}

				// Consume closing parenthesis
				_, err = p.consume(lexer.RightParen, "expected ')' after del targets")
				if err != nil {
					return nil, err
				}

				return &ast.TupleExpr{
					Elements: elements,

					Span: lexer.Span{Start: startPos, End: p.previous().End()},
				}, nil
			}

			// No comma, so it's a grouped expression
			_, err = p.consume(lexer.RightParen, "expected ')' after target")
			if err != nil {
				return nil, err
			}
			return &ast.GroupExpr{
				Expression: target,

				Span: lexer.Span{Start: startPos, End: p.previous().End()},
			}, nil
		}

		// Restore position and try as del_targets (tuple)
		p.Current = nextPos

		// Parse del_targets as a sequence
		var elements []ast.Expr
		for !p.check(lexer.RightParen) {
			target, err := p.delTarget()
			if err != nil {
				return nil, err
			}
			elements = append(elements, target)

			if !p.match(lexer.Comma) {
				break
			}

			// Allow trailing comma
			if p.check(lexer.RightParen) {
				break
			}
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

		// Parse del_targets as a sequence
		var elements []ast.Expr
		for !p.check(lexer.RightBracket) {
			target, err := p.delTarget()
			if err != nil {
				return nil, err
			}
			elements = append(elements, target)

			if !p.match(lexer.Comma) {
				break
			}

			// Allow trailing comma
			if p.check(lexer.RightBracket) {
				break
			}
		}

		// Consume the closing bracket
		_, err := p.consume(lexer.RightBracket, "expected ']' after list targets")
		if err != nil {
			return nil, err
		}

		return &ast.ListExpr{
			Elements: elements,

			Span: lexer.Span{Start: startPos, End: p.previous().End()},
		}, nil
	}

	return nil, p.error(p.peek(), "expected NAME, '(' or '[' in del target atom")
}

// delTargets parses del targets as per the grammar:
// del_targets: ','.del_target+ [',']
func (p *Parser) delTargets() (ast.Expr, error) {
	// Parse the first del_target
	target, err := p.delTarget()
	if err != nil {
		return nil, err
	}

	// If there's no comma, return the single target
	if !p.match(lexer.Comma) {
		return target, nil
	}

	// We have a comma, so this is a tuple of targets
	elements := []ast.Expr{target}

	// Parse additional targets if any
	for !p.check(lexer.Newline) && !p.check(lexer.Semicolon) && !p.isAtEnd() {
		// Allow trailing comma
		if p.check(lexer.Newline) || p.check(lexer.Semicolon) || p.isAtEnd() {
			break
		}

		target, err = p.delTarget()
		if err != nil {
			return nil, err
		}
		elements = append(elements, target)

		// Expect a comma after each target except possibly the last
		if !p.match(lexer.Comma) {
			break
		}
	}

	// Create a tuple expression with the targets
	return &ast.TupleExpr{
		Elements: elements,

		Span: lexer.Span{
			Start: elements[0].GetSpan().Start,
			End:   elements[len(elements)-1].GetSpan().End,
		},
	}, nil
}
