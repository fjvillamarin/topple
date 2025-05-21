package parser

import (
	"biscuit/compiler/ast"
	"biscuit/compiler/lexer"
)

func (p *Parser) tuple() (ast.Expr, error) {
	// Expect opening parenthesis
	leftParen, err := p.consume(lexer.LeftParen, "expected '('")
	if err != nil {
		return nil, err
	}

	// Empty tuple
	if p.check(lexer.RightParen) {
		rightParen, err := p.consume(lexer.RightParen, "expected ')'")
		if err != nil {
			return nil, err
		}
		return &ast.TupleExpr{
			Elements: []ast.Expr{},

			Span: lexer.Span{Start: leftParen.Start(), End: rightParen.End()},
		}, nil
	}

	// Parse the first expression (could be a star expression or named expression)
	expr, err := p.starNamedExpression()
	if err != nil {
		return nil, err
	}

	// If the next token is 'yield', it's a group
	if p.checkNext(lexer.Yield) {
		expr, err := p.yieldExpression()
		if err != nil {
			return nil, err
		}

		rightParen, err := p.consume(lexer.RightParen, "expected ')'")
		if err != nil {
			return nil, err
		}
		return &ast.GroupExpr{
			Expression: expr,

			Span: lexer.Span{Start: leftParen.Start(), End: rightParen.End()},
		}, nil
	}

	// If there's a comma, it's a tuple
	if p.match(lexer.Comma) {
		elements := []ast.Expr{expr}

		// Parse additional elements if present
		if !p.check(lexer.RightParen) {
			for {
				nextExpr, err := p.starNamedExpression()
				if err != nil {
					return nil, err
				}
				elements = append(elements, nextExpr)

				if !p.match(lexer.Comma) {
					break
				}

				// Allow trailing comma
				if p.check(lexer.RightParen) {
					break
				}
			}
		}

		rightParen, err := p.consume(lexer.RightParen, "expected ')'")
		if err != nil {
			return nil, err
		}
		return &ast.TupleExpr{
			Elements: elements,

			Span: lexer.Span{Start: leftParen.Start(), End: rightParen.End()},
		}, nil
	} else {
		// No comma, so it's a group
		// Groups can only contain named expressions, not star expressions
		_, isStarExpr := expr.(*ast.StarExpr)
		if isStarExpr {
			return nil, p.error(p.previous(), "starred expression cannot appear in a group")
		}

		rightParen, err := p.consume(lexer.RightParen, "expected ')'")
		if err != nil {
			return nil, err
		}
		return &ast.GroupExpr{
			Expression: expr,

			Span: lexer.Span{Start: leftParen.Start(), End: rightParen.End()},
		}, nil
	}
}
