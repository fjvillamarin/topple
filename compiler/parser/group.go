package parser

import (
	"github.com/fjvillamarin/topple/compiler/ast"
	"github.com/fjvillamarin/topple/compiler/lexer"
)

func (p *Parser) group() (ast.Expr, error) {
	// Expect opening parenthesis
	leftParen, err := p.consume(lexer.LeftParen, "expected '('")
	if err != nil {
		return nil, err
	}

	// Check for yield expression
	if p.check(lexer.Yield) {
		expr, err := p.yieldExpression()
		if err != nil {
			return nil, err
		}

		// Expect closing parenthesis
		rightParen, err := p.consume(lexer.RightParen, "expected ')'")
		if err != nil {
			return nil, err
		}

		return &ast.GroupExpr{
			Expression: expr,

			Span: lexer.Span{Start: leftParen.Start(), End: rightParen.End()},
		}, nil
	}

	// Parse named expression
	expr, err := p.namedExpression()
	if err != nil {
		return nil, err
	}

	// Expect closing parenthesis
	rightParen, err := p.consume(lexer.RightParen, "expected ')'")
	if err != nil {
		return nil, err
	}

	return &ast.GroupExpr{
		Expression: expr,

		Span: lexer.Span{Start: leftParen.Start(), End: rightParen.End()},
	}, nil
}
