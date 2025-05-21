package parser

import (
	"biscuit/compiler/ast"
	"biscuit/compiler/lexer"
)

func (p *Parser) await() (ast.Expr, error) {
	// Check if the current token is an await keyword
	if p.match(lexer.Await) {
		awaitToken := p.previous()
		expr, err := p.primary()
		if err != nil {
			return nil, err
		}
		return ast.NewAwaitExpr(expr, lexer.Span{Start: awaitToken.Start(), End: expr.Span().End}), nil
	}

	return p.primary()
}

// primary parses a primary expression.
func (p *Parser) primary() (ast.Expr, error) {
	// Parse the initial atom
	expr, err := p.atom()
	if err != nil {
		return nil, err
	}

	// Keep parsing postfix operations while they exist
	for {
		if p.match(lexer.Dot) {
			// Handle attribute access: expr.NAME
			name, err := p.consume(lexer.Identifier, "expected identifier after '.'")
			if err != nil {
				return nil, err
			}
			expr = ast.NewAttribute(expr, name, lexer.Span{Start: expr.Span().Start, End: name.End()})
		} else if p.match(lexer.LeftParen) {
			// Handle function call: expr(args)
			expr, err = p.finishCall(expr)
			if err != nil {
				return nil, err
			}
		} else if p.match(lexer.LeftBracket) {
			// Handle subscript access: expr[index] or expr[slice]
			indices, err := p.slices()
			if err != nil {
				return nil, err
			}

			right, err := p.consume(lexer.RightBracket, "expected ']' after index")
			if err != nil {
				return nil, err
			}
			expr = ast.NewSubscript(expr, indices, lexer.Span{Start: expr.Span().Start, End: right.End()})
		} else {
			// No more postfix operations
			break
		}
	}

	return expr, nil
}

func (p *Parser) finishCall(callee ast.Expr) (ast.Expr, error) {
	args := []ast.Expr{}
	if !p.check(lexer.RightParen) {
		// Parse first argument
		arg, err := p.expression()
		if err != nil {
			return nil, err
		}
		args = append(args, arg)

		// Parse additional arguments
		for p.match(lexer.Comma) {
			// If there's a right parenthesis after the comma,
			// it's a trailing comma, so we're done parsing arguments
			if p.check(lexer.RightParen) {
				break
			}

			arg, err := p.expression()
			if err != nil {
				return nil, err
			}
			args = append(args, arg)
		}
	}

	right, err := p.consume(lexer.RightParen, "expected ')' after arguments")
	if err != nil {
		return nil, err
	}

	return ast.NewCall(callee, args, lexer.Span{Start: callee.Span().Start, End: right.End()}), nil
}

// atom parses an atom.
func (p *Parser) atom() (ast.Expr, error) {
	if p.match(lexer.False) {
		return ast.NewLiteral(p.previous(), false, lexer.Span{Start: p.previous().Start(), End: p.previous().End()}), nil
	}

	if p.match(lexer.True) {
		return ast.NewLiteral(p.previous(), true, lexer.Span{Start: p.previous().Start(), End: p.previous().End()}), nil
	}

	if p.match(lexer.None) {
		return ast.NewLiteral(p.previous(), nil, lexer.Span{Start: p.previous().Start(), End: p.previous().End()}), nil
	}

	if p.match(lexer.Number, lexer.String) {
		return ast.NewLiteral(p.previous(), p.previous().Literal, lexer.Span{Start: p.previous().Start(), End: p.previous().End()}), nil
	}

	if p.match(lexer.Ellipsis) {
		return ast.NewLiteral(p.previous(), nil, lexer.Span{Start: p.previous().Start(), End: p.previous().End()}), nil
	}

	if p.match(lexer.Identifier) {
		return ast.NewName(p.previous(), lexer.Span{Start: p.previous().Start(), End: p.previous().End()}), nil
	}

	if p.check(lexer.LeftParen) {
		// This could be either a tuple or a group (parenthesized expression)
		return p.tuple()
	}

	if p.check(lexer.LeftBracket) {
		// List literal
		return p.list()
	}

	if p.check(lexer.LeftBrace) {
		// Set literal
		return p.set()
	}

	if p.check(lexer.Yield) {
		// Yield expression
		return p.yieldExpression()
	}

	return nil, p.error(p.peek(), "unexpected token")
}
