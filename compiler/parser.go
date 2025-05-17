package compiler

// The scaffold parses only *one-line expression statements* so that you can
// start writing tests immediately and grow the grammar feature-by-feature.

type Parser struct {
	Tokens  []Token
	Current int
	Errors  []error
}

// NewParser returns a new parser instance.
func NewParser(tokens []Token) *Parser {
	return &Parser{
		Tokens:  tokens,
		Current: 0,
		Errors:  []error{},
	}
}

// Parse parses the tokens and returns a list of statements.
// It will attempt to recover from errors and return all encountered errors.
func (p *Parser) Parse() (*Module, []error) {
	stmts := []Stmt{}

	for !p.isAtEnd() {
		// Skip over any blank lines between statements. A blank line is just a
		// NEWLINE token that is not part of any real statement. By consuming
		// these eagerly we avoid producing ExprStmt nodes whose Value is nil
		// when the source line is empty.
		for p.check(Newline) {
			p.advance()
		}

		// If we have reached EOF after skipping newlines, break out of the
		// loop early so we don't attempt to parse a statement starting at EOF.
		if p.isAtEnd() {
			break
		}

		stmt, err := p.statement()
		if err != nil {
			p.Errors = append(p.Errors, err)
			return nil, p.Errors
		}
		stmts = append(stmts, stmt)
	}

	return &Module{Body: stmts}, p.Errors
}

// ----------------------------------------------------------------------------
// Statements
// ----------------------------------------------------------------------------

// statement parses a single statement.
func (p *Parser) statement() (Stmt, error) {
	return p.expressionStatement()
}

// expressionStatement parses an expression statement.
func (p *Parser) expressionStatement() (Stmt, error) {
	expr, err := p.expression()
	if err != nil {
		return nil, err
	}

	// Consume semicolon if it's there (optional)
	if p.check(Semicolon) {
		p.advance()
	}

	// If we're at the end of the file, we're done
	if p.isAtEnd() {
		return NewExprStmt(expr, expr.Start(), expr.End()), nil
	}

	// Consume the newline
	_, err = p.consume(Newline, "expected newline after expression")
	if err != nil {
		return nil, err
	}

	return NewExprStmt(expr, expr.Start(), expr.End()), nil
}

// ----------------------------------------------------------------------------
// Expressions
// ----------------------------------------------------------------------------

// expression parses an expression.
func (p *Parser) expression() (Expr, error) {
	return p.disjunction()
}

// disjunction parses a disjunction expression.
func (p *Parser) disjunction() (Expr, error) {
	expr, err := p.conjunction()
	if err != nil {
		return nil, err
	}

	for p.match(Or) {
		operator := p.previous()
		right, err := p.conjunction()
		if err != nil {
			return nil, err
		}
		expr = NewBinary(expr, operator, right, expr.Start(), right.End())
	}

	return expr, nil
}

// conjunction parses a conjunction expression.
func (p *Parser) conjunction() (Expr, error) {
	expr, err := p.inversion()
	if err != nil {
		return nil, err
	}

	for p.match(And) {
		operator := p.previous()
		right, err := p.inversion()
		if err != nil {
			return nil, err
		}
		expr = NewBinary(expr, operator, right, expr.Start(), right.End())
	}

	return expr, nil
}

// inversion parses an inversion expression.
func (p *Parser) inversion() (Expr, error) {
	if p.match(Not) {
		operator := p.previous()
		right, err := p.inversion()
		if err != nil {
			return nil, err
		}
		return NewUnary(operator, right, operator.Start(), right.End()), nil
	}

	return p.comparison()
}

// comparison parses a comparison expression.
func (p *Parser) comparison() (Expr, error) {
	left, err := p.bitwiseOr()
	if err != nil {
		return nil, err
	}

	// First check for any comparison operator to start a comparison chain
	if isComparisonOperator(p.peek().Type) {
		// We're going to build a chain of comparisons
		operands := []Expr{left}
		operators := []Token{}

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
			return NewBinary(operands[0], operators[0], operands[1], operands[0].Start(), operands[1].End()), nil
		}

		// Handle chained comparisons (a < b < c becomes (a < b) and (b < c))
		var result Expr
		for i := 0; i < len(operators); i++ {
			comparison := NewBinary(operands[i], operators[i], operands[i+1],
				operands[i].Start(), operands[i+1].End())

			if i == 0 {
				result = comparison
			} else {
				// Create an AND expression linking the comparisons
				andToken := Token{Type: And, Lexeme: "and"}
				result = NewBinary(result, andToken, comparison,
					result.Start(), comparison.End())
			}
		}
		return result, nil
	}

	return left, nil
}

// Helper function to check if a token type is a comparison operator
func isComparisonOperator(tokenType TokenType) bool {
	return tokenType == EqualEqual || tokenType == BangEqual ||
		tokenType == Less || tokenType == LessEqual ||
		tokenType == Greater || tokenType == GreaterEqual ||
		tokenType == In || tokenType == Is || tokenType == IsNot || tokenType == NotIn
}

// bitwise_or parses a bitwise OR expression.
func (p *Parser) bitwiseOr() (Expr, error) {
	expr, err := p.bitwiseXor()
	if err != nil {
		return nil, err
	}

	for p.match(Pipe) {
		operator := p.previous()
		right, err := p.bitwiseXor()
		if err != nil {
			return nil, err
		}
		expr = NewBinary(expr, operator, right, expr.Start(), right.End())
	}

	return expr, nil
}

// bitwiseXor parses a bitwise XOR expression.
func (p *Parser) bitwiseXor() (Expr, error) {
	expr, err := p.bitwiseAnd()
	if err != nil {
		return nil, err
	}

	for p.match(Caret) {
		operator := p.previous()
		right, err := p.bitwiseAnd()
		if err != nil {
			return nil, err
		}
		expr = NewBinary(expr, operator, right, expr.Start(), right.End())
	}

	return expr, nil
}

// bitwiseAnd parses a bitwise AND expression.
func (p *Parser) bitwiseAnd() (Expr, error) {
	expr, err := p.shiftExpr()
	if err != nil {
		return nil, err
	}

	for p.match(Ampersand) {
		operator := p.previous()
		right, err := p.shiftExpr()
		if err != nil {
			return nil, err
		}
		expr = NewBinary(expr, operator, right, expr.Start(), right.End())
	}

	return expr, nil
}

// shiftExpr parses a shift expression.
func (p *Parser) shiftExpr() (Expr, error) {
	expr, err := p.sum()
	if err != nil {
		return nil, err
	}

	for p.match(LessLess, GreaterGreater) {
		operator := p.previous()
		right, err := p.sum()
		if err != nil {
			return nil, err
		}
		expr = NewBinary(expr, operator, right, expr.Start(), right.End())
	}

	return expr, nil
}

// sum parses a sum expression.
func (p *Parser) sum() (Expr, error) {
	expr, err := p.term()
	if err != nil {
		return nil, err
	}

	for p.match(Plus, Minus) {
		operator := p.previous()
		right, err := p.term()
		if err != nil {
			return nil, err
		}
		expr = NewBinary(expr, operator, right, expr.Start(), right.End())
	}

	return expr, nil
}

// term parses a term expression.
func (p *Parser) term() (Expr, error) {
	expr, err := p.factor()
	if err != nil {
		return nil, err
	}

	for p.match(Star, Slash, SlashSlash, Percent, At) {
		operator := p.previous()
		right, err := p.factor()
		if err != nil {
			return nil, err
		}
		expr = NewBinary(expr, operator, right, expr.Start(), right.End())
	}

	return expr, nil
}

// factor parses a factor expression.
func (p *Parser) factor() (Expr, error) {
	if p.match(Plus, Minus, Tilde) {
		operator := p.previous()
		right, err := p.factor()
		if err != nil {
			return nil, err
		}
		return NewUnary(operator, right, operator.Start(), right.End()), nil
	}

	return p.power()
}

// power parses a power expression.
func (p *Parser) power() (Expr, error) {
	expr, err := p.primary()
	if err != nil {
		return nil, err
	}

	if p.match(StarStar) {
		operator := p.previous()
		right, err := p.factor()
		if err != nil {
			return nil, err
		}
		return NewBinary(expr, operator, right, expr.Start(), right.End()), nil
	}

	return expr, nil
}

// primary parses a primary expression.
func (p *Parser) primary() (Expr, error) {
	// Parse the initial atom
	expr, err := p.atom()
	if err != nil {
		return nil, err
	}

	// Keep parsing postfix operations while they exist
	for {
		if p.match(Dot) {
			// Handle attribute access: expr.NAME
			name, err := p.consume(Identifier, "expected identifier after '.'")
			if err != nil {
				return nil, err
			}
			expr = NewAttribute(expr, name, expr.Start(), name.End())
		} else if p.match(LeftParen) {
			// Handle function call: expr(args)
			expr, err = p.finishCall(expr)
			if err != nil {
				return nil, err
			}
		} else if p.match(LeftBracket) {
			// Handle subscript access: expr[index]
			index, err := p.expression()
			if err != nil {
				return nil, err
			}

			right, err := p.consume(RightBracket, "expected ']' after index")
			if err != nil {
				return nil, err
			}
			expr = NewSubscript(expr, index, expr.Start(), right.End())
		} else {
			// No more postfix operations
			break
		}
	}

	return expr, nil
}

func (p *Parser) finishCall(callee Expr) (Expr, error) {
	args := []Expr{}
	if !p.check(RightParen) {
		// Parse first argument
		arg, err := p.expression()
		if err != nil {
			return nil, err
		}
		args = append(args, arg)

		// Parse additional arguments
		for p.match(Comma) {
			// If there's a right parenthesis after the comma,
			// it's a trailing comma, so we're done parsing arguments
			if p.check(RightParen) {
				break
			}

			arg, err := p.expression()
			if err != nil {
				return nil, err
			}
			args = append(args, arg)
		}
	}

	right, err := p.consume(RightParen, "expected ')' after arguments")
	if err != nil {
		return nil, err
	}

	return NewCall(callee, args, callee.Start(), right.End()), nil
}

// atom parses an atom.
func (p *Parser) atom() (Expr, error) {
	if p.match(False) {
		return NewLiteral(p.previous(), false, p.previous().Start(), p.previous().End()), nil
	}

	if p.match(True) {
		return NewLiteral(p.previous(), true, p.previous().Start(), p.previous().End()), nil
	}

	if p.match(None) {
		return NewLiteral(p.previous(), nil, p.previous().Start(), p.previous().End()), nil
	}

	if p.match(Number, String) {
		return NewLiteral(p.previous(), p.previous().Literal, p.previous().Start(), p.previous().End()), nil
	}

	if p.match(Ellipsis) {
		return NewLiteral(p.previous(), nil, p.previous().Start(), p.previous().End()), nil
	}

	if p.match(Identifier) {
		return NewName(p.previous(), p.previous().Start(), p.previous().End()), nil
	}

	// TODO: Support tuples, groups, lists, etc.

	return nil, p.error(p.peek(), "unexpected token")
}

// ----------------------------------------------------------------------------
// Helper functions
// ----------------------------------------------------------------------------

func (p *Parser) consume(t TokenType, message string) (Token, error) {
	if p.check(t) {
		return p.advance(), nil
	}

	return Token{}, p.error(p.peek(), message)
}

func (p *Parser) error(token Token, message string) error {
	return &ParseError{Token: token, Message: message}
}

// match checks if the current token is one of the given types.
func (p *Parser) match(types ...TokenType) bool {
	for _, t := range types {
		if p.check(t) {
			p.advance()
			return true
		}
	}
	return false
}

func (p *Parser) check(t TokenType) bool {
	if p.isAtEnd() {
		return false
	}
	return p.peek().Type == t
}

func (p *Parser) advance() Token {
	if !p.isAtEnd() {
		p.Current++
	}
	return p.previous()
}

func (p *Parser) isAtEnd() bool {
	return p.peek().Type == EOF
}

func (p *Parser) peek() Token {
	return p.Tokens[p.Current]
}

func (p *Parser) previous() Token {
	return p.Tokens[p.Current-1]
}
