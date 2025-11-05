package parser

import (
	"fmt"
	"topple/compiler/ast"
	"topple/compiler/lexer"
)

// The scaffold parses only *one-line expression statements* so that you can
// start writing tests immediately and grow the grammar feature-by-feature.

type Parser struct {
	Tokens         []lexer.Token
	Current        int
	Errors         []error
	tempVarCounter int
}

// NewParser returns a new parser instance.
func NewParser(tokens []lexer.Token) *Parser {
	return &Parser{
		Tokens:         tokens,
		Current:        0,
		Errors:         []error{},
		tempVarCounter: 0,
	}
}

// Parse parses the tokens and returns a list of statements.
// It will attempt to recover from errors and return all encountered errors.
func (p *Parser) Parse() (*ast.Module, []error) {
	stmts := []ast.Stmt{}

	for !p.isAtEnd() {
		// Skip over any blank lines between statements. A blank line is just a
		// NEWLINE token that is not part of any real statement. By consuming
		// these eagerly we avoid producing ExprStmt nodes whose Value is nil
		// when the source line is empty.
		for p.check(lexer.Newline) {
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
		// Unwrap MultiStmt nodes at module level
		unwrapped := unwrapMultiStmt(stmt)
		stmts = append(stmts, unwrapped...)
	}

	return &ast.Module{Body: stmts}, p.Errors
}

// ParseError is an error that occurs in the parser.
type ParseError struct {
	Token   lexer.Token
	Message string
}

// Error returns a string representation of the ParseError.
func (e *ParseError) Error() string {
	if e.Token.Type == lexer.EOF {
		return fmt.Sprintf("at end: %s (position %s)", e.Message, e.Span())
	}
	return fmt.Sprintf("at '%s': %s (position %s)", e.Token.Lexeme, e.Message, e.Span())
}

// Span returns the span of the token that caused the error.
func (e *ParseError) Span() lexer.Span {
	return e.Token.Span
}

// NewParseError creates a new ParseError.
func NewParseError(token lexer.Token, message string) *ParseError {
	return &ParseError{Token: token, Message: message}
}

// unwrapMultiStmt takes a statement and returns a slice of statements.
// If the statement is a MultiStmt, it returns its inner statements.
// Otherwise, it returns a slice containing just the original statement.
func unwrapMultiStmt(stmt ast.Stmt) []ast.Stmt {
	if multi, ok := stmt.(*ast.MultiStmt); ok {
		return multi.Stmts
	}
	return []ast.Stmt{stmt}
}

// generateTempVarName generates a unique temporary variable name
func (p *Parser) generateTempVarName() string {
	p.tempVarCounter++
	return fmt.Sprintf("_chain_tmp_%d", p.tempVarCounter)
}
