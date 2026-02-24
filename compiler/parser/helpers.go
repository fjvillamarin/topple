package parser

import (
	"github.com/fjvillamarin/topple/compiler/lexer"
)

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

// consumeNameOrKeyword consumes a token that can be used as a name.
// This includes Identifier tokens and also keywords when they appear in
// contexts where Python allows keywords as names (e.g., dotted import paths,
// attribute access). Python allows keywords like "class", "for", etc. as
// attribute names: obj.class, obj.for, import module.class
func (p *Parser) consumeNameOrKeyword(context string) (lexer.Token, error) {
	token := p.peek()

	// Check if it's an identifier
	if token.Type == lexer.Identifier {
		return p.advance(), nil
	}

	// Check if it's a keyword that can be used as a name
	// Python allows all keywords as attribute names in dotted paths
	if p.isKeyword(token.Type) {
		return p.advance(), nil
	}

	return lexer.Token{}, p.error(token, "expected identifier"+context)
}

// isKeyword returns true if the token type is a keyword
func (p *Parser) isKeyword(t lexer.TokenType) bool {
	// Python keywords
	switch t {
	case lexer.And, lexer.As, lexer.Assert, lexer.Async, lexer.Await,
		lexer.Break, lexer.Class, lexer.Continue, lexer.Def, lexer.Del,
		lexer.Elif, lexer.Else, lexer.Except, lexer.False, lexer.Finally,
		lexer.For, lexer.From, lexer.Global, lexer.If, lexer.Import,
		lexer.In, lexer.Is, lexer.Lambda, lexer.Match, lexer.None,
		lexer.Nonlocal, lexer.Not, lexer.Or, lexer.Pass, lexer.Raise,
		lexer.Return, lexer.True, lexer.Try, lexer.While, lexer.With,
		lexer.Yield, lexer.Case, lexer.Type:
		return true
	// PSX-specific keywords
	case lexer.View, lexer.Component:
		return true
	default:
		return false
	}
}
