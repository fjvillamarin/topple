package parser

import (
	"biscuit/compiler/ast"
	"biscuit/compiler/lexer"
)

// matchStmt parses a match statement according to the grammar:
// match_stmt: "match" subject_expr ':' NEWLINE INDENT case_block+ DEDENT
func (p *Parser) matchStmt() (ast.Stmt, error) {
	matchToken, err := p.consume(lexer.Match, "expected 'match'")
	if err != nil {
		return nil, err
	}

	// Parse subject expression
	subject, err := p.subjectExpr()
	if err != nil {
		return nil, err
	}

	// Consume ':'
	_, err = p.consume(lexer.Colon, "expected ':' after match subject")
	if err != nil {
		return nil, err
	}

	// Consume NEWLINE
	_, err = p.consume(lexer.Newline, "expected newline after ':'")
	if err != nil {
		return nil, err
	}

	// Consume all remaining newlines (handles comments)
	for p.check(lexer.Newline) {
		p.advance()
	}

	// Consume INDENT
	_, err = p.consume(lexer.Indent, "expected indented block after match")
	if err != nil {
		return nil, err
	}

	// Parse one or more case blocks
	var cases []ast.CaseBlock
	for p.check(lexer.Case) {
		caseBlock, err := p.caseBlock()
		if err != nil {
			return nil, err
		}
		cases = append(cases, caseBlock)

		// Skip newlines between case blocks
		for p.check(lexer.Newline) {
			p.advance()
		}
	}

	if len(cases) == 0 {
		return nil, p.error(p.peek(), "expected at least one case block")
	}

	// Consume DEDENT
	dedentToken, err := p.consume(lexer.Dedent, "expected dedent after match cases")
	if err != nil {
		return nil, err
	}

	return &ast.MatchStmt{
		Subject: subject,
		Cases:   cases,
		Span:    lexer.Span{Start: matchToken.Start(), End: dedentToken.End()},
	}, nil
}

// subjectExpr parses the subject expression for a match statement:
// subject_expr: star_named_expression ',' star_named_expressions? | named_expression
func (p *Parser) subjectExpr() (ast.Expr, error) {
	// Parse the first expression
	first, err := p.starNamedExpression()
	if err != nil {
		return nil, err
	}

	// Check if there's a comma (indicating multiple expressions)
	if p.match(lexer.Comma) {
		elements := []ast.Expr{first}

		// Parse additional expressions if present
		if !p.check(lexer.Colon) {
			for {
				expr, err := p.starNamedExpression()
				if err != nil {
					return nil, err
				}
				elements = append(elements, expr)

				if !p.match(lexer.Comma) {
					break
				}

				// Allow trailing comma
				if p.check(lexer.Colon) {
					break
				}
			}
		}

		// Return as tuple expression
		return &ast.TupleExpr{
			Elements: elements,
			Span:     lexer.Span{Start: first.GetSpan().Start, End: elements[len(elements)-1].GetSpan().End},
		}, nil
	}

	return first, nil
}

// caseBlock parses a case block according to the grammar:
// case_block: "case" patterns guard? ':' block
func (p *Parser) caseBlock() (ast.CaseBlock, error) {
	caseToken, err := p.consume(lexer.Case, "expected 'case'")
	if err != nil {
		return ast.CaseBlock{}, err
	}

	// Parse patterns
	patterns, err := p.patterns()
	if err != nil {
		return ast.CaseBlock{}, err
	}

	// Parse optional guard
	var guard ast.Expr
	if p.match(lexer.If) {
		guard, err = p.namedExpression()
		if err != nil {
			return ast.CaseBlock{}, err
		}
	}

	// Consume ':'
	_, err = p.consume(lexer.Colon, "expected ':' after case pattern")
	if err != nil {
		return ast.CaseBlock{}, err
	}

	// Parse block
	body, err := p.block()
	if err != nil {
		return ast.CaseBlock{}, err
	}

	return ast.CaseBlock{
		Patterns: patterns,
		Guard:    guard,
		Body:     body,
		Span:     lexer.Span{Start: caseToken.Start(), End: body[len(body)-1].GetSpan().End},
	}, nil
}

// patterns parses patterns according to the grammar:
// patterns: open_sequence_pattern | pattern
func (p *Parser) patterns() ([]ast.Pattern, error) {
	// For now, we'll simplify and just parse or_pattern which handles alternatives
	pattern, err := p.orPattern()
	if err != nil {
		return nil, err
	}

	// Check if this is an or_pattern with multiple alternatives
	if orPat, ok := pattern.(*ast.OrPattern); ok {
		return orPat.Patterns, nil
	}

	return []ast.Pattern{pattern}, nil
}

// orPattern parses or patterns according to the grammar:
// or_pattern: '|'.closed_pattern+
func (p *Parser) orPattern() (ast.Pattern, error) {
	patterns := []ast.Pattern{}

	// Parse the first pattern
	pattern, err := p.asPattern()
	if err != nil {
		return nil, err
	}
	patterns = append(patterns, pattern)

	// Parse additional patterns separated by '|'
	for p.match(lexer.Pipe) {
		pattern, err := p.asPattern()
		if err != nil {
			return nil, err
		}
		patterns = append(patterns, pattern)
	}

	// If there's only one pattern, return it directly
	if len(patterns) == 1 {
		return patterns[0], nil
	}

	// Return as OrPattern
	return &ast.OrPattern{
		Patterns: patterns,
		Span:     lexer.Span{Start: patterns[0].GetSpan().Start, End: patterns[len(patterns)-1].GetSpan().End},
	}, nil
}

// asPattern parses as patterns according to the grammar:
// as_pattern: or_pattern 'as' pattern_capture_target
func (p *Parser) asPattern() (ast.Pattern, error) {
	pattern, err := p.closedPattern()
	if err != nil {
		return nil, err
	}

	// Check for 'as' keyword
	if p.match(lexer.As) {
		// Parse capture target
		if !p.check(lexer.Identifier) {
			return nil, p.error(p.peek(), "expected identifier after 'as'")
		}

		nameToken, err := p.consume(lexer.Identifier, "expected identifier")
		if err != nil {
			return nil, err
		}

		target := &ast.Name{
			Token: nameToken,
			Span:  lexer.Span{Start: nameToken.Start(), End: nameToken.End()},
		}

		return &ast.AsPattern{
			Pattern: pattern,
			Target:  target,
			Span:    lexer.Span{Start: pattern.GetSpan().Start, End: target.GetSpan().End},
		}, nil
	}

	return pattern, nil
}

// closedPattern parses closed patterns according to the grammar:
// closed_pattern: literal_pattern | capture_pattern | wildcard_pattern | value_pattern | group_pattern | sequence_pattern | mapping_pattern | class_pattern
func (p *Parser) closedPattern() (ast.Pattern, error) {
	switch {
	// Wildcard pattern: _
	case p.check(lexer.Identifier) && p.peek().Lexeme == "_":
		token, _ := p.consume(lexer.Identifier, "")
		return &ast.WildcardPattern{
			Span: lexer.Span{Start: token.Start(), End: token.End()},
		}, nil

	// Literal patterns: numbers, strings, None, True, False
	case p.check(lexer.Number), p.check(lexer.String), p.check(lexer.None), p.check(lexer.True), p.check(lexer.False):
		return p.literalPattern()

	// Group pattern or tuple pattern: (pattern)
	case p.check(lexer.LeftParen):
		return p.groupOrSequencePattern(true)

	// List pattern: [pattern]
	case p.check(lexer.LeftBracket):
		return p.groupOrSequencePattern(false)

	// Mapping pattern: {pattern}
	case p.check(lexer.LeftBrace):
		return p.mappingPattern()

	// Identifier: could be capture pattern or value pattern
	case p.check(lexer.Identifier):
		return p.identifierPattern()

	default:
		return nil, p.error(p.peek(), "expected pattern")
	}
}

// literalPattern parses literal patterns
func (p *Parser) literalPattern() (ast.Pattern, error) {
	var expr ast.Expr

	switch {
	case p.check(lexer.Number):
		token, _ := p.consume(lexer.Number, "")
		expr = &ast.Literal{
			Value: token.Literal,
			Token: token,
			Span:  lexer.Span{Start: token.Start(), End: token.End()},
		}
	case p.check(lexer.String):
		token, _ := p.consume(lexer.String, "")
		expr = &ast.Literal{
			Value: token.Literal,
			Token: token,
			Span:  lexer.Span{Start: token.Start(), End: token.End()},
		}
	case p.check(lexer.None):
		token, _ := p.consume(lexer.None, "")
		expr = &ast.Literal{
			Value: nil,
			Token: token,
			Span:  lexer.Span{Start: token.Start(), End: token.End()},
		}
	case p.check(lexer.True):
		token, _ := p.consume(lexer.True, "")
		expr = &ast.Literal{
			Value: true,
			Token: token,
			Span:  lexer.Span{Start: token.Start(), End: token.End()},
		}
	case p.check(lexer.False):
		token, _ := p.consume(lexer.False, "")
		expr = &ast.Literal{
			Value: false,
			Token: token,
			Span:  lexer.Span{Start: token.Start(), End: token.End()},
		}
	default:
		return nil, p.error(p.peek(), "expected literal")
	}

	return &ast.LiteralPattern{
		Value: expr,
		Span:  expr.GetSpan(),
	}, nil
}

// identifierPattern determines if an identifier is a capture pattern or value pattern
func (p *Parser) identifierPattern() (ast.Pattern, error) {
	nameToken, err := p.consume(lexer.Identifier, "expected identifier")
	if err != nil {
		return nil, err
	}

	name := &ast.Name{
		Token: nameToken,
		Span:  lexer.Span{Start: nameToken.Start(), End: nameToken.End()},
	}

	// Check if this is a dotted name (value pattern)
	if p.check(lexer.Dot) {
		// Parse as attribute access (value pattern)
		expr := ast.Expr(name)
		for p.match(lexer.Dot) {
			attrName, err := p.consume(lexer.Identifier, "expected identifier after '.'")
			if err != nil {
				return nil, err
			}
			expr = &ast.Attribute{
				Object: expr,
				Name:   attrName,
				Span:   lexer.Span{Start: expr.GetSpan().Start, End: attrName.End()},
			}
		}
		return &ast.ValuePattern{
			Value: expr,
			Span:  expr.GetSpan(),
		}, nil
	}

	// Simple identifier - capture pattern
	return &ast.CapturePattern{
		Name: name,
		Span: name.GetSpan(),
	}, nil
}

// groupOrSequencePattern parses group patterns (parentheses) or sequence patterns (brackets)
func (p *Parser) groupOrSequencePattern(isParens bool) (ast.Pattern, error) {
	var leftToken, rightToken lexer.Token
	var err error

	if isParens {
		leftToken, err = p.consume(lexer.LeftParen, "expected '('")
		if err != nil {
			return nil, err
		}
	} else {
		leftToken, err = p.consume(lexer.LeftBracket, "expected '['")
		if err != nil {
			return nil, err
		}
	}

	// Empty sequence
	if (isParens && p.check(lexer.RightParen)) || (!isParens && p.check(lexer.RightBracket)) {
		if isParens {
			rightToken, err = p.consume(lexer.RightParen, "expected ')'")
		} else {
			rightToken, err = p.consume(lexer.RightBracket, "expected ']'")
		}
		if err != nil {
			return nil, err
		}

		return &ast.SequencePattern{
			Patterns: []ast.Pattern{},
			IsTuple:  isParens,
			Span:     lexer.Span{Start: leftToken.Start(), End: rightToken.End()},
		}, nil
	}

	// Parse patterns
	var patterns []ast.Pattern
	for {
		pattern, err := p.closedPattern()
		if err != nil {
			return nil, err
		}
		patterns = append(patterns, pattern)

		if !p.match(lexer.Comma) {
			break
		}

		// Allow trailing comma
		if (isParens && p.check(lexer.RightParen)) || (!isParens && p.check(lexer.RightBracket)) {
			break
		}
	}

	if isParens {
		rightToken, err = p.consume(lexer.RightParen, "expected ')'")
	} else {
		rightToken, err = p.consume(lexer.RightBracket, "expected ']'")
	}
	if err != nil {
		return nil, err
	}

	// Single pattern in parentheses is a group pattern, not a sequence
	if isParens && len(patterns) == 1 {
		return &ast.GroupPattern{
			Pattern: patterns[0],
			Span:    lexer.Span{Start: leftToken.Start(), End: rightToken.End()},
		}, nil
	}

	return &ast.SequencePattern{
		Patterns: patterns,
		IsTuple:  isParens,
		Span:     lexer.Span{Start: leftToken.Start(), End: rightToken.End()},
	}, nil
}

// mappingPattern parses mapping patterns (dictionaries)
func (p *Parser) mappingPattern() (ast.Pattern, error) {
	leftBrace, err := p.consume(lexer.LeftBrace, "expected '{'")
	if err != nil {
		return nil, err
	}

	// Empty mapping
	if p.check(lexer.RightBrace) {
		rightBrace, err := p.consume(lexer.RightBrace, "expected '}'")
		if err != nil {
			return nil, err
		}
		return &ast.MappingPattern{
			Pairs:   []ast.MappingPatternPair{},
			HasRest: false,
			Span:    lexer.Span{Start: leftBrace.Start(), End: rightBrace.End()},
		}, nil
	}

	var pairs []ast.MappingPatternPair
	var doubleStar ast.Pattern
	hasRest := false

	for {
		// Check for **pattern
		if p.match(lexer.StarStar) {
			if hasRest {
				return nil, p.error(p.previous(), "only one **pattern allowed in mapping pattern")
			}
			pattern, err := p.closedPattern()
			if err != nil {
				return nil, err
			}
			doubleStar = pattern
			hasRest = true
		} else {
			// Parse key: pattern
			key, err := p.expression()
			if err != nil {
				return nil, err
			}

			_, err = p.consume(lexer.Colon, "expected ':' after mapping pattern key")
			if err != nil {
				return nil, err
			}

			pattern, err := p.closedPattern()
			if err != nil {
				return nil, err
			}

			pairs = append(pairs, ast.MappingPatternPair{
				Key:     key,
				Pattern: pattern,
				Span:    lexer.Span{Start: key.GetSpan().Start, End: pattern.GetSpan().End},
			})
		}

		if !p.match(lexer.Comma) {
			break
		}

		// Allow trailing comma
		if p.check(lexer.RightBrace) {
			break
		}
	}

	rightBrace, err := p.consume(lexer.RightBrace, "expected '}'")
	if err != nil {
		return nil, err
	}

	return &ast.MappingPattern{
		Pairs:      pairs,
		DoubleStar: doubleStar,
		HasRest:    hasRest,
		Span:       lexer.Span{Start: leftBrace.Start(), End: rightBrace.End()},
	}, nil
}
