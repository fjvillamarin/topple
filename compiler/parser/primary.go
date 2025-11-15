package parser

import (
	"topple/compiler/ast"
	"topple/compiler/lexer"
)

func (p *Parser) await() (ast.Expr, error) {
	// Check if the current token is an await keyword
	if p.match(lexer.Await) {
		awaitToken := p.previous()
		expr, err := p.primary()
		if err != nil {
			return nil, err
		}
		return &ast.AwaitExpr{
			Expr: expr,

			Span: lexer.Span{Start: awaitToken.Start(), End: expr.GetSpan().End},
		}, nil
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
			expr = &ast.Attribute{
				Object: expr,
				Name:   name,

				Span: lexer.Span{Start: expr.GetSpan().Start, End: name.End()},
			}
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
			expr = &ast.Subscript{
				Object:  expr,
				Indices: indices,

				Span: lexer.Span{Start: expr.GetSpan().Start, End: right.End()},
			}
		} else {
			// No more postfix operations
			break
		}
	}

	return expr, nil
}

func (p *Parser) finishCall(callee ast.Expr) (ast.Expr, error) {
	// Use the proper arguments parser that handles *args and **kwargs
	args, err := p.arguments()
	if err != nil {
		return nil, err
	}

	right, err := p.consume(lexer.RightParen, "expected ')' after arguments")
	if err != nil {
		return nil, err
	}

	// Convert []ast.Argument to []ast.Expr for the Call node

	return &ast.Call{
		Callee:    callee,
		Arguments: args,

		Span: lexer.Span{Start: callee.GetSpan().Start, End: right.End()},
	}, nil
}

// atom parses an atom.
func (p *Parser) atom() (ast.Expr, error) {
	if p.match(lexer.False) {
		return &ast.Literal{
			Value: false,
			Type:  ast.LiteralTypeBool,

			Span: lexer.Span{Start: p.previous().Start(), End: p.previous().End()},
		}, nil
	}

	if p.match(lexer.True) {
		return &ast.Literal{
			Value: true,
			Type:  ast.LiteralTypeBool,

			Span: lexer.Span{Start: p.previous().Start(), End: p.previous().End()},
		}, nil
	}

	if p.match(lexer.None) {
		return &ast.Literal{
			Value: nil,
			Type:  ast.LiteralTypeNone,

			Span: lexer.Span{Start: p.previous().Start(), End: p.previous().End()},
		}, nil
	}

	if p.match(lexer.Number, lexer.String) {
		return &ast.Literal{
			Token: p.previous(),
			Value: p.previous().Literal,
			Type:  ast.LiteralTypeString,

			Span: lexer.Span{Start: p.previous().Start(), End: p.previous().End()},
		}, nil
	}

	if p.check(lexer.FStringStart) {
		// Backtrack since we consumed the FStringStart token
		return p.fstring()
	}

	if p.match(lexer.Ellipsis) {
		return &ast.Literal{
			Value: nil,
			Type:  ast.LiteralTypeNone,

			Span: lexer.Span{Start: p.previous().Start(), End: p.previous().End()},
		}, nil
	}

	if p.match(lexer.Identifier) {
		return &ast.Name{
			Token: p.previous(),

			Span: lexer.Span{Start: p.previous().Start(), End: p.previous().End()},
		}, nil
	}

	if p.check(lexer.LeftParen) {
		// This could be either a tuple or a group (parenthesized expression)
		return p.tuple()
	}

	if p.check(lexer.LeftBracket) {
		// Could be either a list literal or list comprehension
		return p.listOrListComp()
	}

	if p.check(lexer.LeftBrace) {
		// Could be either a dictionary or a set - we need to look ahead
		return p.dictOrSet()
	}

	if p.check(lexer.Yield) {
		// Yield expression
		return p.yieldExpression()
	}

	return nil, p.error(p.peek(), "unexpected token")
}
