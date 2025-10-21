package parser

import (
	"topple/compiler/ast"
	"topple/compiler/lexer"
)

// bitwise_or parses a bitwise OR expression.
func (p *Parser) bitwiseOr() (ast.Expr, error) {
	expr, err := p.bitwiseXor()
	if err != nil {
		return nil, err
	}

	for p.match(lexer.Pipe) {
		operator := p.previous()
		right, err := p.bitwiseXor()
		if err != nil {
			return nil, err
		}
		expr = &ast.Binary{
			Left:     expr,
			Operator: operator,
			Right:    right,

			Span: lexer.Span{
				Start: expr.GetSpan().Start,
				End:   right.GetSpan().End,
			},
		}
	}

	return expr, nil
}

// bitwiseXor parses a bitwise XOR expression.
func (p *Parser) bitwiseXor() (ast.Expr, error) {
	expr, err := p.bitwiseAnd()
	if err != nil {
		return nil, err
	}

	for p.match(lexer.Caret) {
		operator := p.previous()
		right, err := p.bitwiseAnd()
		if err != nil {
			return nil, err
		}
		expr = &ast.Binary{
			Left:     expr,
			Operator: operator,
			Right:    right,

			Span: lexer.Span{
				Start: expr.GetSpan().Start,
				End:   right.GetSpan().End,
			},
		}
	}

	return expr, nil
}

// bitwiseAnd parses a bitwise AND expression.
func (p *Parser) bitwiseAnd() (ast.Expr, error) {
	expr, err := p.shiftExpr()
	if err != nil {
		return nil, err
	}

	for p.match(lexer.Ampersand) {
		operator := p.previous()
		right, err := p.shiftExpr()
		if err != nil {
			return nil, err
		}
		expr = &ast.Binary{
			Left:     expr,
			Operator: operator,
			Right:    right,

			Span: lexer.Span{
				Start: expr.GetSpan().Start,
				End:   right.GetSpan().End,
			},
		}
	}

	return expr, nil
}

// shiftExpr parses a shift expression.
func (p *Parser) shiftExpr() (ast.Expr, error) {
	expr, err := p.sum()
	if err != nil {
		return nil, err
	}

	for p.match(lexer.LessLess, lexer.GreaterGreater) {
		operator := p.previous()
		right, err := p.sum()
		if err != nil {
			return nil, err
		}
		expr = &ast.Binary{
			Left:     expr,
			Operator: operator,
			Right:    right,

			Span: lexer.Span{
				Start: expr.GetSpan().Start,
				End:   right.GetSpan().End,
			},
		}
	}

	return expr, nil
}
