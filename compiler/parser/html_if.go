package parser

import (
	"biscuit/compiler/ast"
	"biscuit/compiler/lexer"
)

// htmlIfStatement parses an if statement within HTML content context.
// Uses the same AST node as regular if statements but uses HTML content block parsing.
func (p *Parser) htmlIfStatement() (ast.Stmt, error) {
	// Consume the 'if' token
	ifToken, err := p.consume(lexer.If, "expected 'if'")
	if err != nil {
		return nil, err
	}

	// Parse the condition expression
	condition, err := p.namedExpression()
	if err != nil {
		return nil, err
	}

	// Expect colon
	_, err = p.consume(lexer.Colon, "expected ':' after if condition")
	if err != nil {
		return nil, err
	}

	// Parse the body using HTML content block parsing
	body, err := p.htmlContentBlock()
	if err != nil {
		return nil, err
	}

	// Check for elif
	if p.match(lexer.Elif) {
		// Parse elif as a separate if statement
		elifStmt, err := p.htmlElifStatement()
		if err != nil {
			return nil, err
		}

		// Return If with elif as the Else branch
		return &ast.If{
			Condition: condition,
			Body:      body,
			Else:      []ast.Stmt{elifStmt},
			Span:      lexer.Span{Start: ifToken.Start(), End: elifStmt.GetSpan().End},
		}, nil
	}

	// Check for else
	elseBody, err := p.htmlElseBlock()
	if err != nil {
		return nil, err
	}

	// Determine the end position
	var endPos lexer.Position
	if len(elseBody) > 0 {
		endPos = elseBody[len(elseBody)-1].GetSpan().End
	} else if len(body) > 0 {
		endPos = body[len(body)-1].GetSpan().End
	} else {
		endPos = p.previous().End()
	}

	return &ast.If{
		Condition: condition,
		Body:      body,
		Else:      elseBody,
		Span:      lexer.Span{Start: ifToken.Start(), End: endPos},
	}, nil
}

// htmlElifStatement parses an elif part of an if statement within HTML content context.
func (p *Parser) htmlElifStatement() (ast.Stmt, error) {
	// Already consumed 'elif' token
	elifToken := p.previous()

	// Parse condition
	condition, err := p.namedExpression()
	if err != nil {
		return nil, err
	}

	// Expect colon
	_, err = p.consume(lexer.Colon, "expected ':' after elif condition")
	if err != nil {
		return nil, err
	}

	// Parse body using HTML content block parsing
	body, err := p.htmlContentBlock()
	if err != nil {
		return nil, err
	}

	// Check for another elif
	if p.match(lexer.Elif) {
		// Parse the next elif
		nextElif, err := p.htmlElifStatement()
		if err != nil {
			return nil, err
		}

		// Return If with elif as the Else branch
		return &ast.If{
			Condition: condition,
			Body:      body,
			Else:      []ast.Stmt{nextElif},
			Span:      lexer.Span{Start: elifToken.Start(), End: nextElif.GetSpan().End},
		}, nil
	}

	// Check for else
	elseBody, err := p.htmlElseBlock()
	if err != nil {
		return nil, err
	}

	// Determine the end position
	var endPos lexer.Position
	if len(elseBody) > 0 {
		endPos = elseBody[len(elseBody)-1].GetSpan().End
	} else if len(body) > 0 {
		endPos = body[len(body)-1].GetSpan().End
	} else {
		endPos = p.previous().End()
	}

	return &ast.If{
		Condition: condition,
		Body:      body,
		Else:      elseBody,
		Span:      lexer.Span{Start: elifToken.Start(), End: endPos},
	}, nil
}

// htmlElseBlock parses an else block within HTML content context.
// Returns empty slice if no else block is present.
func (p *Parser) htmlElseBlock() ([]ast.Stmt, error) {
	if !p.match(lexer.Else) {
		return []ast.Stmt{}, nil
	}

	// Expect colon
	_, err := p.consume(lexer.Colon, "expected ':' after else")
	if err != nil {
		return nil, err
	}

	// Parse else body using HTML content block parsing
	return p.htmlContentBlock()
}
