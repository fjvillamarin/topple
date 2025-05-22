package parser

import (
	"biscuit/compiler/ast"
	"biscuit/compiler/lexer"
)

// withStatement parses a with statement.
func (p *Parser) withStatement() (ast.Stmt, error) {
	// Check for async
	isAsync := false
	var startToken lexer.Token

	if p.match(lexer.Async) {
		isAsync = true
		startToken = p.previous()
		_, err := p.consume(lexer.With, "expected 'with' after 'async'")
		if err != nil {
			return nil, err
		}
	} else {
		// Consume the 'with' token
		var err error
		startToken, err = p.consume(lexer.With, "expected 'with'")
		if err != nil {
			return nil, err
		}
	}

	// Check if we have parenthesized with items
	hasParens := p.match(lexer.LeftParen)

	// Parse the with items
	items, err := p.withItems(hasParens)
	if err != nil {
		return nil, err
	}

	// If we had parentheses, expect closing paren
	if hasParens {
		_, err = p.consume(lexer.RightParen, "expected ')' after with items")
		if err != nil {
			return nil, err
		}
	}

	// Expect colon
	_, err = p.consume(lexer.Colon, "expected ':' after with items")
	if err != nil {
		return nil, err
	}

	// Parse the body
	body, err := p.block()
	if err != nil {
		return nil, err
	}

	// Determine the end position
	var endPos lexer.Position
	if len(body) > 0 {
		endPos = body[len(body)-1].GetSpan().End
	} else {
		endPos = p.previous().End()
	}

	return &ast.With{
		Items:   items,
		IsAsync: isAsync,
		Body:    body,
		Span:    lexer.Span{Start: startToken.Start(), End: endPos},
	}, nil
}

// withItems parses a comma-separated list of with items.
// If inParens is true, it allows a trailing comma.
func (p *Parser) withItems(inParens bool) ([]ast.WithItem, error) {
	items := []ast.WithItem{}

	// Parse the first with item
	item, err := p.withItem()
	if err != nil {
		return nil, err
	}
	items = append(items, item)

	// Parse any additional with items separated by commas
	for p.match(lexer.Comma) {
		// If we're in parentheses and find a trailing comma, we're done
		if inParens && p.check(lexer.RightParen) {
			break
		}

		item, err = p.withItem()
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	return items, nil
}

// withItem parses a single with item: expression ['as' star_target]
func (p *Parser) withItem() (ast.WithItem, error) {
	// Parse the expression
	expr, err := p.expression()
	if err != nil {
		return ast.WithItem{}, err
	}

	// Check for 'as' target
	var asExpr ast.Expr
	if p.match(lexer.As) {
		asExpr, err = p.starTarget()
		if err != nil {
			return ast.WithItem{}, err
		}

		// Check that it's followed by one of ',', ')', or ':'
		if !p.check(lexer.Comma) && !p.check(lexer.RightParen) && !p.check(lexer.Colon) {
			return ast.WithItem{}, p.error(p.peek(), "expected ',' or ':' after 'as' target")
		}
	}

	return ast.WithItem{
		Expr: expr,
		As:   asExpr,
	}, nil
}
