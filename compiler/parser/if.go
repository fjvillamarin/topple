package parser

import (
	"biscuit/compiler/ast"
	"biscuit/compiler/lexer"
	"fmt"
)

// ifStatement parses an if statement.
func (p *Parser) ifStatement() (ast.Stmt, error) {
	// Consume the 'if' token
	ifToken, err := p.consume(lexer.If, "expected 'if'")
	if err != nil {
		return nil, err
	}

	fmt.Println("ifStatement")

	// Parse the condition expression
	condition, err := p.namedExpression()
	if err != nil {
		return nil, err
	}

	fmt.Println("condition", condition)

	// Expect colon
	_, err = p.consume(lexer.Colon, "expected ':' after if condition")
	if err != nil {
		return nil, err
	}

	fmt.Println("colon")

	// Parse the body
	body, err := p.block()
	if err != nil {
		return nil, err
	}

	fmt.Println("body", body)

	// Check for elif
	if p.match(lexer.Elif) {
		// Parse elif as a separate if statement
		elifStmt, err := p.elifStatement()
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
	var elseBody []ast.Stmt
	var endPos lexer.Position

	if p.match(lexer.Else) {
		// Expect colon
		_, err = p.consume(lexer.Colon, "expected ':' after 'else'")
		if err != nil {
			return nil, err
		}

		// Parse else body
		elseBody, err = p.block()
		if err != nil {
			return nil, err
		}

		// Set end position to end of else body
		if len(elseBody) > 0 {
			endPos = elseBody[len(elseBody)-1].GetSpan().End
		} else {
			endPos = p.previous().End()
		}
	} else {
		// No else, end is end of if body
		if len(body) > 0 {
			endPos = body[len(body)-1].GetSpan().End
		} else {
			endPos = p.previous().End()
		}
	}

	return &ast.If{
		Condition: condition,
		Body:      body,
		Else:      elseBody,
		Span:      lexer.Span{Start: ifToken.Start(), End: endPos},
	}, nil
}

// elifStatement parses an elif part of an if statement.
func (p *Parser) elifStatement() (ast.Stmt, error) {
	// Condition already consumed 'elif' token
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

	// Parse body
	body, err := p.block()
	if err != nil {
		return nil, err
	}

	// Check for another elif
	if p.match(lexer.Elif) {
		// Parse the next elif
		nextElif, err := p.elifStatement()
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
	var elseBody []ast.Stmt
	var endPos lexer.Position

	if p.match(lexer.Else) {
		// Expect colon
		_, err = p.consume(lexer.Colon, "expected ':' after 'else'")
		if err != nil {
			return nil, err
		}

		// Parse else body
		elseBody, err = p.block()
		if err != nil {
			return nil, err
		}

		// Set end position to end of else body
		if len(elseBody) > 0 {
			endPos = elseBody[len(elseBody)-1].GetSpan().End
		} else {
			endPos = p.previous().End()
		}
	} else {
		// No else, end is end of elif body
		if len(body) > 0 {
			endPos = body[len(body)-1].GetSpan().End
		} else {
			endPos = p.previous().End()
		}
	}

	return &ast.If{
		Condition: condition,
		Body:      body,
		Else:      elseBody,
		Span:      lexer.Span{Start: elifToken.Start(), End: endPos},
	}, nil
}
