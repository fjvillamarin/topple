package parser

import (
	"biscuit/compiler/ast"
	"biscuit/compiler/lexer"
)

// htmlWhileStatement parses a while statement within HTML content context.
// Uses the same AST node as regular while statements but uses HTML content block parsing.
func (p *Parser) htmlWhileStatement() (ast.Stmt, error) {
	// Consume the 'while' token
	whileToken, err := p.consume(lexer.While, "expected 'while'")
	if err != nil {
		return nil, err
	}

	// Parse the condition expression
	condition, err := p.namedExpression()
	if err != nil {
		return nil, err
	}

	// Expect colon
	_, err = p.consume(lexer.Colon, "expected ':' after while condition")
	if err != nil {
		return nil, err
	}

	// Parse the body using HTML content block parsing
	body, err := p.htmlContentBlock()
	if err != nil {
		return nil, err
	}

	// Check for else using while-specific function
	elseBody, err := p.htmlWhileElseBlock()
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

	return &ast.While{
		Test: condition,
		Body: body,
		Else: elseBody,
		Span: lexer.Span{Start: whileToken.Start(), End: endPos},
	}, nil
}

// htmlWhileElseBlock parses an else block for while statements within HTML content context.
// Returns empty slice if no else block is present.
func (p *Parser) htmlWhileElseBlock() ([]ast.Stmt, error) {
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
