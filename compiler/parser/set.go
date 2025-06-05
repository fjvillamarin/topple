package parser

import (
	"sylfie/compiler/ast"
	"sylfie/compiler/lexer"
)

func (p *Parser) set() (ast.Expr, error) {
	// Expect opening brace
	leftBrace, err := p.consume(lexer.LeftBrace, "expected '{'")
	if err != nil {
		return nil, err
	}

	elements := []ast.Expr{}

	// Parse elements
	// First element
	expr, err := p.starNamedExpression()
	if err != nil {
		return nil, err
	}
	elements = append(elements, expr)

	// Parse additional elements separated by commas
	for p.match(lexer.Comma) {
		// Allow trailing comma
		if p.check(lexer.RightBrace) {
			break
		}

		expr, err := p.starNamedExpression()
		if err != nil {
			return nil, err
		}
		elements = append(elements, expr)
	}

	// Expect closing brace
	rightBrace, err := p.consume(lexer.RightBrace, "expected '}'")
	if err != nil {
		return nil, err
	}

	return &ast.SetExpr{
		Elements: elements,

		Span: lexer.Span{Start: leftBrace.Start(), End: rightBrace.End()},
	}, nil
}
