package parser

import (
	"github.com/fjvillamarin/topple/compiler/ast"
	"github.com/fjvillamarin/topple/compiler/lexer"
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
	// Check for open sequence pattern (comma-separated patterns)
	if p.checkOpenSequencePattern() {
		return p.openSequencePattern()
	}

	// Parse single pattern
	pattern, err := p.pattern()
	if err != nil {
		return nil, err
	}

	return []ast.Pattern{pattern}, nil
}

// checkOpenSequencePattern checks if we have an open sequence pattern
func (p *Parser) checkOpenSequencePattern() bool {
	// Look ahead to see if we have a pattern followed by a comma
	// This is a simplified check - in practice we'd need more sophisticated lookahead
	return false // For now, we'll handle this in the main patterns function
}

// openSequencePattern parses open sequence patterns:
// open_sequence_pattern: maybe_star_pattern ',' maybe_sequence_pattern?
func (p *Parser) openSequencePattern() ([]ast.Pattern, error) {
	var patterns []ast.Pattern

	// Parse first pattern
	pattern, err := p.maybeStarPattern()
	if err != nil {
		return nil, err
	}
	patterns = append(patterns, pattern)

	// Consume comma
	if !p.match(lexer.Comma) {
		return patterns, nil
	}

	// Parse remaining patterns if present
	if !p.check(lexer.Colon) && !p.check(lexer.If) {
		remaining, err := p.maybeSequencePattern()
		if err != nil {
			return nil, err
		}
		patterns = append(patterns, remaining...)
	}

	return patterns, nil
}

// maybeSequencePattern parses maybe_sequence_pattern:
// maybe_sequence_pattern: ','.maybe_star_pattern+ ','?
func (p *Parser) maybeSequencePattern() ([]ast.Pattern, error) {
	var patterns []ast.Pattern

	for {
		pattern, err := p.maybeStarPattern()
		if err != nil {
			return nil, err
		}
		patterns = append(patterns, pattern)

		if !p.match(lexer.Comma) {
			break
		}

		// Allow trailing comma
		if p.check(lexer.Colon) || p.check(lexer.If) {
			break
		}
	}

	return patterns, nil
}

// maybeStarPattern parses maybe_star_pattern:
// maybe_star_pattern: star_pattern | pattern
func (p *Parser) maybeStarPattern() (ast.Pattern, error) {
	if p.check(lexer.Star) {
		return p.starPattern()
	}
	return p.pattern()
}

// starPattern parses star patterns:
// star_pattern: '*' (capture_pattern | wildcard_pattern)
func (p *Parser) starPattern() (ast.Pattern, error) {
	starToken, err := p.consume(lexer.Star, "expected '*'")
	if err != nil {
		return nil, err
	}

	var pattern ast.Pattern

	if p.check(lexer.Identifier) && p.peek().Lexeme == "_" {
		// Wildcard pattern
		token, _ := p.consume(lexer.Identifier, "")
		pattern = &ast.WildcardPattern{
			Span: lexer.Span{Start: token.Start(), End: token.End()},
		}
	} else if p.check(lexer.Identifier) {
		// Capture pattern
		nameToken, err := p.consume(lexer.Identifier, "expected identifier")
		if err != nil {
			return nil, err
		}
		name := &ast.Name{
			Token: nameToken,
			Span:  lexer.Span{Start: nameToken.Start(), End: nameToken.End()},
		}
		pattern = &ast.CapturePattern{
			Name: name,
			Span: name.GetSpan(),
		}
	} else {
		return nil, p.error(p.peek(), "expected identifier or '_' after '*'")
	}

	return &ast.StarPattern{
		Pattern: pattern,
		Span:    lexer.Span{Start: starToken.Start(), End: pattern.GetSpan().End},
	}, nil
}

// pattern parses a single pattern:
// pattern: as_pattern | or_pattern
func (p *Parser) pattern() (ast.Pattern, error) {
	return p.asPattern()
}

// orPattern parses or patterns according to the grammar:
// or_pattern: '|'.closed_pattern+
func (p *Parser) orPattern() (ast.Pattern, error) {
	patterns := []ast.Pattern{}

	// Parse the first pattern
	pattern, err := p.closedPattern()
	if err != nil {
		return nil, err
	}
	patterns = append(patterns, pattern)

	// Parse additional patterns separated by '|'
	for p.match(lexer.Pipe) {
		pattern, err := p.closedPattern()
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
	pattern, err := p.orPattern()
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

	// Signed number (negative numbers)
	case p.check(lexer.Minus):
		return p.signedNumberPattern()

	// Group pattern or tuple pattern: (pattern)
	case p.check(lexer.LeftParen):
		return p.groupOrSequencePattern(true)

	// List pattern: [pattern]
	case p.check(lexer.LeftBracket):
		return p.sequencePattern(false)

	// Mapping pattern: {pattern}
	case p.check(lexer.LeftBrace):
		return p.mappingPattern()

	// Identifier: could be capture pattern, value pattern, or class pattern
	case p.check(lexer.Identifier):
		return p.identifierPattern()

	default:
		return nil, p.error(p.peek(), "expected pattern")
	}
}

// signedNumberPattern parses signed number patterns (negative numbers and complex numbers)
func (p *Parser) signedNumberPattern() (ast.Pattern, error) {
	minusToken, err := p.consume(lexer.Minus, "expected '-'")
	if err != nil {
		return nil, err
	}

	if !p.check(lexer.Number) {
		return nil, p.error(p.peek(), "expected number after '-'")
	}

	numberToken, err := p.consume(lexer.Number, "expected number")
	if err != nil {
		return nil, err
	}

	// Create a negative number literal
	var expr ast.Expr = &ast.Unary{
		Operator: minusToken,
		Right: &ast.Literal{
			Value: numberToken.Literal,
			Token: numberToken,
			Span:  lexer.Span{Start: numberToken.Start(), End: numberToken.End()},
		},
		Span: lexer.Span{Start: minusToken.Start(), End: numberToken.End()},
	}

	// Check for complex number patterns (signed_number '+' NUMBER or signed_number '-' NUMBER)
	if p.check(lexer.Plus) || p.check(lexer.Minus) {
		opToken := p.advance()
		if p.check(lexer.Number) {
			imagToken, err := p.consume(lexer.Number, "expected imaginary number")
			if err != nil {
				return nil, err
			}

			// Create complex number expression
			imagLiteral := &ast.Literal{
				Value: imagToken.Literal,
				Token: imagToken,
				Span:  lexer.Span{Start: imagToken.Start(), End: imagToken.End()},
			}

			var rightOperand ast.Expr = imagLiteral
			if opToken.Type == lexer.Minus {
				rightOperand = &ast.Unary{
					Operator: opToken,
					Right:    imagLiteral,
					Span:     lexer.Span{Start: opToken.Start(), End: imagToken.End()},
				}
			}

			expr = &ast.Binary{
				Left:     expr,
				Operator: opToken,
				Right:    rightOperand,
				Span:     lexer.Span{Start: minusToken.Start(), End: imagToken.End()},
			}
		}
	}

	return &ast.LiteralPattern{
		Value: expr,
		Span:  expr.GetSpan(),
	}, nil
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

		// Check for complex number patterns (NUMBER '+' NUMBER or NUMBER '-' NUMBER)
		if p.check(lexer.Plus) || p.check(lexer.Minus) {
			opToken := p.advance()
			if p.check(lexer.Number) {
				imagToken, err := p.consume(lexer.Number, "expected imaginary number")
				if err != nil {
					return nil, err
				}

				imagLiteral := &ast.Literal{
					Value: imagToken.Literal,
					Token: imagToken,
					Span:  lexer.Span{Start: imagToken.Start(), End: imagToken.End()},
				}

				var rightOperand ast.Expr = imagLiteral
				if opToken.Type == lexer.Minus {
					rightOperand = &ast.Unary{
						Operator: opToken,
						Right:    imagLiteral,
						Span:     lexer.Span{Start: opToken.Start(), End: imagToken.End()},
					}
				}

				expr = &ast.Binary{
					Left:     expr,
					Operator: opToken,
					Right:    rightOperand,
					Span:     lexer.Span{Start: token.Start(), End: imagToken.End()},
				}
			}
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

// identifierPattern determines if an identifier is a capture pattern, value pattern, or class pattern
func (p *Parser) identifierPattern() (ast.Pattern, error) {
	nameToken, err := p.consume(lexer.Identifier, "expected identifier")
	if err != nil {
		return nil, err
	}

	name := &ast.Name{
		Token: nameToken,
		Span:  lexer.Span{Start: nameToken.Start(), End: nameToken.End()},
	}

	// Check if this is a dotted name (value pattern) or class pattern
	if p.check(lexer.Dot) {
		// Parse as attribute access (value pattern or class name)
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

		// Check if this is a class pattern
		if p.check(lexer.LeftParen) {
			return p.classPattern(expr)
		}

		// Value pattern
		return &ast.ValuePattern{
			Value: expr,
			Span:  expr.GetSpan(),
		}, nil
	}

	// Check if this is a class pattern
	if p.check(lexer.LeftParen) {
		return p.classPattern(name)
	}

	// Simple identifier - capture pattern
	return &ast.CapturePattern{
		Name: name,
		Span: name.GetSpan(),
	}, nil
}

// classPattern parses class patterns:
// class_pattern: name_or_attr '(' [pattern_arguments ','?] ')'
func (p *Parser) classPattern(className ast.Expr) (ast.Pattern, error) {
	_, err := p.consume(lexer.LeftParen, "expected '('")
	if err != nil {
		return nil, err
	}

	var positionalPatterns []ast.Pattern
	var keywordPatterns []ast.KwdPatternPair

	// Empty class pattern
	if p.check(lexer.RightParen) {
		rightParen, err := p.consume(lexer.RightParen, "expected ')'")
		if err != nil {
			return nil, err
		}
		return &ast.ClassPattern{
			Class:       className,
			Patterns:    positionalPatterns,
			KwdPatterns: keywordPatterns,
			Span:        lexer.Span{Start: className.GetSpan().Start, End: rightParen.End()},
		}, nil
	}

	// Parse pattern arguments
	for {
		// Check for keyword pattern (NAME '=' pattern)
		if p.check(lexer.Identifier) && p.peekN(1).Type == lexer.Equal {
			nameToken, err := p.consume(lexer.Identifier, "expected identifier")
			if err != nil {
				return nil, err
			}
			_, err = p.consume(lexer.Equal, "expected '='")
			if err != nil {
				return nil, err
			}
			pattern, err := p.pattern()
			if err != nil {
				return nil, err
			}

			name := &ast.Name{
				Token: nameToken,
				Span:  lexer.Span{Start: nameToken.Start(), End: nameToken.End()},
			}

			keywordPatterns = append(keywordPatterns, ast.KwdPatternPair{
				Name:    name,
				Pattern: pattern,
				Span:    lexer.Span{Start: nameToken.Start(), End: pattern.GetSpan().End},
			})
		} else {
			// Positional pattern
			if len(keywordPatterns) > 0 {
				return nil, p.error(p.peek(), "positional patterns must come before keyword patterns")
			}
			pattern, err := p.pattern()
			if err != nil {
				return nil, err
			}
			positionalPatterns = append(positionalPatterns, pattern)
		}

		if !p.match(lexer.Comma) {
			break
		}

		// Allow trailing comma
		if p.check(lexer.RightParen) {
			break
		}
	}

	rightParen, err := p.consume(lexer.RightParen, "expected ')'")
	if err != nil {
		return nil, err
	}

	return &ast.ClassPattern{
		Class:       className,
		Patterns:    positionalPatterns,
		KwdPatterns: keywordPatterns,
		Span:        lexer.Span{Start: className.GetSpan().Start, End: rightParen.End()},
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
		pattern, err := p.maybeStarPattern()
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
		// Check if it's a star pattern or has a comma - then it's a sequence
		if _, isStar := patterns[0].(*ast.StarPattern); isStar {
			return &ast.SequencePattern{
				Patterns: patterns,
				IsTuple:  isParens,
				Span:     lexer.Span{Start: leftToken.Start(), End: rightToken.End()},
			}, nil
		}
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

// sequencePattern parses sequence patterns for brackets
func (p *Parser) sequencePattern(isParens bool) (ast.Pattern, error) {
	return p.groupOrSequencePattern(isParens)
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
			pattern, err := p.capturePattern()
			if err != nil {
				return nil, err
			}
			doubleStar = pattern
			hasRest = true
		} else {
			// Parse key: pattern
			// Key must be a literal or value pattern
			var key ast.Expr
			if p.check(lexer.Number) || p.check(lexer.String) || p.check(lexer.None) || p.check(lexer.True) || p.check(lexer.False) {
				// Literal key
				literalPattern, err := p.literalPattern()
				if err != nil {
					return nil, err
				}
				if litPat, ok := literalPattern.(*ast.LiteralPattern); ok {
					key = litPat.Value
				} else {
					return nil, p.error(p.peek(), "expected literal key")
				}
			} else if p.check(lexer.Identifier) {
				// Value pattern key (dotted name)
				nameToken, err := p.consume(lexer.Identifier, "expected identifier")
				if err != nil {
					return nil, err
				}
				name := &ast.Name{
					Token: nameToken,
					Span:  lexer.Span{Start: nameToken.Start(), End: nameToken.End()},
				}
				key = name

				// Handle dotted names
				for p.match(lexer.Dot) {
					attrName, err := p.consume(lexer.Identifier, "expected identifier after '.'")
					if err != nil {
						return nil, err
					}
					key = &ast.Attribute{
						Object: key,
						Name:   attrName,
						Span:   lexer.Span{Start: key.GetSpan().Start, End: attrName.End()},
					}
				}
			} else {
				return nil, p.error(p.peek(), "expected literal or identifier for mapping key")
			}

			_, err = p.consume(lexer.Colon, "expected ':' after mapping pattern key")
			if err != nil {
				return nil, err
			}

			pattern, err := p.pattern()
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

// capturePattern parses a capture pattern (identifier that's not '_')
func (p *Parser) capturePattern() (ast.Pattern, error) {
	if !p.check(lexer.Identifier) {
		return nil, p.error(p.peek(), "expected identifier")
	}

	nameToken, err := p.consume(lexer.Identifier, "expected identifier")
	if err != nil {
		return nil, err
	}

	if nameToken.Lexeme == "_" {
		return nil, p.error(nameToken, "cannot use '_' as capture pattern")
	}

	name := &ast.Name{
		Token: nameToken,
		Span:  lexer.Span{Start: nameToken.Start(), End: nameToken.End()},
	}

	return &ast.CapturePattern{
		Name: name,
		Span: name.GetSpan(),
	}, nil
}
