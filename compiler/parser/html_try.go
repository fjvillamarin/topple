package parser

import (
	"biscuit/compiler/ast"
	"biscuit/compiler/lexer"
)

// htmlTryStatement parses a try statement within HTML content context.
// Uses the same AST node as regular try statements but uses HTML content block parsing.
func (p *Parser) htmlTryStatement() (ast.Stmt, error) {
	// Consume the 'try' keyword
	tryToken, err := p.consume(lexer.Try, "expected 'try'")
	if err != nil {
		return nil, err
	}

	// Expect colon
	_, err = p.consume(lexer.Colon, "expected ':' after 'try'")
	if err != nil {
		return nil, err
	}

	// Parse the try body using HTML content block parsing
	body, err := p.htmlContentBlock()
	if err != nil {
		return nil, err
	}

	// Check what follows - except, except*, or finally
	var excepts []ast.Except
	var elseBody []ast.Stmt
	var finallyBody []ast.Stmt
	var endPos lexer.Position

	// Parse except blocks if present
	if p.check(lexer.Except) {
		// Parse one or more except blocks
		for p.match(lexer.Except) {
			except, err := p.htmlParseExceptBlock()
			if err != nil {
				return nil, err
			}
			excepts = append(excepts, except)
		}

		// Parse optional else block
		if p.match(lexer.Else) {
			elseBody, err = p.htmlParseElseBlock()
			if err != nil {
				return nil, err
			}
		}

		// Parse optional finally block
		if p.match(lexer.Finally) {
			finallyBody, err = p.htmlParseFinallyBlock()
			if err != nil {
				return nil, err
			}
		}
	} else if p.match(lexer.Finally) {
		// Just a finally block without any except blocks
		finallyBody, err = p.htmlParseFinallyBlock()
		if err != nil {
			return nil, err
		}
	} else {
		return nil, p.error(p.peek(), "expected 'except' or 'finally' after try block")
	}

	// Determine the end position of the entire try statement
	if len(finallyBody) > 0 {
		endPos = finallyBody[len(finallyBody)-1].GetSpan().End
	} else if len(elseBody) > 0 {
		endPos = elseBody[len(elseBody)-1].GetSpan().End
	} else if len(excepts) > 0 {
		endPos = excepts[len(excepts)-1].Body[len(excepts[len(excepts)-1].Body)-1].GetSpan().End
	} else {
		endPos = body[len(body)-1].GetSpan().End
	}

	return &ast.Try{
		Body:    body,
		Excepts: excepts,
		Else:    elseBody,
		Finally: finallyBody,
		Span:    lexer.Span{Start: tryToken.Start(), End: endPos},
	}, nil
}

// htmlParseExceptBlock parses an except block within HTML content context.
func (p *Parser) htmlParseExceptBlock() (ast.Except, error) {
	// The 'except' token has already been consumed by the caller
	exceptToken := p.previous()

	// Check if this is an 'except*' form
	isStar := p.match(lexer.Star)

	// Parse the exception type, if present
	var exceptionType ast.Expr
	var err error
	var asName *ast.Name

	// Empty except: has no expression
	if !p.check(lexer.Colon) {
		// Parse the exception type expression
		exceptionType, err = p.expression()
		if err != nil {
			return ast.Except{}, err
		}

		// Check for 'as NAME'
		if p.match(lexer.As) {
			nameToken, err := p.consume(lexer.Identifier, "expected identifier after 'as'")
			if err != nil {
				return ast.Except{}, err
			}
			asName = &ast.Name{
				Token: nameToken,
				Span:  lexer.Span{Start: nameToken.Start(), End: nameToken.End()},
			}
		}
	}

	// Expect colon
	_, err = p.consume(lexer.Colon, "expected ':' after except clause")
	if err != nil {
		return ast.Except{}, err
	}

	// Parse the except body using HTML content block parsing
	body, err := p.htmlContentBlock()
	if err != nil {
		return ast.Except{}, err
	}

	// Determine the end position
	endPos := body[len(body)-1].GetSpan().End

	return ast.Except{
		Type:   exceptionType,
		Name:   asName,
		Body:   body,
		IsStar: isStar,
		Span:   lexer.Span{Start: exceptToken.Start(), End: endPos},
	}, nil
}

// htmlParseElseBlock parses an else block within HTML content context. The 'else' token has already been consumed.
func (p *Parser) htmlParseElseBlock() ([]ast.Stmt, error) {
	// Expect colon
	_, err := p.consume(lexer.Colon, "expected ':' after 'else'")
	if err != nil {
		return nil, err
	}

	// Parse the else body using HTML content block parsing
	return p.htmlContentBlock()
}

// htmlParseFinallyBlock parses a finally block within HTML content context. The 'finally' token has already been consumed.
func (p *Parser) htmlParseFinallyBlock() ([]ast.Stmt, error) {
	// Expect colon
	_, err := p.consume(lexer.Colon, "expected ':' after 'finally'")
	if err != nil {
		return nil, err
	}

	// Parse the finally body using HTML content block parsing
	return p.htmlContentBlock()
}
