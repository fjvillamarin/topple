package parser

import (
	"github.com/fjvillamarin/topple/compiler/ast"
	"github.com/fjvillamarin/topple/compiler/lexer"
)

func (p *Parser) list() (ast.Expr, error) {
	// Expect opening bracket
	leftBracket, err := p.consume(lexer.LeftBracket, "expected '['")
	if err != nil {
		return nil, err
	}

	// Check for empty list
	if p.check(lexer.RightBracket) {
		rightBracket, err := p.consume(lexer.RightBracket, "expected ']'")
		if err != nil {
			return nil, err
		}
		return &ast.ListExpr{
			Elements: []ast.Expr{},
			Span:     lexer.Span{Start: leftBracket.Start(), End: rightBracket.End()},
		}, nil
	}

	// Parse the first expression
	expr, err := p.starNamedExpression()
	if err != nil {
		return nil, err
	}

	// Check if this is a comprehension (look for 'for' or 'async')
	if p.check(lexer.For) || p.check(lexer.Async) {
		// It's a list comprehension
		return p.listComp(expr, leftBracket)
	}

	// It's a regular list
	elements := []ast.Expr{expr}

	// Parse additional elements separated by commas
	for p.match(lexer.Comma) {
		// Allow trailing comma
		if p.check(lexer.RightBracket) {
			break
		}

		expr, err := p.starNamedExpression()
		if err != nil {
			return nil, err
		}
		elements = append(elements, expr)
	}

	// Expect closing bracket
	rightBracket, err := p.consume(lexer.RightBracket, "expected ']'")
	if err != nil {
		return nil, err
	}

	return &ast.ListExpr{
		Elements: elements,
		Span:     lexer.Span{Start: leftBracket.Start(), End: rightBracket.End()},
	}, nil
}
