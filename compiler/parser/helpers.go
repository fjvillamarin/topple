package parser

import "biscuit/compiler/lexer"

func (p *Parser) consume(t lexer.TokenType, message string) (lexer.Token, error) {
	if p.check(t) {
		return p.advance(), nil
	}

	return lexer.Token{}, p.error(p.peek(), message)
}

func (p *Parser) error(token lexer.Token, message string) error {
	return &ParseError{Token: token, Message: message}
}

// match checks if the current token is one of the given types.
func (p *Parser) match(types ...lexer.TokenType) bool {
	for _, t := range types {
		if p.check(t) {
			p.advance()
			return true
		}
	}
	return false
}

func (p *Parser) check(t lexer.TokenType) bool {
	if p.isAtEnd() {
		return false
	}
	return p.peek().Type == t
}

func (p *Parser) checkNext(t lexer.TokenType) bool {
	if p.isAtEnd() {
		return false
	}
	return p.peekN(1).Type == t
}

func (p *Parser) advance() lexer.Token {
	if !p.isAtEnd() {
		p.Current++
	}
	return p.previous()
}

func (p *Parser) isAtEnd() bool {
	return p.peek().Type == lexer.EOF
}

func (p *Parser) peek() lexer.Token {
	return p.Tokens[p.Current]
}

func (p *Parser) peekN(n int) lexer.Token {
	return p.Tokens[p.Current+n]
}

func (p *Parser) previous() lexer.Token {
	return p.Tokens[p.Current-1]
}
