package parser

import (
	"github.com/fjvillamarin/topple/compiler/ast"
	"github.com/fjvillamarin/topple/compiler/lexer"
)

// tryStatement parses a try statement according to the grammar:
//
// try_stmt:
//
//	| 'try' ':' block finally_block
//	| 'try' ':' block except_block+ [else_block] [finally_block]
//	| 'try' ':' block except_star_block+ [else_block] [finally_block]
func (p *Parser) tryStatement() (ast.Stmt, error) {
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

	// Parse the try body
	body, err := p.block()
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
			except, err := p.parseExceptBlock()
			if err != nil {
				return nil, err
			}
			excepts = append(excepts, except)
		}

		// Parse optional else block
		if p.match(lexer.Else) {
			elseBody, err = p.parseElseBlock()
			if err != nil {
				return nil, err
			}
		}

		// Parse optional finally block
		if p.match(lexer.Finally) {
			finallyBody, err = p.parseFinallyBlock()
			if err != nil {
				return nil, err
			}

			// Check for invalid clauses after finally
			if p.check(lexer.Except) {
				return nil, p.error(p.peek(), "'except' clause cannot appear after 'finally'")
			}
			if p.check(lexer.Else) {
				return nil, p.error(p.peek(), "'else' clause cannot appear after 'finally'")
			}
		}
	} else if p.match(lexer.Finally) {
		// Just a finally block without any except blocks
		finallyBody, err = p.parseFinallyBlock()
		if err != nil {
			return nil, err
		}

		// Check for invalid clauses after finally
		if p.check(lexer.Except) {
			return nil, p.error(p.peek(), "'except' clause cannot appear after 'finally'")
		}
		if p.check(lexer.Else) {
			return nil, p.error(p.peek(), "'else' clause cannot appear after 'finally'")
		}
	} else if p.check(lexer.Else) {
		// Else without except is invalid
		return nil, p.error(p.peek(), "'else' clause requires at least one 'except' clause")
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

// parseExceptBlock parses an except block according to the grammar:
//
// except_block:
//
//	| 'except' expression ['as' NAME] ':' block
//	| 'except' ':' block
//
// except_star_block:
//
//	| 'except' '*' expression ['as' NAME] ':' block
func (p *Parser) parseExceptBlock() (ast.Except, error) {
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

	// Parse the except body
	body, err := p.block()
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

// parseElseBlock parses an else block. The 'else' token has already been consumed.
func (p *Parser) parseElseBlock() ([]ast.Stmt, error) {
	// Expect colon
	_, err := p.consume(lexer.Colon, "expected ':' after 'else'")
	if err != nil {
		return nil, err
	}

	// Parse the else body
	body, err := p.block()
	if err != nil {
		return nil, err
	}

	return body, nil
}

// parseFinallyBlock parses a finally block. The 'finally' token has already been consumed.
func (p *Parser) parseFinallyBlock() ([]ast.Stmt, error) {
	// Expect colon
	_, err := p.consume(lexer.Colon, "expected ':' after 'finally'")
	if err != nil {
		return nil, err
	}

	// Parse the finally body
	body, err := p.block()
	if err != nil {
		return nil, err
	}

	return body, nil
}
