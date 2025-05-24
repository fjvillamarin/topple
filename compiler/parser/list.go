package parser

import (
	"biscuit/compiler/ast"
	"biscuit/compiler/lexer"
	"fmt"
)

func (p *Parser) list() (ast.Expr, error) {
	// Expect opening bracket
	leftBracket, err := p.consume(lexer.LeftBracket, "expected '['")
	if err != nil {
		return nil, err
	}

	elements := []ast.Expr{}

	// Parse elements if the list is not empty
	if !p.check(lexer.RightBracket) {
		// Parse star named expressions
		expr, err := p.starNamedExpression()
		if err != nil {
			return nil, err
		}
		elements = append(elements, expr)

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
	}

	fmt.Println("elements", elements)

	// Expect closing bracket
	rightBracket, err := p.consume(lexer.RightBracket, "expected ']'")
	if err != nil {
		return nil, err
	}

	return &ast.ListExpr{
		Elements: elements,

		Span: lexer.Span{Start: leftBracket.Start(), End: rightBracket.End()},
	}, nil
}
