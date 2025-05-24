package parser

import (
	"biscuit/compiler/ast"
	"biscuit/compiler/lexer"
)

// dictOrSet determines whether to parse a dictionary or set by looking ahead
func (p *Parser) dictOrSet() (ast.Expr, error) {
	// Save current position
	startPos := p.Current

	// Consume the opening '{'
	_, err := p.consume(lexer.LeftBrace, "expected '{'")
	if err != nil {
		return nil, err
	}

	// Check for empty braces - this is an empty dictionary
	if p.check(lexer.RightBrace) {
		// Restore position and parse as dictionary
		p.Current = startPos
		return p.dict()
	}

	// Check for ** (dictionary unpacking)
	if p.check(lexer.StarStar) {
		// Restore position and parse as dictionary
		p.Current = startPos
		return p.dict()
	}

	// We need to look ahead to see if there's a colon after an expression
	// Try to parse an expression and see if it's followed by a colon
	_, err = p.expression()
	if err != nil {
		// If we can't parse an expression, restore and try set
		p.Current = startPos
		return p.set()
	}

	// Check if there's a colon after the expression
	if p.check(lexer.Colon) {
		// It's a dictionary (key: value)
		p.Current = startPos
		return p.dict()
	}

	// Otherwise, it's a set
	p.Current = startPos
	return p.set()
}

// dict parses a dictionary literal according to the grammar:
// dict: '{' [double_starred_kvpairs] '}'
func (p *Parser) dict() (ast.Expr, error) {
	// Consume the opening '{'
	leftBrace, err := p.consume(lexer.LeftBrace, "expected '{'")
	if err != nil {
		return nil, err
	}

	var pairs []ast.DictPair

	// Check for empty dictionary
	if p.check(lexer.RightBrace) {
		// Empty dictionary
		rightBrace, err := p.consume(lexer.RightBrace, "expected '}'")
		if err != nil {
			return nil, err
		}

		return &ast.DictExpr{
			Pairs: pairs,
			Span:  lexer.Span{Start: leftBrace.Start(), End: rightBrace.End()},
		}, nil
	}

	// Parse double_starred_kvpairs
	pairs, err = p.doubleStarredKvpairs()
	if err != nil {
		return nil, err
	}

	// Consume the closing '}'
	rightBrace, err := p.consume(lexer.RightBrace, "expected '}'")
	if err != nil {
		return nil, err
	}

	return &ast.DictExpr{
		Pairs: pairs,
		Span:  lexer.Span{Start: leftBrace.Start(), End: rightBrace.End()},
	}, nil
}

// doubleStarredKvpairs parses a comma-separated list of dictionary pairs:
// double_starred_kvpairs: ','.double_starred_kvpair+ [',']
func (p *Parser) doubleStarredKvpairs() ([]ast.DictPair, error) {
	var pairs []ast.DictPair

	// Parse the first pair
	pair, err := p.doubleStarredKvpair()
	if err != nil {
		return nil, err
	}
	pairs = append(pairs, pair)

	// Parse additional pairs separated by commas
	for p.match(lexer.Comma) {
		// Allow trailing comma
		if p.check(lexer.RightBrace) {
			break
		}

		pair, err := p.doubleStarredKvpair()
		if err != nil {
			return nil, err
		}
		pairs = append(pairs, pair)
	}

	return pairs, nil
}

// doubleStarredKvpair parses either a starred expression or a key-value pair:
// double_starred_kvpair: '**' bitwise_or | kvpair
func (p *Parser) doubleStarredKvpair() (ast.DictPair, error) {
	// Check for double-starred expression (**expression)
	if p.match(lexer.StarStar) {
		starToken := p.previous()

		// Parse the expression after **
		expr, err := p.bitwiseOr()
		if err != nil {
			return nil, err
		}

		return &ast.DoubleStarredPair{
			Expr: expr,
			Span: lexer.Span{Start: starToken.Start(), End: expr.GetSpan().End},
		}, nil
	}

	// Otherwise, parse as key-value pair
	return p.kvpair()
}

// kvpair parses a key-value pair:
// kvpair: expression ':' expression
func (p *Parser) kvpair() (ast.DictPair, error) {
	// Parse the key expression
	key, err := p.expression()
	if err != nil {
		return nil, err
	}

	// Consume the colon
	_, err = p.consume(lexer.Colon, "expected ':' after dictionary key")
	if err != nil {
		return nil, err
	}

	// Parse the value expression
	value, err := p.expression()
	if err != nil {
		return nil, err
	}

	return &ast.KeyValuePair{
		Key:   key,
		Value: value,
		Span:  lexer.Span{Start: key.GetSpan().Start, End: value.GetSpan().End},
	}, nil
}
