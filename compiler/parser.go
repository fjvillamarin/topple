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
	atom, err := p.primary()
	if err != nil {
		return nil, err
	}

	return atom, nil
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
