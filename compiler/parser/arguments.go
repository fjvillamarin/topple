package parser

import (
	"biscuit/compiler/ast"
	"biscuit/compiler/lexer"
)

// arguments parses function call arguments according to the grammar:
//
// arguments:
//
//	| args [','] &')'
//
// This is used when parsing function calls in expressions.
func (p *Parser) arguments() ([]ast.Argument, error) {
	// Empty argument list
	if p.check(lexer.RightParen) {
		return []ast.Argument{}, nil
	}

	return p.args()
}

// args parses a sequence of arguments according to the grammar:
//
// args:
//
//	| ','.(starred_expression | ( assignment_expression | expression !':=') !'=')+ [',' kwargs]
//	| kwargs
func (p *Parser) args() ([]ast.Argument, error) {
	// Try to parse as kwargs first
	if p.check(lexer.StarStar) || (p.check(lexer.Identifier) && p.checkNext(lexer.Equal)) {
		return p.kwargs()
	}

	// Parse regular arguments or starred expressions
	args := []ast.Argument{}

	// Parse the first argument
	arg, err := p.parseArg()
	if err != nil {
		return nil, err
	}
	args = append(args, arg)

	// Parse additional arguments separated by commas
	for p.match(lexer.Comma) {
		// If there's a trailing comma, we're done
		if p.check(lexer.RightParen) {
			break
		}

		// Check if we've hit keyword arguments
		if p.check(lexer.StarStar) || (p.check(lexer.Identifier) && p.checkNext(lexer.Equal)) {
			// Parse the kwargs and combine with existing args
			kwArgs, err := p.kwargs()
			if err != nil {
				return nil, err
			}
			args = append(args, kwArgs...)
			break
		}

		// Parse the next argument
		arg, err = p.parseArg()
		if err != nil {
			return nil, err
		}
		args = append(args, arg)
	}

	return args, nil
}

// kwargs parses keyword arguments according to the grammar:
//
// kwargs:
//
//	| ','.kwarg_or_starred+ ',' ','.kwarg_or_double_starred+
//	| ','.kwarg_or_starred+
//	| ','.kwarg_or_double_starred+
func (p *Parser) kwargs() ([]ast.Argument, error) {
	args := []ast.Argument{}

	// Check if we start with a double-starred expression
	if p.check(lexer.StarStar) {
		// Parse a kwarg_or_double_starred
		arg, err := p.parseKwargOrDoubleStar()
		if err != nil {
			return nil, err
		}
		args = append(args, arg)

		// Parse additional kwarg_or_double_starred separated by commas
		for p.match(lexer.Comma) {
			// If there's a trailing comma, we're done
			if p.check(lexer.RightParen) {
				break
			}

			arg, err = p.parseKwargOrDoubleStar()
			if err != nil {
				return nil, err
			}
			args = append(args, arg)
		}
	} else {
		// Parse a kwarg_or_starred
		arg, err := p.parseKwargOrStar()
		if err != nil {
			return nil, err
		}
		args = append(args, arg)

		// Parse additional kwarg_or_starred separated by commas
		hasKwargOrDoubleStar := false
		for p.match(lexer.Comma) {
			// If there's a trailing comma, we're done
			if p.check(lexer.RightParen) {
				break
			}

			// Check if we've switched to kwarg_or_double_starred
			if !hasKwargOrDoubleStar && p.check(lexer.StarStar) {
				hasKwargOrDoubleStar = true
				// Parse a kwarg_or_double_starred
				arg, err = p.parseKwargOrDoubleStar()
				if err != nil {
					return nil, err
				}
			} else if hasKwargOrDoubleStar {
				// Continue parsing kwarg_or_double_starred
				arg, err = p.parseKwargOrDoubleStar()
				if err != nil {
					return nil, err
				}
			} else {
				// Continue parsing kwarg_or_starred
				arg, err = p.parseKwargOrStar()
				if err != nil {
					return nil, err
				}
			}
			args = append(args, arg)
		}
	}

	return args, nil
}

// parseArg parses a single argument (non-keyword)
// This handles: starred_expression | expression
func (p *Parser) parseArg() (ast.Argument, error) {
	startPos := p.peek().Start()

	// Check for starred expression: '*' expression
	if p.match(lexer.Star) {
		star := p.previous()
		expr, err := p.expression()
		if err != nil {
			return ast.Argument{}, err
		}

		return ast.Argument{
			Value:        expr,
			IsStar:       true,
			IsDoubleStar: false,
			Span:         lexer.Span{Start: star.Start(), End: expr.GetSpan().End},
		}, nil
	}

	// Regular expression
	expr, err := p.expression()
	if err != nil {
		return ast.Argument{}, err
	}

	// Check that it's not followed by '=' (which would make it a keyword arg)
	if p.check(lexer.Equal) {
		return ast.Argument{}, p.error(p.peek(), "unexpected '=' in argument")
	}

	return ast.Argument{
		Value:        expr,
		IsStar:       false,
		IsDoubleStar: false,
		Span:         lexer.Span{Start: startPos, End: expr.GetSpan().End},
	}, nil
}

// parseKwargOrStar parses a kwarg_or_starred according to the grammar:
//
// kwarg_or_starred:
//
//	| NAME '=' expression
//	| starred_expression
func (p *Parser) parseKwargOrStar() (ast.Argument, error) {
	startPos := p.peek().Start()

	// Check for starred expression: '*' expression
	if p.match(lexer.Star) {
		star := p.previous()
		expr, err := p.expression()
		if err != nil {
			return ast.Argument{}, err
		}

		return ast.Argument{
			Value:        expr,
			IsStar:       true,
			IsDoubleStar: false,
			Span:         lexer.Span{Start: star.Start(), End: expr.GetSpan().End},
		}, nil
	}

	// Must be a keyword argument: NAME '=' expression
	if !p.check(lexer.Identifier) {
		return ast.Argument{}, p.error(p.peek(), "expected identifier or '*' in argument")
	}

	nameToken := p.advance()
	name := &ast.Name{
		Token: nameToken,
		Span:  lexer.Span{Start: nameToken.Start(), End: nameToken.End()},
	}

	_, err := p.consume(lexer.Equal, "expected '=' after identifier in keyword argument")
	if err != nil {
		return ast.Argument{}, err
	}

	expr, err := p.expression()
	if err != nil {
		return ast.Argument{}, err
	}

	return ast.Argument{
		Name:         name,
		Value:        expr,
		IsStar:       false,
		IsDoubleStar: false,
		Span:         lexer.Span{Start: startPos, End: expr.GetSpan().End},
	}, nil
}

// parseKwargOrDoubleStar parses a kwarg_or_double_starred according to the grammar:
//
// kwarg_or_double_starred:
//
//	| NAME '=' expression
//	| '**' expression
func (p *Parser) parseKwargOrDoubleStar() (ast.Argument, error) {
	startPos := p.peek().Start()

	// Check for double-starred expression: '**' expression
	if p.match(lexer.StarStar) {
		star := p.previous()
		expr, err := p.expression()
		if err != nil {
			return ast.Argument{}, err
		}

		return ast.Argument{
			Value:        expr,
			IsStar:       false,
			IsDoubleStar: true,
			Span:         lexer.Span{Start: star.Start(), End: expr.GetSpan().End},
		}, nil
	}

	// Must be a keyword argument: NAME '=' expression
	if !p.check(lexer.Identifier) {
		return ast.Argument{}, p.error(p.peek(), "expected identifier or '**' in argument")
	}

	nameToken := p.advance()
	name := &ast.Name{
		Token: nameToken,
		Span:  lexer.Span{Start: nameToken.Start(), End: nameToken.End()},
	}

	_, err := p.consume(lexer.Equal, "expected '=' after identifier in keyword argument")
	if err != nil {
		return ast.Argument{}, err
	}

	expr, err := p.expression()
	if err != nil {
		return ast.Argument{}, err
	}

	return ast.Argument{
		Name:         name,
		Value:        expr,
		IsStar:       false,
		IsDoubleStar: false,
		Span:         lexer.Span{Start: startPos, End: expr.GetSpan().End},
	}, nil
}
