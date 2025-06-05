package parser

import (
	"sylfie/compiler/ast"
	"sylfie/compiler/lexer"
)

// comparison parses a comparison expression.
func (p *Parser) comparison() (ast.Expr, error) {
	left, err := p.bitwiseOr()
	if err != nil {
		return nil, err
	}

	// First check for any comparison operator to start a comparison chain
	if isComparisonOperator(p.peek().Type) {
		// We're going to build a chain of comparisons
		operands := []ast.Expr{left}
		operators := []lexer.Token{}

		// Keep consuming comparison operators and their right operands
		for isComparisonOperator(p.peek().Type) {
			// Regular comparison operator - now includes IsNot and NotIn
			p.advance()
			operators = append(operators, p.previous())

			// Parse the right operand
			right, err := p.bitwiseOr()
			if err != nil {
				return nil, err
			}
			operands = append(operands, right)
		}

		// Handle a single comparison (most common case)
		if len(operands) == 2 {
			return &ast.Binary{
				Left:     operands[0],
				Operator: operators[0],
				Right:    operands[1],

				Span: lexer.Span{Start: operands[0].GetSpan().Start, End: operands[1].GetSpan().End},
			}, nil
		}

		// Handle chained comparisons (a < b < c becomes (a < b) and (b < c))
		var result ast.Expr
		for i := 0; i < len(operators); i++ {
			comparison := &ast.Binary{
				Left:     operands[i],
				Operator: operators[i],
				Right:    operands[i+1],

				Span: lexer.Span{Start: operands[i].GetSpan().Start, End: operands[i+1].GetSpan().End},
			}

			if i == 0 {
				result = comparison
			} else {
				// Create an AND expression linking the comparisons
				andToken := lexer.Token{Type: lexer.And, Lexeme: "and"}
				result = &ast.Binary{
					Left:     result,
					Operator: andToken,
					Right:    comparison,

					Span: lexer.Span{Start: result.GetSpan().Start, End: comparison.GetSpan().End},
				}
			}
		}
		return result, nil
	}

	return left, nil
}

// Helper function to check if a token type is a comparison operator
func isComparisonOperator(tokenType lexer.TokenType) bool {
	return tokenType == lexer.EqualEqual || tokenType == lexer.BangEqual ||
		tokenType == lexer.Less || tokenType == lexer.LessEqual ||
		tokenType == lexer.Greater || tokenType == lexer.GreaterEqual ||
		tokenType == lexer.In || tokenType == lexer.Is || tokenType == lexer.IsNot || tokenType == lexer.NotIn
}
