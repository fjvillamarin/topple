package parser

import (
	"topple/compiler/ast"
	"topple/compiler/lexer"
)

// sum parses a sum expression.
func (p *Parser) sum() (ast.Expr, error) {
	left, err := p.term()
	if err != nil {
		return nil, err
	}

	for p.match(lexer.Plus, lexer.Minus) {
		operator := p.previous()
		right, err := p.term()
		if err != nil {
			return nil, err
		}
		left = &ast.Binary{
			Left:     left,
			Operator: operator,
			Right:    right,

			Span: lexer.Span{
				Start: left.GetSpan().Start,
				End:   right.GetSpan().End,
			},
		}
	}

	return left, nil
}

// term parses a term expression.
func (p *Parser) term() (ast.Expr, error) {
	left, err := p.factor()
	if err != nil {
		return nil, err
	}

	for p.match(lexer.Star, lexer.Slash, lexer.SlashSlash, lexer.Percent, lexer.At) {
		operator := p.previous()
		right, err := p.factor()
		if err != nil {
			return nil, err
		}
		left = &ast.Binary{
			Left:     left,
			Operator: operator,
			Right:    right,

			Span: lexer.Span{
				Start: left.GetSpan().Start,
				End:   right.GetSpan().End,
			},
		}
	}
	return left, nil
}

// factor parses a factor expression.
func (p *Parser) factor() (ast.Expr, error) {
	if p.match(lexer.Plus, lexer.Minus, lexer.Tilde) {
		operator := p.previous()
		right, err := p.factor()
		if err != nil {
			return nil, err
		}
		return &ast.Unary{
			Operator: operator,
			Right:    right,

			Span: lexer.Span{
				Start: operator.Start(),
				End:   right.GetSpan().End,
			},
		}, nil
	}
	return p.power()
}

// power parses a power expression.
func (p *Parser) power() (ast.Expr, error) {
	left, err := p.await()
	if err != nil {
		return nil, err
	}

	if p.match(lexer.StarStar) {
		operator := p.previous()
		right, err := p.factor()
		if err != nil {
			return nil, err
		}
		return &ast.Binary{
			Left:     left,
			Operator: operator,
			Right:    right,

			Span: lexer.Span{
				Start: left.GetSpan().Start,
				End:   right.GetSpan().End,
			},
		}, nil
	}

	return left, nil
}
