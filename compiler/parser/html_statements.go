package parser

import (
	"biscuit/compiler/ast"
	"biscuit/compiler/lexer"
	"fmt"
)

// htmlContentStatement parses statements that can appear inside HTML content
// This excludes function definitions which are only allowed at view level
func (p *Parser) htmlContentStatement() (ast.Stmt, error) {
	fmt.Printf("[DEBUG] htmlContentStatement()\n")
	// Check for HTML elements first
	if p.check(lexer.TagOpen) {
		return p.htmlElement()
	}

	// Check for compound statements (excluding function definitions)
	switch p.peek().Type {
	case lexer.If:
		return p.htmlIfStatement()
	case lexer.While:
		return p.htmlWhileStatement()
	case lexer.For:
		return p.htmlForStatement()
	case lexer.Try:
		return p.htmlTryStatement()
	case lexer.Match:
		return p.htmlMatchStatement()
	case lexer.Async:
		// For async statements, we need to look ahead
		if p.checkNext(lexer.For) {
			return p.htmlForStatement()
		}
		if p.checkNext(lexer.With) {
			// TODO: Implement async with in views when needed
			return nil, p.error(p.peek(), "async with statements not implemented in views yet")
		}
		if p.checkNext(lexer.Def) {
			return nil, p.error(p.peek(), "function definitions are not allowed inside HTML content")
		}
		// Fall through to simple statements for other async cases
		return p.simpleStatement()
	case lexer.Def:
		return nil, p.error(p.peek(), "function definitions are not allowed inside HTML content")
	}

	// Check for simple statements
	switch p.peek().Type {
	case lexer.Return:
		return p.returnStatement()
	case lexer.Pass:
		return p.passStatement()
	// TODO: Add more simple statements as needed
	default:
		// Fall back to regular statement parsing for Python statements
		return p.simpleStatement()
	}
}

// htmlContentBlock parses a block of statements within HTML content context
func (p *Parser) htmlContentBlock() ([]ast.Stmt, error) {
	// Check if this is a simple statement block (single line)
	if !p.check(lexer.Newline) {
		stmt, err := p.htmlContentStatement()
		if err != nil {
			return nil, err
		}
		return []ast.Stmt{stmt}, nil
	}

	// Otherwise expect NEWLINE INDENT statements DEDENT
	_, err := p.consume(lexer.Newline, "expected newline")
	if err != nil {
		return nil, err
	}

	// Consume all remaining newlines
	for p.check(lexer.Newline) {
		p.advance()
	}

	_, err = p.consume(lexer.Indent, "expected indented block")
	if err != nil {
		return nil, err
	}

	statements := []ast.Stmt{}
	for !p.isAtEnd() && !p.check(lexer.Dedent) {
		stmt, err := p.htmlContentStatement()
		if err != nil {
			return nil, err
		}

		statements = append(statements, stmt)

		// Consume all the newlines we see
		for p.check(lexer.Newline) {
			p.advance()
		}
	}

	_, err = p.consume(lexer.Dedent, "expected dedent at end of block")
	if err != nil {
		return nil, err
	}

	return statements, nil
}
