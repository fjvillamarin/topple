package parser

import (
	"github.com/fjvillamarin/topple/compiler/ast"
	"github.com/fjvillamarin/topple/compiler/lexer"
)

// whileStatement parses a while statement.
func (p *Parser) whileStatement() (ast.Stmt, error) {
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

	// Parse the body
	body, err := p.block()
	if err != nil {
		return nil, err
	}

	// Check for else
	elseBody, err := p.elseBlock()
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
