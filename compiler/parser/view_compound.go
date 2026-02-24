package parser

import (
	"github.com/fjvillamarin/topple/compiler/ast"
	"github.com/fjvillamarin/topple/compiler/lexer"
)

// viewIfStatement parses an if statement within a view context.
// Uses the same AST node as regular if statements but parses view blocks.
func (p *Parser) viewIfStatement() (ast.Stmt, error) {
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

	// Parse the body using view block parsing
	body, err := p.viewBlock()
	if err != nil {
		return nil, err
	}

	// Check for elif
	if p.match(lexer.Elif) {
		// Parse elif as a separate if statement
		elifStmt, err := p.viewElifStatement()
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
	elseBody, err := p.viewElseBlock()
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

// viewElifStatement parses an elif part of an if statement within a view context.
func (p *Parser) viewElifStatement() (ast.Stmt, error) {
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

	// Parse body using view block parsing
	body, err := p.viewBlock()
	if err != nil {
		return nil, err
	}

	// Check for another elif
	if p.match(lexer.Elif) {
		// Parse the next elif
		nextElif, err := p.viewElifStatement()
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
	elseBody, err := p.viewElseBlock()
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

// viewElseBlock parses an else block within a view context.
// Returns empty slice if no else block is present.
func (p *Parser) viewElseBlock() ([]ast.Stmt, error) {
	if !p.match(lexer.Else) {
		return []ast.Stmt{}, nil
	}

	// Expect colon
	_, err := p.consume(lexer.Colon, "expected ':' after else")
	if err != nil {
		return nil, err
	}

	// Parse else body using view block parsing
	return p.viewBlock()
}

// viewForStatement parses a for statement within a view context.
// Uses the same AST node as regular for statements but parses view blocks.
func (p *Parser) viewForStatement() (ast.Stmt, error) {
	// Check for async
	isAsync := false
	var startToken lexer.Token

	if p.match(lexer.Async) {
		isAsync = true
		startToken = p.previous()
		_, err := p.consume(lexer.For, "expected 'for' after 'async'")
		if err != nil {
			return nil, err
		}
	} else {
		// Consume the 'for' token
		var err error
		startToken, err = p.consume(lexer.For, "expected 'for'")
		if err != nil {
			return nil, err
		}
	}

	// Parse the target (star_targets)
	target, err := p.starTargets()
	if err != nil {
		return nil, err
	}

	// Expect 'in' keyword
	_, err = p.consume(lexer.In, "expected 'in' after target in for statement")
	if err != nil {
		return nil, err
	}

	// Parse the iterable (star_expressions)
	iterable, err := p.starExpressions()
	if err != nil {
		return nil, err
	}

	// Expect colon
	_, err = p.consume(lexer.Colon, "expected ':' after iterable in for statement")
	if err != nil {
		return nil, err
	}

	// Parse the body using view block parsing
	body, err := p.viewBlock()
	if err != nil {
		return nil, err
	}

	// Check for else
	elseBody, err := p.viewElseBlock()
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

	// Handle the case where target is a list of expressions
	var targetExpr ast.Expr
	if len(target) == 1 {
		targetExpr = target[0]
	} else {
		// If multiple targets, create a tuple
		targetExpr = &ast.TupleExpr{
			Elements: target,
			Span:     lexer.Span{Start: target[0].GetSpan().Start, End: target[len(target)-1].GetSpan().End},
		}
	}

	return &ast.For{
		Target:   targetExpr,
		Iterable: iterable,
		Body:     body,
		Else:     elseBody,
		IsAsync:  isAsync,
		Span:     lexer.Span{Start: startToken.Start(), End: endPos},
	}, nil
}

// viewWhileStatement parses a while statement within a view context.
// Uses the same AST node as regular while statements but parses view blocks.
func (p *Parser) viewWhileStatement() (ast.Stmt, error) {
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

	// Parse the body using view block parsing
	body, err := p.viewBlock()
	if err != nil {
		return nil, err
	}

	// Check for else
	elseBody, err := p.viewElseBlock()
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

// viewTryStatement parses a try statement within a view context.
// Uses the same AST node as regular try statements but parses view blocks.
func (p *Parser) viewTryStatement() (ast.Stmt, error) {
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

	// Parse the try body using view block parsing
	body, err := p.viewBlock()
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
			except, err := p.viewParseExceptBlock()
			if err != nil {
				return nil, err
			}
			excepts = append(excepts, except)
		}

		// Parse optional else block
		if p.match(lexer.Else) {
			elseBody, err = p.viewParseElseBlock()
			if err != nil {
				return nil, err
			}
		}

		// Parse optional finally block
		if p.match(lexer.Finally) {
			finallyBody, err = p.viewParseFinallyBlock()
			if err != nil {
				return nil, err
			}
		}
	} else if p.match(lexer.Finally) {
		// Just a finally block without any except blocks
		finallyBody, err = p.viewParseFinallyBlock()
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

// viewParseExceptBlock parses an except block within a view context.
func (p *Parser) viewParseExceptBlock() (ast.Except, error) {
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

	// Parse the except body using view block parsing
	body, err := p.viewBlock()
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

// viewParseElseBlock parses an else block within a view context. The 'else' token has already been consumed.
func (p *Parser) viewParseElseBlock() ([]ast.Stmt, error) {
	// Expect colon
	_, err := p.consume(lexer.Colon, "expected ':' after 'else'")
	if err != nil {
		return nil, err
	}

	// Parse the else body using view block parsing
	return p.viewBlock()
}

// viewParseFinallyBlock parses a finally block within a view context. The 'finally' token has already been consumed.
func (p *Parser) viewParseFinallyBlock() ([]ast.Stmt, error) {
	// Expect colon
	_, err := p.consume(lexer.Colon, "expected ':' after 'finally'")
	if err != nil {
		return nil, err
	}

	// Parse the finally body using view block parsing
	return p.viewBlock()
}

// viewMatchStatement parses a match statement within a view context.
// Uses the same AST node as regular match statements but parses view blocks.
func (p *Parser) viewMatchStatement() (ast.Stmt, error) {
	matchToken, err := p.consume(lexer.Match, "expected 'match'")
	if err != nil {
		return nil, err
	}

	// Parse subject expression
	subject, err := p.subjectExpr()
	if err != nil {
		return nil, err
	}

	// Consume ':'
	_, err = p.consume(lexer.Colon, "expected ':' after match subject")
	if err != nil {
		return nil, err
	}

	// Consume NEWLINE
	_, err = p.consume(lexer.Newline, "expected newline after ':'")
	if err != nil {
		return nil, err
	}

	// Consume all remaining newlines (handles comments)
	for p.check(lexer.Newline) {
		p.advance()
	}

	// Consume INDENT
	_, err = p.consume(lexer.Indent, "expected indented block after match")
	if err != nil {
		return nil, err
	}

	// Parse one or more case blocks
	var cases []ast.CaseBlock
	for p.check(lexer.Case) {
		caseBlock, err := p.viewCaseBlock()
		if err != nil {
			return nil, err
		}
		cases = append(cases, caseBlock)

		// Skip newlines between case blocks
		for p.check(lexer.Newline) {
			p.advance()
		}
	}

	if len(cases) == 0 {
		return nil, p.error(p.peek(), "expected at least one case block")
	}

	// Consume DEDENT
	dedentToken, err := p.consume(lexer.Dedent, "expected dedent after match cases")
	if err != nil {
		return nil, err
	}

	return &ast.MatchStmt{
		Subject: subject,
		Cases:   cases,
		Span:    lexer.Span{Start: matchToken.Start(), End: dedentToken.End()},
	}, nil
}

// viewCaseBlock parses a case block within a view context.
func (p *Parser) viewCaseBlock() (ast.CaseBlock, error) {
	caseToken, err := p.consume(lexer.Case, "expected 'case'")
	if err != nil {
		return ast.CaseBlock{}, err
	}

	// Parse patterns
	patterns, err := p.patterns()
	if err != nil {
		return ast.CaseBlock{}, err
	}

	// Parse optional guard
	var guard ast.Expr
	if p.match(lexer.If) {
		guard, err = p.namedExpression()
		if err != nil {
			return ast.CaseBlock{}, err
		}
	}

	// Consume ':'
	_, err = p.consume(lexer.Colon, "expected ':' after case pattern")
	if err != nil {
		return ast.CaseBlock{}, err
	}

	// Parse block using view block parsing
	body, err := p.viewBlock()
	if err != nil {
		return ast.CaseBlock{}, err
	}

	return ast.CaseBlock{
		Patterns: patterns,
		Guard:    guard,
		Body:     body,
		Span:     lexer.Span{Start: caseToken.Start(), End: body[len(body)-1].GetSpan().End},
	}, nil
}

// viewFunctionDef parses a function definition within a view context.
// Uses the same AST node as regular function definitions but ensures proper context handling.
func (p *Parser) viewFunctionDef() (ast.Stmt, error) {
	// Track if this is an async function
	isAsync := false
	var startToken lexer.Token

	// Check for 'async' keyword
	if p.match(lexer.Async) {
		isAsync = true
		startToken = p.previous()
		_, err := p.consume(lexer.Def, "expected 'def' after 'async'")
		if err != nil {
			return nil, err
		}
	} else {
		// Consume the 'def' keyword
		var err error
		startToken, err = p.consume(lexer.Def, "expected 'def'")
		if err != nil {
			return nil, err
		}
	}

	// Parse the function name
	nameToken, err := p.consume(lexer.Identifier, "expected function name")
	if err != nil {
		return nil, err
	}
	name := &ast.Name{
		Token: nameToken,
		Span:  lexer.Span{Start: nameToken.Start(), End: nameToken.End()},
	}

	// Parse optional type parameters
	var typeParams []*ast.TypeParam
	if p.check(lexer.LeftBracket) {
		// Get the type parameters using the existing parser function
		paramExprs, err := p.typeParams()
		if err != nil {
			return nil, err
		}

		// Convert from Expr to *TypeParam
		typeParams = make([]*ast.TypeParam, 0, len(paramExprs))
		for _, expr := range paramExprs {
			if tp, ok := expr.(*ast.TypeParam); ok {
				typeParams = append(typeParams, tp)
			}
		}
	}

	// Expect opening parenthesis for parameters
	_, err = p.consume(lexer.LeftParen, "expected '(' after function name")
	if err != nil {
		return nil, err
	}

	// Parse function parameters (if any)
	var parameterList *ast.ParameterList
	if !p.check(lexer.RightParen) {
		parameterList, err = p.params()
		if err != nil {
			return nil, err
		}
	} else {
		// Empty parameter list
		parameterList = &ast.ParameterList{
			Parameters:  []*ast.Parameter{},
			SlashIndex:  -1,
			VarArgIndex: -1,
			KwArgIndex:  -1,
			Span:        lexer.Span{Start: p.previous().End(), End: p.peek().Start()},
		}
	}

	// Expect closing parenthesis
	_, err = p.consume(lexer.RightParen, "expected ')' after function parameters")
	if err != nil {
		return nil, err
	}

	// Parse optional return type annotation
	var returnType ast.Expr
	if p.match(lexer.Arrow) {
		returnType, err = p.expression()
		if err != nil {
			return nil, err
		}
	}

	// Expect colon
	_, err = p.consume(lexer.Colon, "expected ':' after function header")
	if err != nil {
		return nil, err
	}

	// Parse function body using regular block() since function bodies are regular Python code
	// The key difference from viewBlock() is that function bodies should NOT contain HTML
	body, err := p.block()
	if err != nil {
		return nil, err
	}

	// Determine the end position
	var endPos lexer.Position
	if len(body) > 0 {
		endPos = body[len(body)-1].GetSpan().End
	} else {
		endPos = p.previous().End()
	}

	// Create and return the Function node
	return &ast.Function{
		Name:           name,
		TypeParameters: typeParams,
		ReturnType:     returnType,
		Parameters:     parameterList,
		Body:           body,
		IsAsync:        isAsync,
		Span:           lexer.Span{Start: startToken.Start(), End: endPos},
	}, nil
}
