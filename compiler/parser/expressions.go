package parser

import (
	"github.com/fjvillamarin/topple/compiler/ast"
	"github.com/fjvillamarin/topple/compiler/lexer"
)

// expression parses an expression.
func (p *Parser) expression() (ast.Expr, error) {
	// Check for lambda expressions first
	if p.check(lexer.Lambda) {
		return p.lambdef()
	}

	expr, err := p.disjunction()
	if err != nil {
		return nil, err
	}

	if p.match(lexer.If) {
		condition, err := p.disjunction()
		if err != nil {
			return nil, err
		}
		_, err = p.consume(lexer.Else, "expected 'else' after condition")
		if err != nil {
			return nil, err
		}
		elseExpr, err := p.expression()
		if err != nil {
			return nil, err
		}
		return &ast.TernaryExpr{
			Condition: condition,
			TrueExpr:  expr,
			FalseExpr: elseExpr,

			Span: lexer.Span{Start: expr.GetSpan().Start, End: elseExpr.GetSpan().End},
		}, nil
	}

	return expr, nil
}

// disjunction parses a disjunction expression.
func (p *Parser) disjunction() (ast.Expr, error) {
	expr, err := p.conjunction()
	if err != nil {
		return nil, err
	}

	for p.match(lexer.Or) {
		operator := p.previous()
		right, err := p.conjunction()
		if err != nil {
			return nil, err
		}
		expr = &ast.Binary{
			Left:     expr,
			Operator: operator,
			Right:    right,

			Span: lexer.Span{Start: expr.GetSpan().Start, End: right.GetSpan().End},
		}
	}

	return expr, nil
}

// conjunction parses a conjunction expression.
func (p *Parser) conjunction() (ast.Expr, error) {
	expr, err := p.inversion()
	if err != nil {
		return nil, err
	}

	for p.match(lexer.And) {
		operator := p.previous()
		right, err := p.inversion()
		if err != nil {
			return nil, err
		}
		expr = &ast.Binary{
			Left:     expr,
			Operator: operator,
			Right:    right,

			Span: lexer.Span{Start: expr.GetSpan().Start, End: right.GetSpan().End},
		}
	}

	return expr, nil
}

// inversion parses an inversion expression.
func (p *Parser) inversion() (ast.Expr, error) {
	if p.match(lexer.Not) {
		operator := p.previous()
		right, err := p.inversion()
		if err != nil {
			return nil, err
		}
		return &ast.Unary{
			Operator: operator,
			Right:    right,

			Span: lexer.Span{Start: operator.Start(), End: right.GetSpan().End},
		}, nil
	}

	return p.comparison()
}

// parseNameList is a helper function to parse a comma-separated list of identifiers
func (p *Parser) parseNameList() ([]*ast.Name, error) {
	names := []*ast.Name{}

	// Parse first identifier
	nameToken, err := p.consume(lexer.Identifier, "expected identifier")
	if err != nil {
		return nil, err
	}

	// Create a Name node from the token
	name := &ast.Name{
		Token: nameToken,

		Span: lexer.Span{Start: nameToken.Start(), End: nameToken.End()},
	}
	names = append(names, name)

	// Parse additional identifiers separated by commas
	for p.match(lexer.Comma) {
		nameToken, err = p.consume(lexer.Identifier, "expected identifier after ','")
		if err != nil {
			return nil, err
		}

		// Create a Name node from the token
		name = &ast.Name{
			Token: nameToken,

			Span: lexer.Span{Start: nameToken.Start(), End: nameToken.End()},
		}
		names = append(names, name)
	}

	return names, nil
}

func (p *Parser) yieldExpression() (ast.Expr, error) {
	// Expect 'yield' keyword
	yieldToken, err := p.consume(lexer.Yield, "expected 'yield'")
	if err != nil {
		return nil, err
	}

	// Check for 'yield from' form
	isFrom := false
	if p.match(lexer.From) {
		isFrom = true
		// Parse the expression after 'yield from'
		expr, err := p.expression()
		if err != nil {
			return nil, err
		}
		return &ast.YieldExpr{
			IsFrom: isFrom,
			Value:  expr,

			Span: lexer.Span{Start: yieldToken.Start(), End: expr.GetSpan().End},
		}, nil
	}

	// Check if there's an expression after 'yield'
	if p.isAtEnd() || p.check(lexer.Newline) || p.check(lexer.Semicolon) || p.check(lexer.RightParen) || p.check(lexer.RightBraceF) || p.check(lexer.Comma) {
		// No expression, yield on its own
		return &ast.YieldExpr{
			IsFrom: isFrom,
			Value:  nil,

			Span: lexer.Span{Start: yieldToken.Start(), End: yieldToken.End()},
		}, nil
	}

	// Parse star expressions after 'yield'
	expr, err := p.starExpressions()
	if err != nil {
		return nil, err
	}
	return &ast.YieldExpr{
		IsFrom: isFrom,
		Value:  expr,

		Span: lexer.Span{Start: yieldToken.Start(), End: expr.GetSpan().End},
	}, nil
}

func (p *Parser) starExpressions() (ast.Expr, error) {
	// Parse the first star expression
	expr, err := p.starExpression()
	if err != nil {
		return nil, err
	}

	// If there's no comma, return the expression as is
	if !p.match(lexer.Comma) {
		return expr, nil
	}

	// We have a comma, so this is a tuple of expressions
	elements := []ast.Expr{expr}

	// Allow trailing comma with no following expression
	if !p.check(lexer.Newline) && !p.check(lexer.RightParen) && !p.check(lexer.RightBracket) && !p.check(lexer.RightBrace) && !p.check(lexer.Semicolon) && !p.isAtEnd() {
		// Parse subsequent expressions
		for {
			expr, err := p.starExpression()
			if err != nil {
				return nil, err
			}
			elements = append(elements, expr)

			if !p.match(lexer.Comma) {
				break
			}

			// Allow trailing comma
			if p.check(lexer.Newline) || p.check(lexer.RightParen) || p.check(lexer.RightBracket) || p.check(lexer.RightBrace) || p.check(lexer.Semicolon) || p.isAtEnd() {
				break
			}
		}
	}

	// Create a tuple with the collected expressions
	return &ast.TupleExpr{
		Elements: elements,

		Span: lexer.Span{Start: elements[0].GetSpan().Start, End: elements[len(elements)-1].GetSpan().End},
	}, nil
}

func (p *Parser) starExpression() (ast.Expr, error) {
	if p.match(lexer.Star) {
		// This is a starred expression like *args
		star := p.previous()
		expr, err := p.bitwiseOr() // According to the grammar, star expressions use bitwise_or
		if err != nil {
			return nil, err
		}

		return &ast.StarExpr{
			Expr: expr,

			Span: lexer.Span{Start: star.Start(), End: expr.GetSpan().End},
		}, nil
	}

	// Not a star expression, parse as a regular expression
	return p.expression()
}

func (p *Parser) namedExpression() (ast.Expr, error) {
	// Handle assignment expressions first (the walrus operator)
	if p.check(lexer.Identifier) && p.checkNext(lexer.Walrus) {
		name, err := p.consume(lexer.Identifier, "expected identifier") // Consume the identifier
		if err != nil {
			return nil, err
		}
		_, err = p.consume(lexer.Walrus, "expected ':=' after identifier") // Consume the ':=' operator
		if err != nil {
			return nil, err
		}

		value, err := p.expression()
		if err != nil {
			return nil, err
		}

		return &ast.AssignExpr{
			Left: &ast.Name{
				Token: name,

				Span: lexer.Span{Start: name.Start(), End: name.End()},
			},
			Right: value,

			Span: lexer.Span{Start: name.Start(), End: value.GetSpan().End},
		}, nil
	}

	// If not an assignment expression, parse a regular expression
	return p.expression()
}

func (p *Parser) starNamedExpression() (ast.Expr, error) {
	if p.match(lexer.Star) {
		// This is a starred expression like *args
		star := p.previous()
		expr, err := p.bitwiseOr() // According to the grammar, star expressions use bitwise_or
		if err != nil {
			return nil, err
		}

		return &ast.StarExpr{
			Expr: expr,

			Span: lexer.Span{Start: star.Start(), End: expr.GetSpan().End},
		}, nil
	}

	// Not a star expression, parse as a regular expression
	return p.namedExpression()
}
