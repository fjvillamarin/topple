package parser

import (
	"topple/compiler/ast"
	"topple/compiler/lexer"
)

// slice parses a single slice element as per the grammar:
// slice:
//
//	| [expression] ':' [expression] [':' [expression] ]
//	| named_expression
func (p *Parser) slice() (ast.Expr, error) {
	// Check if this is a slice notation or just an expression
	// We need to look ahead to see if there's a colon after the first expression (if any)
	startPos := p.peek().Start()

	// Empty slice is allowed (:)
	if p.check(lexer.Colon) {
		// No start expression, consume the colon
		p.advance()

		var end ast.Expr
		var err error

		// Check for end expression after colon
		if !p.check(lexer.Colon) && !p.check(lexer.RightBracket) && !p.check(lexer.Comma) {
			end, err = p.expression()
			if err != nil {
				return nil, err
			}
		}

		// Check for step (second colon)
		var step ast.Expr
		if p.match(lexer.Colon) {
			// Parse optional step
			if !p.check(lexer.RightBracket) && !p.check(lexer.Comma) {
				step, err = p.expression()
				if err != nil {
					return nil, err
				}
			}
		}

		endPos := p.previous().End()
		if step != nil {
			endPos = step.GetSpan().End
		} else if end != nil {
			endPos = end.GetSpan().End
		}

		return &ast.Slice{
			StartIndex: nil,
			EndIndex:   end,
			Step:       step,

			Span: lexer.Span{Start: startPos, End: endPos},
		}, nil
	}

	// There's an expression before any potential colon
	expr, err := p.expression()
	if err != nil {
		return nil, err
	}

	// If there's no colon after the expression, this is just a regular index
	if !p.match(lexer.Colon) {
		return expr, nil
	}

	// We have a slice with a start expression
	var end ast.Expr

	// Check for end expression after colon
	if !p.check(lexer.Colon) && !p.check(lexer.RightBracket) && !p.check(lexer.Comma) {
		end, err = p.expression()
		if err != nil {
			return nil, err
		}
	}

	// Check for step (second colon)
	var step ast.Expr
	if p.match(lexer.Colon) {
		// Parse optional step
		if !p.check(lexer.RightBracket) && !p.check(lexer.Comma) {
			step, err = p.expression()
			if err != nil {
				return nil, err
			}
		}
	}

	endPos := expr.GetSpan().End
	if step != nil {
		endPos = step.GetSpan().End
	} else if end != nil {
		endPos = end.GetSpan().End
	}

	return &ast.Slice{
		StartIndex: expr,
		EndIndex:   end,
		Step:       step,

		Span: lexer.Span{Start: startPos, End: endPos},
	}, nil
}

// slices parses one or more slice elements as per the grammar:
// slices:
//
//	| slice !','
//	| ','.(slice | starred_expression)+ [',']
func (p *Parser) slices() ([]ast.Expr, error) {
	// Parse the first slice
	first, err := p.slice()
	if err != nil {
		return nil, err
	}

	// Initialize the slice with the first element
	indices := []ast.Expr{first}

	// If there's no comma, it's just a single index/slice
	if !p.match(lexer.Comma) {
		return indices, nil
	}

	// Parse additional slice or starred_expression elements
	for !p.check(lexer.RightBracket) {
		// Handle starred expressions
		if p.match(lexer.Star) {
			// This is a starred expression like *args
			star := p.previous()
			expr, err := p.bitwiseOr() // According to the grammar, star expressions use bitwise_or
			if err != nil {
				return nil, err
			}
			indices = append(indices, &ast.StarExpr{
				Expr: expr,

				Span: lexer.Span{Start: star.Start(), End: expr.GetSpan().End},
			})
		} else {
			// Regular slice expression
			expr, err := p.slice()
			if err != nil {
				return nil, err
			}
			indices = append(indices, expr)
		}

		// Break if no more commas
		if !p.match(lexer.Comma) {
			break
		}

		// Allow trailing comma
		if p.check(lexer.RightBracket) {
			break
		}
	}

	return indices, nil
}
