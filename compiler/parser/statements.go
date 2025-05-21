package parser

import (
	"biscuit/compiler/ast"
	"biscuit/compiler/lexer"
)

// statement parses a single statement.
func (p *Parser) statement() (ast.Stmt, error) {
	return p.simpleStatement()
}

// simpleStatement parses an expression statement.
func (p *Parser) simpleStatement() (ast.Stmt, error) {
	// Check for keywords first
	switch p.peek().Type {
	case lexer.Type:
		return p.typeAlias()
	case lexer.Return:
		return p.returnStatement()
	case lexer.Import, lexer.From:
		return p.importStatement()
	case lexer.Raise:
		return p.raiseStatement()
	case lexer.Pass:
		return p.passStatement()
	case lexer.Del:
		return p.delStatement()
	case lexer.Yield:
		return p.yieldStatement()
	case lexer.Assert:
		return p.assertStatement()
	case lexer.Break:
		return p.breakStatement()
	case lexer.Continue:
		return p.continueStatement()
	case lexer.Global:
		return p.globalStatement()
	case lexer.Nonlocal:
		return p.nonlocalStatement()
	}

	// Check for assignment before expression
	if p.check(lexer.Identifier) || p.check(lexer.LeftParen) || p.check(lexer.LeftBracket) || p.check(lexer.Star) {
		// Save current position
		currentPos := p.Current

		// Try to parse as assignment
		stmt, err := p.assignment()
		if err == nil {
			return stmt, nil
		}

		// If assignment parsing failed, restore position and try as expression
		p.Current = currentPos
	}

	expr, err := p.starExpressions()
	if err != nil {
		return nil, err
	}

	return ast.NewExprStmt(expr, lexer.Span{Start: expr.Span().Start, End: expr.Span().End}), nil
}

// returnStatement parses a return statement.
func (p *Parser) returnStatement() (ast.Stmt, error) {
	// Consume the 'return' keyword
	returnToken, err := p.consume(lexer.Return, "expected 'return'")
	if err != nil {
		return nil, err
	}

	// Exit early if there's no return expression
	if p.isAtEnd() || p.check(lexer.Newline) || p.check(lexer.Semicolon) {
		p.advance()
		return ast.NewReturnStmt(nil, lexer.Span{Start: returnToken.Start(), End: returnToken.End()}), nil
	}

	// Parse the return expression
	expr, err := p.starExpressions()
	if err != nil {
		return nil, err
	}

	return ast.NewReturnStmt(expr, lexer.Span{Start: returnToken.Start(), End: expr.Span().End}), nil
}

func (p *Parser) raiseStatement() (ast.Stmt, error) {
	// Consume the 'raise' keyword
	raiseToken, err := p.consume(lexer.Raise, "expected 'raise'")
	if err != nil {
		return nil, err
	}

	// Check if there's an expression after 'raise'
	if p.isAtEnd() || p.check(lexer.Newline) || p.check(lexer.Semicolon) {
		// Just a 'raise' with no exception
		endPos := raiseToken.End()
		return ast.NewRaiseStmt(nil, nil, false, false, lexer.Span{Start: raiseToken.Start(), End: endPos}), nil
	}

	// Parse the exception expression
	exception, err := p.expression()
	if err != nil {
		return nil, err
	}

	// Check if there's a 'from' clause
	hasFrom := false
	var fromExpr ast.Expr = nil
	endPos := exception.Span().End

	if p.match(lexer.From) {
		hasFrom = true
		// Parse the 'from' expression
		fromExpr, err = p.expression()
		if err != nil {
			return nil, err
		}
		endPos = fromExpr.Span().End
	}

	return ast.NewRaiseStmt(exception, fromExpr, true, hasFrom, lexer.Span{Start: raiseToken.Start(), End: endPos}), nil
}

func (p *Parser) passStatement() (ast.Stmt, error) {
	// Consume the 'pass' keyword
	passToken := p.advance()
	return ast.NewPassStmt(lexer.Span{Start: passToken.Start(), End: passToken.End()}), nil
}

func (p *Parser) yieldStatement() (ast.Stmt, error) {
	expr, err := p.yieldExpression()
	if err != nil {
		return nil, err
	}
	return ast.NewYieldStmt(expr, lexer.Span{Start: expr.Span().Start, End: expr.Span().End}), nil
}

func (p *Parser) assertStatement() (ast.Stmt, error) {
	// Consume the 'assert' keyword
	assertToken, err := p.consume(lexer.Assert, "expected 'assert'")
	if err != nil {
		return nil, err
	}

	// Parse the test expression
	test, err := p.expression()
	if err != nil {
		return nil, err
	}

	// Check for optional message expression
	var message ast.Expr = nil
	endPos := test.Span().End

	if p.match(lexer.Comma) {
		// Parse the message expression
		message, err = p.expression()
		if err != nil {
			return nil, err
		}
		endPos = message.Span().End
	}

	return ast.NewAssertStmt(test, message, lexer.Span{Start: assertToken.Start(), End: endPos}), nil
}

func (p *Parser) breakStatement() (ast.Stmt, error) {
	// Consume the 'break' keyword
	breakToken := p.advance()
	return ast.NewBreakStmt(lexer.Span{Start: breakToken.Start(), End: breakToken.End()}), nil
}

func (p *Parser) continueStatement() (ast.Stmt, error) {
	// Consume the 'continue' keyword
	continueToken := p.advance()
	return ast.NewContinueStmt(lexer.Span{Start: continueToken.Start(), End: continueToken.End()}), nil
}

func (p *Parser) globalStatement() (ast.Stmt, error) {
	// Consume the 'global' keyword
	globalToken, err := p.consume(lexer.Global, "expected 'global'")
	if err != nil {
		return nil, err
	}

	// Parse at least one identifier
	names, err := p.parseNameList()
	if err != nil {
		return nil, err
	}

	if len(names) == 0 {
		return nil, p.error(p.previous(), "expected at least one identifier after 'global'")
	}

	// Get the end position from the last name
	endPos := names[len(names)-1].Span().End
	return ast.NewGlobalStmt(names, lexer.Span{Start: globalToken.Start(), End: endPos}), nil
}

func (p *Parser) nonlocalStatement() (ast.Stmt, error) {
	// Consume the 'nonlocal' keyword
	nonlocalToken, err := p.consume(lexer.Nonlocal, "expected 'nonlocal'")
	if err != nil {
		return nil, err
	}

	// Parse at least one identifier
	names, err := p.parseNameList()
	if err != nil {
		return nil, err
	}

	if len(names) == 0 {
		return nil, p.error(p.previous(), "expected at least one identifier after 'nonlocal'")
	}

	// Get the end position from the last name
	endPos := names[len(names)-1].Span().End
	return ast.NewNonlocalStmt(names, lexer.Span{Start: nonlocalToken.Start(), End: endPos}), nil
}
