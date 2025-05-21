package parser

import (
	"biscuit/compiler/ast"
	"biscuit/compiler/lexer"
)

// sum parses a sum expression.
func (p *Parser) sum() (ast.Expr, error) {
	expr, err := p.term()
	if err != nil {
		return nil, err
	}

	for p.match(lexer.Plus, lexer.Minus) {
		operator := p.previous()
		right, err := p.term()
		if err != nil {
			return nil, err
		}
		expr = ast.NewBinary(expr, operator, right, lexer.Span{Start: expr.Span().Start, End: right.Span().End})
	}

	return expr, nil
}

// term parses a term expression.
func (p *Parser) term() (ast.Expr, error) {
	expr, err := p.factor()
	if err != nil {
		return nil, err
	}

	for p.match(lexer.Star, lexer.Slash, lexer.SlashSlash, lexer.Percent, lexer.At) {
		operator := p.previous()
		right, err := p.factor()
		if err != nil {
			return nil, err
		}
		expr = ast.NewBinary(expr, operator, right, lexer.Span{Start: expr.Span().Start, End: right.Span().End})
	}

	return expr, nil
}

// factor parses a factor expression.
func (p *Parser) factor() (ast.Expr, error) {
	if p.match(lexer.Plus, lexer.Minus, lexer.Tilde) {
		operator := p.previous()
		right, err := p.factor()
		if err != nil {
			return nil, err
		}
		return ast.NewUnary(operator, right, lexer.Span{Start: operator.Start(), End: right.Span().End}), nil
	}

	return p.power()
}

// power parses a power expression.
func (p *Parser) power() (ast.Expr, error) {
	expr, err := p.await()
	if err != nil {
		return nil, err
	}

	if p.match(lexer.StarStar) {
		operator := p.previous()
		right, err := p.factor()
		if err != nil {
			return nil, err
		}
		return ast.NewBinary(expr, operator, right, lexer.Span{Start: expr.Span().Start, End: right.Span().End}), nil
	}

	return expr, nil
}
